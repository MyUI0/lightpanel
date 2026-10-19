package handlers

import (
	"lightpanel/config"
	"lightpanel/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReviewRenameRebasesInternalWorkDir(t *testing.T) {
	setupInstallTest(t)
	oldPath := filepath.Join(config.AppsDir, "before")
	oldWork := filepath.Join(oldPath, "sub")
	if err := os.MkdirAll(oldWork, 0700); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(config.ConfigApps, map[string]models.Project{"before": {Path: oldPath, WorkDir: oldWork, Cmd: "echo ready"}}); err != nil {
		t.Fatal(err)
	}
	values := url.Values{"name": {"after"}, "path": {oldPath}, "work_dir": {oldWork}, "cmd": {"echo ready"}}
	r := httptest.NewRequest("POST", "/edit/before", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	editAppHandler(w, r)
	apps, err := loadAppsStrict()
	if err != nil {
		t.Fatal(err)
	}
	got, ok := apps["after"]
	if !ok {
		t.Fatalf("rename failed: %d %s", w.Code, w.Header().Get("Location"))
	}
	want := filepath.Join(config.AppsDir, "after", "sub")
	if got.WorkDir != want {
		t.Fatalf("renamed directory but stale workdir: got %q want %q", got.WorkDir, want)
	}
	if _, err := os.Stat(got.WorkDir); err != nil {
		t.Fatal(err)
	}
}

func TestReviewLegacyDefaultLoginIsForcedToChange(t *testing.T) {
	setupTestAccount(t)
	oldSessions := sessionDir
	sessionDir = t.TempDir()
	t.Cleanup(func() { sessionDir = oldSessions })
	if err := WriteJSON(config.ConfigUsr, models.UserConfig{Username: "admin", Password: hash("admin")}); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/login/auth", strings.NewReader(url.Values{"username": {"admin"}, "password": {"admin"}}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	loginHandler(w, r)
	if w.Code != 302 {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var token, csrf string
	for _, c := range w.Result().Cookies() {
		if c.Name == "lp_session" {
			token = c.Value
		}
		if c.Name == "lp_csrf" {
			csrf = c.Value
		}
	}
	sess := getSessionData(token)
	if sess == nil || !sess.FirstLogin {
		t.Fatal("legacy default password did not get restricted session")
	}
	var user models.UserConfig
	if err := LoadJSON(config.ConfigUsr, &user); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(user.Password, "pbkdf2-sha256$") || !user.MustChangePassword {
		t.Fatal("legacy upgrade lost change-password requirement")
	}
	for _, tc := range []struct {
		path     string
		mutation bool
		want     int
	}{{"/", false, 303}, {"/setting", false, 204}, {"/create", true, 403}, {"/setting/save-account", true, 204}} {
		req := httptest.NewRequest("POST", tc.path, nil)
		req.AddCookie(&http.Cookie{Name: "lp_session", Value: token})
		req.Header.Set("X-CSRF-Token", csrf)
		out := httptest.NewRecorder()
		handler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }
		if tc.mutation {
			authWithCSRF(handler)(out, req)
		} else {
			auth(handler)(out, req)
		}
		if out.Code != tc.want {
			t.Fatalf("%s returned %d want %d", tc.path, out.Code, tc.want)
		}
	}
}

func TestReviewChangedManifestCannotStartInstall(t *testing.T) {
	setupInstallTest(t)
	oldSrc := config.ConfigSrc
	config.ConfigSrc = filepath.Join(config.ConfigDir, "sources.json")
	t.Cleanup(func() { config.ConfigSrc = oldSrc })
	source := "https://store.example/list"
	if err := WriteJSON(config.ConfigSrc, []models.StoreSource{{URL: source}}); err != nil {
		t.Fatal(err)
	}
	oldCache := storeCache
	t.Cleanup(func() { storeCache = oldCache })
	offered := models.StoreApp{Name: "demo", Cmd: "echo old"}
	changed := offered
	changed.Cmd = "echo changed"
	storeCache.apps = []models.StoreApp{changed}
	storeCache.key = source
	storeCache.err = ""
	storeCache.lastInit = time.Now()
	for _, endpoint := range []string{"/install/0?source=0", "/install/confirm/0?source=0"} {
		req := httptest.NewRequest("POST", endpoint, strings.NewReader(url.Values{"manifest": {storeDigest(source, offered)}}.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		out := httptest.NewRecorder()
		if strings.HasPrefix(endpoint, "/install/confirm/") {
			confirmInstallWithParams(out, req)
		} else {
			startStoreInstall(out, req)
		}
		if out.Code != 303 || !strings.Contains(out.Header().Get("Location"), "err=") {
			t.Fatalf("changed manifest accepted: %d %s", out.Code, out.Header().Get("Location"))
		}
	}
	apps, err := loadAppsStrict()
	if err != nil || len(apps) != 0 {
		t.Fatalf("unexpected application: %v %v", apps, err)
	}
}

func TestReviewRenameRejectsReservedFinalTarget(t *testing.T) {
	setupInstallTest(t)
	oldPath := filepath.Join(config.AppsDir, "before")
	target := filepath.Join(config.AppsDir, "after")
	if err := os.MkdirAll(oldPath, 0700); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(config.ConfigApps, map[string]models.Project{"before": {Path: oldPath, Cmd: "echo ready"}}); err != nil {
		t.Fatal(err)
	}
	appOpMu.Lock()
	installReservations[pathKey(target)] = "review-reservation"
	appOpMu.Unlock()
	t.Cleanup(func() { appOpMu.Lock(); delete(installReservations, pathKey(target)); appOpMu.Unlock() })
	values := url.Values{"name": {"after"}, "path": {oldPath}, "cmd": {"echo ready"}}
	r := httptest.NewRequest("POST", "/edit/before", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	editAppHandler(w, r)
	apps, err := loadAppsStrict()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := apps["before"]; !ok {
		t.Fatal("rename consumed pending installation target")
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatal("original directory moved despite reservation")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("reserved target was created")
	}
}
