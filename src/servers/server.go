package servers

import (
	"net/http"
)

type Server interface {
}

type HttpServer struct {
	server *http.Server
}
