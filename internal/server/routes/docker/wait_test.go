package docker

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/joyrex2001/kubedock/internal/model"
	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/server/routes/common"
)

type waitResponse struct {
	StatusCode int
	Error      *struct{ Message string }
}

func newWaitServer(t *testing.T) (*httptest.Server, *model.Database) {
	waitInterval = 10 * time.Millisecond
	db, err := model.New()
	if err != nil {
		t.Fatalf("unexpected error creating database: %s", err)
	}
	// a limiter that never allows keeps UpdateContainerStatus away from the backend
	cr := &common.ContextRouter{DB: db, Limiter: rate.NewLimiter(0, 0)}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/containers/:id/wait", func(c *gin.Context) { ContainerWait(cr, c) })
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv, db
}

func newWaitContainer(t *testing.T, db *model.Database, tainr *types.Container) *types.Container {
	if err := db.SaveContainer(tainr); err != nil {
		t.Fatalf("unexpected error saving container: %s", err)
	}
	return tainr
}

// startWait will post a wait request and return once the response headers
// are received, which must happen before the condition is met.
func startWait(t *testing.T, srv *httptest.Server, id, cond string) *http.Response {
	url := srv.URL + "/containers/" + id + "/wait"
	if cond != "" {
		url += "?condition=" + cond
	}
	type result struct {
		res *http.Response
		err error
	}
	ch := make(chan result, 1)
	go func() {
		res, err := http.Post(url, "application/json", nil)
		ch <- result{res, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("unexpected error posting wait: %s", r.err)
		}
		return r.res
	case <-time.After(2 * time.Second):
		t.Fatalf("no response headers received for wait %s", cond)
	}
	return nil
}

func readWait(t *testing.T, res *http.Response) waitResponse {
	defer res.Body.Close()
	type result struct {
		body []byte
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		body, err := io.ReadAll(res.Body)
		ch <- result{body, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("unexpected error reading wait body: %s", r.err)
		}
		out := waitResponse{}
		if err := json.Unmarshal(r.body, &out); err != nil {
			t.Fatalf("unexpected wait body %q: %s", r.body, err)
		}
		return out
	case <-time.After(2 * time.Second):
		t.Fatalf("no wait body received")
	}
	return waitResponse{}
}

func TestContainerWaitStreamsExitCode(t *testing.T) {
	srv, db := newWaitServer(t)
	tainr := newWaitContainer(t, db, &types.Container{Name: "tb303", Running: true})

	res := startWait(t, srv, tainr.ID, "next-exit")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, but got %d", res.StatusCode)
	}

	time.Sleep(50 * time.Millisecond)
	tainr.Running = false
	tainr.Completed = true
	tainr.ExitCode = 3

	out := readWait(t, res)
	if out.StatusCode != 3 || out.Error != nil {
		t.Errorf("expected exit code 3 without error, but got %+v", out)
	}
}

func TestContainerWaitConditions(t *testing.T) {
	srv, db := newWaitServer(t)

	created := newWaitContainer(t, db, &types.Container{Name: "created"})
	if out := readWait(t, startWait(t, srv, created.ID, "")); out.StatusCode != 0 {
		t.Errorf("expected not-running to return for a created container, but got %+v", out)
	}

	failed := newWaitContainer(t, db, &types.Container{Name: "failed", Failed: true, ExitCode: 128})
	if out := readWait(t, startWait(t, srv, failed.ID, "next-exit")); out.StatusCode != 128 || out.Error == nil {
		t.Errorf("expected exit code 128 with error for failed container, but got %+v", out)
	}

	autorm := newWaitContainer(t, db, &types.Container{Name: "autorm", Stopped: true, AutoRemove: true, ExitCode: 7})
	if out := readWait(t, startWait(t, srv, autorm.ID, "removed")); out.StatusCode != 7 {
		t.Errorf("expected exit code 7 for auto-removed container, but got %+v", out)
	}
	if _, err := db.GetContainer(autorm.ID); err == nil {
		t.Errorf("expected auto-removed container to be deleted")
	}

	keep := newWaitContainer(t, db, &types.Container{Name: "keep", Completed: true, ExitCode: 1})
	res := startWait(t, srv, keep.ID, "removed")
	time.Sleep(50 * time.Millisecond)
	if _, err := db.GetContainer(keep.ID); err != nil {
		t.Errorf("expected container without AutoRemove to be kept")
	}
	if err := db.DeleteContainer(keep); err != nil {
		t.Fatalf("unexpected error deleting container: %s", err)
	}
	if out := readWait(t, res); out.StatusCode != 1 {
		t.Errorf("expected exit code 1 once removed, but got %+v", out)
	}
}

func TestContainerWaitErrors(t *testing.T) {
	srv, db := newWaitServer(t)
	tainr := newWaitContainer(t, db, &types.Container{Name: "tb303"})

	tests := []struct {
		id   string
		cond string
		code int
	}{
		{id: "missing", code: http.StatusNotFound},
		{id: tainr.ID, cond: "bogus", code: http.StatusBadRequest},
	}
	for i, tst := range tests {
		res := startWait(t, srv, tst.id, tst.cond)
		res.Body.Close()
		if res.StatusCode != tst.code {
			t.Errorf("failed test %d - expected %d, but got %d", i, tst.code, res.StatusCode)
		}
	}
}
