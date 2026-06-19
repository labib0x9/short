#!/bin/bash

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting test suite...${NC}"

# Unit tests
echo -e "\n${YELLOW}Running unit tests...${NC}"
if go test -v ./...; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi

# Worker tests with coverage
echo -e "\n${YELLOW}Running worker tests with coverage...${NC}"
if go test -v -coverprofile=worker.out ./internal/worker; then
    echo -e "${GREEN}✓ Worker tests passed${NC}"
    go tool cover -func=worker.out
else
    echo -e "${RED}✗ Worker tests failed${NC}"
    exit 1
fi

# Overall coverage
echo -e "\n${YELLOW}Generating overall coverage report...${NC}"
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

echo -e "\n${GREEN}All tests completed successfully!${NC}"
