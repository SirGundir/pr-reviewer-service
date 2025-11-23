#!/bin/bash

echo "=== Load Testing PR Reviewer Service ==="
echo ""

if ! command -v hey &> /dev/null; then
    echo "Installing hey..."
    go install github.com/rakyll/hey@latest
fi

BASE_URL="http://localhost:8080"

#Health check
echo "1. Health Check Load Test (1000 requests, 50 concurrent)"
hey -n 1000 -c 50 $BASE_URL/health
echo ""

#Team creation
echo "2. Creating test team..."
curl -s -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "load-test-team",
    "members": [
      {"user_id": "lt1", "username": "LoadUser1", "is_active": true},
      {"user_id": "lt2", "username": "LoadUser2", "is_active": true},
      {"user_id": "lt3", "username": "LoadUser3", "is_active": true}
    ]
  }' > /dev/null
echo "Team created"
echo ""

#PR creation
echo "3. PR Creation Load Test (100 requests, 10 concurrent)"
for i in {1..100}; do
  echo "{\"pull_request_id\":\"load-pr-$i\",\"pull_request_name\":\"Load Test PR $i\",\"author_id\":\"lt1\"}"
done | hey -n 100 -c 10 -m POST -H "Content-Type: application/json" -D /dev/stdin $BASE_URL/pullRequest/create
echo ""

#Stats endpoint
echo "4. Stats Endpoint Load Test (500 requests, 25 concurrent)"
hey -n 500 -c 25 $BASE_URL/stats/users
echo ""

echo "=== Load Test Complete ==="