package web

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	router *mux.Router
}

func NewServer() *Server {
	s := &Server{
		router: mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/solve", s.handleSolve).Methods("POST")
	api.HandleFunc("/exec", s.handleExec).Methods("POST")
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Static files
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./internal/web/static/"))))

	// Serve main page and terminal
	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
	s.router.HandleFunc("/terminal", s.handleTerminal).Methods("GET")
}

func (s *Server) Start(addr string) error {
	log.Printf("Server starting on %s", addr)
	return http.ListenAndServe(addr, s.router)
}
