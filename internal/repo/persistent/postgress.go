package persistent

import "pr-reviewer-service/pkg/postgres"

// Repositories содержит все репозитории
type Repositories struct {
	User        *UserRepo
	Team        *TeamRepo
	PullRequest *PullRequestRepo
}

// NewRepositories инициализирует все репозитории
func NewRepositories(pg *postgres.Postgres) *Repositories {
	userRepo := NewUserRepo(pg)
	teamRepo := NewTeamRepo(pg, userRepo)
	prRepo := NewPullRequestRepo(pg)

	return &Repositories{
		User:        userRepo,
		Team:        teamRepo,
		PullRequest: prRepo,
	}
}
