package handlers

import (
	"bytes"
	"lightpanel/config"
	"lightpanel/models"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManagedProcessHelper(t *testing.T) {
	if os.Getenv("LP_TEST_HELPER") != "1" {
		return
	}
	if os.Args[len(os.Args)-1] == "crash" {
		os.Exit(1)
	}
	time.Sleep(time.Minute)
	os.Exit(0)
}

func setupLifecycle(t *testing.T, mode string) models.Project {
	t.Helper()
	t.Setenv("LP_TEST_HELPER", "1")
	old := config.ConfigApps
	dir := t.TempDir()
	config.ConfigApps = filepath.Join(dir, "apps.json")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	app := models.Project{Path: dir, Cmd: `"` + executable + `" -test.run=TestManagedProcessHelper -- ` + mode, AutoStart: true}
	if err := WriteJSON(config.ConfigApps, map[string]models.Project{"test-app": app}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { appOpMu.Lock(); killAppByName("test-app"); appOpMu.Unlock(); config.ConfigApps = old })
	return app
}

func TestManualStopDoesNotRestartAutoStartApp(t *testing.T) {
	app := setupLifecycle(t, "wait")
	appOpMu.Lock()
	started := launchApp("test-app", &app)
	appOpMu.Unlock()
	if !started {
		t.Fatal("launch failed")
	}
	appOpMu.Lock()
	killAppByName("test-app")
	appOpMu.Unlock()
	time.Sleep(3500 * time.Millisecond)
	checkAppStatus(&app)
	if app.PID != 0 {
		t.Fatal("manually stopped app restarted")
	}
}

func TestCrashingAppStopsAfterThreeRetries(t *testing.T) {
	app := setupLifecycle(t, "crash")
	appOpMu.Lock()
	delete(crashCounts, "test-app")
	started := launchApp("test-app", &app)
	appOpMu.Unlock()
	if !started {
		t.Fatal("launch failed")
	}
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		appOpMu.Lock()
		count := crashCounts["test-app"].count
		appOpMu.Unlock()
		if count == 4 {
			checkAppStatus(&app)
			if app.PID != 0 {
				t.Fatal("retry limit left process running")
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("crash retry limit did not take effect")
}

func TestCSRFHeaderPreservesMultipartFields(t *testing.T) {
	setupTestAccount(t)
	old := sessionDir
	sessionDir = t.TempDir()
	defer func() { sessionDir = old }()
	session := httptest.NewRecorder()
	createSession(session, false)
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	form.WriteField("name", "demo")
	form.Close()
	req := httptest.NewRequest("POST", "/create/manual", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	for _, cookie := range session.Result().Cookies() {
		req.AddCookie(cookie)
		if cookie.Name == "lp_csrf" {
			req.Header.Set("X-CSRF-Token", cookie.Value)
		}
	}
	out := httptest.NewRecorder()
	authWithCSRF(func(w http.ResponseWriter, r *http.Request) {
		if r.Form.Get("name") != "demo" {
			t.Error("multipart name not parsed")
		}
		w.WriteHeader(204)
	})(out, req)
	if out.Code != 204 {
		t.Fatalf("status %d: %s", out.Code, out.Body.String())
	}
}
