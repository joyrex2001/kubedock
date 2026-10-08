package docker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

func TestVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(httputil.VersionAliasMiddleware(router))
	router.GET("/version", func(c *gin.Context) { Version(&common.ContextRouter{}, c) })

	for _, path := range []string{"/version", "/v1.25/version", "/v" + config.DockerAPIVersion + "/version"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, but got %d", path, w.Code)
			continue
		}
		res := struct {
			APIVersion    string `json:"ApiVersion"`
			MinAPIVersion string
		}{}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Errorf("unexpected error decoding %s: %s", path, err)
		}
		if res.APIVersion != config.DockerAPIVersion || res.MinAPIVersion != config.DockerMinAPIVersion {
			t.Errorf("unexpected versions for %s: %s", path, w.Body.String())
		}
	}
}
