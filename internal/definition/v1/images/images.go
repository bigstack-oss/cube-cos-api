package images

const (
	Module = "images"
)

var (
	Visibilitise = []string{"public", "private"}
	Destinations = []string{"CubeStorage"}
	Oses         = []string{
		"CentOS",
		"Fedora",
		"Ubuntu",
		"Debian",
		"Windows",
		"Rocky",
		"FreeBSD",
		"CoreOS",
		"Arch",
		"Others",
	}
)

type Reserved struct {
	File                        string `json:"file" bson:"file"`
	Name                        string `json:"name" bson:"name"`
	Os                          string `json:"os" bson:"os"`
	Destination                 string `json:"destination" bson:"destination"`
	Domain                      string `json:"domain" bson:"domain"`
	SourceFromAnotherHypervisor bool   `json:"sourceFromAnotherHypervisor" bson:"sourceFromAnotherHypervisor"`
	Visibility                  string `json:"visibility" bson:"visibility"`
}
