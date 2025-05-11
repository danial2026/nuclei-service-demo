# Nuclei Service

A demo service showcasing Nuclei vulnerability scanning capabilities.

## Overview

This service provides a robust API for managing Nuclei vulnerability scans, template management, and result storage. It's designed to handle high-volume scanning operations while maintaining performance and reliability. The service includes a background worker for processing scans asynchronously and storing results in PostgreSQL.

## Key Features

- REST API for scans and templates
- PostgreSQL storage
- Docker deployment
- Zap logging
- Demo vulnerable routes
- Configurable scans
- Async processing
- Status tracking
- Build automation

## Requirements

- Go 1.21+
- PostgreSQL 12+
- Docker & Docker Compose
- Nuclei (for local dev)

## Quick Start

1. Clone and setup:
```bash
git clone https://github.com/danial2026/nuclei-service-demo.git
cd nuclei-service-demo
./setup.bash
```

2. Configure your environment:
```bash
# Edit .env with your settings
vim .env
```

3. Launch with Docker:
```bash
docker-compose up --build
```

## Project Structure

```markdown
.
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── model/            # Data models
│   ├── repository/       # Database operations
│   ├── service/          # Business logic
│   └── server/           # HTTP server and handlers
├── docker/               # Docker-related files
├── scripts/              # Utility scripts
├── build.sh             # Build script
├── setup.bash           # Setup script
├── Dockerfile           # Main service Dockerfile
└── docker-compose.yml   # Docker Compose configuration
```

## API Reference

### Templates

#### List Templates
```http
GET /api/v1/templates
```

Query Parameters:
- `tags`: Filter by template tags
- `author`: Filter by template author
- `severity`: Filter by severity level
- `type`: Filter by template type

Response:
```json
[
  {
    "id": "string",
    "name": "string",
    "author": "string",
    "tags": ["string"],
    "severity": "string",
    "type": "string",
    "description": "string",
    "created_at": "string",
    "updated_at": "string"
  }
]
```

#### Get Template Details
```http
GET /api/v1/templates/{id}
```

#### Refresh Template Cache
```http
POST /api/v1/templates/refresh
```

### Scans

#### List Scans
```http
GET /api/v1/scans
```

Query Parameters:
- `status`: Filter by scan status
- `target`: Filter by target URL
- `template_id`: Filter by template ID

#### Start New Scan
```http
POST /api/v1/scans
```

Request:
```json
{
  "target": "string",
  "template_ids": ["string"],
  "tags": ["string"],
  "options": {
    "concurrency": 10,
    "rate_limit": 100,
    "timeout": 30,
    "retries": 3,
    "headless": false
  }
}
```

#### Get Scan Results
```http
GET /api/v1/scans/{id}/results
```

## Demo Server

The service includes a demo server that exposes intentionally vulnerable endpoints for testing purposes. These endpoints simulate common security vulnerabilities and can be used to test the Nuclei scanner.

### Vulnerable Endpoints

1. **Open Redirect**
```http
GET /vuln/openredirect?redirect=<url>
```
Vulnerable to open redirect attacks.

2. **Oracle Fatwire LFI**
```http
GET /vuln/lfi-fatwire?fn=<file_path>
```
Vulnerable to Local File Inclusion attacks.

3. **HiBoss RCE**
```http
GET /vuln/hiboss-rce?ip=<ip_address>
```
Vulnerable to Remote Command Execution via ping command.

4. **ThinkPHP Arbitrary File Write**
```http
GET /vuln/thinkphp-write?content=<content>
```
Vulnerable to arbitrary file write operations.

5. **Zyxel Unauthenticated LFI**
```http
GET /vuln/zyxel-lfi?path=<file_path>
```
Vulnerable to Local File Inclusion attacks.

6. **Nuxt.js XSS**
```http
GET /vuln/nuxt-xss?stack=<payload>
```
Vulnerable to Cross-Site Scripting (XSS) attacks.

7. **Sick-Beard XSS**
```http
GET /vuln/sickbeard-xss?pattern=<payload>
```
Vulnerable to Cross-Site Scripting (XSS) attacks.

8. **Fastjson Deserialization RCE**
```http
POST /vuln/fastjson-rce
```
Vulnerable to Remote Code Execution via JSON deserialization.

9. **BeyondTrust XSS**
```http
GET /vuln/beyondtrust-xss?input=<payload>
```
Vulnerable to Cross-Site Scripting (XSS) attacks.

10. **WordPress Brandfolder Open Redirect**
```http
GET /vuln/brandfolder-redirect?url=<redirect_url>
```
Vulnerable to open redirect attacks.

> **Warning**: These endpoints are intentionally vulnerable and should only be used in controlled testing environments. Do not expose the demo server to production or untrusted networks.

## Development

1. Setup:
```bash
./setup.bash
```

2. Run tests:
```bash
go test ./...
```

3. Build the service:
```bash
./build.sh
```

4. Start service:
```bash
go run cmd/server/main.go
```

## Deployment

The service is containerized for easy deployment. The stack includes:

- API service
- PostgreSQL database
- Nuclei scanner service
- Background worker for scan processing

Deploy with:
```bash
docker-compose up -d
```

## Contributing

Found a bug? Have a feature request? Contributions are welcome!

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m 'Add your feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

## Contact

- Website: [danials.space](https://danials.space)
- Email: [danial@danials.space](mailto:danial@danials.space)

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Built with ❤️ by [Danial](https://danials.space) | May 2025