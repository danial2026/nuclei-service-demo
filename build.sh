#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${YELLOW}==>${NC} $1"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Build Go binary
print_status "Building Go binary..."
CGO_ENABLED=0 GOOS=linux go build -o nuclei-service-demo ./cmd/main.go
if [ $? -eq 0 ]; then
    print_success "Go binary built successfully"
else
    print_error "Failed to build Go binary"
    exit 1
fi

print_success "Build completed successfully!"
