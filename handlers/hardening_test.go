package handlers

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func setupTestAccount(t *testing.T) {
	t.Helper()
	old := config.ConfigUsr
	config.ConfigUsr = filepath.Join(t.TempDir(), "user.json")
	if err := WriteJSON(config.ConfigUsr, models.UserConfig{Username: "admin", Password: hash("old-password-long")}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { config.ConfigUsr = old })
}
func setupInstallTest(t *testing.T) {
	t.Helper()
	oldData, oldApps, oldCfg, oldDir := config.DataDir, config.AppsDir, config.ConfigApps, config.ConfigDir
	config.DataDir = t.TempDir()
	config.AppsDir = filepath.Join(config.DataDir, "apps")
	config.ConfigDir = filepath.Join(config.DataDir, "config")
	config.ConfigApps = filepath.Join(config.ConfigDir, "apps.json")
	os.MkdirAll(config.AppsDir, 0700)
	if err := WriteJSON(config.ConfigApps, map[string]models.Project{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		installJobs.Range(func(k, v any) bool { _ = v.(*installJob).action("delete"); return true })
		config.DataDir, config.AppsDir, config.ConfigApps, config.ConfigDir = oldData, oldApps, oldCfg, oldDir
	})
}

type regressionTransport func(*http.Request) (*http.Response, error)

func (f regressionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func mockRemote(t *testing.T, f regressionTransport) {
	t.Helper()
	old := publicTransport
	publicTransport = f
	t.Cleanup(func() { publicTransport = old })
}
func response(r *http.Request, data string) *http.Response {
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data)), ContentLength: int64(len(data)), Request: r}
}
func waitInstall(t *testing.T, j *installJob, status string) {
	t.Helper()
	j.mu.Lock()
	done := j.done
	j.mu.Unlock()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not stop")
	}
	if got := j.task.GetStatus(); got != status {
		t.Fatalf("status=%s want=%s: %s", got, status, j.task.Snapshot().Message)
	}
}
func basePlan(name string) installPlan {
	return installPlan{App: models.StoreApp{Name: name, URL: "https://packages.example/app.bin", Cmd: "echo ready"}}
}

func TestInstallCollisionPreservesExistingFiles(t *testing.T) {
	setupInstallTest(t)
	target := filepath.Join(config.AppsDir, "demo-app")
	os.MkdirAll(target, 0700)
	marker := filepath.Join(target, "important.txt")
	os.WriteFile(marker, []byte("keep"), 0600)
	WriteJSON(config.ConfigApps, map[string]models.Project{"demo-app": {Path: target}})
	if _, err := submitInstall(basePlan("demo app"), nil); err == nil {
		t.Fatal("accepted occupied directory")
	}
	if b, err := os.ReadFile(marker); err != nil || string(b) != "keep" {
		t.Fatal("existing file changed")
	}
}
func TestInstallRejectsNullConfigAndReleasesLock(t *testing.T) {
	setupInstallTest(t)
	os.WriteFile(config.ConfigApps, []byte("null"), 0600)
	if _, err := submitInstall(basePlan("demo"), nil); err == nil {
		t.Fatal("accepted null config")
	}
	if !appOpMu.TryLock() {
		t.Fatal("mutex leaked")
	}
	appOpMu.Unlock()
}
func TestInstallRejectsUnsafeNamesAndPrivateURL(t *testing.T) {
	setupInstallTest(t)
	for _, name := range []string{"../outside", "..\\outside", "a/b", "a:b", ""} {
		if _, err := submitInstall(basePlan(name), nil); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	p := basePlan("demo")
	p.App.URL = "http://127.0.0.1/file"
	if _, err := submitInstall(p, nil); err == nil {
		t.Fatal("accepted private URL")
	}
}
func TestInstallCancelAndPauseWaitForWorker(t *testing.T) {
	for _, action := range []string{"delete", "pause"} {
		t.Run(action, func(t *testing.T) {
			setupInstallTest(t)
			entered := make(chan struct{})
			calls := 0
			mockRemote(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					close(entered)
					<-r.Context().Done()
					return nil, r.Context().Err()
				}
				return response(r, "complete package"), nil
			})
			j, err := submitInstall(basePlan("demo"), nil)
			if err != nil {
				t.Fatal(err)
			}
			<-entered
			if _, err := submitInstall(basePlan("demo"), nil); err == nil {
				t.Fatal("duplicate reservation accepted")
			}
			if err := j.action(action); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(j.target); !os.IsNotExist(err) {
				t.Fatal("cancelled worker committed directory")
			}
			apps, _ := loadAppsStrict()
			if len(apps) != 0 {
				t.Fatal("cancelled worker created application")
			}
			if action == "pause" {
				if err := j.action("resume"); err != nil {
					t.Fatal(err)
				}
				waitInstall(t, j, "completed")
				if calls != 2 {
					t.Fatalf("unexpected workers: %d", calls)
				}
			}
		})
	}
}
func TestInstallErrorsNeverCommit(t *testing.T) {
	for _, kind := range []string{"http", "limit", "archive", "short"} {
		t.Run(kind, func(t *testing.T) {
			setupInstallTest(t)
			old := config.MaxDownBytes
			config.MaxDownBytes = 16
			t.Cleanup(func() { config.MaxDownBytes = old })
			mockRemote(t, func(r *http.Request) (*http.Response, error) {
				resp := response(r, "not an archive")
				switch kind {
				case "http":
					resp.StatusCode = 404
				case "limit":
					resp = response(r, strings.Repeat("x", 32))
					resp.ContentLength = -1
				case "short":
					resp.ContentLength = 16
				}
				return resp, nil
			})
			p := basePlan("demo")
			p.App.AutoExtract = kind == "archive"
			j, err := submitInstall(p, nil)
			if err != nil {
				t.Fatal(err)
			}
			waitInstall(t, j, "error")
			apps, _ := loadAppsStrict()
			if len(apps) != 0 {
				t.Fatal("failed install committed config")
			}
			if _, err := os.Stat(j.target); !os.IsNotExist(err) {
				t.Fatal("failed install retained final directory")
			}
			entries, _ := os.ReadDir(config.AppsDir)
			if len(entries) != 0 {
				t.Fatal("staging leaked")
			}
		})
	}
}
func TestInstallMetadataAndAtomicCommit(t *testing.T) {
	setupInstallTest(t)
	mockRemote(t, func(r *http.Request) (*http.Response, error) { return response(r, "package"), nil })
	p := basePlan("demo app")
	p.Source = "https://store.example/manifest.json"
	p.App.SetupCmd = "echo setup"
	p.App.WorkDir = "sub"
	p.App.Version = "2"
	p.App.Port = 8080
	p.App.Icon = "https://example.com/icon.png"
	j, err := submitInstall(p, nil)
	if err != nil {
		t.Fatal(err)
	}
	waitInstall(t, j, "completed")
	apps, _ := loadAppsStrict()
	a := apps["demo-app"]
	if a.SourceURL != p.Source || a.PackageURL != p.App.URL || a.SetupCmd != p.App.SetupCmd || a.WorkDir != filepath.Join(j.target, "sub") || a.Port != 8080 || a.Version != "2" {
		t.Fatalf("lost metadata: %+v", a)
	}
	if err := j.action("delete"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(j.target); err != nil {
		t.Fatal("deleting completed record removed application")
	}
}
func TestPublicClientChecksRedirectAndAddresses(t *testing.T) {
	mockRemote(t, func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"http://127.0.0.1/private"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	})
	if _, err := remoteClient(time.Second).Get("https://public.example/file"); err == nil {
		t.Fatal("followed private redirect")
	}
	for _, raw := range []string{"http://localhost/x", "http://[::]/x", "ftp://example.com/a", "http://user:secret@example.com/a"} {
		if validateRemoteURL(raw) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if conn, err := publicDial(ctx, "tcp", "127.0.0.1:1"); err == nil {
		conn.Close()
		t.Fatal("dialed nonpublic IP")
	}
}
func TestStoreCacheAndManifestBinding(t *testing.T) {
	old := storeCache
	t.Cleanup(func() { storeCache = old })
	storeCache.lastInit = time.Time{}
	mockRemote(t, func(r *http.Request) (*http.Response, error) {
		return response(r, `[{"name":"`+r.URL.Hostname()+`"}]`), nil
	})
	srcs := []models.StoreSource{{URL: "https://a.example/list"}, {URL: "https://b.example/list"}}
	a, _, _ := getStoreApps(srcs, 0)
	b, _, _ := getStoreApps(srcs, 1)
	if a[0].Name == b[0].Name {
		t.Fatal("cache reused wrong source")
	}
	if storeDigest(srcs[0].URL, a[0]) == storeDigest(srcs[1].URL, a[0]) {
		t.Fatal("digest does not bind source")
	}
	changed := a[0]
	changed.Cmd = "echo changed"
	if storeDigest(srcs[0].URL, a[0]) == storeDigest(srcs[0].URL, changed) {
		t.Fatal("digest does not bind command")
	}
}
func TestRestoreExcludesNormalizedSessionPaths(t *testing.T) {
	for _, name := range []string{"config/./sessions/token.json", "apps/../config/sessions/token.json", "data/config//sessions/token.json"} {
		var b bytes.Buffer
		tw := tar.NewWriter(&b)
		tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: 2})
		tw.Write([]byte("{}"))
		tw.Close()
		stage := t.TempDir()
		if err := stageRestore(&b, stage); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(stage, "config", "sessions", "token.json")); !os.IsNotExist(err) {
			t.Fatalf("restored %s", name)
		}
	}
}
func TestPasswordChangeInvalidatesEverySession(t *testing.T) {
	setupTestAccount(t)
	oldDir := sessionDir
	sessionDir = t.TempDir()
	t.Cleanup(func() { sessionDir = oldDir })
	w := httptest.NewRecorder()
	first := createSession(w, false)
	second := createSession(httptest.NewRecorder(), false)
	r := httptest.NewRequest("POST", "/setting/save-account", strings.NewReader(url.Values{"old": {"old-password-long"}, "new": {"a-new-long-password"}}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(&http.Cookie{Name: "lp_session", Value: first})
	out := httptest.NewRecorder()
	saveAccount(out, r)
	if isValidSession(first) || isValidSession(second) {
		t.Fatal("old session remains valid")
	}
	var user models.UserConfig
	LoadJSON(config.ConfigUsr, &user)
	if !strings.HasPrefix(user.Password, "pbkdf2-sha256$") || !passwordMatches(user.Password, "a-new-long-password") {
		t.Fatal("new password not saved securely")
	}
	if passwordMatches(user.Password, "wrong") {
		t.Fatal("wrong password accepted")
	}
}
func TestLogoutAnonymousAndSessionWriteFailure(t *testing.T) {
	out := httptest.NewRecorder()
	logoutHandler(out, httptest.NewRequest("GET", "/logout", nil))
	if out.Code != 302 {
		t.Fatal(out.Code)
	}
	setupTestAccount(t)
	old := sessionDir
	sessionDir = filepath.Join(t.TempDir(), "file")
	os.WriteFile(sessionDir, []byte("x"), 0600)
	defer func() { sessionDir = old }()
	out = httptest.NewRecorder()
	if createSession(out, false) != "" || len(out.Result().Cookies()) > 0 {
		t.Fatal("reported session despite storage failure")
	}
}
func TestCustomPageBoundaryAndSandbox(t *testing.T) {
	setupInstallTest(t)
	oldSessions := sessionDir
	sessionDir = t.TempDir()
	defer func() { sessionDir = oldSessions }()
	setupTestAccount(t)
	os.MkdirAll(filepath.Join(config.DataDir, "pages"), 0700)
	os.MkdirAll(filepath.Join(config.DataDir, "pages-other"), 0700)
	os.WriteFile(filepath.Join(config.DataDir, "pages", "demo.html"), []byte("<script>/* fixture */</script>"), 0600)
	os.WriteFile(filepath.Join(config.DataDir, "pages-other", "keep.html"), []byte("keep"), 0600)
	for _, name := range []string{"../pages-other/keep.html", "..\\pages-other\\keep.html"} {
		out := httptest.NewRecorder()
		serveCustomPage(out, httptest.NewRequest("GET", "/page/"+url.PathEscape(name), nil))
		if out.Code == 200 {
			t.Fatal("page escaped")
		}
		deletePage(httptest.NewRecorder(), httptest.NewRequest("POST", "/page/del/"+url.PathEscape(name), nil))
	}
	if _, err := os.Stat(filepath.Join(config.DataDir, "pages-other", "keep.html")); err != nil {
		t.Fatal("deleted sibling")
	}
	out := httptest.NewRecorder()
	serveCustomPage(out, httptest.NewRequest("GET", "/page/demo.html", nil))
	if out.Code != 200 || !strings.Contains(out.Header().Get("Content-Security-Policy"), "sandbox allow-scripts;") {
		t.Fatal("missing opaque-origin sandbox")
	}
}
func TestAnalyzerNoHTMLSinkAndNoTemporaryScript(t *testing.T) {
	setupInstallTest(t)
	mockRemote(t, func(r *http.Request) (*http.Response, error) { return response(r, "apt install <test-markup>"), nil })
	out := httptest.NewRecorder()
	analyzeScriptHandler(out, httptest.NewRequest("GET", "/api/analyze?url=https://scripts.example/a.sh", nil))
	var info ScriptInfo
	if err := json.Unmarshal(out.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(config.DataDir, "scripts")); !os.IsNotExist(err) {
		t.Fatal("analyzer wrote script to disk")
	}
	b, err := os.ReadFile("web/templates/analyze.html")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), ".innerHTML") || !strings.Contains(string(b), "tag.textContent=String(value)") {
		t.Fatal("analyzer does not use text rendering")
	}
}
func TestInstallCommitConfigFailureLeavesNoApp(t *testing.T) {
	setupInstallTest(t)
	entered, release := make(chan struct{}), make(chan struct{})
	mockRemote(t, func(r *http.Request) (*http.Response, error) {
		close(entered)
		<-release
		return response(r, "package"), nil
	})
	j, err := submitInstall(basePlan("demo"), nil)
	if err != nil {
		t.Fatal(err)
	}
	<-entered
	os.WriteFile(config.ConfigApps, []byte("null"), 0600)
	close(release)
	waitInstall(t, j, "error")
	if !appOpMu.TryLock() {
		t.Fatal("commit leaked mutex")
	}
	appOpMu.Unlock()
	if _, err := os.Stat(j.target); !os.IsNotExist(err) {
		t.Fatal(fmt.Sprintf("failed commit retained target: %v", err))
	}
}

func TestSecureSessionAndFirstLoginGate(t *testing.T) {
	setupTestAccount(t)
	old := sessionDir
	sessionDir = t.TempDir()
	defer func() { sessionDir = old }()
	w := httptest.NewRecorder()
	token := createSession(w, true, true)
	for _, cookie := range w.Result().Cookies() {
		if !cookie.Secure {
			t.Fatal("TLS session cookie missing Secure")
		}
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "lp_session", Value: token})
	out := httptest.NewRecorder()
	called := false
	auth(func(http.ResponseWriter, *http.Request) { called = true })(out, r)
	if called || out.Code != 303 {
		t.Fatal("first login allowed management page")
	}
}

func TestProcessIdentityRejectsMismatchedStart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pid.pid")
	if err := writeProcessIdentity(path, os.Getpid()); err != nil {
		t.Fatal(err)
	}
	app := models.Project{Path: dir}
	if !managedIdentityMatches(os.Getpid(), &app) {
		t.Fatal("current process identity rejected")
	}
	id := readProcessIdentity(&app)
	id.Started--
	WriteJSON(path, id)
	if managedIdentityMatches(os.Getpid(), &app) {
		t.Fatal("mismatched start time accepted")
	}
	os.WriteFile(path, []byte(fmt.Sprint(os.Getpid())), 0600)
	if managedIdentityMatches(os.Getpid(), &app) {
		t.Fatal("legacy unverified PID accepted")
	}
}

func TestInitialAccountUsesUniqueCredential(t *testing.T) {
	setupInstallTest(t)
	oldUsr, oldSrc, oldBg, oldSession := config.ConfigUsr, config.ConfigSrc, config.ConfigBg, sessionDir
	config.ConfigUsr = filepath.Join(config.ConfigDir, "user.json")
	config.ConfigSrc = filepath.Join(config.ConfigDir, "sources.json")
	config.ConfigBg = filepath.Join(config.ConfigDir, "bg.json")
	t.Cleanup(func() {
		config.ConfigUsr, config.ConfigSrc, config.ConfigBg, sessionDir = oldUsr, oldSrc, oldBg, oldSession
	})
	t.Setenv("LIGHTPANEL_INITIAL_PASSWORD", "")
	if err := InitConfig(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(config.ConfigDir, "initial-password.txt"))
	if err != nil {
		t.Fatal(err)
	}
	password := strings.TrimSpace(string(b))
	var user models.UserConfig
	LoadJSON(config.ConfigUsr, &user)
	if len(password) != 24 || !user.MustChangePassword || !passwordMatches(user.Password, password) || passwordMatches(user.Password, "admin") {
		t.Fatal("initial account was not hardened")
	}
}
