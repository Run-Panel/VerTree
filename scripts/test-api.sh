#!/bin/bash

# Universal Update Service API Test Script
# This script tests the basic functionality of the update service APIs

BASE_URL="http://localhost:8080"
ADMIN_API="$BASE_URL/admin/api/v1"
CLIENT_API="$BASE_URL/api/v1"

echo "🚀 Universal Update Service API Test"
echo "=================================="

# Function to make HTTP requests and show results
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4
    
    echo ""
    echo "📡 $description"
    echo "   $method $url"
    
    if [ -n "$data" ]; then
        echo "   Data: $data"
        response=$(curl -s -X $method -H "Content-Type: application/json" -d "$data" "$url")
    else
        response=$(curl -s -X $method "$url")
    fi
    
    echo "   Response: $response" | jq . 2>/dev/null || echo "   Response: $response"
}

# Check if jq is available for pretty JSON formatting
if ! command -v jq &> /dev/null; then
    echo "⚠️  jq not found. JSON responses will not be pretty-printed."
fi

# Test 1: Health Check
make_request "GET" "$BASE_URL/health" "" "Health Check"

# Test 2: Get Channels
make_request "GET" "$ADMIN_API/channels" "" "Get All Channels"

# Test 3: Create a Test Version
version_data='{
    "version": "v1.0.0-test",
    "channel": "alpha",
    "title": "Test Version 1.0.0",
    "description": "This is a test version created by the API test script",
    "release_notes": "## Test Release\n\n- Added test functionality\n- Fixed test bugs",
    "breaking_changes": "",
    "min_upgrade_version": "",
    "file_url": "https://example.com/releases/v1.0.0-test/app",
    "file_size": 1024000,
    "file_checksum": "sha256:test123456789abcdef",
    "is_forced": false
}'

make_request "POST" "$ADMIN_API/versions" "$version_data" "Create Test Version"

# Test 4: Get Versions List
make_request "GET" "$ADMIN_API/versions?limit=10" "" "Get Versions List"

# Test 5: Client Update Check
client_data='{
    "current_version": "v0.9.0",
    "channel": "alpha",
    "client_id": "test-client-123",
    "region": "global",
    "arch": "amd64",
    "os": "linux"
}'

make_request "POST" "$CLIENT_API/check-update" "$client_data" "Client Update Check"

# Test 6: Record Download Start
download_data='{
    "version": "v1.0.0-test",
    "client_id": "test-client-123"
}'

make_request "POST" "$CLIENT_API/download-started" "$download_data" "Record Download Started"

# Test 7: Record Install Result
install_data='{
    "version": "v1.0.0-test",
    "client_id": "test-client-123",
    "success": true,
    "error_message": ""
}'

make_request "POST" "$CLIENT_API/install-result" "$install_data" "Record Install Result"

# Test 8: Get Statistics
make_request "GET" "$ADMIN_API/stats?period=7d&action=all" "" "Get Statistics"

# Test 9: Get Version Distribution
make_request "GET" "$ADMIN_API/stats/distribution?period=7d" "" "Get Version Distribution"

# Test 10: Get Region Distribution
make_request "GET" "$ADMIN_API/stats/regions?period=30d" "" "Get Region Distribution"

echo ""
echo "✅ API Testing Complete!"
echo ""
echo "💡 Tips:"
echo "   - Access the admin interface at: $BASE_URL/admin"
echo "   - View API documentation in the README.md file"
echo "   - Check server logs for any issues"
echo ""
echo "🔧 Next Steps:"
echo "   1. Open the admin interface in your browser"
echo "   2. Create and publish a real version"
echo "   3. Test with actual client applications"
