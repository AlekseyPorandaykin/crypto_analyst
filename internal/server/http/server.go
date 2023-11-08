package http

import (
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/price"
	"github.com/labstack/echo/v4"
)

type Server struct {
	e *echo.Echo
}

func NewServer(calculate *price.Calculate) *Server {
	e := echo.New()
	h := NewHandler(calculate)
	h.RegistrationRoute(e)
	return &Server{
		e: e,
	}
}

func (s *Server) Run(address string) error {
	return s.e.Start(address)
}

func (s *Server) Close() {
	_ = s.e.Close()
}
