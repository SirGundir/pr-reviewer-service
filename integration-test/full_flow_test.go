package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected 201, got %d", resp.StatusCode)
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

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected 201, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		prData := result["pr"].(map[string]interface{})
		reviewers := prData["assigned_reviewers"].([]interface{})

		if len(reviewers) < 1 {
			t.Error("Expected at least 1 reviewer")
		}
	})

	//Get statistics
	t.Run("GetStats", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/stats/users")
		if err != nil {
			t.Fatalf("Failed to get stats: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
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

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
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

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Idempotent merge should return 200, got %d", resp.StatusCode)
		}
	})
}

func TestConcurrentPRCreation(t *testing.T) {
	teamName := fmt.Sprintf("concurrent-team-%d", time.Now().Unix())

	//Create team
	team := map[string]interface{}{
		"team_name": teamName,
		"members": []map[string]interface{}{
			{"user_id": "c1", "username": "User1", "is_active": true},
			{"user_id": "c2", "username": "User2", "is_active": true},
			{"user_id": "c3", "username": "User3", "is_active": true},
		},
	}

	body, _ := json.Marshal(team)
	http.Post(baseURL+"/team/add", "application/json", bytes.NewBuffer(body))

	//Create 10 PRs concurrently
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

	//Wait
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
}
