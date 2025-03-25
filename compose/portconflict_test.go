package compose

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectPortConflicts(t *testing.T) {
	// Create a logger for testing
	logger := logrus.New().WithField("test", true)

	// Test case 1: No conflicts
	t.Run("no conflicts", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
			"app2": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "8080", Target: 8080},
				},
			},
		}

		conflicts := DetectPortConflicts(services, logger)
		assert.Empty(t, conflicts, "Expected no conflicts")
	})

	// Test case 2: Simple conflict
	t.Run("simple conflict", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
			"app2": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
		}

		conflicts := DetectPortConflicts(services, logger)
		assert.Len(t, conflicts, 1, "Expected one conflict")
		assert.Contains(t, conflicts, uint32(80), "Expected conflict on port 80")
		assert.Len(t, conflicts[80], 2, "Expected two services in conflict")
		assert.Contains(t, conflicts[80], "app1")
		assert.Contains(t, conflicts[80], "app2")
	})

	// Test case 3: Multiple conflicts
	t.Run("multiple conflicts", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
					{Published: "443", Target: 443},
				},
			},
			"app2": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
					{Published: "443", Target: 443},
				},
			},
		}

		conflicts := DetectPortConflicts(services, logger)
		assert.Len(t, conflicts, 2, "Expected two conflicts")
		assert.Contains(t, conflicts, uint32(80), "Expected conflict on port 80")
		assert.Contains(t, conflicts, uint32(443), "Expected conflict on port 443")
	})

	// Test case 4: Invalid port format
	t.Run("invalid port format", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "invalid", Target: 80},
				},
			},
		}

		conflicts := DetectPortConflicts(services, logger)
		assert.Empty(t, conflicts, "Expected no conflicts for invalid port format")
	})
}

func TestResolvePortConflicts(t *testing.T) {
	// Create a logger for testing
	logger := logrus.New().WithField("test", true)

	// Test case 1: No conflicts
	t.Run("no conflicts", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
			"app2": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "8080", Target: 8080},
				},
			},
		}

		err := ResolvePortConflicts(services, 100, logger)
		require.NoError(t, err)

		// Verify ports remain unchanged
		assert.Equal(t, "80", services["app1"].Ports[0].Published)
		assert.Equal(t, "8080", services["app2"].Ports[0].Published)
	})

	// Test case 2: Simple conflict resolution
	t.Run("simple conflict resolution", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
			"app2": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
		}

		err := ResolvePortConflicts(services, 100, logger)
		require.NoError(t, err)

		// Verify ports are adjusted
		assert.Equal(t, "80", services["app1"].Ports[0].Published)  // First service keeps original port
		assert.Equal(t, "180", services["app2"].Ports[0].Published) // Second service gets offset
	})

	// Test case 3: Multiple conflicts resolution
	t.Run("multiple conflicts resolution", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
					{Published: "443", Target: 443},
				},
			},
			"app2": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
					{Published: "443", Target: 443},
				},
			},
		}

		err := ResolvePortConflicts(services, 100, logger)
		require.NoError(t, err)

		// Verify ports are adjusted
		assert.Equal(t, "80", services["app1"].Ports[0].Published) // First service keeps original ports
		assert.Equal(t, "443", services["app1"].Ports[1].Published)
		assert.Equal(t, "180", services["app2"].Ports[0].Published) // Second service gets offset
		assert.Equal(t, "543", services["app2"].Ports[1].Published)
	})

	// Test case 4: Unresolvable conflict
	t.Run("unresolvable conflict", func(t *testing.T) {
		services := types.Services{
			"app1": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
			"app2": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "80", Target: 80},
				},
			},
			"app3": types.ServiceConfig{
				Ports: []types.ServicePortConfig{
					{Published: "180", Target: 80}, // This port would conflict with app2's new port
				},
			},
		}

		err := ResolvePortConflicts(services, 100, logger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to resolve all port conflicts")
	})
}
