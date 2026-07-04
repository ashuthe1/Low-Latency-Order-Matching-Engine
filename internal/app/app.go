package app

import (
	"net/http"

	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/api"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/config"
	"github.com/ashuthe1/Low-Latency-Order-Matching-Engine/internal/engine"
)

type App struct {
	Config *config.Config
	Router http.Handler
	Engine engine.Engine
}

func New(cfg *config.Config) (*App, error) {

	eng := engine.New()

	router := api.NewRouter()

	return &App{
		Config: cfg,
		Router: router,
		Engine: eng,
	}, nil
}
