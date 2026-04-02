package handlers

import (
	"net/http"

	"lightpanel/config"
	"lightpanel/models"
)

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var usr models.UserConfig
		_ = LoadJSON(config.ConfigUsr, &usr)
		user, pwd, ok := r.BasicAuth()
		if !ok || user != usr.Username || pwd != usr.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="LightPanel"`)
			w.WriteHeader(401)
			_, _ = w.Write([]byte("请登录"))
			return
		}
		next(w, r)
	}
}

func requirePOST(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			_, _ = w.Write([]byte("Method Not Allowed"))
			return
		}
		next(w, r)
	}
}
