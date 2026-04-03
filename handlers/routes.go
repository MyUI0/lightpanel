package handlers

import (
	"net/http"
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

	mux.HandleFunc("/login", loginPage)
	mux.HandleFunc("/login/auth", requirePOST(loginHandler))
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/", auth(indexPage))
	mux.HandleFunc("/system", auth(systemPage))
	mux.HandleFunc("/store", auth(storePage))
	mux.HandleFunc("/source", auth(sourcePage))
	mux.HandleFunc("/source/add", authWithCSRF(addSource))
	mux.HandleFunc("/source/del/", authWithCSRF(delSource))
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
	mux.HandleFunc("/page/", auth(serveCustomPage))
	mux.HandleFunc("/create", authWithCSRF(createApp))
	mux.HandleFunc("/create/manual", authWithCSRF(createManualApp))
	mux.HandleFunc("/create/progress/", auth(apiCreateProgress))
	mux.HandleFunc("/start/", authWithCSRF(startApp))
	mux.HandleFunc("/stop/", authWithCSRF(stopApp))
	mux.HandleFunc("/restart/", authWithCSRF(restartApp))
	mux.HandleFunc("/edit/", auth(editAppHandler))
	mux.HandleFunc("/detect/", auth(detectApp))
	mux.HandleFunc("/toggle-auto/", authWithCSRF(toggleAutoStart))
	mux.HandleFunc("/log/", auth(logPage))
	mux.HandleFunc("/log/clear/", authWithCSRF(clearLog))
	mux.HandleFunc("/delete/", authWithCSRF(deleteApp))
	mux.HandleFunc("/install/", authWithCSRF(startStoreInstall))
	mux.HandleFunc("/kill/", authWithCSRF(killSystemProc))
	mux.HandleFunc("/analyze", auth(scriptAnalyzePage))
	mux.HandleFunc("/api/analyze", auth(analyzeScriptHandler))

	http.Handle("/", securityHeaders(mux))
}

func ListenAndServe() error {
	go func() {
		time.Sleep(500 * time.Millisecond)
		autoStartApps()
	}()
	return http.ListenAndServe(":"+config.Port, nil)
}
