# altituder

A Go-based terrain elevation and raycast intersection service.

## Features

- **Height Lookup**: Get terrain elevation at any geographic coordinate
- **Intersection Search**: Perform raycast to find terrain intersections (e.g., for drone camera views)
- **Tile Caching**: Efficient multi-level caching (memory + disk)
- **Reusable API**: Business logic separated from HTTP handlers for maximum reusability

## Architecture

The codebase has been refactored to provide clean separation between HTTP handlers and business logic:

- **`SearchIntersection`**: Reusable function for terrain raycast operations
- **`PickHeight`**: Reusable function for elevation lookups

These functions can be used in:
- HTTP API endpoints (current implementation)
- CLI tools
- Batch processors
- gRPC or other API protocols
- Internal service-to-service calls

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed documentation and usage examples.

## Quick Start

```bash
# Build
go build -o altituder .

# Run server
./altituder serve

# Query height
curl "http://localhost:8080/height?lat=25.0&lon=55.0&z=14"

# Query intersection
curl "http://localhost:8080/intersection?cam_lat=25.001&cam_lon=55.729&cam_alt=177.72&quat=0.8581,0.0776,-0.1359,0.4899"
```

## Testing

```bash
go test ./...
```
