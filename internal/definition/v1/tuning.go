package v1

import (
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
	json "github.com/json-iterator/go"
)

const (
	Tunings = "tunings"
)

var (
	tuningSpecs = sync.Map{}
)

type Policy struct {
	Name    string   `json:"name" yaml:"name"`
	Version string   `json:"version" yaml:"version"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tunings []Tuning `json:"tunings" yaml:"tunings"`
}

type TuningSpec struct {
	Name         string `json:"name"`
	ExampleValue `json:"exampleValue"`
	Description  string  `json:"description"`
	Roles        []*Role `json:"roles"`
	Selector     `json:"selector"`
}

type ExampleValue struct {
	Type    string      `json:"type"`
	Default interface{} `json:"default"`
	Min     interface{} `json:"min"`
	Max     interface{} `json:"max"`
}

type Tuning struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Name    string `json:"name" yaml:"name"`
	Value   string `json:"value" yaml:"value"`

	Node   `json:"node,omitempty" yaml:"node,omitempty" bson:"node,omitempty"`
	Status status.Details `json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"`
}

func SetSpecToTuning(tuningName string, tuningSpec *TuningSpec) {
	tuningSpecs.Store(tuningName, tuningSpec)
}

func GetRolesToHandleTuning(tuningName string) ([]*Role, bool) {
	val, loaded := tuningSpecs.Load(tuningName)
	if !loaded {
		return nil, false
	}

	return val.(*TuningSpec).Roles, true
}

func GetAllTunings() *sync.Map {
	return &tuningSpecs
}

func ShouldCurrentRoleHandleTheTuning(tuningName string, roleName string) bool {
	val, loaded := tuningSpecs.Load(tuningName)
	if !loaded {
		return false
	}

	for _, r := range val.([]*Role) {
		if r.Name == roleName {
			return true
		}
	}

	return false
}

func (t *Tuning) Bytes() ([]byte, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (t *Tuning) SetNodeInfo(role, address string) {
	t.Node = Node{
		Role:     role,
		ID:       HostID,
		Hostname: Hostname,
		Address:  address,
	}
}

func (t *Policy) AppendTunings(tunings []Tuning) {
	t.Tunings = append(t.Tunings, tunings...)
}

func (t *Policy) DeleteTuning(tuningName string) {
	newTunings := []Tuning{}
	for _, tuning := range t.Tunings {
		if tuning.Name != tuningName {
			newTunings = append(newTunings, tuning)
		}
	}

	t.Tunings = newTunings
}
