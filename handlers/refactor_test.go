package handlers

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"lightpanel/config"
	"lightpanel/models"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestFormatSizePreservesFraction(t *testing.T) {
	if got := formatSize(1536); got != "1.5 KB" {
		t.Fatalf("got %q", got)
	}
}

func TestConcurrentJSONWritesRemainValid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := WriteJSON(path, map[string]int{"value": i}); err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	var got map[string]int
	if err := LoadJSON(path, &got); err != nil {
		t.Fatal(err)
	}
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".json-*"))
	if len(leftovers) > 0 {
		t.Fatalf("temporary files leaked: %v", leftovers)
	}
}

func TestArchiveRejectsEscapes(t *testing.T) {
	root := filepath.Join(t.TempDir(), "apps")
	os.MkdirAll(root, 0755)
	for _, name := range []string{"../apps-other/pwn", "../pwn", "/tmp/pwn", "C:/pwn", "..\\pwn"} {
		if _, err := archiveTarget(root, name); err == nil {
			t.Errorf("accepted %q", name)
		}
		var b bytes.Buffer
		tw := tar.NewWriter(&b)
		tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: 1})
		tw.Write([]byte("x"))
		tw.Close()
		if extractTarReader(&b, root) {
			t.Errorf("tar accepted %q", name)
		}
	}
	if _, err := archiveTarget(root, "nested/app"); err != nil {
		t.Fatal(err)
	}
}

func TestZipRejectsPrefixSibling(t *testing.T) {
	dir := t.TempDir()
	var data bytes.Buffer
	zw := zip.NewWriter(&data)
	f, _ := zw.Create("../apps-evil/file")
	f.Write([]byte("x"))
	zw.Close()
	archive := filepath.Join(dir, "archive.zip")
	os.WriteFile(archive, data.Bytes(), 0600)
	if extractZip(archive, filepath.Join(dir, "apps")) {
		t.Fatal("zip escaped destination")
	}
}

func TestArchiveRejectsExistingSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip("symlinks unavailable: " + err.Error())
	}
	if _, err := archiveTarget(root, "link/file"); err == nil {
		t.Fatal("accepted symlink ancestor")
	}
}

func TestBackupRoundTripAndLegacyPrefix(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "config", "sessions"), 0700)
	os.MkdirAll(filepath.Join(root, "apps", "demo"), 0755)
	os.WriteFile(filepath.Join(root, "config", "apps.json"), []byte(`{"demo":{"cmd":"./run"}}`), 0600)
	os.WriteFile(filepath.Join(root, "config", "sessions", "secret.json"), []byte(`{}`), 0600)
	os.WriteFile(filepath.Join(root, "apps", "demo", "run"), []byte("hello"), 0755)
	os.WriteFile(filepath.Join(root, "apps", "demo", "pid.pid"), []byte("123"), 0600)
	var buffer bytes.Buffer
	if err := writeBackup(&buffer, root); err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(&buffer)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	stage := t.TempDir()
	if err := stageRestore(gz, stage); err != nil {
		t.Fatal(err)
	}
	value, err := os.ReadFile(filepath.Join(stage, "apps", "demo", "run"))
	if err != nil || string(value) != "hello" {
		t.Fatalf("missing restored binary: %s %v", value, err)
	}
	for _, name := range []string{"config/sessions/secret.json", "apps/demo/pid.pid", "data"} {
		if _, err := os.Stat(filepath.Join(stage, name)); !os.IsNotExist(err) {
			t.Fatalf("unexpected backup entry %s", name)
		}
	}
	var legacy bytes.Buffer
	tw := tar.NewWriter(&legacy)
	payload := []byte(`{"legacy":{"cmd":"./run"}}`)
	tw.WriteHeader(&tar.Header{Name: "data/config/apps.json", Mode: 0600, Size: int64(len(payload))})
	tw.Write(payload)
	tw.Close()
	stage = t.TempDir()
	if err := stageRestore(&legacy, stage); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(stage, "config", "apps.json")); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreRejectsInvalidConfig(t *testing.T) {
	var b bytes.Buffer
	tw := tar.NewWriter(&b)
	tw.WriteHeader(&tar.Header{Name: "config/apps.json", Size: 1, Mode: 0600})
	tw.Write([]byte("{"))
	tw.Close()
	if err := stageRestore(&b, t.TempDir()); err == nil {
		t.Fatal("invalid JSON accepted")
	}
}

func TestAllTemplatesRender(t *testing.T) {
	InitTemplates()
	data := map[string]any{"Apps": map[string]models.Project{"a\"<b": {Cmd: "./run --x <value>", Status: "已停止"}}, "Cpu": 1, "Mem": 2, "Disk": 3, "ProcNum": 4, "LockTime": 0, "Sources": []models.StoreSource{}, "Tasks": []models.DownloadTask{}, "History": []models.DownloadHistory{}, "Procs": []models.ProcInfo{}, "Pages": []string{}, "Name": "a\"<b", "App": models.StoreApp{}, "CSRFToken": "test"}
	for _, name := range []string{"index", "login", "store", "store_params", "source", "system", "log_list", "log", "setting", "edit", "downloads", "analyze"} {
		t.Run(name, func(t *testing.T) {
			local := make(map[string]any)
			for k, v := range data {
				local[k] = v
			}
			if name == "store" {
				local["Deployed"] = map[string]bool{}
				local["Apps"] = []models.StoreApp{{Name: "Example", Desc: "Demo"}}
			}
			var out bytes.Buffer
			if err := htmlRender.ExecuteTemplate(&out, name, local); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "ZgotmplZ") {
				t.Fatal("template contextual escaping failure")
			}
		})
	}
}

func TestCSRFRequiresExplicitTokenAndPost(t *testing.T) {
	setupTestAccount(t)
	old := sessionDir
	sessionDir = t.TempDir()
	defer func() { sessionDir = old }()
	record := httptest.NewRecorder()
	createSession(record, false)
	cookies := record.Result().Cookies()
	token := ""
	for _, cookie := range cookies {
		if cookie.Name == "lp_csrf" {
			token = cookie.Value
		}
	}
	handler := authWithCSRF(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	for _, tc := range []struct {
		method   string
		explicit bool
		want     int
	}{{"POST", false, 403}, {"POST", true, 204}, {"GET", true, 405}} {
		req := httptest.NewRequest(tc.method, "/start/demo", nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		if tc.explicit {
			req.Header.Set("X-CSRF-Token", token)
		}
		out := httptest.NewRecorder()
		handler(out, req)
		if out.Code != tc.want {
			t.Errorf("%s explicit=%v: %d != %d", tc.method, tc.explicit, out.Code, tc.want)
		}
	}
}

func TestForwardedIPRequiresTrustedProxy(t *testing.T) {
	t.Setenv("LIGHTPANEL_TRUST_PROXY", "")
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.1.2.3:42"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	if getClientIP(req) != "10.1.2.3" {
		t.Fatal("trusted spoofed forwarded IP")
	}
}

func TestMissingAppActionReportsFailure(t *testing.T) {
	old := config.ConfigApps
	config.ConfigApps = filepath.Join(t.TempDir(), "apps.json")
	defer func() { config.ConfigApps = old }()
	WriteJSON(config.ConfigApps, map[string]models.Project{})
	req := httptest.NewRequest("POST", "/start/missing", nil)
	req.Header.Set("Accept", "application/json")
	out := httptest.NewRecorder()
	startApp(out, req)
	if out.Code == 200 {
		t.Fatal("missing application reported success")
	}
	var result map[string]any
	if json.Unmarshal(out.Body.Bytes(), &result) != nil || result["ok"] != false {
		t.Fatal(out.Body.String())
	}
}

func TestStaticAssetsEmbedded(t *testing.T) {
	for _, name := range []string{"panel.css", "shell.js", "apps.js", "create.js"} {
		r := httptest.NewRequest("GET", "/static/"+name, nil)
		w := httptest.NewRecorder()
		staticHandler().ServeHTTP(w, r)
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Fatalf("missing %s", name)
		}
	}
}
