package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (app *application) serve(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// for multiple goroutines
	eg, ctx := errgroup.WithContext(ctx)

	// Start the server
	eg.Go(func() error {
		app.logger.Info("starting server",zap.String("addr", addr))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	// Wait for termination signal
	<-ctx.Done()
	app.logger.Info("shutting down server",zap.Error(ctx.Err()))

	// Context for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	eg.Go(func() error {
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		app.logger.Error("error during shutdown", zap.Error(err))
	
		return err
	}

	app.logger.Info("stopped server", zap.String("addr", addr))
	return nil
}
