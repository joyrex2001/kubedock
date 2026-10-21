package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joyrex2001/kubedock/internal/config"
	"github.com/joyrex2001/kubedock/internal/model"
	"github.com/joyrex2001/kubedock/internal/server/httputil"
)

func TestImageJSONPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := model.New()
	if err != nil {
		t.Fatalf("could not create model database: %s", err)
	}
	cr := &ContextRouter{DB: db}
	router := gin.New()
	router.Use(httputil.VersionAliasMiddleware(router))
	router.GET("/images/:image/*json", func(c *gin.Context) { ImageJSON(cr, c) })

	platform := url.QueryEscape(`{"os":"` + config.GOOS + `","architecture":"` + config.GOARCH + `"}`)
	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{name: "default platform", query: "", wantStatus: http.StatusOK},
		{name: "matching platform", query: "platform=" + platform, wantStatus: http.StatusOK},
		{name: "platform with manifests disabled", query: "manifests=false&platform=" + platform, wantStatus: http.StatusOK},
		{name: "invalid platform json", query: "platform=not-json", wantStatus: http.StatusBadRequest},
		{name: "platform requires os and architecture", query: "platform=" + url.QueryEscape(`{"os":"linux"}`), wantStatus: http.StatusBadRequest},
		{name: "unsupported platform", query: "platform=" + url.QueryEscape(`{"os":"unsupported","architecture":"`+config.GOARCH+`"}`), wantStatus: http.StatusNotFound},
		{name: "manifests unsupported", query: "manifests=true", wantStatus: http.StatusNotImplemented},
		{name: "manifests and platform are exclusive", query: "manifests=true&platform=" + platform, wantStatus: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v"+config.DockerAPIVersion+"/images/platform-test/json?"+test.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != test.wantStatus {
				t.Fatalf("expected status %d, got %d: %s", test.wantStatus, w.Code, w.Body.String())
			}
			if w.Code == http.StatusOK {
				res := struct {
					Architecture string
					Os           string
				}{}
				if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
					t.Fatalf("could not decode image inspect response: %s", err)
				}
				if res.Architecture != config.GOARCH || res.Os != config.GOOS {
					t.Errorf("unexpected image platform: os=%q architecture=%q", res.Os, res.Architecture)
				}
			}
		})
	}
}
