package images

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/bigstack-oss/bigstack-dependency-go/pkg/mongo"
	"github.com/bigstack-oss/bigstack-dependency-go/pkg/openstack/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// These tests drive the real handlers against a fake Keystone and a fake
// Glance. They use only symbols the handlers had before the authorization
// change, so the same file runs against the old code as a negative control.

type fakeProject struct {
	id, name, domainId string
}

type fakeImage struct {
	id, owner, visibility string
	members               map[string]string
}

type fakeIdentity struct {
	user    string
	project string
	roles   []string
}

var (
	fakeDomains = map[string]string{"default": "Default", "other": "Other"}

	fakeProjects = []fakeProject{
		{id: "p-admin", name: "admin", domainId: "default"},
		{id: "p-alpha", name: "alpha", domainId: "default"},
		{id: "p-beta", name: "beta", domainId: "default"},
	}

	fakeImages = []fakeImage{
		{id: "img-a", owner: "p-alpha", visibility: "private"},
		{id: "img-b", owner: "p-beta", visibility: "private"},
		{id: "img-pub", owner: "p-beta", visibility: "public"},
		{id: "img-shared", owner: "p-beta", visibility: "shared", members: map[string]string{"p-alpha": "accepted"}},
		{id: "img-shared2", owner: "p-beta", visibility: "shared"},
		{id: "img-comm", owner: "p-beta", visibility: "community"},
	}

	fakeTokens = map[string]fakeIdentity{
		"admin-tok":  {user: "admin", project: "p-admin", roles: []string{"admin", "member", "reader"}},
		"alice-tok":  {user: "alice", project: "p-alpha", roles: []string{"member", "reader"}},
		"reader-tok": {user: "rita", project: "p-alpha", roles: []string{"reader"}},
	}

	glancePatches = struct {
		sync.Mutex
		ids []string
	}{}

	testRouter *gin.Engine
)

func TestMain(m *testing.M) {
	srv := httptest.NewServer(http.HandlerFunc(serveFakeCloud))
	defer srv.Close()
	baseUrl = srv.URL

	os.Setenv("OS_AUTH_TYPE", "env")
	os.Setenv("OS_AUTH_URL", srv.URL+"/v3")
	os.Setenv("OS_USERNAME", "cos-api")
	os.Setenv("OS_PASSWORD", "secret")
	os.Setenv("OS_DOMAIN_NAME", "Default")
	os.Setenv("OS_PROJECT_NAME", "admin")
	os.Setenv("OS_PROJECT_DOMAIN_NAME", "Default")

	err := openstack.NewGlobalHelper()
	if err != nil {
		fmt.Printf("failed to init the fake openstack helper(%v)\n", err)
		os.Exit(1)
	}

	err = mongo.NewGlobalHelper(
		mongo.Uri("mongodb://127.0.0.1:1/?serverSelectionTimeoutMS=300"),
		mongo.AuthEnable(false),
	)
	if err != nil {
		fmt.Printf("failed to init the mongo helper(%v)\n", err)
		os.Exit(1)
	}

	gin.SetMode(gin.TestMode)
	testRouter = gin.New()
	testRouter.Use(fakeAuthMiddleware)
	for _, handler := range Handlers {
		testRouter.Handle(handler.Method, handler.Path, handler.Func)
	}

	os.Exit(m.Run())
}

var baseUrl string

// fakeAuthMiddleware stands in for the runtime middleware, which has already
// authenticated the request and recorded how.
func fakeAuthMiddleware(c *gin.Context) {
	c.Set("reqId", "test")
	authType := c.GetHeader("X-Test-Auth")
	if authType != "" {
		c.Set("authType", authType)
	}

	c.Next()
}

func serveFakeCloud(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/v3/"):
		serveFakeKeystone(w, r)
	case strings.HasPrefix(r.URL.Path, "/glance/v2/"):
		serveFakeGlance(w, r)
	case strings.Count(r.URL.Path, "/") == 2 && strings.HasSuffix(r.URL.Path, "/"):
		// gophercloud discovers the major versions at each service root.
		writeFakeJson(w, http.StatusOK, map[string]any{"versions": []map[string]any{
			{"id": "v1.0", "status": "CURRENT"},
			{"id": "v2.0", "status": "CURRENT"},
			{"id": "v3.0", "status": "CURRENT"},
		}})
	default:
		http.NotFound(w, r)
	}
}

func writeFakeJson(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}

func fakeProjectBody(p fakeProject) map[string]any {
	return map[string]any{
		"id": p.id, "name": p.name, "domain_id": p.domainId, "enabled": true, "description": "",
	}
}

func findFakeProject(id string) (fakeProject, bool) {
	for _, p := range fakeProjects {
		if p.id == id {
			return p, true
		}
	}

	return fakeProject{}, false
}

func fakeTokenBody(identity fakeIdentity) map[string]any {
	project, _ := findFakeProject(identity.project)
	domain := map[string]any{"id": project.domainId, "name": fakeDomains[project.domainId]}
	roles := []map[string]any{}
	for _, role := range identity.roles {
		roles = append(roles, map[string]any{"id": "r-" + role, "name": role})
	}

	return map[string]any{
		"token": map[string]any{
			"expires_at": "2099-01-01T00:00:00.000000Z",
			"issued_at":  "2026-01-01T00:00:00.000000Z",
			"methods":    []string{"password"},
			"user":       map[string]any{"id": "u-" + identity.user, "name": identity.user, "domain": domain},
			"project":    map[string]any{"id": project.id, "name": project.name, "domain": domain},
			"roles":      roles,
			"catalog":    fakeCatalog(),
		},
	}
}

func fakeCatalog() []map[string]any {
	urls := map[string]string{
		"identity":           baseUrl + "/v3",
		"compute":            baseUrl + "/compute/",
		"image":              baseUrl + "/glance/",
		"network":            baseUrl + "/network/",
		"dns":                baseUrl + "/dns/",
		"load-balancer":      baseUrl + "/lb/",
		"block-storage":      baseUrl + "/volume/",
		"shared-file-system": baseUrl + "/share/",
		"object-store":       baseUrl + "/object/",
	}

	catalog := []map[string]any{}
	for serviceType, url := range urls {
		catalog = append(catalog, map[string]any{
			"id":   serviceType,
			"type": serviceType,
			"name": serviceType,
			"endpoints": []map[string]any{
				{"id": serviceType + "-public", "interface": "public", "region": "RegionOne", "region_id": "RegionOne", "url": url},
			},
		})
	}

	return catalog
}

func serveFakeKeystone(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v3/")
	switch {
	case path == "auth/tokens" && r.Method == http.MethodPost:
		w.Header().Set("X-Subject-Token", "service-tok")
		writeFakeJson(w, http.StatusCreated, fakeTokenBody(fakeTokens["admin-tok"]))
	case path == "auth/tokens" && r.Method == http.MethodGet:
		identity, found := fakeTokens[r.Header.Get("X-Subject-Token")]
		if !found {
			writeFakeJson(w, http.StatusNotFound, map[string]any{})
			return
		}
		writeFakeJson(w, http.StatusOK, fakeTokenBody(identity))
	case path == "projects":
		list := []map[string]any{}
		for _, p := range fakeProjects {
			name := r.URL.Query().Get("name")
			if name == "" || name == p.name {
				list = append(list, fakeProjectBody(p))
			}
		}
		writeFakeJson(w, http.StatusOK, map[string]any{"projects": list, "links": map[string]any{}})
	case strings.HasPrefix(path, "projects/"):
		p, found := findFakeProject(strings.TrimPrefix(path, "projects/"))
		if !found {
			writeFakeJson(w, http.StatusNotFound, map[string]any{})
			return
		}
		writeFakeJson(w, http.StatusOK, map[string]any{"project": fakeProjectBody(p)})
	case path == "domains":
		list := []map[string]any{}
		for id, name := range fakeDomains {
			wanted := r.URL.Query().Get("name")
			if wanted == "" || wanted == name {
				list = append(list, map[string]any{"id": id, "name": name, "enabled": true})
			}
		}
		writeFakeJson(w, http.StatusOK, map[string]any{"domains": list, "links": map[string]any{}})
	default:
		http.NotFound(w, r)
	}
}

func fakeImageBody(image fakeImage) map[string]any {
	return map[string]any{
		"id":          image.id,
		"name":        image.id,
		"owner":       image.owner,
		"visibility":  image.visibility,
		"status":      "active",
		"disk_format": "qcow2",
		"size":        1048576,
		"created_at":  "2026-01-01T00:00:00Z",
		"updated_at":  "2026-01-01T00:00:00Z",
	}
}

func findFakeImage(id string) (fakeImage, bool) {
	for _, image := range fakeImages {
		if image.id == id {
			return image, true
		}
	}

	return fakeImage{}, false
}

func serveFakeGlance(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/glance/v2/"), "/")
	switch {
	case len(parts) == 1 && parts[0] == "images":
		query := r.URL.Query()
		list := []map[string]any{}
		for _, image := range fakeImages {
			if query.Get("id") != "" && query.Get("id") != image.id {
				continue
			}
			if query.Get("owner") != "" && query.Get("owner") != image.owner {
				continue
			}
			if query.Get("visibility") != "" && query.Get("visibility") != image.visibility {
				continue
			}
			list = append(list, fakeImageBody(image))
		}
		writeFakeJson(w, http.StatusOK, map[string]any{"images": list})
	case len(parts) == 2 && parts[0] == "images":
		image, found := findFakeImage(parts[1])
		if !found {
			writeFakeJson(w, http.StatusNotFound, map[string]any{})
			return
		}
		if r.Method == http.MethodPatch {
			glancePatches.Lock()
			glancePatches.ids = append(glancePatches.ids, image.id)
			glancePatches.Unlock()
		}
		writeFakeJson(w, http.StatusOK, fakeImageBody(image))
	case len(parts) == 4 && parts[0] == "images" && parts[2] == "members":
		image, found := findFakeImage(parts[1])
		status, isMember := image.members[parts[3]]
		if !found || !isMember {
			writeFakeJson(w, http.StatusNotFound, map[string]any{})
			return
		}
		writeFakeJson(w, http.StatusOK, map[string]any{
			"image_id": image.id, "member_id": parts[3], "status": status, "schema": "/v2/schemas/member",
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
		})
	default:
		http.NotFound(w, r)
	}
}

func patchedImages() []string {
	glancePatches.Lock()
	defer glancePatches.Unlock()
	return append([]string{}, glancePatches.ids...)
}

func resetPatchedImages() {
	glancePatches.Lock()
	defer glancePatches.Unlock()
	glancePatches.ids = nil
}

type testReq struct {
	method string
	path   string
	token  string
	auth   string
	body   string
}

func (r testReq) do() *httptest.ResponseRecorder {
	req := httptest.NewRequest(r.method, r.path, strings.NewReader(r.body))
	if r.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	auth := r.auth
	if r.token != "" {
		req.Header.Set("X-Auth-Token", r.token)
		if auth == "" {
			auth = "openstack"
		}
	}
	if auth != "" {
		req.Header.Set("X-Test-Auth", auth)
	}

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	return rec
}

func listedImageIds(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	body := struct {
		Data struct {
			Images []struct {
				Id string `json:"id"`
			} `json:"images"`
		} `json:"data"`
	}{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	ids := []string{}
	for _, image := range body.Data.Images {
		ids = append(ids, image.Id)
	}

	return ids
}

func importPath(query string) string {
	return "/images?file=new.qcow2&name=new-img&os=Ubuntu&destination=cube-storage&" + query
}

func TestListImagesOfAnotherProjectIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images?project=beta", token: "alice-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestListImagesCsvOfAnotherProjectIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images.csv?project=beta", token: "alice-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestListImagesOfOwnProjectForMember(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images?project=alpha", token: "alice-tok"}.do()
	require.ElementsMatch(t, []string{"img-a"}, listedImageIds(t, rec))
}

func TestListImagesWithoutProjectShowsOnlyVisibleImagesToMember(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images", token: "alice-tok"}.do()
	require.ElementsMatch(t, []string{"img-a", "img-pub", "img-shared", "img-comm"}, listedImageIds(t, rec))
}

func TestListImagesOfAnotherProjectForAdmin(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images?project=beta", token: "admin-tok"}.do()
	require.Contains(t, listedImageIds(t, rec), "img-b")
}

func TestListImagesWithoutProjectShowsAllImagesToAdmin(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images", token: "admin-tok"}.do()
	require.ElementsMatch(
		t,
		[]string{"img-a", "img-b", "img-pub", "img-shared", "img-shared2", "img-comm"},
		listedImageIds(t, rec),
	)
}

func TestListImagesShowsOnlyPublicImagesToSamlCaller(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images", auth: "saml"}.do()
	require.ElementsMatch(t, []string{"img-pub", "img-comm"}, listedImageIds(t, rec))
}

func TestListImagesOfProjectIsForbiddenForSamlCaller(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images?project=alpha", auth: "saml"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestListImagesWithoutAuthTypeIsForbidden(t *testing.T) {
	rec := testReq{method: http.MethodGet, path: "/images"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestImportToAnotherProjectIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodPost, path: importPath("project=beta&domain=Default"), token: "alice-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestImportToAnotherDomainIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodPost, path: importPath("project=alpha&domain=Other"), token: "alice-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestImportPublicImageIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodPost, path: importPath("project=alpha&domain=Default&visibility=public"), token: "alice-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestImportFromAnotherHypervisorIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodPost, path: importPath("project=alpha&domain=Default&sourceFromAnotherHypervisor=true"), token: "alice-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestImportReservedImageIsForbiddenForMember(t *testing.T) {
	path := "/images?file=amphora-x.qcow2&name=amphora&os=Ubuntu&destination=cube-storage&project=alpha&domain=Default"
	rec := testReq{method: http.MethodPost, path: path, token: "alice-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestImportToOwnProjectIsForbiddenForReader(t *testing.T) {
	rec := testReq{method: http.MethodPost, path: importPath("project=alpha&domain=Default"), token: "reader-tok"}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

// The import passes authorization and then fails on the unreachable mongo,
// which the handler answers with 400.
func TestImportToOwnProjectIsAllowedForMember(t *testing.T) {
	rec := testReq{method: http.MethodPost, path: importPath("project=alpha&domain=Default&visibility=private"), token: "alice-tok"}.do()
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
}

func TestImportToAnotherProjectIsAllowedForAdmin(t *testing.T) {
	rec := testReq{method: http.MethodPost, path: importPath("project=beta&domain=Default"), token: "admin-tok"}.do()
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
}

func TestUpdateImageOfAnotherProjectIsForbiddenForMember(t *testing.T) {
	resetPatchedImages()
	rec := testReq{method: http.MethodPatch, path: "/images/img-b", token: "alice-tok", body: `{"name":"x"}`}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	require.Empty(t, patchedImages())
}

func TestUpdateMissingImageIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodPatch, path: "/images/nope", token: "alice-tok", body: `{"name":"x"}`}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}

func TestUpdateImageOfOwnProjectForMember(t *testing.T) {
	resetPatchedImages()
	rec := testReq{method: http.MethodPatch, path: "/images/img-a", token: "alice-tok", body: `{"name":"x"}`}.do()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, []string{"img-a"}, patchedImages())
}

func TestPublishImageOfOwnProjectIsForbiddenForMember(t *testing.T) {
	resetPatchedImages()
	rec := testReq{method: http.MethodPatch, path: "/images/img-a", token: "alice-tok", body: `{"visibility":"public"}`}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	require.Empty(t, patchedImages())
}

func TestUpdateImageOfOwnProjectIsForbiddenForReader(t *testing.T) {
	resetPatchedImages()
	rec := testReq{method: http.MethodPatch, path: "/images/img-a", token: "reader-tok", body: `{"name":"x"}`}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	require.Empty(t, patchedImages())
}

func TestUpdateImageOfAnotherProjectForAdmin(t *testing.T) {
	resetPatchedImages()
	rec := testReq{method: http.MethodPatch, path: "/images/img-b", token: "admin-tok", body: `{"name":"x"}`}.do()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, []string{"img-b"}, patchedImages())
}

func TestUpdateMissingImageIsNotFoundForAdmin(t *testing.T) {
	rec := testReq{method: http.MethodPatch, path: "/images/nope", token: "admin-tok", body: `{"name":"x"}`}.do()
	require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}

func TestUpdateImageTaskIsForbiddenForMember(t *testing.T) {
	rec := testReq{method: http.MethodPatch, path: "/images/tasks", token: "alice-tok", body: `{"id":"img-a","name":"img-a"}`}.do()
	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
}
