package handlers

import (
	"net/http"

	"lightpanel/config"
)

func SetupRoutes() {
	http.HandleFunc("/", auth(indexPage))
	http.HandleFunc("/system", auth(systemPage))
	http.HandleFunc("/store", auth(storePage))
	http.HandleFunc("/source", auth(sourcePage))
	http.HandleFunc("/source/add", auth(requirePOST(addSource)))
	http.HandleFunc("/source/del/", auth(requirePOST(delSource)))
	http.HandleFunc("/downloads", auth(downloadPage))
	http.HandleFunc("/dl/api", auth(apiDownloads))
	http.HandleFunc("/dl/action/", auth(requirePOST(apiDownloadAction)))
	http.HandleFunc("/setting", auth(settingPage))
	http.HandleFunc("/setting/save", auth(saveSetting))
	http.HandleFunc("/create", auth(createApp))
	http.HandleFunc("/start/", auth(requirePOST(startApp)))
	http.HandleFunc("/stop/", auth(requirePOST(stopApp)))
	http.HandleFunc("/restart/", auth(requirePOST(restartApp)))
	http.HandleFunc("/edit/", auth(editAppHandler))
	http.HandleFunc("/toggle-auto/", auth(requirePOST(toggleAutoStart)))
	http.HandleFunc("/log/", auth(logPage))
	http.HandleFunc("/log/clear/", auth(requirePOST(clearLog)))
	http.HandleFunc("/delete/", auth(requirePOST(deleteApp)))
	http.HandleFunc("/install/", auth(requirePOST(startStoreInstall)))
	http.HandleFunc("/kill/", auth(requirePOST(killSystemProc)))
}

func ListenAndServe() error {
	go autoStartApps()
	return http.ListenAndServe(":"+config.Port, nil)
}
