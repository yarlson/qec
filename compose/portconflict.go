package compose

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
)

// PortConflict represents a port mapping conflict between services
type PortConflict struct {
	HostPort uint32
	Services []string
}

// DetectPortConflicts scans through service configurations and identifies host port collisions
func DetectPortConflicts(services types.Services, logger *logrus.Entry) map[uint32][]string {
	// Map to store port -> service name mappings
	portMap := make(map[uint32][]string)

	// Iterate through all services
	for name, service := range services {
		// Skip services without port mappings
		if len(service.Ports) == 0 {
			continue
		}

		// Check each port mapping
		for _, port := range service.Ports {
			// Skip if no host port is specified (using container port)
			if port.Published == "" {
				continue
			}

			// Parse the published port string to uint32
			hostPort, err := strconv.ParseUint(port.Published, 10, 32)
			if err != nil {
				logger.Warnf("Invalid port format for service %s: %s", name, port.Published)
				continue
			}

			// Add service to the port mapping
			portMap[uint32(hostPort)] = append(portMap[uint32(hostPort)], name)
			logger.Debugf("Service %s maps to host port %d", name, hostPort)
		}
	}

	// Filter out ports with only one service (no conflicts)
	conflicts := make(map[uint32][]string)
	for port, serviceNames := range portMap {
		if len(serviceNames) > 1 {
			// Sort service names to ensure consistent order
			sort.Strings(serviceNames)
			conflicts[port] = serviceNames
			logger.Warnf("Port conflict detected on port %d between services: %v", port, serviceNames)
		}
	}

	return conflicts
}

// ResolvePortConflicts attempts to resolve port conflicts by applying an offset
func ResolvePortConflicts(services types.Services, offset uint32, logger *logrus.Entry) error {
	// First detect all conflicts
	conflicts := DetectPortConflicts(services, logger)

	// If no conflicts, we're done
	if len(conflicts) == 0 {
		return nil
	}

	// Create a map to track used ports after resolution
	usedPorts := make(map[uint32]bool)

	// Iterate through all services and adjust conflicting ports
	for name, service := range services {
		if len(service.Ports) == 0 {
			continue
		}

		// Check each port mapping
		for i := range service.Ports {
			port := &service.Ports[i]
			if port.Published == "" {
				continue
			}

			// Parse the published port string to uint32
			hostPort, err := strconv.ParseUint(port.Published, 10, 32)
			if err != nil {
				logger.Warnf("Invalid port format for service %s: %s", name, port.Published)
				continue
			}

			// Check if this port is in conflict
			if conflictingServices, hasConflict := conflicts[uint32(hostPort)]; hasConflict {
				// Find the service's position in the conflict list
				serviceIndex := -1
				for j, s := range conflictingServices {
					if s == name {
						serviceIndex = j
						break
					}
				}

				if serviceIndex == -1 {
					continue
				}

				// Calculate new port by adding offset * service index
				newPort := uint32(hostPort) + (offset * uint32(serviceIndex))

				// Check if the new port is already in use
				if usedPorts[newPort] {
					return fmt.Errorf("unable to resolve port conflict: port %d is already in use after applying offset", newPort)
				}

				// Update the port and mark it as used
				logger.Infof("Adjusting port for service %s from %d to %d", name, hostPort, newPort)
				port.Published = strconv.FormatUint(uint64(newPort), 10)
				usedPorts[newPort] = true
			}
		}
		services[name] = service
	}

	// Check for any remaining conflicts after resolution
	remainingConflicts := DetectPortConflicts(services, logger)
	if len(remainingConflicts) > 0 {
		return fmt.Errorf("unable to resolve all port conflicts: %v", remainingConflicts)
	}

	return nil
}
