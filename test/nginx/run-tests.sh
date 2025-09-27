#!/bin/bash
# Run automated tests against hey-fpm + Nginx

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:8080"

echo "=== Hey-Codex FPM + Nginx Integration Tests ==="
echo ""

# Check if services are running
if ! lsof -i :9000 &> /dev/null; then
    echo "Error: hey-fpm is not running on port 9000"
    echo "Please run: $SCRIPT_DIR/start-test.sh"
    exit 1
fi

if ! lsof -i :8080 &> /dev/null; then
    echo "Error: Nginx is not running on port 8080"
    echo "Please run: $SCRIPT_DIR/start-test.sh"
    exit 1
fi

echo "✓ Services are running"
echo ""

PASSED=0
FAILED=0

# Test function
test_url() {
    local name="$1"
    local url="$2"
    local expected_code="${3:-200}"

    echo -n "Testing $name... "

    response=$(curl -s -w "\n%{http_code}" "$url")
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

# Test content contains string
test_content() {
    local name="$1"
    local url="$2"
    local expected="$3"

    echo -n "Testing $name... "

    response=$(curl -s "$url")

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

# Test header present
test_header() {
    local name="$1"
    local url="$2"
    local header="$3"

    echo -n "Testing $name... "

    response=$(curl -s -I "$url")

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
echo "Running HTTP tests..."
echo ""

test_url "Main page" "$BASE_URL/" 200
test_content "Main page content" "$BASE_URL/" "Hey-Codex FPM is working"

test_url "Info page" "$BASE_URL/info.php" 200
test_content "PHP version" "$BASE_URL/info.php" "PHP"

test_url "Headers page" "$BASE_URL/headers.php" 200
test_header "Custom header" "$BASE_URL/headers.php" "X-Custom-Header"
test_header "Powered by header" "$BASE_URL/headers.php" "X-Powered-By"

test_url "JSON API" "$BASE_URL/json.php" 200
test_header "JSON content type" "$BASE_URL/json.php" "Content-Type: application/json"
test_content "JSON response" "$BASE_URL/json.php" '"status":"success"'

test_url "Cookies page" "$BASE_URL/cookies.php" 200
test_header "Set-Cookie header" "$BASE_URL/cookies.php" "Set-Cookie"

# Test 404
echo -n "Testing 404 error... "
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/nonexistent.php")
http_code=$(echo "$response" | tail -n1)
if [ "$http_code" = "404" ]; then
    echo "✓ PASS (HTTP 404)"
    PASSED=$((PASSED + 1))
else
    echo "✗ FAIL (Expected HTTP 404, got $http_code)"
    FAILED=$((FAILED + 1))
fi

# Summary
echo ""
echo "=== Test Summary ==="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✓ All tests passed!"
    exit 0
else
    echo "✗ Some tests failed"
    exit 1
fi