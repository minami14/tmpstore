package main

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/minami14/tmpstore"
)

const (
	mb          = 1 << 20
	maxFileSize = 100 * mb
	duration    = time.Hour
	lifetime    = 24 * time.Hour
)

func main() {
	s := tmpstore.New()
	s.SetDirectory("tmpstore")
	s.SetMaxFileSize(maxFileSize)
	s.SetDuration(duration)
	s.SetLifetime(lifetime)
	go s.Run()
	defer s.Clear()

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.PUT("/", func(context echo.Context) error {
		body, err := ioutil.ReadAll(context.Request().Body)
		if err != nil {
			return err
		}

		name := uuid.New().String()
		if err := s.Store(name, body); err != nil {
			return err
		}

		return context.String(http.StatusCreated, name)
	})

	e.GET("/", func(context echo.Context) error {
		name := context.QueryParam("name")
		if err := s.UpdateTime(name); err != nil {
			return err
		}

		return context.File(s.Dir() + name)
	})

	e.DELETE("/", func(context echo.Context) error {
		name := context.QueryParam("name")
		if err := s.Remove(name); err != nil {
			return err
		}

		return context.NoContent(http.StatusNoContent)
	})

	e.GET("/test", func(context echo.Context) error {
		return context.String(http.StatusOK, "test")
	})

	e.Logger.Fatal(e.Start(":80"))
}
