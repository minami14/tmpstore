package main

import (
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/minami14/tmpstore"
)

const (
	mb = 1 << 20
	maxFileSize = 100 * mb
)

func main() {
	s := tmpstore.New("tmpstore")
	s.SetMaxFileSize(maxFileSize)
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

		return context.String(http.StatusOK, name)
	})

	e.GET("/", func(context echo.Context) error {
		name := context.QueryParam("name")
		if err := s.UpdateTime(name); err != nil {
			return err
		}

		return context.File(s.Dir() + name)
	})

	e.GET("/test", func(context echo.Context) error {
		return context.String(http.StatusOK, "test")
	})

	e.Logger.Fatal(e.Start(":80"))
}
