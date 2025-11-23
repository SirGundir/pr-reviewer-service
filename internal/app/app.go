package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pr-reviewer-service/config"
	"pr-reviewer-service/internal/controller/http/middleware"
	v1 "pr-reviewer-service/internal/controller/http/v1"
	"pr-reviewer-service/internal/repo/persistent"
	"pr-reviewer-service/internal/usecase"
	"pr-reviewer-service/pkg/postgres"
)

func Run(cfg *config.Config) {

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//Postgres
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		log.Fatalf("app - Run - postgres.New: %v", err)
	}
	defer pg.Close()

	//Repositories & Use Cases
	userRepo := persistent.NewUserRepo(pg)
	teamRepo := persistent.NewTeamRepo(pg, userRepo)
	prRepo := persistent.NewPullRequestRepo(pg)

	reviewerSelector := usecase.NewReviewerSelector()
	teamUC := usecase.NewTeamUseCase(teamRepo, userRepo)
	userUC := usecase.NewUserUseCase(userRepo, prRepo)
	prUC := usecase.NewPullRequestUseCase(prRepo, userRepo, reviewerSelector)
	statsUC := usecase.NewStatsUseCase(prRepo, userRepo)

	//HTTP Server
	mux := http.NewServeMux()
	v1.NewRouter(mux, teamUC, userUC, prUC, statsUC)

	//Middleware: Recovery, Logger, CORS
	handler := middleware.CORS(middleware.Recovery(middleware.Logger(mux)))

	server := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: handler,
	}

	go func() {
		log.Printf("app starting: listening on port %s, PG=%s", cfg.HTTP.Port, cfg.PG.URL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("app - ListenAndServe error: %v", err)
		}
	}()

	//Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	log.Printf("app - shutting down, signal: %s", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("app - server shutdown error: %v", err)
	} else {
		log.Println("app - server gracefully stopped")
	}
}
