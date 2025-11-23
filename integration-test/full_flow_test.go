package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

func TestFullWorkflow(t *testing.T) {
	//Health check
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	teamName := fmt.Sprintf("test-team-%d", time.Now().Unix())

	//Create team
	t.Run("CreateTeam", func(t *testing.T) {
		team := map[string]interface{}{
			"team_name": teamName,
			"members": []map[string]interface{}{
				{"user_id": "user1", "username": "Alice", "is_active": true},
				{"user_id": "user2", "username": "Bob", "is_active": true},
				{"user_id": "user3", "username": "Charlie", "is_active": true},
			},
		}

		body, _ := json.Marshal(team)
		resp, err := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create team: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Response status: %d", resp.StatusCode)
		t.Logf("Response body: %s", string(bodyBytes))

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected 201, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	//Create PR
	prID := fmt.Sprintf("pr-%d", time.Now().Unix())
	t.Run("CreatePR", func(t *testing.T) {
		pr := map[string]interface{}{
			"pull_request_id":   prID,
			"pull_request_name": "Test Feature",
			"author_id":         "user1",
		}

		body, _ := json.Marshal(pr)
		resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create PR: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Logf("Response status: %d", resp.StatusCode)
		t.Logf("Response body: %s", string(bodyBytes))

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected 201, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
			return
		}

		var result map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if result["pr"] == nil {
			t.Fatal("Response doesn't contain 'pr' field")
		}

		prData, ok := result["pr"].(map[string]interface{})
		if !ok {
			t.Fatalf("pr field is not a map, got: %T", result["pr"])
		}

		reviewersRaw, exists := prData["assigned_reviewers"]
		if !exists {
			t.Fatal("PR doesn't have 'assigned_reviewers' field")
		}

		if reviewersRaw == nil {
			t.Log("No reviewers assigned (empty array)")
			return
		}

		reviewers, ok := reviewersRaw.([]interface{})
		if !ok {
			t.Fatalf("assigned_reviewers is not an array, got: %T, value: %v", reviewersRaw, reviewersRaw)
		}

		t.Logf("Assigned reviewers count: %d", len(reviewers))

		if len(reviewers) < 1 || len(reviewers) > 2 {
			t.Errorf("Expected 1-2 reviewers, got %d", len(reviewers))
		}
	})

	//Get statistics
	t.Run("GetStats", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/stats/users")
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		} else {
			t.Logf("User stats: %s", string(bodyBytes))
		}
	})

	//Merge PR
	t.Run("MergePR", func(t *testing.T) {
		mergeReq := map[string]interface{}{
			"pull_request_id": prID,
		}

		body, _ := json.Marshal(mergeReq)
		resp, err := http.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to merge PR: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	//Test idempotency
	t.Run("MergeIdempotency", func(t *testing.T) {
		mergeReq := map[string]interface{}{
			"pull_request_id": prID,
		}

		body, _ := json.Marshal(mergeReq)
		resp, err := http.Post(baseURL+"/pullRequest/merge", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to merge PR again: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Idempotent merge should return 200, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
		}
	})
}

func TestConcurrentPRCreation(t *testing.T) {
	teamName := fmt.Sprintf("concurrent-team-%d", time.Now().Unix())

	// Create team
	team := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "c1", "username": "User1", "is_active": true},
			{"user_id": "c2", "username": "User2", "is_active": true},
			{"user_id": "c3", "username": "User3", "is_active": true},
		},
	}

	body, _ := json.Marshal(team)
	resp, _ := http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))
	if resp != nil {
		resp.Body.Close()
	}

	// Wait a bit for team to be created
	time.Sleep(100 * time.Millisecond)

	// Create 10 PRs concurrently
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			pr := map[string]interface{}{
				"pull_request_id":   fmt.Sprintf("concurrent-pr-%d-%d", time.Now().UnixNano(), index),
				"pull_request_name": fmt.Sprintf("Concurrent PR %d", index),
				"author_id":         "c1",
			}

			body, _ := json.Marshal(pr)
			resp, err := http.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewBuffer(body))
			if err != nil {
				errors <- err
				done <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				errors <- fmt.Errorf("Expected 201, got %d", resp.StatusCode)
				done <- false
				return
			}

			done <- true
		}(i)
	}

	// Wait for all
	successCount := 0
	for i := 0; i < 10; i++ {
		if <-done {
			successCount++
		}
	}

	close(errors)
	for err := range errors {
		t.Logf("Error: %v", err)
	}

	if successCount < 8 {
		t.Errorf("Expected at least 8 successful creations, got %d", successCount)
	}

	t.Logf("Successfully created %d/10 PRs concurrently", successCount)
}
