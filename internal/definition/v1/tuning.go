package v1

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/blevesearch/bleve/v2"
	json "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	Tunings = "tunings"
)

var (
	tuningSpecs  = sync.Map{}
	localTunings = sync.Map{}

	tuningSearcher bleve.Index

	CreateRecordIfNotExist = options.Update().SetUpsert(true)
)

func init() {

}

type TuningPolicy struct {
	Name    string   `json:"name" yaml:"name"`
	Version string   `json:"version" yaml:"version"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tunings []Tuning `json:"tunings" yaml:"tunings"`
}

type TuningSpec struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Limitation  TuningLimitation `json:"limitation"`
	Roles       []*Role          `json:"roles"`
	Selector    `json:"-"`
}

type TuningLimitation struct {
	Type    string `json:"type"`
	Default any    `json:"default"`
	Min     int    `json:"min,omitempty"`
	Max     int    `json:"max,omitempty"`
}

type Tuning struct {
	Name        string           `json:"name" yaml:"name"`
	Value       any              `json:"value" yaml:"value"`
	Description string           `json:"description" yaml:"description"`
	Enabled     bool             `json:"enabled" yaml:"enabled"`
	IsModified  bool             `json:"isModified" yaml:"isModified"`
	Limitation  TuningLimitation `json:"limitation" yaml:"limitation"`

	*Node `json:"node,omitempty" yaml:"node,omitempty" bson:"node,omitempty"`
	Hosts []Host `json:"hosts" yaml:"hosts"`

	Status *status.Details `json:"status,omitempty" yaml:"status"`
}

type ListTuningOptions struct {
	AllNodes bool
}

func (t *Tuning) SetUpdating() {
	t.Status.Current = status.Updating
}

func (t *Tuning) SetUpdated() {
	t.Status.Current = status.Updated
}

func (t *Tuning) CopyAndOverrideHost(node Node) Tuning {
	return Tuning{
		Name:        t.Name,
		Value:       t.Value,
		Description: t.Description,
		Enabled:     t.Enabled,
		IsModified:  t.IsModified,
		Limitation:  t.Limitation,
		Hosts:       []Host{{Name: node.Hostname, Ip: node.Ip}},
	}
}

func (t *Tuning) SearchKey() string {
	return t.Name + "|" + fmt.Sprintf("%v", t.Value) + "|" + strconv.FormatBool(t.Enabled) + "|" + strconv.FormatBool(t.IsModified)
}

func CheckTuningSpec(tuning *Tuning) error {
	spec, loaded := tuningSpecs.Load(tuning.Name)
	if !loaded {
		return cuberr.TuningNotFound
	}

	if !isTuningValueValid(spec.(*TuningSpec)) {
		return cuberr.TuningValueInvalid
	}

	return nil
}

func isTuningValueValid(spec *TuningSpec) bool {
	switch spec.Limitation.Type {
	case "int":
		value, ok := spec.Limitation.Default.(int)
		if !ok {
			return false
		}

		if value <= spec.Limitation.Max && value >= spec.Limitation.Min {
			return true
		}
	case "string":
		_, ok := spec.Limitation.Default.(string)
		if !ok {
			return false
		}
	case "bool":
		_, ok := spec.Limitation.Default.(bool)
		if !ok {
			return false
		}
	}

	return false
}

func SetTuningSpec(name string, spec *TuningSpec) {
	tuningSpecs.Store(name, spec)
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
	tuningSpecs.Range(func(key, value any) bool {
		spec := value.(*TuningSpec)
		specs = append(specs, *spec)
		return true
	})

	return specs
}

func GetLocalTunings() *sync.Map {
	return &localTunings
}

func GetLocalTuning(name string) Tuning {
	val, loaded := localTunings.Load(name)
	if !loaded {
		return Tuning{}
	}

	return val.(Tuning)
}

func SetLocalTuning(tuning Tuning) {
	localTunings.Store(tuning.Name, tuning)
}

func ListLocalTunings() []Tuning {
	tunings := []Tuning{}
	localTunings.Range(func(key, value any) bool {
		tunings = append(tunings, value.(Tuning))
		return true
	})

	return tunings
}

func ShouldIHandleTheTuning(name string) bool {
	spec, loaded := tuningSpecs.Load(name)
	if !loaded {
		return false
	}

	for _, r := range spec.([]*Role) {
		if r.Name == CurrentRole {
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
	t.Node = &Node{
		Role:     role,
		Id:       HostID,
		Hostname: Hostname,
		Address:  address,
	}
}

func (t *TuningPolicy) AppendTunings(tunings []Tuning) {
	t.Tunings = append(t.Tunings, tunings...)
}

func (t *TuningPolicy) DeleteTuning(tuningName string) {
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

func InitTuningSearchIndex() error {
	if tuningSearcher != nil {
		return nil
	}

	var err error
	mapping := bleve.NewIndexMapping()
	tuningSearcher, err = bleve.NewMemOnly(mapping)
	return err
}

func GetTuningSearcher() bleve.Index {
	return tuningSearcher
}
