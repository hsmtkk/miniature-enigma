package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/hsmtkk/miniature-enigma/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
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

	h := newHandler(backURL, collection)

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
	backURL    string
	collection string
}

func newHandler(backURL, collection string) *handler {
	return &handler{backURL, collection}
}

var cities = []string{"Tokyo", "Osaka", "Nagoya", "Fukuoka", "Kyoto", "Sapporo", "Sendai", "Naha", "Hiroshima", "HogeFuga"}

// Handler
func (h *handler) root(c echo.Context) error {
	city := cities[rand.Intn(len(cities))]

	projectID, err := util.GetProjectID(c.Request().Context())
	if err != nil {
		return err
	}

	result, err := h.accessBack(city)
	if err != nil {
		return err
	}

	fmt.Printf("response from back: %s\n", string(result))

	var decoded map[string]interface{}
	if err := json.Unmarshal(result, &decoded); err != nil {
		return fmt.Errorf("json.Unmarshal failed; %w", err)
	}

	if err := firestoreSave(c.Request().Context(), projectID, h.collection, decoded); err != nil {
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
