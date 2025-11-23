package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"pr-reviewer-service/internal/entity"
	"pr-reviewer-service/internal/usecase"
)

type teamRoutes struct {
	t *usecase.TeamUseCase
}

func newTeamRoutes(mux *http.ServeMux, t *usecase.TeamUseCase) {
	r := &teamRoutes{t}

	mux.HandleFunc("POST /team/add", r.create)
	mux.HandleFunc("GET /team/get", r.get)
}

type createTeamRequest struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	} `json:"members"`
}

func (r *teamRoutes) create(w http.ResponseWriter, req *http.Request) {
	var input createTeamRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	members := make([]entity.User, len(input.Members))
	for i, m := range input.Members {
		members[i] = entity.User{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	team, err := r.t.CreateTeam(req.Context(), input.TeamName, members)
	if err != nil {
		if errors.Is(err, entity.ErrTeamAlreadyExists) {
			respondError(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{"team": team})
}

func (r *teamRoutes) get(w http.ResponseWriter, req *http.Request) {
	teamName := req.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	team, err := r.t.GetTeam(req.Context(), teamName)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			respondError(w, http.StatusNotFound, "NOT_FOUND", "team not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, team)
}
