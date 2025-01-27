package healths

import (
	"encoding/json"

	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/gin-gonic/gin"
)

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

func genRepairReq(c *gin.Context) *cubecos.Health {
	h := &cubecos.Health{}
	err := json.NewDecoder(c.Request.Body).Decode(&h)
	if err != nil {
		h = &cubecos.Health{}
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
