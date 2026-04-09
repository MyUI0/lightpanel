package handlers

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lightpanel/config"
)

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func SetupRoutes() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","version":"` + config.Version + `"}`))
	})

	mux.HandleFunc("/login", loginPage)
	mux.HandleFunc("/login/auth", requirePOST(loginHandler))
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/", auth(indexPage))
	mux.HandleFunc("/apps", auth(appsPage))
	mux.HandleFunc("/system", auth(systemPage))
	mux.HandleFunc("/store", auth(storePage))
	mux.HandleFunc("/source", auth(sourcePage))
	mux.HandleFunc("/source/add", authWithCSRF(addSource))
	mux.HandleFunc("/source/del/", authWithCSRF(delSource))
	mux.HandleFunc("/source/edit/", authWithCSRF(editSource))
	mux.HandleFunc("/source/test/", authWithCSRF(testSource))
	mux.HandleFunc("/downloads", auth(downloadPage))
	mux.HandleFunc("/dl/api", auth(apiDownloads))
	mux.HandleFunc("/dl/action/", authWithCSRF(apiDownloadAction))
	mux.HandleFunc("/setting", auth(settingPage))
	mux.HandleFunc("/setting/save-account", authWithCSRF(saveAccount))
	mux.HandleFunc("/setting/save-bg", authWithCSRF(saveBg))
	mux.HandleFunc("/setting/save-logo", authWithCSRF(saveLogo))
	mux.HandleFunc("/setting/save-bark", authWithCSRF(saveBark))
	mux.HandleFunc("/setting/backup", auth(backupData))
	mux.HandleFunc("/setting/restore", authWithCSRF(restoreData))
	mux.HandleFunc("/setting/upload-page", authWithCSRF(uploadPage))
	mux.HandleFunc("/page/", auth(serveCustomPage))
	mux.HandleFunc("/page/del/", authWithCSRF(deletePage))
	mux.HandleFunc("/create", authWithCSRF(createApp))
	mux.HandleFunc("/create/manual", authWithCSRF(createManualApp))
	mux.HandleFunc("/create/progress/", auth(apiCreateProgress))
	mux.HandleFunc("/start/", authWithCSRF(startApp))
	mux.HandleFunc("/stop/", authWithCSRF(stopApp))
	mux.HandleFunc("/restart/", authWithCSRF(restartApp))
	mux.HandleFunc("/edit/", authWithCSRF(editAppHandler))
	mux.HandleFunc("/detect/", authWithCSRF(detectApp))
	mux.HandleFunc("/toggle-auto/", authWithCSRF(toggleAutoStart))
	mux.HandleFunc("/log/", auth(logPage))
	mux.HandleFunc("/log/clear/", authWithCSRF(clearLog))
	mux.HandleFunc("/delete/", authWithCSRF(deleteApp))
	mux.HandleFunc("/install/", authWithCSRF(startStoreInstall))
	mux.HandleFunc("/install/params/", auth(storeParamsPage))
	mux.HandleFunc("/install/confirm/", authWithCSRF(confirmInstallWithParams))
	mux.HandleFunc("/kill/", authWithCSRF(killSystemProc))
	mux.HandleFunc("/analyze", auth(scriptAnalyzePage))
	mux.HandleFunc("/api/analyze", auth(analyzeScriptHandler))
	mux.HandleFunc("/api/updates", auth(apiCheckUpdates))

	http.Handle("/", mux)
}

func ListenAndServe() error {
	go func() {
		time.Sleep(500 * time.Millisecond)
		autoStartApps()
	}()

	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: securityHeaders(http.DefaultServeMux),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	return srv.ListenAndServe()
}
