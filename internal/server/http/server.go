package http

import (
	"github.com/labstack/echo/v4"
)

type Handler interface {
	RegistrationRoute(e *echo.Echo)
}

type Server struct {
	e *echo.Echo
}

func NewServer() *Server {
	return &Server{
		e: echo.New(),
	}
}

func (s *Server) Registration(h Handler) {
	h.RegistrationRoute(s.e)
}

func (s *Server) Run(address string) error {
	return s.e.Start(address)
}

func (s *Server) Close() {
	_ = s.e.Close()
}
