#!/bin/bash

set -e

echo "=== Load Testing PR Reviewer Service ==="
echo ""

BASE_URL="http://localhost:8080"

if ! curl -s $BASE_URL/health > /dev/null; then
    echo "Error: Service is not running on $BASE_URL"
    echo "Run: make docker-up"
    exit 1
fi

echo "Service is running"
echo ""

if ! command -v hey &> /dev/null; then
    echo "Installing hey"
    go install github.com/rakyll/hey@latest
    echo "hey installed"
    echo ""
fi

HEY_CMD="hey"
if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    GOPATH=$(go env GOPATH)
    HEY_CMD="$GOPATH/bin/hey.exe"
fi

# Health check load test
echo "Health Check Load Test"
echo "   Requests: 1000, Concurrency: 50"
echo "   ----------------------------------------"
$HEY_CMD -n 1000 -c 50 -q 1 $BASE_URL/health | grep -E "Status|Requests/sec|Average|Fastest|Slowest"
echo ""

echo "Creating test team..."
TEAM_RESPONSE=$(curl -s -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "load-test-team",
    "members": [
      {"user_id": "lt1", "username": "LoadUser1", "is_active": true},
      {"user_id": "lt2", "username": "LoadUser2", "is_active": true},
      {"user_id": "lt3", "username": "LoadUser3", "is_active": true}
    ]
  }')

if echo "$TEAM_RESPONSE" | grep -q "team_name"; then
    echo "Team created successfully"
else
    echo "Team might already exist (continuing...)"
fi
echo ""

# PR Creation Load Test
echo "PR Creation Load Test"
echo "   Requests: 100, Concurrency: 10"

TEMP_FILE=$(mktemp)
for i in {1..100}; do
  echo "{\"pull_request_id\":\"load-pr-$(date +%s%N)-$i\",\"pull_request_name\":\"Load Test PR $i\",\"author_id\":\"lt1\"}" >> $TEMP_FILE
done

$HEY_CMD -n 100 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -D $TEMP_FILE \
  $BASE_URL/pullRequest/create | grep -E "Status|Requests/sec|Average|Fastest|Slowest"

rm $TEMP_FILE
echo ""

# Stats endpoint load test
echo "Stats Endpoint Load Test"
echo "   Requests: 500, Concurrency: 25"
echo "   ----------------------------------------"
$HEY_CMD -n 500 -c 25 -q 1 $BASE_URL/stats/users | grep -E "Status|Requests/sec|Average|Fastest|Slowest"
echo ""

# Get Team Load Test
echo "Get Team Load Test"
echo "   Requests: 200, Concurrency: 20"
echo "   ----------------------------------------"
$HEY_CMD -n 200 -c 20 "$BASE_URL/team/get?team_name=load-test-team" | grep -E "Status|Requests/sec|Average|Fastest|Slowest"
echo ""

echo "=== Load Test Complete ==="
echo ""
echo "All endpoints tested"
echo ""
