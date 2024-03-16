package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	api "github.com/rimdian/rimdian/internal/api/router"
	"github.com/rimdian/rimdian/internal/api/service"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
)

func NewAPIServer(ctx context.Context, log *logrus.Logger, svc service.Service) (*http.Server, error) {

	cfg := svc.GetConfig()

	if cfg.OPEN_CENSUS_EXPORTER == "stackdriver" {
		// GCP service account
		serviceAccount := option.WithCredentialsJSON([]byte(cfg.GCLOUD_JSON_CREDENTIALS))

		// https://opencensus.io/exporters/supported-exporters/go/stackdriver/
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:               cfg.GCLOUD_PROJECT,
			MonitoringClientOptions: []option.ClientOption{serviceAccount},
			TraceClientOptions:      []option.ClientOption{serviceAccount},
			// https://cloud.google.com/trace/docs/setup/go#oc-upload-fail
			ReportingInterval:        60 * time.Second,  // Stackdriver’s minimum stats reporting period must be >= 60 seconds
			TraceSpansBufferMaxBytes: 100 * 1000 * 1000, // 100 MB
			BundleCountThreshold:     500,
		})
		if err != nil {
			log.Printf("stackdriver trace init error: %s", err)
			return nil, err
		}
		defer exporter.Flush()

		// start the metrics exporter
		exporter.StartMetricsExporter()
		defer exporter.StopMetricsExporter()

		// Subscribe client + server views to see stats in Stackdriver Monitoring.

		if err := view.Register(
			ochttp.ClientSentBytesDistribution,
			ochttp.ClientReceivedBytesDistribution,
			ochttp.ClientRoundtripLatencyDistribution,
			ochttp.ClientCompletedCount,
		); err != nil {
			log.Printf("stackdriver register client stats: %s", err)
			return nil, err
		}

		if err := view.Register(
			ochttp.ServerRequestCountView,
			ochttp.ServerRequestBytesView,
			ochttp.ServerResponseBytesView,
			ochttp.ServerLatencyView,
			ochttp.ServerRequestCountByMethod,
			ochttp.ServerResponseCountByStatusCode,
		); err != nil {
			log.Printf("stackdriver register server stats: %s", err)
			return nil, err
		}

		// export traces in production
		if cfg.ENV == "production" {
			trace.RegisterExporter(exporter)
			trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(0.03)})
			// trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		}
	}

	api := api.NewAPI(cfg, svc, log)

	return &http.Server{
		Addr: fmt.Sprintf(":%v", cfg.API_PORT),
		// wrap handler with OpenCensus
		Handler: &ochttp.Handler{
			Handler:     api.Handler,
			Propagation: &propagation.HTTPFormat{},
		},
		ReadHeaderTimeout: time.Second * 10,
		ReadTimeout:       time.Second * 30,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}, nil
}
