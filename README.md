# qec - Quantum Entanglement Communicator for Docker Compose

`qec` solves common Docker Compose pain points when working with microservices spread across multiple directories. It eliminates manual path adjustments, naming conflicts, and port collision headaches that developers face when combining multiple compose files.

## Developer Experience Problems Solved

### 1. No More Path Juggling
**Problem**: When combining compose files from different directories, Docker Compose uses the first file's directory as the base for all build contexts, breaking builds in other directories.

**Solution**: `qec` automatically handles build contexts:
```yaml
# web/docker-compose.yml
services:
  api:
    build: ./api  # Works! Uses web/api

# db/docker-compose.yml
services:
  worker:
    build: ./worker  # Also works! Uses db/worker
```

### 2. Automatic Conflict Resolution
**Problem**: Services from different compose files often have the same names and port mappings, requiring manual renaming and port adjustments.

**Solution**: `qec` automatically prefixes resources and adjusts ports:
```yaml
# Before (conflicting names and ports)
# web/docker-compose.yml
services:
  api:
    ports: ["3000:3000"]
# db/docker-compose.yml
services:
  api:
    ports: ["3000:3000"]

# After (automatically resolved)
services:
  web_api:
    ports: ["3000:3000"]
  db_api:
    ports: ["3100:3000"]
```

### 3. Consistent Volume Management
**Problem**: Volume names clash between different compose files, leading to data mixing or manual prefixing.

**Solution**: `qec` handles volume prefixing and reference updates:
```yaml
# Before (volumes would conflict)
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

# After (automatically namespaced)
services:
  web_api:
    volumes: ["web_data:/app/data"]
  db_postgres:
    volumes: ["db_data:/var/lib/postgresql/data"]
volumes:
  web_data:
  db_data:
```

### 4. Dependency Management Made Easy
**Problem**: Service dependencies break when combining files due to name conflicts and manual prefixing.

**Solution**: `qec` automatically updates all references:
```yaml
# Before
services:
  api:
    depends_on: ["redis", "postgres"]
  redis:
    image: redis
  postgres:
    image: postgres

# After
services:
  cache_api:
    depends_on: ["cache_redis", "db_postgres"]
  cache_redis:
    image: redis
  db_postgres:
    image: postgres
```

## Usage

Simple drop-in replacement for docker-compose:

```bash
# Instead of:
docker-compose -f web/docker-compose.yml -f db/docker-compose.yml up

# Use:
qec -f web/docker-compose.yml -f db/docker-compose.yml up
```

### Preview Changes

See exactly what `qec` will do before running:

```bash
qec -f web/docker-compose.yml -f db/docker-compose.yml --dry-run --verbose up
```

### Options

- `-f, --file FILE`: Multiple compose files (maintains Docker Compose compatibility)
- `-d, --detach`: Run in background
- `--dry-run`: Preview changes
- `--verbose`: Show detailed adjustments
- `--command`: Any Docker Compose command (`up`, `down`, `logs`, etc.)
- `-h, --help`: Show help

## Installation

```bash
go install github.com/yarlson/qec@latest
```

## Technical Details

### Automatic Adjustments

- **Build Contexts**: Converts relative paths to absolute based on each file's location
- **Resource Names**: Prefixes with directory name (e.g., `web_`, `db_`)
- **Port Conflicts**: Resolves by adding offset to later services
- **Volume References**: Updates all volume mounts to match prefixed names
- **Dependencies**: Updates `depends_on` and `links` to use prefixed names

### Safety Features

- **Dry Run Mode**: Preview all changes before applying
- **Validation**: Checks for invalid configurations
- **Clear Logging**: Shows exactly what's being changed
- **Error Handling**: Clear messages for common issues

## Contributing

Found a DX issue we haven't solved? Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.
