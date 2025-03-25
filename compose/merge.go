package compose

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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
	project, err := options.LoadProject(context.Background())
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

// MergeComposeFiles merges multiple compose files
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

	// Get prefix from base directory name
	basePrefix := filepath.Base(files[0].BaseDir)
	if err := files[0].prefixResourceNames(basePrefix); err != nil {
		return nil, fmt.Errorf("failed to prefix resource names for %s: %w", files[0].Path, err)
	}

	// Merge additional files
	for i := 1; i < len(files); i++ {
		cf := files[i]

		// Adjust build contexts for the current file
		if err := cf.adjustBuildContexts(); err != nil {
			return nil, fmt.Errorf("failed to adjust build contexts for %s: %w", cf.Path, err)
		}

		// Get prefix from directory name
		prefix := filepath.Base(cf.BaseDir)
		if err := cf.prefixResourceNames(prefix); err != nil {
			return nil, fmt.Errorf("failed to prefix resource names for %s: %w", cf.Path, err)
		}

		// Merge services (they are already prefixed)
		for name, service := range cf.Project.Services {
			baseProject.Services[name] = service
		}

		// Merge volumes (they are already prefixed)
		if cf.Project.Volumes != nil {
			if baseProject.Volumes == nil {
				baseProject.Volumes = make(types.Volumes)
			}
			for name, volume := range cf.Project.Volumes {
				baseProject.Volumes[name] = volume
			}
		}

		// Merge networks
		if cf.Project.Networks != nil {
			if baseProject.Networks == nil {
				baseProject.Networks = make(types.Networks)
			}
			for name, network := range cf.Project.Networks {
				baseProject.Networks[name] = network
			}
		}

		// Merge configs (they are already prefixed)
		if cf.Project.Configs != nil {
			if baseProject.Configs == nil {
				baseProject.Configs = make(types.Configs)
			}
			for name, config := range cf.Project.Configs {
				baseProject.Configs[name] = config
			}
		}

		// Merge secrets (they are already prefixed)
		if cf.Project.Secrets != nil {
			if baseProject.Secrets == nil {
				baseProject.Secrets = make(types.Secrets)
			}
			for name, secret := range cf.Project.Secrets {
				baseProject.Secrets[name] = secret
			}
		}
	}

	// After merging all files, resolve any port conflicts
	logger := files[0].logger.WithField("component", "port_resolver")
	if err := ResolvePortConflicts(baseProject.Services, 100, logger); err != nil {
		return nil, fmt.Errorf("failed to resolve port conflicts: %w", err)
	}

	return baseProject, nil
}

// prefixResourceNames prefixes all resource names (services, volumes, configs, secrets) with the given prefix
func (cf *ComposeFile) prefixResourceNames(prefix string) error {
	// Create a map to store old name to new name mappings for dependency updates
	nameMap := make(map[string]string)

	// Prefix services
	newServices := make(types.Services)
	for name, service := range cf.Project.Services {
		newName := prefix + "_" + name
		nameMap[name] = newName
		newServices[newName] = service
		cf.logger.Debugf("Prefixed service name from %s to %s", name, newName)
	}
	cf.Project.Services = newServices

	// Prefix volumes
	if cf.Project.Volumes != nil {
		newVolumes := make(types.Volumes)
		for name, volume := range cf.Project.Volumes {
			newName := prefix + "_" + name
			nameMap[name] = newName
			newVolumes[newName] = volume
			cf.logger.Debugf("Prefixed volume name from %s to %s", name, newName)
		}
		cf.Project.Volumes = newVolumes
	}

	// Prefix configs
	if cf.Project.Configs != nil {
		newConfigs := make(types.Configs)
		for name, config := range cf.Project.Configs {
			newName := prefix + "_" + name
			nameMap[name] = newName
			newConfigs[newName] = config
			cf.logger.Debugf("Prefixed config name from %s to %s", name, newName)
		}
		cf.Project.Configs = newConfigs
	}

	// Prefix secrets
	if cf.Project.Secrets != nil {
		newSecrets := make(types.Secrets)
		for name, secret := range cf.Project.Secrets {
			newName := prefix + "_" + name
			nameMap[name] = newName
			newSecrets[newName] = secret
			cf.logger.Debugf("Prefixed secret name from %s to %s", name, newName)
		}
		cf.Project.Secrets = newSecrets
	}

	// Update service dependencies to use prefixed names
	for name, service := range cf.Project.Services {
		// Update depends_on references
		if service.DependsOn != nil {
			newDependsOn := make(types.DependsOnConfig)
			for depName, config := range service.DependsOn {
				newName := prefix + "_" + depName
				newDependsOn[newName] = config
				cf.logger.Debugf("Updated dependency from %s to %s", depName, newName)
			}
			service.DependsOn = newDependsOn
			cf.Project.Services[name] = service
		}

		// Update links references
		if service.Links != nil {
			newLinks := make([]string, len(service.Links))
			for i, link := range service.Links {
				parts := strings.Split(link, ":")
				if len(parts) == 2 {
					newName := prefix + "_" + parts[0]
					newLinks[i] = newName + ":" + parts[1]
					cf.logger.Debugf("Updated link from %s to %s", link, newLinks[i])
				} else {
					newName := prefix + "_" + link
					newLinks[i] = newName
					cf.logger.Debugf("Updated link from %s to %s", link, newName)
				}
			}
			service.Links = newLinks
			cf.Project.Services[name] = service
		}
	}

	return nil
}
