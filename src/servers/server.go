package servers

//go:generate statik -f -src=../webapp/build/

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"

	"github.com/hr3lxphr6j/bililive-go/src/instance"
	_ "github.com/hr3lxphr6j/bililive-go/src/servers/statik"
)

const (
	apiRouterPrefix = "/api"
)

type Server struct {
	server *http.Server
}

func initMux(ctx context.Context) *mux.Router {
	m := mux.NewRouter()
	m.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w,
				r.WithContext(
					context.WithValue(
						r.Context(),
						instance.Key,
						instance.GetInstance(ctx),
					),
				),
			)
		})
	}, log)

	// api router
	apiRoute := m.PathPrefix(apiRouterPrefix).Subrouter()
	apiRoute.Use(mux.CORSMethodMiddleware(apiRoute))
	apiRoute.HandleFunc("/info", getInfo).Methods("GET")
	apiRoute.HandleFunc("/config", getConfig).Methods("GET")
	apiRoute.HandleFunc("/config", putConfig).Methods("PUT")
	apiRoute.HandleFunc("/lives", getAllLives).Methods("GET")
	apiRoute.HandleFunc("/lives", addLives).Methods("POST")
	apiRoute.HandleFunc("/lives/{id}", getLive).Methods("GET")
	apiRoute.HandleFunc("/lives/{id}", removeLive).Methods("DELETE")
	apiRoute.HandleFunc("/lives/{id}/{action}", parseLiveAction).Methods("GET")

	statikFS, err := fs.New()
	if err != nil {
		instance.GetInstance(ctx).Logger.Fatal(err)
	}
	m.PathPrefix("/").Handler(http.FileServer(statikFS))

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
		Addr:    config.RPC.Bind,
		Handler: initMux(ctx),
	}
	server := &Server{server: httpServer}
	inst.Server = server
	return server
}

func (s *Server) Start(ctx context.Context) error {
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Add(1)
	go func() {
		switch err := s.server.ListenAndServe(); err {
		case nil, http.ErrServerClosed:
		default:
			inst.Logger.Error(err)
		}
	}()
	inst.Logger.Infof("Server start at %s", s.server.Addr)
	return nil
}

func (s *Server) Close(ctx context.Context) {
	inst := instance.GetInstance(ctx)
	inst.WaitGroup.Done()
	ctx2, cancel := context.WithCancel(ctx)
	if err := s.server.Shutdown(ctx2); err != nil {
		inst.Logger.WithError(err).Error("failed to shutdown server")
	}
	defer cancel()
	inst.Logger.Infof("Server close")
}
