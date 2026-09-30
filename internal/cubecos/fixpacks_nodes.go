package cubecos

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	log "go-micro.dev/v5/logger"
)

// NodeFixpackStatus is one node's fixpack state, read from that node's own history.
type NodeFixpackStatus struct {
	Name      string   `json:"name"`
	Installed []string `json:"installed"`
	Status    string   `json:"status"`
}

// ListFixpackNodeStatus reports each node against version, or against every
// fixpack installed on any node when version is empty.
func ListFixpackNodeStatus(version string) ([]NodeFixpackStatus, error) {
	args := []string{"fixpack_status"}
	if version != "" {
		args = append(args, version)
	}

	out, err := runFixpackSdk(args...)
	if err != nil {
		return nil, err
	}

	return parseFixpackNodeStatus(out), nil
}

func parseFixpackNodeStatus(out string) []NodeFixpackStatus {
	list := []NodeFixpackStatus{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(fields) != 3 {
			continue
		}

		installed := []string{}
		if fields[1] != "-" {
			installed = strings.Fields(fields[1])
		}

		list = append(list, NodeFixpackStatus{Name: fields[0], Installed: installed, Status: fields[2]})
	}

	return list
}

// ListFixpackRollbackNodes lists the nodes whose latest rollback point is version.
func ListFixpackRollbackNodes(version string) ([]string, error) {
	out, err := runFixpackSdk("fixpack_rollback_nodes", version)
	if err != nil {
		return nil, err
	}

	return strings.Fields(out), nil
}

// GetFixpackRollbackTop returns this node's latest rollback point.
func GetFixpackRollbackTop() (string, error) {
	out, err := runFixpackSdk("fixpack_rollback_top")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// fixpack_status exits 1 when a node is missing a fixpack; its output is still valid.
func runFixpackSdk(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(120))
	defer cancel()

	out, err := exec.CommandContext(ctx, "hex_sdk", args...).Output()
	if err != nil && (len(out) == 0 || args[0] != "fixpack_status") {
		err := fmt.Errorf("failed to run hex_sdk %s(%v)", strings.Join(args, " "), err)
		log.Errorf("fixpacks: %v", err)
		return "", err
	}

	return string(out), nil
}
