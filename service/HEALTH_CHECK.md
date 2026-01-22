# Health Check Endpoint

## Overview

A health check endpoint has been added to the BaseService to provide service status information.

## Endpoint

### URL
- `/health` - Main health check endpoint
- `/health/` - Alternative health check endpoint

Both endpoints return the same response.

## Response Format

### Success Response (HTTP 200)

```json
{
  "health": {
    "status": "healthy"
  }
}
```

### Example

```bash
$ curl http://localhost:8000/health
{
  "health": {
    "status": "healthy"
  }
}
```

## Implementation Details

The health check is implemented in `service/service.go` as part of the `ServeHTTP` method in `BaseService`:

```go
func (s *BaseService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/health" || path == "/health/" {
		s.healthCheckHandler(w, r)
		return
	}
	
	// ... rest of request handling
}
```

## Features

- **Simple Status**: Returns simple healthy/unhealthy status
- **JSON Format**: Uses standard `application/json` content type
- **Fast Response**: Minimal computation required
- **No Authentication**: Open endpoint for monitoring tools
- **Automatic**: Automatically checks both `/health` and `/health/` paths

## Usage Examples

### Using curl

```bash
# Check health status
curl http://localhost:8000/health

# With verbose output
curl -v http://localhost:8000/health

# Include HTTP status code
curl -i http://localhost:8000/health
```

### Using wget

```bash
wget -q -O - http://localhost:8000/health
cat -
```

### Monitoring Integration

#### Prometheus-compatible

The health endpoint can be used with monitoring systems:

```yaml
# prometheus.yml example
scrape_configs:
  - job_name: 'tileproxy'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8000']
    metrics_path: '/health'
    metric_relabel_configs:
      - source_labels: ['tileproxy']
```

#### Kubernetes livenessProbe

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8000
    initialDelaySeconds: 10
    periodSeconds: 10
    timeoutSeconds: 5
    successThreshold: 1
```

### Docker HEALTHCHECK

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1
```

## Future Enhancements

Potential improvements:

1. **Detailed Metrics**: Add more detailed service metrics
   ```json
   {
     "status": "healthy",
     "timestamp": "2024-01-22T10:00:00Z",
     "version": "1.0.0",
     "uptime": 3600,
     "services": {
       "tile_proxy": "healthy",
       "cache": "healthy"
     }
   }
   }
   ```

2. **Dependency Checks**: Check external dependencies
   - Cache connectivity
   - External service availability
   - Database connections (if any)

3. **Memory/CPU Metrics**: Add resource usage
   ```json
   {
     "status": "healthy",
     "memory_usage": "128MB",
     "cpu_usage": "45%",
     "goroutines": 42
   }
   ```

4. **Configurable Endpoints**: Allow different health check paths
   - `/health/ready` - Readiness probe
   - `/health/live` - Liveness probe
   - `/health/cache` - Cache service check

## Testing

### Manual Test

```bash
# Start the service
cd demo/osm
go run *.go

# In another terminal, test the endpoint
curl http://localhost:8001/health
```

### Automated Test

```bash
# Simple health check
#!/bin/bash
for url in "http://localhost:8001/health" \
             "http://localhost:8001/health/" \
             "http://localhost:8001/nonexistent"; do
  echo "Testing: $url"
  response=$(curl -s -o /dev/null -w "%{http_code}" "$url")
  echo "  Status: $response"
done
```

## Troubleshooting

### No Response

If no response is received:
```bash
# Check if service is running
ps aux | grep "go run"

# Check if port is open
netstat -tuln | grep 8001

# Check logs
tail -f /var/log/syslog
```

### Unexpected Response

If receiving non-200 status codes:
- **404**: Check endpoint path is not registered (shouldn't happen)
- **500**: Internal service error (check logs)
- **503**: Service is not ready (shouldn't happen for health check)

## Notes

- The health check does NOT authenticate or authorize
- The endpoint does NOT log requests (for performance)
- The endpoint is ALWAYS available once the service starts
- Status is always "healthy" (no conditional checks implemented yet)
