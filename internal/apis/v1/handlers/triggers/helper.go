package triggers

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/gin-gonic/gin"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	http http.Helper

	trigger               triggers.ApiSchema
	toggle                triggers.Toggle
	rawBody               []byte
	isClusterWiseRequired bool
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:       c,
		reqId:   queries.GetReqId(c),
		handler: handler,
		http:    *http.GetGlobalHelper(),
		rawBody: bodies.ParseReq(c),
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listTriggerMaterials() (*materials, error) {
	attribute, err := h.getAttribute()
	if err != nil {
		return nil, err
	}

	response, err := h.getResponse()
	if err != nil {
		return nil, err
	}

	return &materials{
		Attribute: *attribute,
		Response:  *response,
	}, nil
}

func (h *helper) listTriggers() ([]triggers.ApiSchema, error) {
	list := []triggers.ApiSchema{}
	for _, trigger := range triggers.List() {
		h.syncUpdatingInfo(&trigger)
		list = append(list, trigger)
	}

	return list, nil
}

func (h *helper) getTrigger(name string) (*triggers.ApiSchema, error) {
	for _, trigger := range triggers.List() {
		if trigger.Name == name {
			return &trigger, nil
		}
	}

	return nil, fmt.Errorf(
		"trigger(%s): trigger not found",
		name,
	)
}

func (h *helper) checkTaskUpdateReq() error {
	if h.trigger.Name == "" {
		return fmt.Errorf("trigger id is required")
	}

	if h.trigger.Status == nil {
		return fmt.Errorf("trigger status is required")
	}

	return nil
}

func (h *helper) getAttribute() (*Attribute, error) {
	alertTypes, err := cubecos.GetAlertTypes()
	if err != nil {
		return nil, err
	}

	severities, err := cubecos.GetSeverities()
	if err != nil {
		return nil, err
	}

	categories, err := cubecos.GetCategories()
	if err != nil {
		return nil, err
	}

	eventIds, err := cubecos.GetEventIds()
	if err != nil {
		return nil, err
	}

	return &Attribute{
		AlertTypes: alertTypes,
		Severities: severities,
		Categories: categories,
		EventIds:   eventIds,
	}, nil
}

func (h *helper) getResponse() (*Response, error) {
	scriptTypes, err := cubecos.GetScriptTypes()
	if err != nil {
		return nil, err
	}

	notifications, err := h.getNotifications()
	if err != nil {
		return nil, err
	}

	return &Response{
		ScriptTypes:   scriptTypes,
		Notifications: *notifications,
	}, nil
}

func (h *helper) getNotifications() (*Notifications, error) {
	emails, err := h.GetEmails()
	if err != nil {
		return nil, err
	}

	slacks, err := h.GetSlacks()
	if err != nil {
		return nil, err
	}

	return &Notifications{
		Emails: emails,
		Slacks: slacks,
	}, nil
}

func (h *helper) GetEmails() ([]Email, error) {
	receipients, err := cubecos.GetEmailRecipients()
	if err != nil {
		return nil, err
	}

	emails := make([]Email, len(receipients))
	for _, receipient := range receipients {
		emails = append(emails, Email{
			Address: receipient.Address,
			Note:    receipient.Note,
		})
	}

	return emails, nil
}

func (h *helper) GetSlacks() ([]Slack, error) {
	channels, err := cubecos.GetSlackChannels()
	if err != nil {
		return nil, err
	}

	slacks := make([]Slack, len(channels))
	for _, channel := range channels {
		slacks = append(slacks, Slack{
			Name:        channel.Channel,
			Url:         channel.URL,
			Description: channel.Description,
		})
	}

	return slacks, nil
}
