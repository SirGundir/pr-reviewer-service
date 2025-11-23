package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/internal/usecase"
)

type userRoutes struct {
	u *usecase.UserUseCase
}

func newUserRoutes(mux *http.ServeMux, u *usecase.UserUseCase) {
	r := &userRoutes{u}

	mux.HandleFunc("POST /users/setIsActive", r.setIsActive)
	mux.HandleFunc("GET /users/getReview", r.getReviews)
}

type setIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func (r *userRoutes) setIsActive(w http.ResponseWriter, req *http.Request) {
	var input setIsActiveRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := r.u.SetIsActive(req.Context(), input.UserID, input.IsActive)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (r *userRoutes) getReviews(w http.ResponseWriter, req *http.Request) {
	userID := req.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	prs, err := r.u.GetReviews(req.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
