### Overview

The tool will be a Go-based CLI that emulates the Docker Compose CLI interface but with a key difference in how it handles Docker context paths when multiple docker-compose YAML files are provided. Instead of using the first file’s folder as the base context for all files, the tool will merge the files and automatically adjust each file’s build contexts and other path-dependent elements so that each uses its own folder as the base path.

---

### Primary Features

1. **CLI Interface Compatibility**
    - The CLI will mimic the standard Docker Compose CLI interface.
    - It will accept multiple `-f` options (e.g., `docker-compose -f folder1/docker-compose-1.yml -f folder2/docker-compose-2.yml up -d`) and process them as the native docker-compose command.

2. **Folder-Based Context Handling**
    - Each provided docker-compose file will automatically use its own directory as the Docker context base path.
    - For build sections with relative paths, these will be converted to absolute paths based on the location of the YAML file.
    - .env file autodiscovery will be based on each file’s folder.

3. **Dynamic Merging & Adjustments**
    - **Merging Files:**
        - The tool will merge multiple YAML files into one effective configuration.
        - The merging will be done on the fly (in-memory), and the merged YAML will be piped directly into Docker Compose without writing a temporary file to disk.
    - **Name Prefixing:**
        - To avoid name conflicts, services, volumes, and (for best developer experience) configs and secrets will be prefixed with their source folder name.
    - **Dependency Adjustments:**
        - References between services (such as in `depends_on` or `links`) will be updated automatically to include the folder prefix, ensuring internal consistency.
    - **Port Forwarding Conflicts:**
        - The file order (first file has highest priority) determines port mapping priority.
        - If the first file defines a port (e.g., `80:80`) and a subsequent file defines the same mapping, then the conflicting port forwarding in the later file(s) will be removed.
        - For subsequent conflicts (e.g., the second file defines `8080:8080` and a third file defines the same), the third file’s mapping is removed.

4. **Global Application of Adjustments**
    - The merging, context adjustments, and name prefixing will apply to all Docker Compose subcommands (such as `up`, `down`, `logs`, etc.).

5. **Logging & Warnings**
    - The tool will output detailed log messages during the merging process.
    - These logs will include information about:
        - Which port mappings were removed due to conflicts.
        - How service, volume, config, and secret names were prefixed.
        - Adjustments made to build contexts (e.g., conversion from relative to absolute paths).
    - A verbose logging option will be available via a command-line flag (e.g., `--verbose`).

6. **Integration with Docker Compose**
    - The CLI will invoke an external Docker Compose binary under the hood.
    - The final, merged YAML configuration (with all adjustments) will be piped directly into Docker Compose using the `config` command or equivalent.

7. **Error Handling**
    - The tool will attempt to resolve conflicts automatically.
    - If it encounters a conflict or an error it cannot automatically resolve, it will fail fast, outputting a clear error message and exiting.

8. **Command-Line Flags Only**
    - All configuration (including logging levels and naming conventions) will be provided via command-line flags rather than external configuration files.

9. **Absolute Paths Handling**
    - For build contexts that are already specified as absolute paths in the docker-compose file, no changes or validations will be applied.

10. **Version Compatibility**
    - The tool does not need to handle version-specific behaviors for Docker Compose and will assume a standard baseline version.

---

### Example Workflow

- **Input Command:**
  ```
  docker-compose -f folder1/docker-compose-1.yml -f folder2/docker-compose-2.yml up -d
  ```

- **Processing Steps:**
    1. **File Detection:**
        - Identify each docker-compose file and determine its directory.
    2. **Merge Process:**
        - Read both YAML files.
        - Adjust each file’s context paths to use its folder as the base.
        - Convert relative build contexts to absolute paths based on the file’s directory.
        - Prefix all service, volume, and additional resource names (configs, secrets) with the respective folder name.
        - Resolve port conflicts by keeping the first file’s definition and removing subsequent duplicates.
        - Update any service dependencies (like `depends_on`) to include the folder prefix.
    3. **Logging:**
        - Emit verbose logs if the `--verbose` flag is set, detailing every adjustment made.
    4. **Execution:**
        - Pipe the final merged YAML directly into a single invocation of Docker Compose.
        - Use Docker Compose’s `config` command internally to ensure the validity of the merged configuration before execution.

- **Output:**
    - Docker Compose runs using the adjusted configuration ensuring each file’s context is preserved.

---

### Final Notes

- The tool is intended to provide a better developer experience (DX) by automatically managing potential conflicts and adjusting configurations, so the users can continue to work as if they were using standard Docker Compose while benefiting from per-file context isolation.
- Future enhancements might include additional flags to tweak prefix rules or conflict resolution strategies, but the base version will use the outlined defaults.
