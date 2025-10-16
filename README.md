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

### Build

```bash
go build -o altituder .
```

### CLI Commands

#### Get Elevation at a Location

```bash
# Basic usage
./altituder height --lat 25.0 --lon 55.0

# With custom zoom level
./altituder height --lat 25.197525 --lon 55.274288 --zoom 15

# With custom cache directory
./altituder height --lat 25.0 --lon 55.0 --cache-dir /tmp/cache
```

#### Find Terrain Intersection (Raycast)

```bash
# Basic raycast
./altituder intersection \
  --cam-lat 25.001 \
  --cam-lon 55.729 \
  --cam-alt 177.72 \
  --quat 0.8581,0.0776,-0.1359,0.4899

# With custom parameters
./altituder intersection \
  --cam-lat 25.001 \
  --cam-lon 55.729 \
  --cam-alt 177.72 \
  --quat 1,0,0,0 \
  --step 2 \
  --max-dist 1000
```

#### Start HTTP API Server

```bash
# Start server with defaults
./altituder serve

# Start server with custom port
./altituder serve --addr :9000

# Start with custom configuration
./altituder serve --cache-dir /data/cache --zoom 15
```

### HTTP API

Once the server is running:

```bash
# Query height
curl "http://localhost:8080/height?lat=25.0&lon=55.0&z=14"

# Query intersection
curl "http://localhost:8080/intersection?cam_lat=25.001&cam_lon=55.729&cam_alt=177.72&quat=0.8581,0.0776,-0.1359,0.4899"
```

## Configuration

Configuration can be provided in two ways (flags take precedence over environment variables):

### 1. Command-Line Flags

All commands support these global flags:
- `--cache-dir` - Cache directory for DEM tiles (default: `./cache`)
- `--url-template` - URL template for downloading tiles
- `--subdomains` - Comma-separated list of subdomains (default: `a,b,c`)
- `--zoom` - Default zoom level (default: `14`)
- `--max-native-zoom` - Maximum native zoom level (default: `14`)
- `--height-factor` - Height multiplication factor (default: `1.0`)
- `--nodata-values` - Comma-separated list of no-data values

### 2. Environment Variables

Copy `.env.example` to `.env` and adjust values:

```bash
cp .env.example .env
# Edit .env with your preferred settings
```

Available environment variables:
- `DDM_CACHE_DIR`
- `DDM_URL_TEMPLATE`
- `DDM_SUBDOMAINS`
- `DDM_DEFAULT_Z`
- `DDM_MAX_NATIVE_Z`
- `DDM_HEIGHT_FACTOR`
- `DDM_NODATA_CSV`
- `ADDR` (for HTTP server)

## Testing

```bash
go test ./...
```
