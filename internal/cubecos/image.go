package cubecos

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"

func GetReservedImages() []images.Reserved {
	return []images.Reserved{
		{
			File:                        "amphora-x64-haproxy-yoga.qcow2",
			Name:                        "amphora-x64-haproxy",
			Os:                          "Ubuntu",
			Destination:                 "CubeStorage",
			Domain:                      "default",
			SourceFromAnotherHypervisor: false,
			Visibility:                  "private",
		},
		{
			File:                        "manila-service-image_yoga.qcow2",
			Name:                        "manila-service-image",
			Os:                          "Ubuntu",
			Destination:                 "CubeStorage",
			Domain:                      "default",
			SourceFromAnotherHypervisor: false,
			Visibility:                  "private",
		},
	}
}
