# qec - Quantum Entanglement Communicator for Docker Compose

`qec` is a tool that extends Docker Compose to handle multiple compose files with automatic context path adjustments. It solves the common issue of managing multiple Docker Compose files across different directories by automatically handling build contexts and resource naming.

## Features

- **Docker Compose CLI Compatible**: 
  - Supports both standalone `docker-compose` and the newer Docker Compose plugin
  - Preserves Docker Compose's command-line interface and functionality
  - Automatically detects and uses the appropriate Docker Compose executable

- **Automatic Context Path Handling**: 
  - Each compose file uses its own directory as the base path
  - Automatically converts relative build contexts to absolute paths
  - Preserves absolute paths when already specified
  - Supports `.env` file discovery in each compose file's directory

- **Smart Resource Naming**: 
  - Prefixes services, volumes, configs, and secrets with their source folder name
  - Updates all internal references to maintain consistency:
    - Service dependencies (`depends_on`)
    - Service links
    - Volume mounts
    - Config and secret references

- **Port Conflict Resolution**: 
  - Detects port conflicts between services
  - Resolves conflicts by applying an offset to subsequent services
  - Maintains original port mappings for the first occurrence
  - Logs all port adjustments for transparency

- **Logging and Debugging**:
  - Detailed logging of all operations with `--verbose` flag
  - Logs include:
    - Build context adjustments
    - Resource name prefixing
    - Port conflict resolutions
    - Command execution details

- **Dry Run Support**: 
  - Preview changes without applying them using `--dry-run`
  - Shows the merged configuration and planned actions
  - Validates the configuration without executing commands

## Installation

```bash
go install github.com/yarlson/qec@latest
```

## Usage

```bash
qec [OPTIONS] COMMAND [ARGS...]
```

### Options

- `-f, --file FILE`: Path to a docker-compose YAML file (can be specified multiple times)
- `-d, --detach`: Run containers in the background
- `--dry-run`: Simulate configuration without making runtime changes
- `--verbose`: Enable verbose logging
- `--command COMMAND`: Command to execute (default: "up")
- `-h, --help`: Show help text

### Commands

All standard Docker Compose commands are supported:
- `up`: Create and start containers
- `down`: Stop and remove containers, networks, and volumes
- `ps`: List containers
- `logs`: View output from containers
- `build`: Build or rebuild services
- `pull`: Pull service images
- `push`: Push service images
- `config`: Validate and view the merged configuration

## Examples

### Running Services from Multiple Compose Files

```bash
# Start services from multiple directories
qec -f web/docker-compose.yml -f db/docker-compose.yml up -d

# The tool will:
# 1. Prefix web services as web_service_name
# 2. Prefix db services as db_service_name
# 3. Adjust build contexts to be relative to each file's directory
# 4. Update all internal references
```

### Viewing Merged Configuration

```bash
# View the final merged configuration
qec -f web/docker-compose.yml -f db/docker-compose.yml --command config

# Shows:
# - Prefixed service names
# - Adjusted build contexts
# - Updated dependencies
# - Resolved port conflicts
```

### Dry Run with Verbose Logging

```bash
# Preview changes with detailed logging
qec -f web/docker-compose.yml -f db/docker-compose.yml --dry-run --verbose up

# Outputs:
# - Resource prefixing details
# - Build context adjustments
# - Port conflict resolutions
# - Planned Docker Compose commands
```

## Implementation Details

### Resource Prefixing

The tool prefixes resources based on their source directory:

```yaml
# Original (in web/docker-compose.yml):
services:
  frontend:
    image: nginx
  api:
    image: node
volumes:
  data:

# Becomes:
services:
  web_frontend:
    image: nginx
  web_api:
    image: node
volumes:
  web_data:
```

### Build Context Handling

Build contexts are automatically adjusted:

```yaml
# Original (in web/docker-compose.yml):
services:
  api:
    build:
      context: ./api
      dockerfile: Dockerfile

# Becomes:
services:
  web_api:
    build:
      context: /absolute/path/to/web/api
      dockerfile: Dockerfile
```

### Port Conflict Resolution

Port conflicts are resolved by adding an offset:

```yaml
# First file (web/docker-compose.yml):
services:
  nginx:
    ports:
      - "80:80"      # Kept as is

# Second file (api/docker-compose.yml):
services:
  nginx:
    ports:
      - "80:80"      # Changed to "180:80"
```

### Volume References

Volume references are updated to match prefixed names:

```yaml
# Original:
services:
  postgres:
    volumes:
      - db_data:/var/lib/postgresql/data
volumes:
  db_data:

# Becomes:
services:
  db_postgres:
    volumes:
      - db_db_data:/var/lib/postgresql/data
volumes:
  db_db_data:
```

## Error Handling

The tool includes comprehensive error handling:
- Validates Docker Compose installation and version
- Checks for file existence and YAML validity
- Verifies build context paths
- Ensures port conflict resolution is possible
- Provides clear error messages with context

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.
