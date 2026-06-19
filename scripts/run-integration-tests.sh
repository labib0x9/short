#!/bin/bash

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting integration tests...${NC}"

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

# Run integration tests
echo -e "\n${YELLOW}Running integration tests...${NC}"
if go test -v -tags=integration ./...; then
    echo -e "${GREEN}✓ Integration tests passed${NC}"
else
    echo -e "${RED}✗ Integration tests failed${NC}"
    docker-compose -f docker-compose.test.yml down
    exit 1
fi

# Cleanup
echo -e "\n${YELLOW}Stopping services...${NC}"
if docker-compose -f docker-compose.test.yml down; then
    echo -e "${GREEN}✓ Services stopped${NC}"
else
    echo -e "${RED}✗ Failed to stop services${NC}"
    exit 1
fi

echo -e "\n${GREEN}Integration tests completed successfully!${NC}"
