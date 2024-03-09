package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/package/api"
	"github.com/sirupsen/logrus"
)

func main() {

	log := logrus.New()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, log); err != nil {
		log.WithField("component", "api").WithError(err).Fatal("run api server failed")
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logrus.Logger) error {

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, log)
	if err != nil {
		return err
	}

	svc, err := api.NewService(ctx, log, cfg, repo)
	if err != nil {
		return err
	}

	// create & launch webserver
	server, err := api.NewAPIServer(ctx, log, svc)
	if err != nil {
		return err
	}

	// 30 secs timeout graceful shutdown

	errs := make(chan error, 1)
	go func() {
		<-ctx.Done()

		log.WithField("component", "api").Info("Triggering graceful shutdown...")
		// _ = logger.Log("message", "gracefully shutting down")
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), time.Second*30)
		defer cancelShutdown()
		if err := server.Shutdown(ctxShutdown); err != nil {
			errs <- fmt.Errorf("could not shutdown server: %w", err)
		}

		errs <- nil
	}()

	log.WithField("component", "api").Infof("API listening on port %v", server.Addr)

	if cfg.ENV == entity.ENV_DEV {

		// Serve dev server over SSL for CORS compatibility

		if err := server.ListenAndServeTLS(cfg.DEV_SSL_CERT, cfg.DEV_SSL_KEY); err != http.ErrServerClosed {
			close(errs)
			return fmt.Errorf("could not listen and serve: %w", err)
		}

	} else {

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			close(errs)
			return fmt.Errorf("could not listen and serve: %w", err)
		}
	}

	return <-errs
}
