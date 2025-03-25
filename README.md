# qec - Quantum Entanglement Communicator for Docker Compose

`qec` makes working with multiple Docker Compose files simple and predictable. It automatically handles path adjustments, naming conflicts, and port collisions when combining microservices from different directories.

## Common Problems Solved

### 1. Build Context Resolution
When combining compose files from different directories, Docker Compose defaults to using the first file's directory as the base path. This breaks builds in other directories. `qec` fixes this automatically:

```yaml
# web/docker-compose.yml
services:
  api:
    build: ./api  # Uses web/api as expected

# db/docker-compose.yml
services:
  worker:
    build: ./worker  # Uses db/worker as expected
```

### 2. Service Name and Port Conflicts
Multiple services often share the same names or want the same ports. `qec` handles this automatically:

```yaml
# Before
# web/docker-compose.yml
services:
  api:
    ports: ["3000:3000"]
# db/docker-compose.yml
services:
  api:
    ports: ["3000:3000"]

# After (automatically handled)
services:
  web_api:
    ports: ["3000:3000"]
  db_api:
    ports: ["3100:3000"]
```

### 3. Volume Name Conflicts
Shared volume names between different compose files can lead to data mixing. `qec` keeps data isolated:

```yaml
# Before
# web/docker-compose.yml
services:
  api:
    volumes: ["data:/app/data"]
volumes:
  data:

# db/docker-compose.yml
services:
  postgres:
    volumes: ["data:/var/lib/postgresql/data"]
volumes:
  data:

# After (automatically isolated)
services:
  web_api:
    volumes: ["web_data:/app/data"]
  db_postgres:
    volumes: ["db_data:/var/lib/postgresql/data"]
volumes:
  web_data:
  db_data:
```

### 4. Service Dependencies
References between services break when combining files. `qec` maintains all connections:

```yaml
# Before
services:
  api:
    depends_on: ["redis", "postgres"]
  redis:
    image: redis
  postgres:
    image: postgres

# After (all references updated)
services:
  cache_api:
    depends_on: ["cache_redis", "db_postgres"]
  cache_redis:
    image: redis
  db_postgres:
    image: postgres
```

## Quick Start

Replace `docker-compose` with `qec`:

```bash
# Instead of:
docker-compose -f web/docker-compose.yml -f db/docker-compose.yml up

# Use:
qec -f web/docker-compose.yml -f db/docker-compose.yml up
```

### Preview Mode

See what changes will be made before applying them:

```bash
qec -f web/docker-compose.yml -f db/docker-compose.yml --dry-run --verbose up
```

### Available Options

- `-f, --file FILE`: Specify compose files (same as docker-compose)
- `-d, --detach`: Run in background
- `--dry-run`: Preview changes
- `--verbose`: Show detailed adjustments
- `--command`: Any Docker Compose command (`up`, `down`, `logs`, etc.)
- `-h, --help`: Show help

## Installation

```bash
go install github.com/yarlson/qec@latest
```

## How It Works

### Automatic Adjustments
- Converts relative paths to absolute based on file location
- Prefixes resources with directory names (e.g., `web_`, `db_`)
- Resolves port conflicts by adding offsets
- Updates volume mounts to match prefixed names
- Maintains service dependencies and links

### Safety Features
- Preview mode to review changes
- Configuration validation
- Detailed logging
- Clear error messages

## Contributing

Found a problem we haven't solved? Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.
