package entity

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
	ENV                string
	API_PORT           string
	API_ENDPOINT       string // hostname of the exposed API server (i.e: captainmetrics.mybusiness.com)
	COLLECTOR_ENDPOINT string // hostname of the exposed COLLECTOR server (i.e: collector.mybusiness.com)
	SECRET_KEY         string // secret key for signing tokens & data
	LICENSE_PUBLIC_KEY string // public key for verifying tokens
	ORGANIZATION_ID    string // default organization id created on install
	ORGANIZATION_NAME  string // default organization name created on install
	ROOT_EMAIL         string // root account account login
	DB_TYPE            string // singlestore | mysql
	DB_MAINTENANCE     bool   // true if DB is under maintenance and not accessible
	DB_DSN             string // DB connection string, i.e: user:pass@tcp(IP:3306)/?timeout=30s&interpolateParams=true&parseTime=true&tls=skip-verify
	DB_CA_CERT_BASE64  string // DB TLS Certificate Authority
	// DB_TLS_CERT                    string // DB TLS Certificate
	// DB_TLS_KEY                     string // DB TLS Key
	DB_PREFIX               string // DB prefix default: cm_
	DB_MAX_OPEN_CONNS       int    // DB max open connections
	DB_MAX_IDLE_CONNS       int    // DB max idle connections
	SENTRY_DSN              string // Sentry DSN
	SMTP_HOST               string // SMTP host
	SMTP_PORT               int    // SMTP port
	SMTP_USERNAME           string // SMTP user
	SMTP_PASSWORD           string // SMTP pass
	SMTP_FROM               string // SMTP from email
	SMTP_ENCRYPTION         string // SMTP encryption SSLTLS or STARTTLS
	GCLOUD_PROJECT          string // Google Cloud workspace ID
	GCLOUD_JSON_CREDENTIALS string // Google Cloud JSON credentials
	TASK_QUEUE_LOCATION     string // Task Queue location for tasks (i.e: europe-west3)
	CUBEJS_ENDPOINT         string // Cube.js endpoint
	DEV_SSL_CERT            string // ssl certificate for dev server
	DEV_SSL_KEY             string // ssl certificate for dev server
	MANAGED_CM              bool
	OPEN_CENSUS_EXPORTER    string
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

	if cfg.COLLECTOR_ENDPOINT == "" {
		return eris.New("COLLECTOR_ENDPOINT is required")
	}

	if !govalidator.IsRequestURL(cfg.COLLECTOR_ENDPOINT) {
		return fmt.Errorf("COLLECTOR_ENDPOINT is not a valid hostname: %v", cfg.COLLECTOR_ENDPOINT)
	}

	if !strings.HasPrefix(cfg.COLLECTOR_ENDPOINT, "https://") {
		return eris.New("COLLECTOR_ENDPOINT should start with https://")
	}

	if cfg.SECRET_KEY == "" {
		return eris.New("SECRET_KEY is required")
	}

	if len(cfg.SECRET_KEY) != 32 {
		return eris.New("SECRET_KEY should be 32 characters long")
	}

	if cfg.LICENSE_PUBLIC_KEY == "" {
		return eris.New("LICENSE_PUBLIC_KEY is required")
	}

	if cfg.ORGANIZATION_ID == "" {
		return eris.New("ORGANIZATION_ID is required")
	}

	if cfg.ORGANIZATION_NAME == "" {
		return eris.New("ORGANIZATION_NAME is required")
	}

	if !govalidator.IsEmail(cfg.ROOT_EMAIL) {
		return eris.New("ROOT_EMAIL is not a valid email address")
	}

	if cfg.DB_TYPE == "" {
		cfg.DB_TYPE = "singlestore"
	}

	if cfg.DB_TYPE != "singlestore" && cfg.DB_TYPE != "mysql" {
		return eris.New("DB_TYPE should be either singlestore or mysql")
	}

	if !cfg.DB_MAINTENANCE {
		if cfg.DB_DSN == "" {
			return eris.New("DB_DSN is required")
		}
		// if cfg.DB_CA_CERT_BASE64 == "" {
		// 	return eris.New("DB_CA_CERT_BASE64 is required")
		// }
		// if cfg.DB_TLS_CERT == "" {
		// 	return eris.New("DB_TLS_CERT is required")
		// }
		// if cfg.DB_TLS_KEY == "" {
		// 	return eris.New("DB_TLS_KEY is required")
		// }
	}

	if !govalidator.IsHost(cfg.SMTP_HOST) {
		return fmt.Errorf("SMTP_HOST is not a valid hostname: %v", cfg.SMTP_HOST)
	}
	if cfg.SMTP_PORT == 0 {
		return eris.New("SMTP_PORT is required")
	}
	if cfg.SMTP_USERNAME == "" {
		return eris.New("SMTP_USERNAME is required")
	}
	if cfg.SMTP_PASSWORD == "" {
		return eris.New("SMTP_PASSWORD is required")
	}
	if !govalidator.IsEmail(cfg.SMTP_FROM) {
		return eris.New("SMTP_FROM is not a valid email address")
	}
	if !govalidator.IsEmail(cfg.SMTP_FROM) {
		return eris.New("SMTP_FROM is not a valid email address")
	}
	if cfg.SMTP_ENCRYPTION != "" && (cfg.SMTP_ENCRYPTION != "SSLTLS" && cfg.SMTP_ENCRYPTION != "STARTTLS") {
		return eris.New("SMTP_ENCRYPTION should be SSLTLS or STARTTLS")
	}

	if cfg.GCLOUD_PROJECT == "" {
		return eris.New("GCLOUD_PROJECT is required")
	}

	if cfg.GCLOUD_JSON_CREDENTIALS == "" {
		return eris.New("GCLOUD_JSON_CREDENTIALS is required")
	}

	if cfg.TASK_QUEUE_LOCATION == "" {
		return eris.New("TASK_QUEUE_LOCATION is required")
	}

	if cfg.ENV == ENV_DEV && (cfg.DEV_SSL_CERT == "" || cfg.DEV_SSL_KEY == "") {
		return eris.New("DEV_SSL_CERT & DEV_SSL_KEY absolute paths are required in dev")
	}

	return nil
}
