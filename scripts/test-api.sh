#!/bin/bash

echo "🔨 Testing API endpoints..."

BASE_URL="http://localhost:8080"

# Test health endpoint
echo "Testing health endpoint..."
health_response=$(curl -s "$BASE_URL/health")
if [[ $health_response == *"healthy"* ]]; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi

# Test provision endpoint
echo "Testing provision endpoint..."
provision_response=$(curl -s -X POST "$BASE_URL/api/v1/provision" \
    -H "Content-Type: application/json" \
    -d '{
        "client_id": "test-client-001",
        "client_name": "Test Client"
    }')

if [[ $provision_response == *"success"* ]]; then
    echo "✅ Provision endpoint working"
    echo "Response: $provision_response"
else
    echo "❌ Provision endpoint failed"
    echo "Response: $provision_response"
    exit 1
fi
