package nodes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRegisteredRoles(t *testing.T) {
	tests := []struct {
		name       string
		candidates []string
		nodes      []Node
		want       []string
	}{
		{
			name:       "converged-only cloud keeps just control-converged",
			candidates: GetCloudRoles(),
			nodes: []Node{
				{Hostname: "cc-1", Role: RoleControlConverged},
				{Hostname: "cc-2", Role: RoleControlConverged},
				{Hostname: "cc-3", Role: RoleControlConverged},
			},
			want: []string{RoleControlConverged},
		},
		{
			name:       "keeps the candidates' order, not the nodes' order",
			candidates: GetCloudRoles(),
			nodes: []Node{
				{Hostname: "storage-1", Role: RoleStorage},
				{Hostname: "compute-1", Role: RoleCompute},
				{Hostname: "cc-1", Role: RoleControlConverged},
			},
			want: []string{RoleControlConverged, RoleCompute, RoleStorage},
		},
		{
			name:       "drops a role outside the candidates",
			candidates: GetEdgeRoles(),
			nodes: []Node{
				{Hostname: "edge-1", Role: RoleEdgeCore},
				{Hostname: "cc-1", Role: RoleControlConverged},
			},
			want: []string{RoleEdgeCore},
		},
		{
			name:       "no nodes gives an empty, non-nil list",
			candidates: GetCloudRoles(),
			nodes:      nil,
			want:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, GetRegisteredRoles(tt.candidates, tt.nodes))
		})
	}
}
