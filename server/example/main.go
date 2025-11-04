package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/ductran999/shared-pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type App struct {
	HTTPServer server.HttpServer
}

func InitializeHttpServer() (server.HttpServer, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		time.Sleep(time.Second * 10)
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	httpServer, err := server.NewGinHttpServer(router, server.ServerConfig{
		Host: "localhost",
		Port: 9090,
	})
	if err != nil {
		return nil, err
	}

	return httpServer, nil
}

func NewApp() *App {
	httpServer, err := InitializeHttpServer()
	if err != nil {
		log.Fatal().Msgf("failed to initialize HTTP server: %v", err)
	}

	return &App{
		HTTPServer: httpServer,
	}
}

func main() {
	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := NewApp()
	go func() {
		if err := app.HTTPServer.Start(); err != nil {
			log.Fatal().Msgf("failed to start http server: %v", err)
		}
	}()

	<-appCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.HTTPServer.Stop(ctx); err != nil {
		log.Error().Msgf("failed to stop http server: %v", err)
	} else {
		log.Info().Msg("http server stopped successfully")
	}
}
