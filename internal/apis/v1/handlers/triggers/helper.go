package triggers

import (
	"fmt"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/http"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/kubernetes"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/bodies"
	"github.com/bigstack-oss/cube-cos-api/internal/apis/v1/queries"
	"github.com/bigstack-oss/cube-cos-api/internal/cubecos"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/pages"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/triggers"
	"github.com/gin-gonic/gin"
	log "go-micro.dev/v5/logger"
)

type helper struct {
	c       *gin.Context
	reqId   string
	handler string

	mongo      mongo.Helper
	http       http.Helper
	kubernetes kubernetes.Helper

	materials    *materials
	verifyScript map[string]string
	reqOpts      triggers.ReqOpts
	toggle       triggers.Toggle
	rawBody      []byte

	page *pages.Page
}

func initHelper(c *gin.Context, handler string) (*helper, error) {
	h := &helper{
		c:          c,
		reqId:      queries.GetReqId(c),
		handler:    handler,
		mongo:      *mongo.GetGlobalHelper(),
		http:       *http.GetGlobalHelper(),
		kubernetes: *kubernetes.GetGlobalHelper(),
		rawBody:    bodies.ParseReq(c),
	}

	return h, h.parseParamsByHandler()
}

func (h *helper) listMaterials() (*materials, error) {
	attribute, err := h.getAttribute()
	if err != nil {
		return nil, err
	}

	response, err := h.getResponse()
	if err != nil {
		return nil, err
	}

	return &materials{
		Attribute:    *attribute,
		materialResp: *response,
	}, nil
}

func (h *helper) verifyMaterialScript() (string, error) {
	defer h.deleteDryRunArtifacts()
	err := h.createConfigMapWithScript()
	if err != nil {
		return "", err
	}

	err = h.dryRunScript()
	if err != nil {
		return "", err
	}

	result, err := h.waitDryRunResult()
	if err != nil {
		return "", err
	}

	return result, nil
}

func (h *helper) listTriggers() (*triggerPage, error) {
	list := []triggerResp{}
	for _, trigger := range triggers.List() {
		resp := h.convertTrigger(trigger)

		h.syncBuiltInInfo(&resp)
		h.syncResponseTypes(&resp)
		h.syncInProgressInfo(&resp)

		list = append(list, resp)
	}

	h.addCreatingTriggers(&list)
	h.sortTriggers(&list)
	return &triggerPage{
		Triggers: h.paginateTriggers(list),
		Page:     h.genPageInfo(list),
	}, nil
}

func (h *helper) getTrigger(name string) (*triggerResp, error) {
	for _, trigger := range triggers.List() {
		if trigger.Name == name {
			resp := h.convertTrigger(trigger)
			h.syncBuiltInInfo(&resp)
			h.syncResponseTypes(&resp)
			h.syncInProgressInfo(&resp)
			return &resp, nil
		}
	}

	return nil, fmt.Errorf(
		"trigger(%s): trigger not found",
		name,
	)
}

func (h *helper) checkTaskUpdateReq() error {
	if h.reqOpts.Name == "" {
		return fmt.Errorf("trigger name is required")
	}

	return nil
}

func (h *helper) getAttribute() (*triggers.Attribute, error) {
	events, err := cubecos.GetPredefinedEvents()
	if err != nil {
		log.Errorf("triggers(%s): failed to get predefined events(%v)", h.reqId, err)
		return nil, err
	}

	eventIds, err := h.GetEventIds(events)
	if err != nil {
		log.Errorf("triggers(%s): failed to get event ids(%v)", h.reqId, err)
		return nil, err
	}

	alertTypes, err := h.GetAlertTypes(events)
	if err != nil {
		log.Errorf("triggers(%s): failed to get alert types(%v)", h.reqId, err)
		return nil, err
	}

	severities, err := h.GetSeverities(events)
	if err != nil {
		log.Errorf("triggers(%s): failed to get severities(%v)", h.reqId, err)
		return nil, err
	}

	categories, err := h.GetCategories(events)
	if err != nil {
		log.Errorf("triggers(%s): failed to get categories(%v)", h.reqId, err)
		return nil, err
	}

	return &triggers.Attribute{
		EventIds:   eventIds,
		AlertTypes: alertTypes,
		Severities: severities,
		Categories: categories,
	}, nil
}

func (h *helper) getResponse() (*materialResp, error) {
	notifications, err := h.getNotifications()
	if err != nil {
		return nil, err
	}

	return &materialResp{
		ScriptType: ScriptType{
			Language:    "Bash",
			Environment: "Alpine Linux",
			BuiltInVariable: BuiltInVariable{
				Name:        "EVENT",
				Type:        "object",
				Description: "the 'case sensitive' event env variable that triggered the action",
				Value:       builtInVariable,
			},
		},
		Emails: notifications.Emails,
		Slacks: notifications.Slacks,
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

	emails := []Email{}
	for _, receipient := range receipients {
		if receipient.Address == "" {
			continue
		}

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

	slacks := []Slack{}
	for _, channel := range channels {
		if channel.URL == "" {
			continue
		}

		slacks = append(slacks, Slack{
			Name:        channel.Channel,
			Url:         channel.URL,
			Description: channel.Description,
		})
	}

	return slacks, nil
}
