package bodies

import (
	"encoding/json"
	osmath "math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	return c, recorder
}

// A node record that carries NaN cannot be encoded. gin writes the content type
// before it encodes and only records the failure in c.Errors, so the response
// reached the client as 200 with an empty body. Every peer then read it as
// "unexpected end of JSON input".
func TestSetOkAnswersServerErrorWhenTheBodyCannotEncode(t *testing.T) {
	c, recorder := newTestContext(t)

	SetOk(c, "fetch node successfully", map[string]float64{"usedPercent": osmath.NaN()})

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.NotEmpty(t, recorder.Body.Bytes(), "the response body is empty")

	body := map[string]any{}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "internal server error", body[Status])
}

func TestSetOkWritesTheEncodedBody(t *testing.T) {
	c, recorder := newTestContext(t)

	SetOk(c, "fetch node successfully", map[string]string{"hostname": "cn07"})

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

	body := map[string]any{}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "ok", body[Status])
	require.Equal(t, map[string]any{"hostname": "cn07"}, body[Data])
}
