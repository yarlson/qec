### **Iteration 1: Basic CLI Parsing & YAML File Loading**

**Tasks:**
- Create a basic Go CLI that accepts the same command-line arguments as docker-compose (e.g., multiple `-f` flags and a subcommand like `up -d`).
- Parse input arguments and identify each docker-compose file along with its directory.
- Write unit tests to verify:
    - The CLI correctly parses multiple `-f` flags.
    - It correctly identifies and stores the absolute path and the base directory for each provided file.

**Tests:**
- **Test 1.1:** Run the CLI with two `-f` flags and check that each file’s path and corresponding base directory are correctly identified.
- **Test 1.2:** Validate that invalid file paths produce a clear error message.

---

### **Iteration 2: YAML Merging and Context Adjustments**

**Tasks:**
- Implement functionality to read each docker-compose YAML file into a Go structure (e.g., using a YAML parser).
- Merge multiple YAML files in memory.
- Adjust each file’s build section:
    - Convert relative build contexts to absolute paths based on the file’s directory.
- Write tests to ensure:
    - YAML files are correctly loaded and merged.
    - Relative build contexts in each file are correctly converted to absolute paths.
    - If a build context is already absolute, it remains unchanged.

**Tests:**
- **Test 2.1:** Provide sample YAML files with relative build contexts and verify that the merged result has the correct absolute paths.
- **Test 2.2:** Check that an already absolute build context is unchanged in the merged configuration.

---

### **Iteration 3: Name Prefixing and Dependency Updates**

**Tasks:**
- Prefix service, volume, and (optionally) config/secret names with the folder name.
- Update any references in `depends_on`, `links`, or similar sections to match the new prefixed names.
- Write tests to verify:
    - Services, volumes, and other resource names are prefixed as expected.
    - Cross-service dependencies (e.g., `depends_on`) are updated to the prefixed names.

**Tests:**
- **Test 3.1:** Create two YAML files with services having the same names. Confirm that after merging, the service names are prefixed (e.g., `folder1_serviceA` and `folder2_serviceA`).
- **Test 3.2:** Verify that dependencies (like `depends_on`) are updated to refer to the prefixed service names.
- **Test 3.3:** Check that configs and secrets (if implemented) are also correctly prefixed.

---

### **Iteration 4: Port Mapping Conflict Resolution**

**Tasks:**
- Detect conflicting port mappings between files.
- Implement logic that:
    - Keeps the port mapping from the first file.
    - Removes the port mapping in subsequent files when a conflict is detected.
- Write tests to ensure:
    - For a conflict (e.g., both files define `80:80`), only the first file’s port mapping remains.
    - For multiple conflicts (e.g., second file defines `8080:8080` and third file also defines it), the first occurrence is retained and later ones are removed.

**Tests:**
- **Test 4.1:** Provide YAML files with conflicting port mappings and verify that the merged YAML only includes the first occurrence.
- **Test 4.2:** Validate that non-conflicting port mappings remain intact.

---

### **Iteration 5: Verbose Logging & Dry-Run (Config Command) Integration**

**Tasks:**
- Implement detailed logging that outputs messages about:
    - Adjustments to build contexts.
    - Service/volume prefixing details.
    - Port mapping conflict resolutions.
- Integrate the `--verbose` flag to toggle the logging level.
- Utilize the docker-compose `config` command for validating the merged YAML (dry-run style).
- Write tests to ensure:
    - When `--verbose` is active, logs include all expected details.
    - The output of the `config` command reflects the correct merged configuration without writing to disk.

**Tests:**
- **Test 5.1:** Run the CLI with the `--verbose` flag and capture log output to verify that it contains details about every modification.
- **Test 5.2:** Simulate the `config` command to ensure the merged YAML is valid and contains all adjustments.

---

### **Iteration 6: Integration with External Docker Compose Invocation**

**Tasks:**
- Integrate the final merged YAML with an external docker-compose binary.
- Pipe the merged YAML into a single invocation of docker-compose for any subcommand (up, down, logs, etc.).
- Write tests to verify:
    - The CLI correctly invokes docker-compose with the merged YAML.
    - The output/behavior of docker-compose matches expectations based on the merged configuration.
- Add error handling for any failures returned by docker-compose.

**Tests:**
- **Test 6.1:** Mock or simulate the docker-compose binary invocation and verify that the correct command line is executed with the merged configuration.
- **Test 6.2:** Validate that error messages from docker-compose are properly captured and relayed to the user.

---

### **Final Integration and Documentation**

**Tasks:**
- Document the overall CLI usage, including flags and expected behavior.
- Write integration tests that simulate real-world usage scenarios combining all the features.
- Ensure clear error messages and robust error handling throughout the tool.

**Tests:**
- **Test Final:** End-to-end test using sample YAML files that covers parsing, merging, context adjustments, prefixing, conflict resolution, logging, and external docker-compose invocation.

---

This iterative plan follows TDD principles by defining expected behaviors and corresponding tests for each functionality. Each iteration builds on the previous one, ensuring a robust and maintainable codebase that meets the detailed specification.
