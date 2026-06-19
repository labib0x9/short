#!/bin/bash

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting load tests...${NC}"

# Check if server is running
if ! curl -s http://localhost:3000/aB3kZ9 > /dev/null 2>&1; then
    echo -e "${RED}✗ Server is not running at http://localhost:3000${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Server is running${NC}"

# Check if vegeta is installed
if ! command -v vegeta &> /dev/null; then
    echo -e "${YELLOW}vegeta not installed, installing...${NC}"
    go install github.com/tsenart/vegeta@latest
fi

echo -e "\n${YELLOW}Running load test (100 req/s for 30s)...${NC}"

# Run load test
vegeta attack -duration=30s -rate=100 -targets=tests/load_test.txt | vegeta report

echo -e "\n${YELLOW}Generating detailed report...${NC}"
vegeta attack -duration=30s -rate=100 -targets=tests/load_test.txt | vegeta dump > load_results.json
echo -e "${GREEN}✓ Results saved to load_results.json${NC}"

echo -e "\n${GREEN}Load tests completed!${NC}"
