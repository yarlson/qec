package compose

import (
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// PortConflictTestSuite defines the test suite for port conflict functionality
type PortConflictTestSuite struct {
	suite.Suite
	logger *logrus.Entry
}

// SetupTest runs before each test
func (suite *PortConflictTestSuite) SetupTest() {
	suite.logger = logrus.New().WithField("test", true)
}

// TestDetectPortConflicts tests the port conflict detection functionality
func (suite *PortConflictTestSuite) TestDetectPortConflicts() {
	tests := []struct {
		name     string
		services types.Services
		want     map[uint32][]string
	}{
		{
			name: "no conflicts",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "8080", Target: 80},
					},
				},
			},
			want: map[uint32][]string{},
		},
		{
			name: "simple conflict",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
			},
			want: map[uint32][]string{
				80: {"web1", "web2"},
			},
		},
		{
			name: "multiple conflicts",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
						{Published: "443", Target: 443},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
						{Published: "443", Target: 443},
					},
				},
			},
			want: map[uint32][]string{
				80:  {"web1", "web2"},
				443: {"web1", "web2"},
			},
		},
		{
			name: "invalid port format",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "invalid", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "invalid", Target: 80},
					},
				},
			},
			want: map[uint32][]string{},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := DetectPortConflicts(tt.services, suite.logger)
			assert.Equal(suite.T(), tt.want, got)
		})
	}
}

// TestResolvePortConflicts tests the port conflict resolution functionality
func (suite *PortConflictTestSuite) TestResolvePortConflicts() {
	tests := []struct {
		name     string
		services types.Services
		offset   uint32
		want     types.Services
		wantErr  bool
	}{
		{
			name: "no conflicts",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "8080", Target: 80},
					},
				},
			},
			offset: 100,
			want: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "8080", Target: 80},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "simple conflict resolution",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
			},
			offset: 100,
			want: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "180", Target: 80},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple conflicts resolution",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
						{Published: "443", Target: 443},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
						{Published: "443", Target: 443},
					},
				},
			},
			offset: 100,
			want: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
						{Published: "443", Target: 443},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "180", Target: 80},
						{Published: "543", Target: 443},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unresolvable conflict",
			services: types.Services{
				"web1": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web2": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
				"web3": {
					Ports: []types.ServicePortConfig{
						{Published: "80", Target: 80},
					},
				},
			},
			offset:  0,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := ResolvePortConflicts(tt.services, tt.offset, suite.logger)
			if tt.wantErr {
				assert.Error(suite.T(), err)
				assert.Contains(suite.T(), err.Error(), "unable to resolve port conflict: port 80 is already in use after applying offset")
				return
			}
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), tt.want, tt.services)
		})
	}
}

// Run the test suite
func TestPortConflictTestSuite(t *testing.T) {
	suite.Run(t, new(PortConflictTestSuite))
}
