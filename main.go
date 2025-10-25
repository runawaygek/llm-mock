package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"llm-mock/internal"

	"go.uber.org/zap"
)

func init() {

}

func main() {
	internal.InitLogger(zap.InfoLevel)

	configPath := flag.String("c", "config.yaml", "config file")
	flag.Parse()

	internal.InitConfig(*configPath)
	internal.InitTokens()

	appConfig := internal.AppConfig

	internal.Logger.Info("Config initialized",
		zap.Any("config", appConfig))

	engine := internal.NewServer()

	addr := fmt.Sprintf(":%d", internal.AppConfig.Server.Port)

	readtimeout := time.Duration(internal.AppConfig.Server.ReadTimeout) * time.Second
	writetimeout := time.Duration(internal.AppConfig.Server.WriteTimeout) * time.Second

	server := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  readtimeout,
		WriteTimeout: writetimeout,
	}

	internal.Logger.Info("Server initialized",
		zap.String("addr", addr),
		zap.Duration("readtimeout", readtimeout),
		zap.Duration("writetimeout", writetimeout),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		internal.Logger.Info("Starting server: " + addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			internal.Logger.Error("listen error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	stop()
	internal.Logger.Info("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		internal.Logger.Error("server shutdown", zap.Error(err))
	}

	internal.Logger.Info("server exiting")
}
