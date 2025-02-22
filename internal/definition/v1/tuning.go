package v1

import (
	"strings"
	"sync"

	"github.com/bigstack-oss/cube-cos-api/internal/status"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	Tunings = "tunings"
)

var (
	tuningSpecs            = sync.Map{}
	currentTunings         = sync.Map{}
	CreateRecordIfNotExist = options.Update().SetUpsert(true)
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
	Enabled bool        `json:"enabled" yaml:"enabled"`
	Name    string      `json:"name" yaml:"name"`
	Value   interface{} `json:"value" yaml:"value"`

	Node   `json:"node,omitempty" yaml:"node,omitempty" bson:"node,omitempty"`
	Status status.Details `json:"status,omitempty" yaml:"status,omitempty" bson:"status,omitempty"`
}

type ListTuningOptions struct {
	AllNodes bool
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

func GetTuningSpecs() *sync.Map {
	return &tuningSpecs
}

func ListTuningSpecs() []TuningSpec {
	specs := []TuningSpec{}
	tuningSpecs.Range(func(key, value interface{}) bool {
		specs = append(specs, value.(TuningSpec))
		return true
	})

	return specs
}

func GetCurrentTunings() *sync.Map {
	return &currentTunings
}

func GetCurrentTuning(name string) Tuning {
	val, loaded := currentTunings.Load(name)
	if !loaded {
		return Tuning{}
	}

	return val.(Tuning)
}

func SetCurrentTuning(tuning Tuning) {
	currentTunings.Store(tuning.Name, tuning)
}

func ListCurrentTunings() []Tuning {
	tunings := []Tuning{}
	currentTunings.Range(func(key, value interface{}) bool {
		tunings = append(tunings, value.(Tuning))
		return true
	})

	return tunings
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
		Id:       HostID,
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

func TuningDB() string {
	return Tunings
}

func TuningCollection(name string) string {
	return strings.Split(name, ".")[0]
}
