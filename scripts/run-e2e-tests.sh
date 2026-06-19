#!/bin/bash

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting E2E tests...${NC}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running${NC}"
    exit 1
fi

# Start services
echo -e "\n${YELLOW}Starting Docker services...${NC}"
if docker-compose -f docker-compose.test.yml up -d; then
    echo -e "${GREEN}✓ Services started${NC}"
else
    echo -e "${RED}✗ Failed to start services${NC}"
    exit 1
fi

# Wait for services to be healthy
echo -e "\n${YELLOW}Waiting for services to be healthy...${NC}"
sleep 10

# Run migrations
echo -e "\n${YELLOW}Running migrations...${NC}"
if go run ./cmd/migration/main.go -setup && go run ./cmd/migration/main.go -up; then
    echo -e "${GREEN}✓ Migrations completed${NC}"
else
    echo -e "${RED}✗ Migrations failed${NC}"
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Start server in background
echo -e "\n${YELLOW}Starting server...${NC}"
go run ./cmd/short/main.go &
SERVER_PID=$!
sleep 5

# Run E2E tests
echo -e "\n${YELLOW}Running E2E tests...${NC}"

# Test shorten endpoint
echo -e "\n${YELLOW}Testing shorten endpoint...${NC}"
RESPONSE=$(curl -s -X POST http://localhost:3000/short \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.example.com/test"}')

echo $RESPONSE

if echo $RESPONSE | grep -q '"msg":"success"'; then
    echo -e "${GREEN}✓ Shorten endpoint works${NC}"
else
    echo -e "${RED}✗ Shorten endpoint failed${NC}"
    kill $SERVER_PID
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Extract short code
CODE=$(echo $RESPONSE | grep -o '"code":"[^"]*"' | cut -d'"' -f4)
echo "Short code: $CODE"

# Test redirect endpoint
echo -e "\n${YELLOW}Testing redirect endpoint...${NC}"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/$CODE)

if [ $STATUS = "302" ]; then
    echo -e "${GREEN}✓ Redirect endpoint works (HTTP $STATUS)${NC}"
else
    echo -e "${RED}✗ Redirect endpoint failed (HTTP $STATUS)${NC}"
    kill $SERVER_PID
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Test analytics endpoint
echo -e "\n${YELLOW}Testing analytics endpoint...${NC}"
sleep 3  # Wait for worker to process

RESPONSE=$(curl -s http://localhost:3000/$CODE/stat)
echo $RESPONSE

if echo $RESPONSE | grep -q '"short"'; then
    echo -e "${GREEN}✓ Analytics endpoint works${NC}"
else
    echo -e "${RED}✗ Analytics endpoint failed${NC}"
    kill $SERVER_PID
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Cleanup
echo -e "\n${YELLOW}Cleaning up...${NC}"
kill $SERVER_PID 2>/dev/null

if docker-compose -f docker-compose.test.yml down; then
    echo -e "${GREEN}✓ Services stopped${NC}"
else
    echo -e "${RED}✗ Failed to stop services${NC}"
    exit 1
fi

echo -e "\n${GREEN}E2E tests completed successfully!${NC}"
