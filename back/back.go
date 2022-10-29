package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hsmtkk/miniature-enigma/openweather"
	"github.com/hsmtkk/miniature-enigma/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	port, err := util.GetPort()
	if err != nil {
		log.Fatal(err)
	}

	projectID, err := util.GetProjectID()
	if err != nil {
		log.Fatal(err)
	}

	secretID, err := util.RequiredEnv("OPENWEATHER_KEY_SECRET_ID")
	if err != nil {
		log.Fatal(err)
	}

	h := newHandler(projectID, secretID)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", h.root)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

type handler struct {
	projectID string
	secretID  string
}

func newHandler(projectID, secretID string) *handler {
	return &handler{projectID, secretID}
}

type query struct {
	City string `query:"city"`
}

// Handler
func (h *handler) root(c echo.Context) error {
	var q query
	if err := c.Bind(&q); err != nil {
		return fmt.Errorf("echo.Context.Bind failed; %w", err)
	}

	key, err := getOpenweatherKey(c.Request().Context(), h.projectID, h.secretID)
	if err != nil {
		return err
	}

	result, err := openweather.CurrentData(q.City, key)
	if err != nil {
		return err
	}

	var decoded interface{}
	if err := json.Unmarshal(result, &decoded); err != nil {
		return fmt.Errorf("json.Unmarshal failed; %w", err)
	}

	return c.JSON(http.StatusOK, decoded)
}
