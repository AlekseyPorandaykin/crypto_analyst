package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/labstack/echo/v4/middleware"
)

type Handler interface {
	RegistrationRoute(e *echo.Echo)
}

type ApiRouteGroup interface {
	RegistrationRouteApi(g *echo.Group)
}

type Server struct {
	e *echo.Echo
}

func NewServer() *Server {
	e := echo.New()
	e.Use(middleware.Recover(), middleware.CORS())
	return &Server{
		e: e,
	}
}

func (s *Server) Registration(h Handler) {
	h.RegistrationRoute(s.e)
}
func (s *Server) RegistrationApi(h Handler) {
	h.RegistrationRoute(s.e)
}

func (s *Server) Run(address string) error {
	return s.e.Start(address)
}

func (s *Server) Close() {
	_ = s.e.Close()
}
