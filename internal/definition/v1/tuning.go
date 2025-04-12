package v1

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	cuberr "github.com/bigstack-oss/cube-cos-api/internal/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/status"
	"github.com/blevesearch/bleve/v2"
	json "github.com/json-iterator/go"
	"github.com/shirou/gopsutil/v4/host"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	Tunings         = "tunings"
	TuningRecordTTL = 3600
)

var (
	tuningSpecs  = sync.Map{}
	localTunings = sync.Map{}

	tuningSearcher bleve.Index

	CreateRecordIfNotExist = options.Update().SetUpsert(true)
)

type TuningPolicy struct {
	Name    string   `json:"name" yaml:"name"`
	Version string   `json:"version" yaml:"version"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tunings []Tuning `json:"tunings" yaml:"tunings"`
}

type RawTuningSpec struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Limitation  RawTuningLimitation `json:"limitation"`
}

type RawTuningLimitation struct {
	Type    string `json:"type"`
	Default string `json:"default"`
	Min     string `json:"min,omitempty"`
	Max     string `json:"max,omitempty"`
	Regex   string `json:"regex,omitempty"`
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
	Min     *int   `json:"min,omitempty"`
	Max     *int   `json:"max,omitempty"`
	Regex   string `json:"regex,omitempty"`
}

type Tuning struct {
	Id          string           `json:"id,omitempty" yaml:"-" bson:"id"`
	Name        string           `json:"name" yaml:"name" bson:"name"`
	Value       any              `json:"value" yaml:"value" bson:"value"`
	Description string           `json:"description" yaml:"-" bson:"-"`
	Enabled     bool             `json:"enabled" yaml:"enabled" bson:"enabled"`
	IsModified  bool             `json:"isModified" yaml:"-" bson:"-"`
	Limitation  TuningLimitation `json:"limitation" yaml:"-" bson:"-"`

	*Node  `json:"node,omitempty" yaml:"-" bson:"-"`
	Hosts  []Host         `json:"hosts" yaml:"-" bson:"-"`
	Roles  []Role         `json:"roles,omitempty" yaml:"-" bson:"-"`
	Status *status.Tuning `json:"status,omitempty" yaml:"-" bson:"status,omitempty"`
}

type TuningReset struct {
	Hosts []string `json:"hosts"`
}

type TuningUpdate struct {
	Value any      `json:"value"`
	Hosts []string `json:"hosts"`
}

type TuningToggle struct {
	Enable bool     `json:"enable"`
	Hosts  []string `json:"hosts"`
}

type ListTuningOptions struct {
	AllNodes bool
}

func (t *TuningSpec) IsInLimitedRange(value int) bool {
	return value <= *t.Limitation.Max && value >= *t.Limitation.Min
}

func (t *Tuning) GenerateId() string {
	return fmt.Sprintf("%s-%s", t.Name, t.JoinHosts())
}

func (t *Tuning) JoinHosts() string {
	hosts := []string{}
	for _, host := range t.Hosts {
		hosts = append(hosts, host.Name)
	}

	slices.Sort(hosts)
	return strings.Join(hosts, "-")
}

func (t *Tuning) IncludeHost(hostname string) bool {
	for _, host := range t.Hosts {
		if host.Name == hostname {
			return true
		}
	}

	return false
}

func (t *Tuning) IncludeHosts(hosts []string) bool {
	for _, host := range hosts {
		if !t.IncludeHost(host) {
			return false
		}
	}

	return true
}

func (t *Tuning) InitHosts(hosts []string) {
	for _, host := range hosts {
		t.Hosts = append(t.Hosts, Host{Name: host})
	}
}

func (t *Tuning) InitStatus(current, desired string) {
	t.Status = &status.Tuning{
		Current:   current,
		Desired:   desired,
		CreatedAt: TimeLocalRFC3339(time.Now()),
	}
}

func (t *Tuning) InitUpdateStatus() {
	t.Status = &status.Tuning{
		Current:    status.Updating,
		Desired:    status.Updated,
		CreatedAt:  TimeLocal(),
		UpdatedAt:  TimeLocal(),
		IsUpdating: true,
	}
}

func (t *Tuning) InitResetStatus() {
	t.Status = &status.Tuning{
		Current:    status.Updating,
		Desired:    status.Reset,
		CreatedAt:  TimeLocal(),
		UpdatedAt:  TimeLocal(),
		IsUpdating: true,
	}
}

func (t *Tuning) InitOkStatus() {
	t.Status = &status.Tuning{
		Current:    status.Ok,
		IsUpdating: false,
	}

	bootDuration, err := host.BootTime()
	if err != nil {
		t.Status.UpdatedAt = TimeISO8601Z(time.Now())
		return
	}

	bootTime := time.Unix(int64(bootDuration), 0)
	t.Status.UpdatedAt = TimeISO8601Z(bootTime)
}

func (t *Tuning) StrValue() string {
	return fmt.Sprintf("%v", t.Value)
}

func (t *Tuning) SetDesired(status string) {
	t.Status.Desired = status
}

func (t *Tuning) SetUpdating() {
	t.Status.Current = status.Updating
}

func (t *Tuning) SetUpdated() {
	t.Status.Current = status.Updated
	t.Status.IsUpdating = false
}

func (t *Tuning) SetError() {
	t.Status.Current = status.Error
	t.Status.IsUpdating = false
}

func (t *Tuning) SetCompleted() {
	t.Status.Current = status.Ok
	t.Status.IsUpdating = false
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

func (t *Tuning) GenTaskUpdate() Tuning {
	return Tuning{
		Id:     t.Id,
		Name:   t.Name,
		Value:  t.Value,
		Status: t.Status,
	}
}

func (t *Tuning) SearchKey() string {
	return t.Name + "|" + fmt.Sprintf("%v", t.Value) + "|" + strconv.FormatBool(t.Enabled) + "|" + strconv.FormatBool(t.IsModified)
}

func CheckTuningSpec(tuning Tuning) error {
	spec, loaded := tuningSpecs.Load(tuning.Name)
	if !loaded {
		return cuberr.TuningNotFound
	}

	if !isTuningValueValid(tuning, spec.(*TuningSpec)) {
		return cuberr.TuningValueInvalid
	}

	return nil
}

func isTuningValueValid(tuning Tuning, spec *TuningSpec) bool {
	switch spec.Limitation.Type {
	case "int":
		value, ok := tuning.Value.(int)
		if !ok {
			return false
		}

		return spec.IsInLimitedRange(value)
	case "string":
		_, ok := tuning.Value.(string)
		return ok
	case "bool":
		_, ok := tuning.Value.(bool)
		return ok
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

func GetTuningSpec(name string) (*TuningSpec, error) {
	val, loaded := tuningSpecs.Load(name)
	if !loaded {
		return nil, cuberr.TuningNotFound
	}

	return val.(*TuningSpec), nil
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

	return setRoleAndIpToTunings(tunings)
}

func setRoleAndIpToTunings(tunings []Tuning) []Tuning {
	nodeMap, err := HostnameNodeMap()
	if err != nil {
		return tunings
	}

	for i, tuning := range tunings {
		for j, host := range tuning.Hosts {
			node, found := nodeMap[host.Name]
			if !found {
				continue
			}

			tunings[i].Hosts[j].Ip = node.Ip
			tunings[i].Hosts[j].Role = node.Role
		}
	}

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

func (t *TuningPolicy) UpdateOrAppendTuning(tuning Tuning) {
	if !t.existingTuningUpdated(tuning) {
		t.AppendTuning(tuning)
	}
}

func (t *TuningPolicy) existingTuningUpdated(tuning Tuning) bool {
	for i, existing := range t.Tunings {
		if existing.Name == tuning.Name {
			t.Tunings[i].Value = tuning.Value
			t.Tunings[i].Enabled = tuning.Enabled
			return true
		}
	}

	return false
}

func (t *TuningPolicy) AppendTuning(tuning Tuning) {
	t.Tunings = append(t.Tunings, tuning)
}

func (t *TuningPolicy) AppendTunings(tunings []Tuning) {
	t.Tunings = slices.Concat(t.Tunings, tunings)
}

func (t *TuningPolicy) HasMatchedTuning(tuning Tuning) bool {
	for _, existing := range t.Tunings {
		if existing.Name != tuning.Name {
			continue
		}

		if existing.StrValue() != tuning.StrValue() {
			continue
		}

		if existing.Enabled != tuning.Enabled {
			continue
		}

		return true
	}

	return false
}

func (t *TuningPolicy) DeleteTuning(name string) {
	newTunings := []Tuning{}
	for _, tuning := range t.Tunings {
		if tuning.Name != name {
			newTunings = append(newTunings, tuning)
		}
	}

	t.Tunings = newTunings
}

func TuningDB() string {
	return Tunings
}

func TuningReqCollection() string {
	return "requests"
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
