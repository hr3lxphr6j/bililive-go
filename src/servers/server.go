package servers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/gorilla/mux"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
)

type Server struct {
	server *http.Server
}

var authorization string

func initMux(ctx context.Context) *mux.Router {
	m := mux.NewRouter()
	m.Use(
		// log
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				instance.GetInstance(ctx).Logger.WithFields(map[string]interface{}{
					"Method":     r.Method,
					"Path":       r.RequestURI,
					"RemoteAddr": r.RemoteAddr,
				}).Debug("Http Request")
				handler.ServeHTTP(w, r)
			})
		},

		// context
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), instance.InstanceKey, instance.GetInstance(ctx))))
			})
		},

		// CORS
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodOptions {
					w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
					w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT")
					w.Header().Add("Access-Control-Allow-Headers", "Authorization")
					w.Write(nil)
				} else {
					w.Header().Add("Access-Control-Allow-Origin", "*")
					handler.ServeHTTP(w, r)
				}
			})
		},

		// Content-Type
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if prefix := strings.Split(r.RequestURI, "/")[1]; prefix != "files" && prefix != "debug" {
					w.Header().Add("Content-Type", "application/json")
				}
				handler.ServeHTTP(w, r)
			})
		},

		// token verify
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token := instance.GetInstance(ctx).Config.RPC.Token
				if authorization == "" {
					authorization = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("token:%s", token))))
				}
				if token == "" || r.Header.Get("authorization") == authorization || r.FormValue("token") == token {
					handler.ServeHTTP(w, r)
				} else {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"err_no":403,"err_msg":"the token is incorrect","data":null}`))
				}
			})
		})
	// api
	m.HandleFunc("/info", getInfo).Methods("GET", "OPTIONS")
	m.HandleFunc("/config", getConfig).Methods("GET", "OPTIONS")
	m.HandleFunc("/config", putConfig).Methods("PUT", "OPTIONS")
	m.HandleFunc("/lives", getAllLives).Methods("GET", "OPTIONS")
	m.HandleFunc("/lives", addLives).Methods("POST", "OPTIONS")
	m.HandleFunc("/lives/{id}", getLive).Methods("GET", "OPTIONS")
	m.HandleFunc("/lives/{id}", removeLive).Methods("DELETE", "OPTIONS")
	m.HandleFunc("/lives/{id}/{action}", parseLiveAction).Methods("GET", "OPTIONS")

	// file server
	m.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir(instance.GetInstance(ctx).Config.OutPutPath))))

	// pprof
	if instance.GetInstance(ctx).Config.Debug {
		m.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	}
	return m
}

func NewServer(ctx context.Context) *Server {
	inst := instance.GetInstance(ctx)
	config := inst.Config
	httpServer := &http.Server{
		Addr:    config.RPC.Port,
		Handler: initMux(ctx),
	}
	server := &Server{server: httpServer}
	inst.Server = server
	return server
}

func (s *Server) Start(ctx context.Context) error {
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Add(1)
	config := inst.Config
	go func() {
		if config.RPC.TLS.Enable {
			if err := s.server.ListenAndServeTLS(config.RPC.TLS.CertFile, config.RPC.TLS.KeyFile); err != nil {
				if err != http.ErrServerClosed {
					inst.Logger.Error(err)
				}
			}
		} else {
			if err := s.server.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					inst.Logger.Error(err)
				}
			}
		}
	}()
	inst.Logger.Infof("Server start at %s", s.server.Addr)
	return nil
}

func (s *Server) Close(ctx context.Context) {
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Done()
	ctx2, cancel := context.WithCancel(ctx)
	s.server.Shutdown(ctx2)
	defer cancel()
	inst.Logger.Infof("Server close")
}
