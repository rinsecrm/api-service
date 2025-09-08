# API Service

A microservice that provides HTTP API endpoints for the RinseCRM platform.

## Overview

The API service acts as the main entry point for client applications, handling HTTP requests and coordinating with backend services like the Store service via gRPC.

## Architecture

- **Language**: Go 1.25
- **HTTP Framework**: Standard library with custom handlers
- **gRPC Client**: Communicates with Store service
- **Canary Support**: Built-in canary routing via `X-Canary` header

## Development

### Prerequisites

- Go 1.25 or later
- Docker (for containerization)
- Access to the Store service (for gRPC communication)

### Local Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/rinsecrm/api-service.git
   cd api-service
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Build the service**:
   ```bash
   make build
   ```

5. **Run locally**:
   ```bash
   ./bin/api-service
   ```

### Docker Development

```bash
# Build Docker image
docker build -f Dockerfile.ci -t api-service:dev .

# Run with Docker Compose (includes Store service)
docker-compose up
```

## Developer Workflows

### Pull Request Canaries

When you create or update a Pull Request, the CI/CD system automatically:

1. **Builds a canary Docker image** tagged with `pr-{PR_NUMBER}`
2. **Deploys to integration environment** with canary routing
3. **Creates isolated test environment** for your changes

#### Testing Your PR Canary

Once your PR is deployed, you can test it by adding the `X-Canary` header to your requests:

```bash
# Test your PR canary (replace 123 with your PR number)
curl -H "X-Canary: 123" https://api.dev.example.com/health
```

#### PR Canary Lifecycle

- **Created**: When PR is opened or updated
- **Updated**: When you push new commits to the PR
- **Cleaned up**: Automatically when PR is closed
- **Image cleanup**: Old PR images are cleaned up after 4 weeks

### Creating a Release

To create a new release:

1. **Create a Git Tag**:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

2. **Automated Process**:
   The release workflow automatically:
   - Builds and pushes Docker images (`v1.2.3` and `latest`)
   - Creates GitHub Release with details
   - **Integration environment** gets `latest` immediately
   - **Staging environment** gets a PR for `v1.2.3`
   - **Production environment** gets a PR for `v1.2.3`

3. **Deployment Process**:
   - **Integration**: Automatically updated to latest release
   - **Staging**: Review and merge the staging PR
   - **Production**: Review and merge the production PR (after staging is tested)

#### Release PR Titles

- `Staging: Release API Service v1.2.3`
- `Production: Release API Service v1.2.3`

### Environment Strategy

- **Integration**: Always runs `latest` (latest production release)
- **Staging**: Runs specific version (review before production)
- **Production**: Runs specific version (review before deployment)
- **PR Canaries**: Run `pr-{NUMBER}` (isolated from releases)

## Configuration

### Environment Variables

- `STORE_SERVICE_ADDR`: Address of the Store service (default: `store.apps:80`)
- `PORT`: HTTP server port (default: `8080`)

### Canary Headers

- `X-Canary`: PR number for canary routing (e.g., `123`)

## API Endpoints

### Health Check
```
GET /health
```

### Store Operations
```
GET /store/{key}
POST /store
PUT /store/{key}
DELETE /store/{key}
```

## Monitoring

The service includes:
- Health check endpoint for monitoring
- Structured logging
- gRPC client metrics
- Canary request tracking

## Troubleshooting

### Common Issues

1. **Canary not working**: Ensure you're using the correct `X-Canary` header format
2. **Store service connection**: Check that the Store service is running and accessible
3. **PR canary not deployed**: Check the GitHub Actions workflow logs

### Debugging

```bash
# Check service logs
kubectl logs -f deployment/api -n apps

# Check canary routing
kubectl logs -f deployment/api-canary-pr-123 -n apps
```

## Contributing

1. Create a feature branch
2. Make your changes
3. Add tests
4. Create a Pull Request
5. Test your canary deployment
6. Request review and merge

## License

[Add your license information here]
