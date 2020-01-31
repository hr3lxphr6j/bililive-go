package servers

import (
	"net/http"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
)

func log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instance.GetInstance(r.Context()).Logger.WithFields(map[string]interface{}{
			"Method":     r.Method,
			"Path":       r.RequestURI,
			"RemoteAddr": r.RemoteAddr,
		}).Debug("Http Request")
		handler.ServeHTTP(w, r)
	})
}

func cors(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodOptions:
			w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT")
			w.Header().Add("Access-Control-Allow-Headers", "Authorization")
			w.Write(nil)
		default:
			w.Header().Add("Access-Control-Allow-Origin", "*")
			handler.ServeHTTP(w, r)
		}
	})
}
