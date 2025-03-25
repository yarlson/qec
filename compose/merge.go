package compose

import (
	"fmt"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
)

// ComposeFile represents a Docker Compose file with its metadata
type ComposeFile struct {
	Path    string
	BaseDir string
	Project *types.Project
	logger  *logrus.Entry
}

// NewComposeFile creates a new ComposeFile instance
func NewComposeFile(path string, logger *logrus.Entry) (*ComposeFile, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	baseDir := filepath.Dir(absPath)

	// Create project options with the file's base directory
	options, err := cli.NewProjectOptions(
		[]string{absPath},
		cli.WithWorkingDirectory(baseDir),
		cli.WithOsEnv,
		cli.WithDotEnv,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project options: %w", err)
	}

	// Load the project
	project, err := cli.ProjectFromOptions(options)
	if err != nil {
		return nil, fmt.Errorf("failed to load project from %s: %w", path, err)
	}

	return &ComposeFile{
		Path:    absPath,
		BaseDir: baseDir,
		Project: project,
		logger:  logger.WithField("file", absPath),
	}, nil
}

// adjustBuildContexts converts relative build contexts to absolute paths
func (cf *ComposeFile) adjustBuildContexts() error {
	for name, service := range cf.Project.Services {
		if service.Build == nil {
			continue
		}

		// If context is relative, make it absolute using the file's base directory
		if !filepath.IsAbs(service.Build.Context) {
			absContext := filepath.Join(cf.BaseDir, service.Build.Context)
			cf.logger.Debugf("Converting build context for service %s from %s to %s",
				name, service.Build.Context, absContext)
			service.Build.Context = absContext
		}
	}
	return nil
}

// MergeComposeFiles merges multiple compose files, adjusting build contexts and handling overrides
func MergeComposeFiles(files []*ComposeFile) (*types.Project, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no compose files provided")
	}

	// Use the first file's project as the base
	baseProject := files[0].Project

	// Adjust build contexts for the base project
	if err := files[0].adjustBuildContexts(); err != nil {
		return nil, fmt.Errorf("failed to adjust build contexts for %s: %w", files[0].Path, err)
	}

	// Merge additional files
	for i := 1; i < len(files); i++ {
		cf := files[i]

		// Adjust build contexts for the current file
		if err := cf.adjustBuildContexts(); err != nil {
			return nil, fmt.Errorf("failed to adjust build contexts for %s: %w", cf.Path, err)
		}

		// Merge services
		for name, service := range cf.Project.Services {
			if existing, exists := baseProject.Services[name]; exists {
				// Service exists in base project, merge configurations
				merged, err := mergeServices(&existing, &service)
				if err != nil {
					return nil, fmt.Errorf("failed to merge service %s: %w", name, err)
				}
				baseProject.Services[name] = *merged
			} else {
				// New service, add it to base project
				baseProject.Services[name] = service
			}
		}

		// Merge volumes
		for name, volume := range cf.Project.Volumes {
			if _, exists := baseProject.Volumes[name]; !exists {
				if baseProject.Volumes == nil {
					baseProject.Volumes = make(types.Volumes)
				}
				baseProject.Volumes[name] = volume
			}
			// If volume exists, keep the first definition (from base project)
		}

		// Merge networks
		for name, network := range cf.Project.Networks {
			if _, exists := baseProject.Networks[name]; !exists {
				if baseProject.Networks == nil {
					baseProject.Networks = make(types.Networks)
				}
				baseProject.Networks[name] = network
			}
			// If network exists, keep the first definition (from base project)
		}
	}

	return baseProject, nil
}

// mergeServices merges two service configurations
func mergeServices(base, override *types.ServiceConfig) (*types.ServiceConfig, error) {
	// Create a copy of the base service
	merged := *base

	// Override simple fields if they're set in the override
	if override.Image != "" {
		merged.Image = override.Image
	}
	if override.Build != nil {
		merged.Build = override.Build
	}
	if override.Command != nil {
		merged.Command = override.Command
	}

	// Merge environment variables
	for k, v := range override.Environment {
		if merged.Environment == nil {
			merged.Environment = make(types.MappingWithEquals)
		}
		merged.Environment[k] = v
	}

	// Merge ports (keep base ports if there's a conflict)
	portMap := make(map[string]bool)
	for _, port := range merged.Ports {
		key := fmt.Sprintf("%s:%d", port.Published, port.Target)
		portMap[key] = true
	}

	for _, port := range override.Ports {
		key := fmt.Sprintf("%s:%d", port.Published, port.Target)
		if !portMap[key] {
			merged.Ports = append(merged.Ports, port)
		}
	}

	// Merge volumes
	merged.Volumes = append(merged.Volumes, override.Volumes...)

	// Merge depends_on
	if override.DependsOn != nil {
		if merged.DependsOn == nil {
			merged.DependsOn = make(types.DependsOnConfig)
		}
		for service, config := range override.DependsOn {
			merged.DependsOn[service] = config
		}
	}

	return &merged, nil
}
