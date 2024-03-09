package collector

import (
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var ENV_DEV = "development"
var ENV_TEST = "test"
var ENV_PROD = "production"

type Config struct {
	ENV                     string
	COLLECTOR_PORT          string
	API_ENDPOINT            string // hostname of the exposed API server (i.e: rimdian.mybusiness.com)
	SECRET_KEY              string // secret key for signing data
	GCLOUD_PROJECT          string // Google Cloud project ID
	GCLOUD_JSON_CREDENTIALS string // Google Cloud JSON credentials
	TASK_QUEUE_LOCATION     string // Task Queue location for data imports (i.e: europe-west3)
	DEV_SSL_CERT            string // ssl certificate for dev server
	DEV_SSL_KEY             string // ssl certificate for dev server
}

func ValidateConfig(cfg *Config) error {

	if cfg.API_ENDPOINT == "" {
		return eris.New("API_ENDPOINT is required")
	}

	if !govalidator.IsRequestURL(cfg.API_ENDPOINT) {
		return fmt.Errorf("API_ENDPOINT is not a valid hostname: %v", cfg.API_ENDPOINT)
	}

	if !strings.HasPrefix(cfg.API_ENDPOINT, "https://") {
		return eris.New("API_ENDPOINT should start with https://")
	}

	if cfg.SECRET_KEY == "" {
		return eris.New("SECRET_KEY is required")
	}

	if cfg.GCLOUD_PROJECT == "" {
		return eris.New("GCLOUD_PROJECT is required")
	}

	if cfg.GCLOUD_JSON_CREDENTIALS == "" {
		return eris.New("GCLOUD_JSON_CREDENTIALS is required")
	}

	if cfg.ENV == ENV_DEV && (cfg.DEV_SSL_CERT == "" || cfg.DEV_SSL_KEY == "") {
		return eris.New("DEV_SSL_CERT & DEV_SSL_KEY absolute paths are required in dev")
	}

	return nil
}
