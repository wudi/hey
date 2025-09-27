#!/bin/bash
# Docker integration test script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== Hey-Codex Docker Integration Tests ==="
echo ""

cd "$ROOT_DIR"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Error: docker-compose is not installed"
    exit 1
fi

# Determine docker-compose command
if docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "Using: $DOCKER_COMPOSE"
echo ""

# Build and start containers
echo "Building Docker images..."
$DOCKER_COMPOSE build --no-cache

echo ""
echo "Starting containers..."
$DOCKER_COMPOSE up -d

echo ""
echo "Waiting for services to be ready..."
sleep 5

# Check if containers are running
echo ""
echo "Checking container status..."
$DOCKER_COMPOSE ps

echo ""
echo "Checking hey-fpm logs..."
$DOCKER_COMPOSE logs hey-fpm | tail -20

echo ""
echo "=== Running Tests ==="
echo ""

PASSED=0
FAILED=0
BASE_URL="http://localhost:8080"

# Test function
test_url() {
    local name="$1"
    local url="$2"
    local expected_code="${3:-200}"

    echo -n "Testing $name... "

    response=$(curl -s -w "\n%{http_code}" "$url" 2>/dev/null || echo "FAIL\n000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" = "$expected_code" ]; then
        echo "✓ PASS (HTTP $http_code)"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo "✗ FAIL (Expected HTTP $expected_code, got $http_code)"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# Test content
test_content() {
    local name="$1"
    local url="$2"
    local expected="$3"

    echo -n "Testing $name... "

    response=$(curl -s "$url" 2>/dev/null)

    if echo "$response" | grep -q "$expected"; then
        echo "✓ PASS (found '$expected')"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo "✗ FAIL (did not find '$expected')"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# Test header
test_header() {
    local name="$1"
    local url="$2"
    local header="$3"

    echo -n "Testing $name... "

    response=$(curl -s -I "$url" 2>/dev/null)

    if echo "$response" | grep -qi "$header"; then
        echo "✓ PASS (found header '$header')"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo "✗ FAIL (did not find header '$header')"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# Run tests
test_url "Container connectivity" "$BASE_URL/" 200
test_content "Main page content" "$BASE_URL/" "Hey-Codex FPM is working"
test_content "PHP version" "$BASE_URL/" "8.0.30"

test_url "JSON endpoint" "$BASE_URL/json.php" 200
test_header "JSON content type" "$BASE_URL/json.php" "Content-Type: application/json"
test_content "JSON response" "$BASE_URL/json.php" '"status":"success"'

test_url "Headers page" "$BASE_URL/headers.php" 200
test_header "Custom header" "$BASE_URL/headers.php" "X-Custom-Header"
test_header "Powered by" "$BASE_URL/headers.php" "X-Powered-By"

test_url "Cookies page" "$BASE_URL/cookies.php" 200

test_url "404 error" "$BASE_URL/nonexistent.php" 404

# Summary
echo ""
echo "=== Test Summary ==="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo "Total:  $((PASSED + FAILED))"
echo ""

# Show container stats
echo "=== Container Stats ==="
docker stats --no-stream hey-fpm hey-nginx

echo ""
echo "=== Logs ==="
echo ""
echo "To view hey-fpm logs: $DOCKER_COMPOSE logs hey-fpm"
echo "To view nginx logs:   $DOCKER_COMPOSE logs nginx"
echo ""
echo "To stop containers:   $DOCKER_COMPOSE down"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✓ All Docker tests passed!"
    exit 0
else
    echo "✗ Some Docker tests failed"
    echo ""
    echo "Debug commands:"
    echo "  $DOCKER_COMPOSE logs hey-fpm"
    echo "  $DOCKER_COMPOSE logs nginx"
    echo "  $DOCKER_COMPOSE exec hey-fpm sh"
    exit 1
fi