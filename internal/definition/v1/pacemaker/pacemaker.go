package pacemaker

import (
	"encoding/xml"
	"fmt"
	"os/exec"

	log "go-micro.dev/v5/logger"
)

const (
	AlertId        = "alert-to-notify-api"
	AlertScriptDir = "/var/lib/pacemaker/alerts/scripts"
	AlertLogDir    = "/tmp/pcs"
	Notifier       = "notifier.sh"
)

var (
	AlertDirs     = []string{AlertScriptDir, AlertLogDir}
	AlertNotifier = fmt.Sprintf("%s/%s", AlertScriptDir, "notifier.sh")

	AlertRecord = fmt.Sprintf(
		"%s/%s",
		AlertLogDir,
		"changed.txt",
	)
)

type Status struct {
	XMLName   xml.Name `xml:"pacemaker-result"`
	Summary   Summary  `xml:"summary"`
	Nodes     []Node   `xml:"nodes>node"`
	Resources struct {
		Resources []Resource `xml:"resource"`
		Clones    []Clone    `xml:"clone"`
	} `xml:"resources"`
	NodeAttributes []NodeAttribute `xml:"node_attributes>node"`
	NodeHistory    []NodeHistory   `xml:"node_history>node"`
}

type Summary struct {
	Stack           Stack           `xml:"stack"`
	CurrentDC       CurrentDC       `xml:"current_dc"`
	LastUpdate      TimestampOrigin `xml:"last_update"`
	LastChange      TimestampOrigin `xml:"last_change"`
	NodesConfigured CountField      `xml:"nodes_configured"`
	ResourcesConfig ResourcesConfig `xml:"resources_configured"`
	ClusterOptions  ClusterOptions  `xml:"cluster_options"`
}

type Stack struct {
	Type            string `xml:"type,attr"`
	PacemakerdState string `xml:"pacemakerd-state,attr"`
}

type CurrentDC struct {
	Name         string `xml:"name,attr"`
	ID           string `xml:"id,attr"`
	Version      string `xml:"version,attr"`
	Present      string `xml:"present,attr"`
	WithQuorum   string `xml:"with_quorum,attr"`
	MixedVersion string `xml:"mixed_version,attr"`
}

type TimestampOrigin struct {
	Time   string `xml:"time,attr"`
	Origin string `xml:"origin,attr"`
	User   string `xml:"user,attr,omitempty"`
	Client string `xml:"client,attr,omitempty"`
}

type CountField struct {
	Number string `xml:"number,attr"`
}

type ResourcesConfig struct {
	Number   string `xml:"number,attr"`
	Disabled string `xml:"disabled,attr"`
	Blocked  string `xml:"blocked,attr"`
}

type ClusterOptions struct {
	StonithEnabled         string `xml:"stonith-enabled,attr"`
	SymmetricCluster       string `xml:"symmetric-cluster,attr"`
	NoQuorumPolicy         string `xml:"no-quorum-policy,attr"`
	MaintenanceMode        string `xml:"maintenance-mode,attr"`
	StopAllResources       string `xml:"stop-all-resources,attr"`
	StonithTimeoutMs       string `xml:"stonith-timeout-ms,attr"`
	PriorityFencingDelayMs string `xml:"priority-fencing-delay-ms,attr"`
}

type Node struct {
	Name             string `xml:"name,attr"`
	ID               string `xml:"id,attr"`
	Online           string `xml:"online,attr"`
	Standby          string `xml:"standby,attr"`
	Health           string `xml:"health,attr"`
	ResourcesRunning string `xml:"resources_running,attr"`
}

type Resource struct {
	ID             string       `xml:"id,attr"`
	ResourceAgent  string       `xml:"resource_agent,attr"`
	Role           string       `xml:"role,attr"`
	NodesRunningOn string       `xml:"nodes_running_on,attr"`
	NodeList       []NodeOnHost `xml:"node"`
}

type Clone struct {
	ID        string     `xml:"id,attr"`
	Resources []Resource `xml:"resource"`
}

type NodeOnHost struct {
	Name   string `xml:"name,attr"`
	ID     string `xml:"id,attr"`
	Cached string `xml:"cached,attr"`
}

type NodeAttribute struct {
	Name      string     `xml:"name,attr"`
	Attribute []AttrPair `xml:"attribute"`
}

type AttrPair struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type NodeHistory struct {
	Name            string            `xml:"name,attr"`
	ResourceHistory []ResourceHistory `xml:"resource_history"`
}

type ResourceHistory struct {
	ID                 string             `xml:"id,attr"`
	OperationHistories []OperationHistory `xml:"operation_history"`
}

type OperationHistory struct {
	Call         string `xml:"call,attr"`
	Task         string `xml:"task,attr"`
	RC           string `xml:"rc,attr"`
	RCText       string `xml:"rc_text,attr"`
	Interval     string `xml:"interval,attr,omitempty"`
	LastRCChange string `xml:"last-rc-change,attr"`
	ExecTime     string `xml:"exec-time,attr"`
	QueueTime    string `xml:"queue-time,attr"`
}

func (s *Status) GetVipHost() (string, error) {
	for _, resource := range s.Resources.Resources {
		if resource.ID != "vip" {
			continue
		}

		if len(resource.NodeList) == 0 {
			continue
		}

		return resource.NodeList[0].Name, nil
	}

	return "", fmt.Errorf(
		"node: virtual IP resource not found in pacemaker status",
	)
}

func IsVirtualIpOwner(hostname string) bool {
	vipHost, err := GetVirtualIpHost()
	if err != nil {
		log.Errorf("node: failed to get virtual IP host: %v", err)
		return false
	}

	return vipHost == hostname
}

func GetVirtualIpHost() (string, error) {
	status, err := GetStatus()
	if err != nil {
		return "", fmt.Errorf("node: failed to get pacemaker status: %v", err)
	}

	return status.GetVipHost()
}

func GetStatus() (*Status, error) {
	out, err := exec.Command("crm_mon", "--one-shot", "--inactive", "--output-as", "xml").CombinedOutput()
	if err == nil {
		return convertToStatus(out)
	}

	return nil, fmt.Errorf(
		"node: failed to get pacemaker status: %v, output: %s",
		err, string(out),
	)
}

func convertToStatus(data []byte) (*Status, error) {
	status := Status{}
	err := xml.Unmarshal(data, &status)
	if err != nil {
		log.Errorf("node: failed to unmarshal pcs status xml: %v", err)
		return nil, err
	}

	return &status, nil
}
