#!/usr/bin/env bash

# Download templates
git clone https://github.com/projectdiscovery/nuclei-templates.git templates

# Start docker compose
cd docker
docker-compose -f docker-compose.yml up -d --build
