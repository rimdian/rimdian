package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/rimdian/rimdian/internal/collector"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

func main() {

	log := logrus.New()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, log); err != nil {
		// TODO use another logger?
		log.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logrus.Logger) error {

	viper.AutomaticEnv()
	viper.SetDefault("ENV", collector.ENV_PROD)
	viper.SetDefault("COLLECTOR_PORT", "8080")

	// we can use config files in non-production env
	if os.Getenv("ENV") != collector.ENV_PROD {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./")

		// load config.yaml
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	}

	cfg := &collector.Config{
		ENV:                     viper.GetString("ENV"),
		COLLECTOR_PORT:          viper.GetString("COLLECTOR_PORT"),
		API_ENDPOINT:            viper.GetString("API_ENDPOINT"),
		SECRET_KEY:              viper.GetString("SECRET_KEY"),
		GCLOUD_PROJECT:          viper.GetString("GCLOUD_PROJECT"),
		GCLOUD_JSON_CREDENTIALS: viper.GetString("GCLOUD_JSON_CREDENTIALS"),
		TASK_QUEUE_LOCATION:     viper.GetString("TASK_QUEUE_LOCATION"),
		DEV_SSL_CERT:            viper.GetString("DEV_SSL_CERT"),
		DEV_SSL_KEY:             viper.GetString("DEV_SSL_KEY"),
	}

	if err := collector.ValidateConfig(cfg); err != nil {
		return err
	}

	// init task client

	serviceAccount := option.WithCredentialsJSON([]byte(cfg.GCLOUD_JSON_CREDENTIALS))

	cloudTaskClient, err := cloudtasks.NewClient(ctx, serviceAccount)
	if err != nil {
		return err
	}

	// init net client for sync data import

	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	if cfg.ENV == collector.ENV_DEV {
		netTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	netClient := &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}

	// init app

	taskOrchestrator := taskorchestrator.NewClient(cfg.GCLOUD_PROJECT, cfg.ENV, cloudTaskClient)
	api := collector.NewCollector(cfg, taskOrchestrator, netClient, log)

	// launch webserver

	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", cfg.COLLECTOR_PORT),
		Handler:           api.Handler,
		ReadHeaderTimeout: time.Second * 10,
		ReadTimeout:       time.Second * 30,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	// 30 secs timeout graceful shutdown

	errs := make(chan error, 1)
	go func() {
		<-ctx.Done()
		fmt.Println()

		log.WithField("component", "collector").Info("Triggering graceful shutdown...")
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), time.Second*30)
		defer cancelShutdown()
		if err := server.Shutdown(ctxShutdown); err != nil {
			errs <- fmt.Errorf("could not shutdown server: %w", err)
		}

		errs <- nil
	}()

	log.WithField("component", "collector").Infof("Collector listening on port %v", server.Addr)

	if cfg.ENV == collector.ENV_DEV {

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
