package v1

import (
	"net/http"
	"pr-reviewer-service/internal/usecase"
)

func NewRouter(
	mux *http.ServeMux,
	t *usecase.TeamUseCase,
	u *usecase.UserUseCase,
	pr *usecase.PullRequestUseCase,
) {
	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Register routes
	newTeamRoutes(mux, t)
	newUserRoutes(mux, u)
	newPullRequestRoutes(mux, pr)
}
