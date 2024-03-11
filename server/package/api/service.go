package api

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/storage"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rimdian/rimdian/internal/api/service"
	"github.com/rimdian/rimdian/internal/common/mailer"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"google.golang.org/api/option"
)

func NewService(ctx context.Context, logger *logrus.Logger, cfg *entity.Config, repo repository.Repository) (service.Service, error) {

	mailer := mailer.NewMailer(cfg.SMTP_FROM, cfg.SMTP_USERNAME, cfg.SMTP_PASSWORD, cfg.SMTP_HOST, cfg.SMTP_PORT, cfg.SMTP_ENCRYPTION)

	// init task client

	serviceAccount := option.WithCredentialsJSON([]byte(cfg.GCLOUD_JSON_CREDENTIALS))

	cloudTaskClient, err := cloudtasks.NewClient(ctx, serviceAccount)
	if err != nil {
		return nil, err
	}

	taskOrchestrator := taskorchestrator.NewClient(cfg.GCLOUD_PROJECT, cfg.ENV, cloudTaskClient)

	// init storage client

	storageClient, err := storage.NewClient(ctx, serviceAccount)
	if err != nil {
		return nil, err
	}

	// init net client for sync data import
	// wrap the transport for OpenCensus
	base := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{},
	}

	if cfg.ENV == entity.ENV_DEV {
		base.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	netTransport := &ochttp.Transport{
		Base: base,
		// Use Google Cloud propagation format.
		Propagation: &propagation.HTTPFormat{},
	}

	netClient := &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}

	svc := service.NewService(cfg, logger, repo, mailer, taskOrchestrator, storageClient, netClient)

	if _, err := svc.InstallOrVerifyServer(ctx); err != nil {
		return nil, err
	}

	return svc, nil
}
