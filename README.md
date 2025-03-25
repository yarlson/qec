# qec - Quantum Entanglement Communicator for Docker Compose

`qec` is a tool that extends Docker Compose to handle multiple compose files with automatic context path adjustments. It solves the common issue of managing multiple Docker Compose files across different directories by automatically handling build contexts and resource naming.

## Features

- **Docker Compose CLI Compatible**: Uses the same command-line interface as Docker Compose
- **Automatic Context Path Handling**: Each compose file uses its own directory as the base path
- **Smart Resource Naming**: Automatically prefixes services, volumes, and other resources with their source folder name to avoid conflicts
- **Port Conflict Resolution**: Intelligently handles port mapping conflicts between files
- **Dependency Management**: Automatically updates service dependencies to match prefixed names
- **Dry Run Support**: Preview changes without applying them
- **Verbose Logging**: Detailed logging of all adjustments and changes

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

### Commands

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
qec -f folder1/docker-compose.yml -f folder2/docker-compose.yml up -d
```

### Viewing Merged Configuration

```bash
qec -f folder1/docker-compose.yml -f folder2/docker-compose.yml --command config
```

### Dry Run to Preview Changes

```bash
qec -f folder1/docker-compose.yml -f folder2/docker-compose.yml --dry-run up
```

## How It Works

1. **File Detection**: Identifies each docker-compose file and determines its directory.

2. **Merge Process**:
   - Reads all YAML files
   - Adjusts each file's context paths to use its folder as the base
   - Converts relative build contexts to absolute paths
   - Prefixes all resource names with their folder name
   - Resolves port conflicts by keeping the first file's definition
   - Updates service dependencies to include folder prefixes

3. **Execution**:
   - Merges configurations in memory
   - Pipes the final configuration directly to Docker Compose
   - Executes the requested command

## Resource Naming

Resources are automatically prefixed with their source folder name to avoid conflicts:

- Services: `folder1_service1`, `folder2_service1`
- Volumes: `folder1_volume1`, `folder2_volume1`
- Networks: Preserved as-is (shared across files)
- Configs/Secrets: `folder1_config1`, `folder2_config1`

## Port Conflict Resolution

Port conflicts are resolved based on file order:
- First file's port mappings take precedence
- Conflicting mappings in subsequent files are removed
- Non-conflicting ports are preserved

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details 