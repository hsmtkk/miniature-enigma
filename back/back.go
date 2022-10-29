package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/hsmtkk/miniature-enigma/openweather"
	"github.com/hsmtkk/miniature-enigma/trace"
	"github.com/hsmtkk/miniature-enigma/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

const serviceName = "back"

func main() {
	ctx := context.Background()

	projectID, err := util.GetProjectID(ctx)
	if err != nil {
		log.Fatal(err)
	}

	port, err := util.GetPort()
	if err != nil {
		log.Fatal(err)
	}

	secretID, err := util.RequiredEnv("OPENWEATHER_KEY_SECRET_ID")
	if err != nil {
		log.Fatal(err)
	}

	tp, err := trace.GetTraceProvider(ctx, projectID, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(ctx)

	h := newHandler(projectID, secretID)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(otelecho.Middleware(serviceName, otelecho.WithTracerProvider(tp)))

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
	// random delay
	time.Sleep(time.Second * time.Duration(rand.Intn(10)))

	// random error
	if rand.Intn(10) < 3 {
		return fmt.Errorf("something went wrong")
	}

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

	fmt.Printf("response: %v\n", decoded)

	return c.JSON(http.StatusOK, decoded)
}
