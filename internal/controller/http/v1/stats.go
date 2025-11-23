package v1

import (
	"net/http"
	"pr-reviewer-service/internal/usecase"
)

type statsRoutes struct {
	stats *usecase.StatsUseCase
}

func newStatsRoutes(mux *http.ServeMux, stats *usecase.StatsUseCase) {
	r := &statsRoutes{stats}

	mux.HandleFunc("GET /stats/users", r.getUserStats)
	mux.HandleFunc("GET /stats/prs", r.getPRStats)
}

func (r *statsRoutes) getUserStats(w http.ResponseWriter, req *http.Request) {
	stats, err := r.stats.GetUserStats(req.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"users": stats,
	})
}

func (r *statsRoutes) getPRStats(w http.ResponseWriter, req *http.Request) {
	stats, err := r.stats.GetPRStats(req.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
