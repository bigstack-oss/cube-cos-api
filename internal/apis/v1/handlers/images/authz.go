package images

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/wait"
	"github.com/bigstack-oss/cube-cos-api/internal/auths/authz"
	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/images"
	"github.com/gophercloud/gophercloud/v2"
	opsimage "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	"github.com/gophercloud/gophercloud/v2/openstack/image/v2/members"
	log "go-micro.dev/v5/logger"
)

// The image handlers act on Glance and Keystone with the API's own admin
// credential, so the caller's own rights are checked here. The rules follow
// the Glance policy CubeCOS ships: writes need the member role in the image's
// project, and only admins publish images or touch other projects.
var writeRole = "member"

type caller struct {
	user        string
	isAdmin     bool
	projectId   string
	projectName string
	domainId    string
	domainName  string
	roles       []string
}

func (c caller) hasProject() bool {
	return c.projectId != ""
}

func (c caller) canWrite() bool {
	return c.hasProject() && slices.Contains(c.roles, writeRole)
}

type forbiddenError struct {
	reason string
}

func (e *forbiddenError) Error() string {
	return e.reason
}

func isForbidden(err error) bool {
	var forbidden *forbiddenError
	return errors.As(err, &forbidden)
}

func (h *helper) deny(reason string) error {
	log.Warnf(
		"images(%s): denied %s for user %s in project %s(%s)",
		h.reqId, h.handler, h.caller.user, h.caller.projectName, reason,
	)

	return &forbiddenError{reason: reason}
}

func (h *helper) allow() {
	log.Infof(
		"images(%s): allowed %s for user %s in project %s",
		h.reqId, h.handler, h.caller.user, h.caller.projectName,
	)
}

func (h *helper) resolveCaller() error {
	resolved, err := authz.Resolve(h.c)
	if errors.Is(err, authz.ErrUnavailable) {
		return err
	}

	if err != nil {
		return h.deny(err.Error())
	}

	h.caller = caller{
		user:        resolved.User,
		isAdmin:     resolved.IsAdmin,
		projectId:   resolved.ProjectId,
		projectName: resolved.ProjectName,
		domainId:    resolved.DomainId,
		domainName:  resolved.DomainName,
		roles:       resolved.Roles,
	}

	return nil
}

func (h *helper) authorizeList() error {
	if h.caller.isAdmin {
		return nil
	}

	if !h.caller.hasProject() || h.project != h.caller.projectName {
		return h.deny(fmt.Sprintf("listing images of project %s is not allowed", h.project))
	}

	return nil
}

func (h *helper) authorizeImport() error {
	if h.caller.isAdmin {
		h.allow()
		return nil
	}

	if !h.caller.canWrite() {
		return h.deny("importing images needs the member role in a project")
	}

	if h.reqOpts.Reserved.Type != "" {
		return h.deny(fmt.Sprintf("importing reserved image %s is not allowed", h.reqOpts.File))
	}

	if h.reqOpts.Project != h.caller.projectName || h.reqOpts.Domain != h.caller.domainName {
		return h.deny(fmt.Sprintf("importing images to project %s in domain %s is not allowed", h.reqOpts.Project, h.reqOpts.Domain))
	}

	// Images from another hypervisor are always imported as public.
	if h.reqOpts.Visibility == string(opsimage.ImageVisibilityPublic) || h.reqOpts.SourceFromAnotherHypervisor {
		return h.deny("importing public images is not allowed")
	}

	h.allow()
	return nil
}

func (h *helper) authorizeUpdate() error {
	if h.caller.isAdmin {
		h.allow()
		return nil
	}

	if !h.caller.canWrite() {
		return h.deny("updating images needs the member role in a project")
	}

	// A missing image is refused like a foreign one, so the answer does not
	// tell whether an image id exists.
	image, err := h.openstack.GetImage(h.reqOpts.Id)
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return h.deny(fmt.Sprintf("updating image %s is not allowed", h.reqOpts.Id))
	}
	if err != nil {
		log.Errorf("images(%s): failed to get image %s(%v)", h.reqId, h.reqOpts.Id, err)
		return err
	}

	if image.Owner != h.caller.projectId {
		return h.deny(fmt.Sprintf("updating image %s is not allowed", h.reqOpts.Id))
	}

	if h.reqOpts.Visibility == string(opsimage.ImageVisibilityPublic) {
		return h.deny("publishing images is not allowed")
	}

	h.allow()
	return nil
}

func (h *helper) authorizeTask() error {
	if h.caller.isAdmin {
		return nil
	}

	return h.deny("image tasks are reported by nodes only")
}

// visibleImages keeps the images Glance would list for the caller's token.
func (h *helper) visibleImages(list []opsimage.Image) ([]opsimage.Image, error) {
	if h.caller.isAdmin {
		return list, nil
	}

	visible := []opsimage.Image{}
	for _, image := range list {
		isVisible, err := h.isImageVisible(image)
		if err != nil {
			return nil, err
		}
		if isVisible {
			visible = append(visible, image)
		}
	}

	return visible, nil
}

func (h *helper) isImageVisible(image opsimage.Image) (bool, error) {
	switch image.Visibility {
	case opsimage.ImageVisibilityPublic, opsimage.ImageVisibilityCommunity:
		return true, nil
	}

	if !h.caller.hasProject() {
		return false, nil
	}

	if image.Owner == h.caller.projectId {
		return true, nil
	}

	if image.Visibility != opsimage.ImageVisibilityShared {
		return false, nil
	}

	return h.isAcceptedMember(image.ID)
}

// isAcceptedMember treats only Glance's 404 as "not a member". Any other
// failure is returned, so the list fails instead of silently dropping images.
func (h *helper) isAcceptedMember(imageId string) (bool, error) {
	ctx, cancel := context.WithTimeout(wait.CtxSeconds(30))
	defer cancel()

	member, err := members.Get(ctx, h.openstack.Image, imageId, h.caller.projectId).Extract()
	if gophercloud.ResponseCodeIs(err, http.StatusNotFound) {
		return false, nil
	}
	if err != nil {
		log.Errorf("images(%s): failed to get member %s of image %s(%v)", h.reqId, h.caller.projectId, imageId, err)
		return false, err
	}

	return member.Status == "accepted", nil
}

// visibleProcessings keeps the import records of the caller's project. The
// records hold project and domain names, as the import request does.
func (h *helper) visibleProcessings(processings []images.Image) []images.Image {
	if h.caller.isAdmin {
		return processings
	}

	visible := []images.Image{}
	if !h.caller.hasProject() {
		return visible
	}

	for _, processing := range processings {
		if processing.Project == h.caller.projectName && processing.Domain == h.caller.domainName {
			visible = append(visible, processing)
		}
	}

	return visible
}
