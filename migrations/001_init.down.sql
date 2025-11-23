-- Rollback
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP TABLE IF EXISTS pr_reviewers;

DROP INDEX IF EXISTS idx_pr_status;
DROP TABLE IF EXISTS pull_requests;

DROP INDEX IF EXISTS idx_users_team_active;
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS teams;