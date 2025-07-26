package cubecos

import "github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"

func GetReservedImages() []images.ReqOpts {
	return []images.ReqOpts{
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

func ImportImage(opts images.CreateOpts) error {
	// This function would contain the logic to import an image based on the provided options.
	// For now, we just return nil to indicate success.
	return nil
}
