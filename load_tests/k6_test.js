import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
    vus: 20,      
    duration: "30s", 
    thresholds: {
        http_req_duration: ["p(95)<300"], 
        http_req_failed: ["rate<0.01"], 
    },
};

function randID(prefix) {
    return prefix + Math.floor(Math.random() * 100000);
}

export default function () {
    const teamName = "team-" + Math.floor(Math.random() * 1000);
    const createTeamRes = http.post("http://localhost:8080/team/add", JSON.stringify({
        team_name: teamName,
        members: [
            { user_id: randID("u"), username: "Alice", is_active: true },
            { user_id: randID("u"), username: "Bob", is_active: true }
        ]
    }), { headers: { "Content-Type": "application/json" }});
    check(createTeamRes, { "team created": (r) => r.status === 201 || r.status === 400 });

    const getTeamRes = http.get(`http://localhost:8080/team/get?team_name=${teamName}`);
    check(getTeamRes, { "team fetched": (r) => r.status === 200 || r.status === 404 });

    const prID = randID("pr-");
    const authorID = JSON.parse(getTeamRes.body)?.members?.[0]?.user_id || "u1";
    const createPRRes = http.post("http://localhost:8080/pullRequest/create", JSON.stringify({
        pull_request_id: prID,
        pull_request_name: "Feature X",
        author_id: authorID
    }), { headers: { "Content-Type": "application/json" }});
    check(createPRRes, { "PR created": (r) => r.status === 201 || r.status === 404 || r.status === 409 });

    const getReviewsRes = http.get(`http://localhost:8080/users/getReview?user_id=${authorID}`);
    check(getReviewsRes, { "reviews fetched": (r) => r.status === 200 });

    const setActiveRes = http.post("http://localhost:8080/users/setIsActive", JSON.stringify({
        user_id: authorID,
        is_active: false
    }), { headers: { "Content-Type": "application/json" }});
    check(setActiveRes, { "set isActive": (r) => r.status === 200 || r.status === 404 });

    const mergeRes = http.post("http://localhost:8080/pullRequest/merge", JSON.stringify({
        pull_request_id: prID
    }), { headers: { "Content-Type": "application/json" }});
    check(mergeRes, { "PR merged": (r) => r.status === 200 || r.status === 404 });

    const reassignRes = http.post("http://localhost:8080/pullRequest/reassign", JSON.stringify({
        pull_request_id: prID,
        old_user_id: authorID
    }), { headers: { "Content-Type": "application/json" }});
    check(reassignRes, { "reassign attempted": (r) => r.status === 200 || r.status === 404 || r.status === 409 });

    sleep(0.5);
}
