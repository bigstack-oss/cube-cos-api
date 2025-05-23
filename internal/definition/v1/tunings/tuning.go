package tunings

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	ostime "time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/base"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/errors"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/nodes"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/search"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/time"
	log "go-micro.dev/v5/logger"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	Module    = "tunings"
	RecordTTL = 3600
)

var (
	Specs = sync.Map{}
	local = sync.Map{}

	CreateRecordIfNotExist = options.Update().SetUpsert(true)
)

type Policy struct {
	Name    string   `json:"name" yaml:"name"`
	Version string   `json:"version" yaml:"version"`
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Tunings []Tuning `json:"tunings" yaml:"tunings"`
}
type Spec struct {
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Limitation     Limitation    `json:"limitation"`
	Roles          []*nodes.Role `json:"roles"`
	nodes.Selector `json:"-"`
}

type Limitation struct {
	Type    string `json:"type"`
	Default any    `json:"default"`
	Min     *int   `json:"min,omitempty"`
	Max     *int   `json:"max,omitempty"`
	Regex   string `json:"regex,omitempty"`
}

type RawSpec struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Limitation  RawLimitation `json:"limitation"`
}

type RawLimitation struct {
	Type    string `json:"type"`
	Default string `json:"default"`
	Min     string `json:"min,omitempty"`
	Max     string `json:"max,omitempty"`
	Regex   string `json:"regex,omitempty"`
}

type Tuning struct {
	Name             string     `json:"name" yaml:"name" bson:"name"`
	Value            any        `json:"value" yaml:"value" bson:"value"`
	Description      string     `json:"description" yaml:"-" bson:"-"`
	Enabled          bool       `json:"enabled" yaml:"enabled" bson:"enabled"`
	IsModified       bool       `json:"isModified" yaml:"-" bson:"-"`
	IsReportRequired bool       `json:"-" yaml:"-" bson:"-"`
	Limitation       Limitation `json:"limitation" yaml:"-" bson:"-"`
	SortIndex        string     `json:"-" yaml:"-" bson:"-"`

	Node   *nodes.Node    `json:"node,omitempty" yaml:"-" bson:"-"`
	Hosts  []nodes.Host   `json:"hosts" yaml:"-" bson:"-"`
	Status *status.Tuning `json:"status,omitempty" yaml:"-" bson:"status,omitempty"`
}

type Reset struct {
	Hosts []string `json:"hosts"`
}

type Update struct {
	Value   any      `json:"value"`
	Enabled bool     `json:"enabled"`
	Hosts   []string `json:"hosts"`
}

type Toggle struct {
	Enable bool     `json:"enable"`
	Hosts  []string `json:"hosts"`
}

type ListOptions struct {
	AllNodes bool
}

func (s *Spec) IsInLimitedRange(value int) bool {
	return value <= *s.Limitation.Max && value >= *s.Limitation.Min
}

func (t *Tuning) GenerateId() string {
	return fmt.Sprintf("%s-%s", t.Name, t.JoinHosts())
}

func (t *Tuning) GenSearchableOject() Tuning {
	tuning := Tuning{
		Name:        search.NormalizedKeyword(t.Name),
		Value:       search.NormalizedKeyword(fmt.Sprintf("%v", t.Value)),
		Description: search.NormalizedKeyword(t.Description),
		Enabled:     t.Enabled,
		Status:      &status.Tuning{UpdatedAt: search.NormalizedKeyword(t.Status.UpdatedAt)},
	}

	for _, host := range t.Hosts {
		tuning.Hosts = append(
			tuning.Hosts,
			nodes.Host{Name: search.NormalizedKeyword(host.Name)},
		)
	}

	return tuning
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
	hostSlice := []string{}
	for _, host := range t.Hosts {
		hostSlice = append(hostSlice, host.Name)
	}

	for _, host := range hosts {
		if !slices.Contains(hostSlice, host) {
			return false
		}
	}

	return true
}

func (t *Tuning) SetHosts(hosts []string) {
	t.Hosts = []nodes.Host{}
	for _, host := range hosts {
		t.Hosts = append(t.Hosts, nodes.Host{Name: host})
		if host == base.Hostname {
			t.Node = &nodes.Node{Hostname: host}
		}
	}
}

func (t *Tuning) InitStatus(current, desired string) {
	t.Status = &status.Tuning{
		Current:   current,
		Desired:   desired,
		CreatedAt: time.LocalRFC3339(ostime.Now()),
	}
}

func (t *Tuning) SetUpdating() {
	t.Status = &status.Tuning{
		Current:    status.Updating,
		Desired:    status.Updated,
		CreatedAt:  time.NowLocal(),
		UpdatedAt:  time.NowLocal(),
		IsUpdating: true,
	}
}

func (t *Tuning) SetResetting() {
	t.Status = &status.Tuning{
		Current:    status.Updating,
		Desired:    status.Reset,
		CreatedAt:  time.NowLocal(),
		UpdatedAt:  time.NowLocal(),
		IsUpdating: true,
	}
}

func (t *Tuning) SetOk() {
	t.Status = &status.Tuning{
		Current:    status.Ok,
		IsUpdating: false,
		UpdatedAt:  time.Boot(),
	}
}

func (t *Tuning) StrValue() string {
	return fmt.Sprintf("%v", t.Value)
}

func (t *Tuning) SetDesired(status string) {
	t.Status.Desired = status
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

func (t *Tuning) CopyAndOverrideHost(n nodes.Node) Tuning {
	return Tuning{
		Name:        t.Name,
		Value:       t.Value,
		Description: t.Description,
		Enabled:     t.Enabled,
		IsModified:  t.IsModified,
		Limitation:  t.Limitation,
		Hosts:       []nodes.Host{{Name: n.Hostname, Ip: n.Ip}},
	}
}

func (t *Tuning) GenTaskUpdate() Tuning {
	return Tuning{
		Name:   t.Name,
		Value:  t.Value,
		Node:   t.Node,
		Status: t.Status,
	}
}

func (t *Tuning) IndexKey() string {
	return t.Name + "|" +
		fmt.Sprintf("%v", t.Value) + "|" +
		strconv.FormatBool(t.Enabled) + "|" +
		strconv.FormatBool(t.IsModified)
}

func CheckSpec(tuning Tuning) error {
	spec, loaded := Specs.Load(tuning.Name)
	if !loaded {
		return errors.ErrTuningNotFound
	}

	if !isValueValid(tuning, spec.(*Spec)) {
		return errors.ErrTuningValueInvalid
	}

	return nil
}

func isValueValid(tuning Tuning, spec *Spec) bool {
	switch spec.Limitation.Type {
	case "int", "uint":
		return isValidInt(tuning, spec)
	case "str":
		return isValidString(tuning, spec)
	case "bool", "boolean":
		_, ok := tuning.Value.(bool)
		return ok
	}

	return false
}

func isValidInt(tuning Tuning, spec *Spec) bool {
	value, ok := tuning.Value.(float64)
	if !ok {
		log.Errorf("tuning: %s value is not int(%v: %v)", tuning.Name, tuning.Value, reflect.TypeOf(tuning.Value))
		return false
	}

	return spec.IsInLimitedRange(int(value))
}

func isValidString(tuning Tuning, spec *Spec) bool {
	value, ok := tuning.Value.(string)
	if !ok {
		return false
	}

	if spec.Limitation.Regex == "na" {
		return true
	}

	return regexp.
		MustCompile(spec.Limitation.Regex).
		MatchString(value)
}

func SetSpec(name string, spec *Spec) {
	Specs.Store(name, spec)
}

func GetRolesToHandle(tuningName string) ([]*nodes.Role, bool) {
	val, loaded := Specs.Load(tuningName)
	if !loaded {
		return nil, false
	}

	return val.(*Spec).Roles, true
}

func GetSpec(name string) (*Spec, error) {
	val, loaded := Specs.Load(name)
	if !loaded {
		return nil, errors.ErrTuningNotFound
	}

	return val.(*Spec), nil
}

func GetSpecs() *sync.Map {
	return &Specs
}

func ListSpecs() []Spec {
	list := []Spec{}
	Specs.Range(func(key, value any) bool {
		spec := value.(*Spec)
		list = append(list, *spec)
		return true
	})

	return list
}

func GetLocal() *sync.Map {
	return &local
}

func Get(name string) Tuning {
	val, loaded := local.Load(name)
	if !loaded {
		return Tuning{}
	}

	return val.(Tuning)
}

func SetLocal(tuning Tuning) {
	local.Store(tuning.Name, tuning)
}

func ListLocal() []Tuning {
	tunings := []Tuning{}
	local.Range(func(key, value any) bool {
		tunings = append(tunings, value.(Tuning))
		return true
	})

	return setRoleAndIpToTunings(tunings)
}

func setRoleAndIpToTunings(tunings []Tuning) []Tuning {
	nodeMap, err := nodes.HostnameMap()
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

func (t *Tuning) Bytes() ([]byte, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (t *Tuning) SetNodeInfo(role, address string) {
	t.Node = &nodes.Node{
		Role:     role,
		Id:       base.HostID,
		Hostname: base.Hostname,
		Address:  address,
	}
}

func (p *Policy) UpdateOrAppendTuning(tuning Tuning) {
	if !p.existingTuningUpdated(tuning) {
		p.AppendTuning(tuning)
	}
}

func (p *Policy) existingTuningUpdated(tuning Tuning) bool {
	for i, existing := range p.Tunings {
		if existing.Name == tuning.Name {
			p.Tunings[i].Value = tuning.Value
			p.Tunings[i].Enabled = tuning.Enabled
			return true
		}
	}

	return false
}

func (p *Policy) AppendTuning(tuning Tuning) {
	p.Tunings = append(p.Tunings, tuning)
}

func (p *Policy) AppendTunings(tunings []Tuning) {
	p.Tunings = slices.Concat(p.Tunings, tunings)
}

func (p *Policy) HasMatchedTuning(tuning Tuning) bool {
	for _, existing := range p.Tunings {
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

func (p *Policy) DeleteTuning(name string) {
	newTunings := []Tuning{}
	for _, tuning := range p.Tunings {
		if tuning.Name != name {
			newTunings = append(newTunings, tuning)
		}
	}

	p.Tunings = newTunings
}

func DB() string {
	return Module
}

func ReqCollection() string {
	return "requests"
}

func Collection(name string) string {
	return strings.Split(name, ".")[0]
}
