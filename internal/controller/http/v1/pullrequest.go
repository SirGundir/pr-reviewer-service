package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/internal/usecase"
)

type pullRequestRoutes struct {
	pr *usecase.PullRequestUseCase
}

func newPullRequestRoutes(mux *http.ServeMux, pr *usecase.PullRequestUseCase) {
	r := &pullRequestRoutes{pr}

	mux.HandleFunc("POST /pullRequest/create", r.create)
	mux.HandleFunc("POST /pullRequest/merge", r.merge)
	mux.HandleFunc("POST /pullRequest/reassign", r.reassign)
}

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

func (r *pullRequestRoutes) create(w http.ResponseWriter, req *http.Request) {
	var input createPRRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pr, err := r.pr.CreatePR(req.Context(), input.PullRequestID, input.PullRequestName, input.AuthorID)
	if err != nil {
		if errors.Is(err, entity.ErrPRAlreadyExists) {
			respondError(w, http.StatusConflict, "PR_EXISTS", "PR id already exists")
			return
		}
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "author or team not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{"pr": pr})
}

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

func (r *pullRequestRoutes) merge(w http.ResponseWriter, req *http.Request) {
	var input mergePRRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pr, err := r.pr.MergePR(req.Context(), input.PullRequestID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "pull request not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"pr": pr})
}

type reassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

func (r *pullRequestRoutes) reassign(w http.ResponseWriter, req *http.Request) {
	var input reassignReviewerRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pr, newReviewerID, err := r.pr.ReassignReviewer(req.Context(), input.PullRequestID, input.OldUserID)
	if err != nil {
		if errors.Is(err, entity.ErrPRAlreadyMerged) {
			respondError(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
			return
		}
		if errors.Is(err, entity.ErrReviewerNotAssigned) {
			respondError(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
			return
		}
		if errors.Is(err, entity.ErrNoCandidates) {
			respondError(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
			return
		}
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "pull request or user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
