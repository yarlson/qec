# qec - Quantum Entanglement Communicator for Docker Compose

`qec` makes working with multiple Docker Compose files simple and predictable. It automatically handles path adjustments, naming conflicts, and port collisions when combining microservices from different directories.

## About

`qec` is a command-line tool that extends Docker Compose to intelligently handle multiple compose files. It solves common microservices orchestration challenges by:

- Maintaining correct build contexts for each service's directory
- Automatically prefixing service names to prevent conflicts
- Resolving port collisions with smart offset calculation
- Keeping volume data isolated between services
- Preserving service dependencies across files

It acts as a drop-in replacement for `docker-compose`, requiring minimal changes to your existing workflow.

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
Multiple services often share the same names or want the same ports. `qec` handles this automatically by prefixing service names with their directory and offsetting conflicting ports by 100:

```yaml
# Before
# web/docker-compose.yml
services:
  api:
    ports: ["80:80"]
# db/docker-compose.yml
services:
  api:
    ports: ["80:80"]

# After (automatically handled)
services:
  web_api:
    ports: ["80:80"]  # First file keeps original port
  db_api:
    ports: ["180:80"]  # Second file gets offset by 100
```

### 3. Volume Name Conflicts
Shared volume names between different compose files can lead to data mixing. `qec` keeps data isolated by prefixing volume names with their directory name:

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
    volumes: ["web_web_data:/app/data"]  # Directory name is used as prefix
  db_postgres:
    volumes: ["db_db_data:/var/lib/postgresql/data"]  # Directory name is used as prefix
volumes:
  web_web_data:  # Volume names get directory prefix
  db_db_data:
```

### 4. Service Dependencies
References between services break when combining files. `qec` maintains all connections by updating references with directory prefixes:

```yaml
# Before
# web/docker-compose.yml
services:
  api:
    depends_on: ["redis", "postgres"]
  redis:
    image: redis
  postgres:
    image: postgres

# After (all references updated)
services:
  web_api:
    depends_on: ["web_redis", "web_postgres"]  # References updated with directory prefix
  web_redis:
    image: redis
  web_postgres:
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
- Resolves port conflicts by adding offset of 100 to subsequent files
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
