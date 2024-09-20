package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/mhayk/GO-Expert-temp-by-cep-otel/configs"
	"github.com/mhayk/GO-Expert-temp-by-cep-otel/input-api/internal/infra/web"
	"github.com/mhayk/GO-Expert-temp-by-cep-otel/input-api/internal/infra/web/webserver"
	otel_provider "github.com/mhayk/GO-Expert-temp-by-cep-otel/pkg/otel"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

func ConfigureServer(conf *configs.Conf) *webserver.WebServer {
	fmt.Println("Starting web server on port", conf.InputApiHttpPort)

	tracer := otel.Tracer("intput-api-tracer")

	webserver := webserver.NewWebServer(":" + conf.InputApiHttpPort)
	webCEPHandler := web.NewWebCEPHandler(conf, tracer)
	webStatusHandler := web.NewWebStatusHandler()
	webserver.AddHandler("POST /cep", webCEPHandler.Get)
	webserver.AddHandler("GET /status", webStatusHandler.Get)
	return webserver
}

func init() {
	viper.AutomaticEnv()
}

func main() {
	// Load the configurations
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// start the provider
	shutdown, err := otel_provider.InitProvider(configs.InputApiOtelServiceName, configs.OpenTelemetryCollectorExporerEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	// start the web server
	go func() {
		webserver := ConfigureServer(configs)
		webserver.Start()
	}()

	select {
	case <-sigCh:
		log.Println("Shutting down gracefully, CTRL+c pressed...")
	case <-ctx.Done():
		log.Println("Shutting down due other reason...")
	}

	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
}
