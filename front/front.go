package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/hsmtkk/miniature-enigma/trace"
	"github.com/hsmtkk/miniature-enigma/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

const serviceName = "front"

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

	backURL, err := util.RequiredEnv("BACK_URL")
	if err != nil {
		log.Fatal(err)
	}

	collection, err := util.RequiredEnv("COLLECTION")
	if err != nil {
		log.Fatal(err)
	}

	tp, err := trace.GetTraceProvider(ctx, projectID, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Shutdown(ctx)

	h := newHandler(projectID, backURL, collection)

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
	projectID  string
	backURL    string
	collection string
}

func newHandler(projectID, backURL, collection string) *handler {
	return &handler{projectID, backURL, collection}
}

var cities = []string{"Tokyo", "Osaka", "Nagoya", "Fukuoka", "Kyoto", "Sapporo", "Sendai", "Naha", "Hiroshima", "HogeFuga"}

// Handler
func (h *handler) root(c echo.Context) error {
	reqBytes, err := httputil.DumpRequest(c.Request(), false)
	if err != nil {
		return fmt.Errorf("httputil.DumpRequest failed; %w", err)
	}
	fmt.Printf("request dump: %s\n", string(reqBytes))

	// random sleep
	time.Sleep(time.Second * time.Duration(rand.Intn(10)))

	city := cities[rand.Intn(len(cities))]

	result, err := h.accessBack(city)
	if err != nil {
		return err
	}

	fmt.Printf("response from back: %s\n", string(result))

	var decoded map[string]interface{}
	if err := json.Unmarshal(result, &decoded); err != nil {
		return fmt.Errorf("json.Unmarshal failed; %w", err)
	}

	if err := firestoreSave(c.Request().Context(), h.projectID, h.collection, decoded); err != nil {
		return err
	}

	return c.String(http.StatusOK, city)
}

func (h *handler) accessBack(city string) ([]byte, error) {
	url := fmt.Sprintf("%s?city=%s", h.backURL, city)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get failed; %w", err)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed; %w", err)
	}
	return result, nil
}
