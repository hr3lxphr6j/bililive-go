package servers

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"net/http"
	"strings"
)

type Server struct {
	server *http.Server
}

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
					"Token":      r.Header.Get("token"),
				}).Debug("Http Request")
				handler.ServeHTTP(w, r)
			})
		},
		// token verify
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("token") == instance.GetInstance(ctx).Config.RPC.Token {
					handler.ServeHTTP(w, r)
				} else {
					http.Error(w, "Forbidden", http.StatusForbidden)
				}
			})
		},
		// context
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), instance.InstanceKey, instance.GetInstance(ctx))))
			})
		},
		// Content-Type
		func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Split(r.RequestURI, "/")[1] != "files" {
					w.Header().Add("Content-Type", "application/json")
				}
				handler.ServeHTTP(w, r)
			})
		})
	m.HandleFunc("/config", getConfig).Methods("GET")
	m.HandleFunc("/config", putConfig).Methods("PUT")
	m.HandleFunc("/lives", getAllLives).Methods("GET")
	m.HandleFunc("/lives", addLives).Methods("POST")
	m.HandleFunc("/lives/{id}", getLive).Methods("GET")
	m.HandleFunc("/lives/{id}", removeLive).Methods("DELETE")
	m.HandleFunc("/lives/{id}/{action}", parseLiveAction).Methods("GET")
	m.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir(instance.GetInstance(ctx).Config.OutPutPath))))
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
				inst.Logger.Error(err)
			}
		} else {
			if err := s.server.ListenAndServe(); err != nil {
				inst.Logger.Error(err)
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
