package docker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/joyrex2001/kubedock/internal/model"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

func newVolumesRouter(t *testing.T) *gin.Engine {
	db, err := model.New()
	if err != nil {
		t.Fatalf("unexpected error creating database: %s", err)
	}
	cr := &common.ContextRouter{DB: db}
	wrap := func(fn func(*common.ContextRouter, *gin.Context)) gin.HandlerFunc {
		return func(c *gin.Context) { fn(cr, c) }
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/volumes", wrap(VolumesList))
	router.GET("/volumes/:name", wrap(VolumesInfo))
	router.DELETE("/volumes/:name", wrap(VolumesDelete))
	router.POST("/volumes/create", wrap(VolumesCreate))
	router.POST("/volumes/prune", wrap(VolumesPrune))
	return router
}

func doVolumesRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	router.ServeHTTP(w, req)
	return w
}

func listVolumeNames(t *testing.T, router *gin.Engine, filters string) []string {
	path := "/volumes"
	if filters != "" {
		path += "?filters=" + url.QueryEscape(filters)
	}
	w := doVolumesRequest(router, http.MethodGet, path, "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 listing volumes, but got %d", w.Code)
	}
	res := struct {
		Volumes  []struct{ Name string }
		Warnings []string
	}{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unexpected error decoding volume list: %s", err)
	}
	if res.Volumes == nil || res.Warnings == nil {
		t.Errorf("expected non-null Volumes and Warnings, but got %s", w.Body.String())
	}
	names := []string{}
	for _, v := range res.Volumes {
		names = append(names, v.Name)
	}
	return names
}

func TestVolumesLifecycle(t *testing.T) {
	router := newVolumesRouter(t)

	if names := listVolumeNames(t, router, ""); len(names) != 0 {
		t.Errorf("expected no volumes, but got %v", names)
	}

	if w := doVolumesRequest(router, http.MethodGet, "/volumes/data", ""); w.Code != http.StatusNotFound {
		t.Errorf("expected 404 inspecting missing volume, but got %d", w.Code)
	}

	for i := 0; i < 2; i++ {
		w := doVolumesRequest(router, http.MethodPost, "/volumes/create",
			`{"Name":"data","Labels":{"com.docker.compose.project":"proj"}}`)
		if w.Code != http.StatusCreated {
			t.Errorf("expected 201 creating volume (attempt %d), but got %d", i, w.Code)
		}
	}
	doVolumesRequest(router, http.MethodPost, "/volumes/create", `{"Name":"other"}`)

	w := doVolumesRequest(router, http.MethodGet, "/volumes/data", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 inspecting volume, but got %d", w.Code)
	}
	vol := struct {
		Name   string
		Driver string
		Labels map[string]string
	}{}
	if err := json.Unmarshal(w.Body.Bytes(), &vol); err != nil {
		t.Fatalf("unexpected error decoding volume: %s", err)
	}
	if vol.Name != "data" || vol.Driver != "local" || vol.Labels["com.docker.compose.project"] != "proj" {
		t.Errorf("unexpected volume details: %s", w.Body.String())
	}

	if names := listVolumeNames(t, router, ""); len(names) != 2 {
		t.Errorf("expected 2 volumes, but got %v", names)
	}
	if names := listVolumeNames(t, router, `{"label":{"com.docker.compose.project=proj":true}}`); len(names) != 1 || names[0] != "data" {
		t.Errorf("expected only volume data for label filter, but got %v", names)
	}

	if w := doVolumesRequest(router, http.MethodDelete, "/volumes/data", ""); w.Code != http.StatusNoContent {
		t.Errorf("expected 204 deleting volume, but got %d", w.Code)
	}
	if w := doVolumesRequest(router, http.MethodDelete, "/volumes/data", ""); w.Code != http.StatusNotFound {
		t.Errorf("expected 404 deleting missing volume, but got %d", w.Code)
	}

	w = doVolumesRequest(router, http.MethodPost, "/volumes/prune", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 pruning volumes, but got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"other"`) {
		t.Errorf("expected volume other to be pruned, but got %s", w.Body.String())
	}
	if names := listVolumeNames(t, router, ""); len(names) != 0 {
		t.Errorf("expected no volumes after prune, but got %v", names)
	}
}
