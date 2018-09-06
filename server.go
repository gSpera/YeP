package main

import "net/http"

//Server is a YeP server
//Implements http.Handler
type Server struct {
	db  Database
	mux *http.ServeMux
	cfg config
}

//NewServer creates a new server
func NewServer(db Database, cfg config) Server {
	s := Server{
		db:  db,
		mux: http.NewServeMux(),
		cfg: cfg,
	}
	return s
}

func (s *Server) handleRoute(pattern string, r Route) {
	s.mux.HandleFunc(pattern, routeToHandler(r, s))
}

//Route is a Server Route
type Route func(s Server, w http.ResponseWriter, req *http.Request)

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(w, req)
}
