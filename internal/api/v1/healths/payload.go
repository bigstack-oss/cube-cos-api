package healths

import (
	"fmt"
	"time"

	json "github.com/json-iterator/go"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/gin-gonic/gin"
)

// M1 TODO: this will be removed once the real data is available in the COS side
func genFakeHealths() cubecos.Health {
	return cubecos.Health{
		Overall: &cubecos.Overall{
			Status: status.Details{
				Current:     "ng",
				Description: "ceph has 2 ceph_osd down",
			},
		},
		InUse: []definition.Service{
			{
				Name:     "clusterLink",
				Category: "core",
				Status:   status.Details{Current: "ok"},
				Modules: []definition.Module{
					{
						Name:   "link",
						Status: status.Details{Current: "ok"},
					},
					{
						Name:   "clock",
						Status: status.Details{Current: "ok"},
					},
					{
						Name:   "dns",
						Status: status.Details{Current: "ok"},
					},
				},
			},
		},
		Error: []definition.Service{
			{
				Name:     "storage",
				Category: "storage",
				Status: status.Details{
					Current:     "ng",
					Description: "ceph has 2 ceph_osd down",
				},
				Modules: []definition.Module{
					{
						Name:   "ceph",
						Status: status.Details{Current: "ok"},
					},
					{
						Name: "ceph_osd",
						Status: status.Details{
							Current:     "ng",
							Description: "2 osd down",
						},
					},
					{
						Name:   "ceph_mon",
						Status: status.Details{Current: "ok"},
					},
					{
						Name:   "ceph_mgr",
						Status: status.Details{Current: "ok"},
					},
					{
						Name:   "ceph_mds",
						Status: status.Details{Current: "ok"},
					},
				},
			},
		},
		Fixing: []definition.Service{},
	}
}

// M1 TODO: this will be removed once the real data is available in the COS side
func (h *helper) genFakeHealthCheckResult() cubecos.HealthCheckResult {
	now := time.Now().UTC()
	interval := 5 * time.Minute
	history := []cubecos.HealthCheckPoint{}
	count := 0

	for t := h.StartAsTime(); !t.After(h.StopAsTime()); t = t.Add(interval) {
		timestamp := now.Add(-time.Duration(count) * interval).Format(time.RFC3339)
		status := "ok"
		checkResult := cubecos.HealthCheckPoint{Time: timestamp, Status: status}
		if count%5 == 0 {
			h.setFakeError(&checkResult)
		}

		history = append(history, checkResult)
		count++
	}

	return cubecos.HealthCheckResult{
		Category: "cloud computing",
		Service:  "compute",
		Module:   "nova",
		History:  history,
	}
}

func (h *helper) setFakeError(checkResult *cubecos.HealthCheckPoint) {
	checkResult.Status = "ng"
	checkResult.Error = &cubecos.Error{
		Type:        "service down",
		Nodes:       []string{definition.DataCenterName},
		Description: "nova has 1 nodes down",
		Details:     "{ ... the best efforts of error summary / direction ...} ",
		Log: fmt.Sprintf(
			"http://{dataCenter}:8888/log/nova/%s-20250205113459-b3gc.log",
			definition.DataCenterName,
		),
	}
}

func genRepairReq(c *gin.Context) *cubecos.Health {
	h := &cubecos.Health{}
	err := json.NewDecoder(c.Request.Body).Decode(&h)
	if err != nil {
		h.DataCenter = &definition.DataCenter{}
		h.Overall = &cubecos.Overall{}
	}

	h.DataCenter.SetDetailsByInitedInfo()
	h.Overall.Status.SetCurrentToRepairing()
	h.Overall.Status.SetDesiredToOk()
	return h
}

func parseHealthBody(c *gin.Context) (*cubecos.Health, error) {
	h := &cubecos.Health{}
	err := json.NewDecoder(c.Request.Body).Decode(&h)
	if err != nil {
		return nil, err
	}

	return h, nil
}
