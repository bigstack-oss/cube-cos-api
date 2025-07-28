package triggers

import (
	"fmt"
	"strings"
	"time"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/status"
)

type ReqOpts struct {
	Name        string `json:"name" bson:"name"`
	Enabled     bool   `json:"enabled" bson:"enabled"`
	Description string `json:"description" bson:"description"`
	Attribute   `json:"attribute" bson:"attribute"`
	Response    `json:"response" bson:"response"`
	Nodes       []string       `json:"nodes" bson:"nodes"`
	Status      status.Trigger `json:"status" bson:"status"`
}

type Attribute struct {
	AlertTypes []string `json:"alertTypes" bson:"alertTypes"`
	EventIds   []string `json:"eventIds" bson:"eventIds"`
	Severities []string `json:"severities" bson:"severities"`
	Categories []string `json:"categories" bson:"categories"`
}

type Response struct {
	Script `json:"script" bson:"script"`
	Emails []string `json:"emails" bson:"emails"`
	Slacks []string `json:"slacks" bson:"slacks"`
}

type Script struct {
	Name    string `json:"name" bson:"name"`
	Content string `json:"content" bson:"content"`
}

type Toggle struct {
	Enable bool     `json:"enable" yaml:"enable"`
	Nodes  []string `json:"nodes" yaml:"nodes"`
}

func (r *ReqOpts) SetUpdating() {
	r.Status = status.Trigger{
		Current:    status.Updating,
		Desired:    status.Updated,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
		UpdatedAt:  time.Now().Local().Format(time.RFC3339),
		IsUpdating: true,
	}
}

func (r *ReqOpts) SetDeleting() {
	r.Status = status.Trigger{
		Current:    status.Deleting,
		Desired:    status.Deleted,
		CreatedAt:  time.Now().Local().Format(time.RFC3339),
		UpdatedAt:  time.Now().Local().Format(time.RFC3339),
		IsDeleting: true,
	}
}

func (r *ReqOpts) SetCompleted() {
	r.Status.Current = status.Completed
	r.Status.IsCreating = false
}

func (r *ReqOpts) SetError() {
	r.Status.Current = status.Error
	r.Status.IsDeleting = false
}

func (r *ReqOpts) GenMatchRule() string {
	andRules := []string{}
	if len(r.Attribute.AlertTypes) > 0 {
		andRules = append(
			andRules,
			r.GenOrRule("type", r.Attribute.AlertTypes),
		)
	}

	if len(r.Attribute.EventIds) > 0 {
		andRules = append(
			andRules,
			r.GenOrRule("id", r.Attribute.EventIds),
		)
	}

	if len(r.Attribute.Severities) > 0 {
		andRules = append(
			andRules,
			r.GenOrRule("severity", r.Attribute.Severities),
		)
	}

	if len(r.Attribute.Categories) > 0 {
		andRules = append(
			andRules,
			r.GenOrRule("category", r.Attribute.Categories),
		)
	}

	return strings.Join(
		andRules,
		" AND ",
	)
}

func (r *ReqOpts) GenOrRule(key string, attrs []string) string {
	rule := []string{}
	for _, attr := range attrs {
		rule = append(
			rule,
			fmt.Sprintf(`%q == %q`, key, attr),
		)
	}

	orRule := strings.Join(rule, " OR ")
	return fmt.Sprintf(`(%s)`, orRule)
}
