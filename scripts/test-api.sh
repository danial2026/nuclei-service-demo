#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API endpoint
API_URL="http://localhost:3742/api/v1"
DEMO_URL="http://localhost:3743"

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        echo "Response: $3"
    fi
}

# Function to print section header
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

# Function to print info
print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Demo 1: Template Management
print_header "Template Management"

# List all templates
print_info "Listing all available templates..."
response=$(curl -s -w "\n%{http_code}" "${API_URL}/templates")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "List all templates" "$body"

# List templates by tag
print_info "Listing templates with 'vulnerabilities' tag..."
response=$(curl -s -w "\n%{http_code}" "${API_URL}/templates?tag=vulnerabilities")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "List templates by tag" "$body"

# Demo 2: Vulnerability Scanning
print_header "Vulnerability Scanning"

# SQL Injection Scan
print_info "Starting SQL Injection scan..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{
        "target": "'${DEMO_URL}'/vuln/sqli?user=admin'\''--",
        "template_ids": ["http/vulnerabilities/sql-injection.yaml"],
        "tags": ["sqli", "injection", "vulnerabilities"],
        "options": {
            "concurrency": 10,
            "rate_limit": 50,
            "timeout": 15,
            "retries": 2,
            "headless": false
        }
    }' \
    "${API_URL}/scans")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "SQL Injection scan" "$body"

# XSS Scan
print_info "Starting XSS scan..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{
        "target": "'${DEMO_URL}'/vuln/xss?msg=<script>alert(1)</script>",
        "template_ids": ["http/vulnerabilities/xss.yaml"],
        "tags": ["xss", "injection", "vulnerabilities"],
        "options": {
            "concurrency": 10,
            "rate_limit": 50,
            "timeout": 15,
            "retries": 2,
            "headless": false
        }
    }' \
    "${API_URL}/scans")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "XSS scan" "$body"

# SSRF Scan
print_info "Starting SSRF scan..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{
        "target": "'${DEMO_URL}'/vuln/ssrf?url=http://localhost:8080/secret",
        "template_ids": ["http/vulnerabilities/ssrf.yaml"],
        "tags": ["ssrf", "vulnerabilities"],
        "options": {
            "concurrency": 10,
            "rate_limit": 50,
            "timeout": 15,
            "retries": 2,
            "headless": false
        }
    }' \
    "${API_URL}/scans")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "SSRF scan" "$body"

# Command Injection Scan
print_info "Starting Command Injection scan..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{
        "target": "'${DEMO_URL}'/vuln/cmd?cmd=ls%20-la",
        "template_ids": ["http/vulnerabilities/command-injection.yaml"],
        "tags": ["cmd-injection", "vulnerabilities"],
        "options": {
            "concurrency": 10,
            "rate_limit": 50,
            "timeout": 15,
            "retries": 2,
            "headless": false
        }
    }' \
    "${API_URL}/scans")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "Command Injection scan" "$body"

# LFI Scan
print_info "Starting LFI scan..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{
        "target": "'${DEMO_URL}'/vuln/lfi?file=../main.go",
        "template_ids": ["http/vulnerabilities/lfi.yaml"],
        "tags": ["lfi", "vulnerabilities"],
        "options": {
            "concurrency": 10,
            "rate_limit": 50,
            "timeout": 15,
            "retries": 2,
            "headless": false
        }
    }' \
    "${API_URL}/scans")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "LFI scan" "$body"

# Comprehensive Security Scan
print_info "Starting comprehensive security scan..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{
        "target": "'${DEMO_URL}'",
        "template_ids": [
            "http/vulnerabilities/sql-injection.yaml",
            "http/vulnerabilities/xss.yaml",
            "http/vulnerabilities/ssrf.yaml",
            "http/vulnerabilities/command-injection.yaml",
            "http/vulnerabilities/lfi.yaml",
            "http/vulnerabilities/xxe.yaml",
            "http/vulnerabilities/csrf.yaml",
            "http/vulnerabilities/idor.yaml",
            "http/vulnerabilities/open-redirect.yaml"
        ],
        "tags": ["comprehensive", "vulnerabilities", "security-scan"],
        "options": {
            "concurrency": 25,
            "rate_limit": 150,
            "timeout": 30,
            "retries": 3,
            "headless": true,
            "follow_redirects": true,
            "max_redirects": 5,
            "proxy": "",
            "user_agent": "Mozilla/5.0 (compatible; Nuclei/1.0)",
            "custom_headers": {
                "X-API-Key": "demo-key"
            }
        }
    }' \
    "${API_URL}/scans")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "Comprehensive security scan" "$body"

# Extract scan ID from response
scan_id=$(echo "$body" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$scan_id" ]; then
    # Get scan details
    print_info "Getting scan details..."
    response=$(curl -s -w "\n%{http_code}" "${API_URL}/scans/${scan_id}")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    print_result $status_code "Get scan details" "$body"

    # Get scan results
    print_info "Getting scan results..."
    response=$(curl -s -w "\n%{http_code}" "${API_URL}/scans/${scan_id}/results")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    print_result $status_code "Get scan results" "$body"
else
    echo -e "${RED}✗ Could not extract scan ID from response${NC}"
    echo "Response body: $body"
fi

# Demo 3: Error Handling
print_header "Error Handling"

# Test invalid scan ID
print_info "Testing invalid scan ID..."
response=$(curl -s -w "\n%{http_code}" "${API_URL}/scans/invalid-id")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "Invalid scan ID" "$body"

# Test invalid template
print_info "Testing invalid template..."
response=$(curl -s -w "\n%{http_code}" \
    -H "Content-Type: application/json" \
    -d '{
        "target": "'${DEMO_URL}'",
        "template_ids": ["invalid-template.yaml"],
        "tags": ["test", "invalid"],
        "options": {
            "concurrency": 10,
            "rate_limit": 50
        }
    }' \
    "${API_URL}/scans")
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')
print_result $status_code "Invalid template" "$body"

echo -e "\n${GREEN}API demo completed!${NC}" 