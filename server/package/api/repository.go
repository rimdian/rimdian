package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/go-sql-driver/mysql"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewRepository(ctx context.Context, log *logrus.Logger) (repository.Repository, *entity.Config, error) {

	viper.AutomaticEnv()
	viper.SetDefault("ENV", entity.ENV_PROD)
	viper.SetDefault("API_PORT", "8000")
	viper.SetDefault("DB_MAINTENANCE", false)
	viper.SetDefault("DB_TYPE", "singlestore")
	viper.SetDefault("DB_PREFIX", "rmd_")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 200)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 200)
	viper.SetDefault("CUBEJS_ENDPOINT", "http://localhost:4444")

	// we can use config files in non-production env
	if os.Getenv("ENV") != entity.ENV_PROD {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		currentPath, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}

		configPath := strings.Split(currentPath, "server")[0] + "server/"

		viper.AddConfigPath(configPath)

		// load config.yaml
		err = viper.ReadInConfig()
		if err != nil {
			return nil, nil, err
		}
	}

	cfg := &entity.Config{
		ENV:                viper.GetString("ENV"),
		API_PORT:           viper.GetString("API_PORT"),
		API_ENDPOINT:       viper.GetString("API_ENDPOINT"),
		COLLECTOR_ENDPOINT: viper.GetString("COLLECTOR_ENDPOINT"),
		SECRET_KEY:         viper.GetString("SECRET_KEY"),
		LICENSE_PUBLIC_KEY: viper.GetString("LICENSE_PUBLIC_KEY"),
		ORGANIZATION_ID:    viper.GetString("ORGANIZATION_ID"),
		ORGANIZATION_NAME:  viper.GetString("ORGANIZATION_NAME"),
		ROOT_EMAIL:         viper.GetString("ROOT_EMAIL"),
		DB_MAINTENANCE:     viper.GetBool("DB_MAINTENANCE"),
		DB_TYPE:            viper.GetString("DB_TYPE"),
		DB_DSN:             viper.GetString("DB_DSN"),
		DB_CA_CERT_BASE64:  viper.GetString("DB_CA_CERT_BASE64"),
		// DB_TLS_CERT:                    viper.GetString("DB_TLS_CERT"),
		// DB_TLS_KEY:                     viper.GetString("DB_TLS_KEY"),
		DB_PREFIX:                      viper.GetString("DB_PREFIX"),
		DB_MAX_OPEN_CONNS:              viper.GetInt("DB_MAX_OPEN_CONNS"),
		DB_MAX_IDLE_CONNS:              viper.GetInt("DB_MAX_IDLE_CONNS"),
		SMTP_HOST:                      viper.GetString("SMTP_HOST"),
		SMTP_PORT:                      viper.GetInt("SMTP_PORT"),
		SMTP_USERNAME:                  viper.GetString("SMTP_USERNAME"),
		SMTP_PASSWORD:                  viper.GetString("SMTP_PASSWORD"),
		SMTP_FROM:                      viper.GetString("SMTP_FROM"),
		SMTP_ENCRYPTION:                viper.GetString("SMTP_ENCRYPTION"),
		GCLOUD_PROJECT:                 viper.GetString("GCLOUD_PROJECT"),
		GCLOUD_JSON_CREDENTIALS:        viper.GetString("GCLOUD_JSON_CREDENTIALS"),
		TASK_QUEUE_LOCATION:            viper.GetString("TASK_QUEUE_LOCATION"),
		CUBEJS_ENDPOINT:                viper.GetString("CUBEJS_ENDPOINT"),
		DEV_SSL_CERT:                   viper.GetString("DEV_SSL_CERT"),
		DEV_SSL_KEY:                    viper.GetString("DEV_SSL_KEY"),
		MANAGED_CM:                     viper.GetBool("MANAGED_CM"),
		APP_GOOGLE_API_KEY:             viper.GetString("APP_GOOGLE_API_KEY"),
		APP_GOOGLE_OAUTH_CLIENT_ID:     viper.GetString("APP_GOOGLE_OAUTH_CLIENT_ID"),
		APP_GOOGLE_OAUTH_SECRET:        viper.GetString("APP_GOOGLE_OAUTH_SECRET"),
		APP_GOOGLE_ADS_DEVELOPER_TOKEN: viper.GetString("APP_GOOGLE_ADS_DEVELOPER_TOKEN"),
	}

	if err := entity.ValidateConfig(cfg); err != nil {
		return nil, nil, err
	}

	var workspaceDB *sql.DB
	var systemDB *sql.DB

	if !cfg.DB_MAINTENANCE {

		// add CA cert if provided
		if cfg.DB_CA_CERT_BASE64 != "" {
			caDecoded, err := base64.StdEncoding.DecodeString(cfg.DB_CA_CERT_BASE64)
			if err != nil {
				return nil, nil, errors.New("failed to decode base64 ca")
			}

			// Create a pool of TLS certs and append one
			rootCertPool := x509.NewCertPool()
			if ok := rootCertPool.AppendCertsFromPEM(caDecoded); !ok {
				return nil, nil, errors.New("failed to append ca to sql cert pool")
			}

			// clientCert := make([]tls.Certificate, 0, 1)

			// certs, err := tls.X509KeyPair([]byte(cfg.DB_TLS_CERT), []byte(cfg.DB_TLS_KEY))
			// if err != nil {
			// 	return err
			// }

			// clientCert = append(clientCert, certs)

			mysql.RegisterTLSConfig("custom", &tls.Config{
				RootCAs: rootCertPool,
				// Certificates: clientCert,
			})
		}

		// trace SQL queries with OpenCensus
		// var ocsqlOptions ocsql.TraceOption
		ocsqlOptions := ocsql.WithAllTraceOptions()
		driverName, err := ocsql.Register("mysql", ocsqlOptions)
		if err != nil {
			log.Infof("Failed to register the ocsql driver: %v", err)
			return nil, nil, err
		}

		sqlConfig, err := mysql.ParseDSN(cfg.DB_DSN)

		if err != nil {
			return nil, nil, err
		}

		sqlConfig.ParseTime = true

		// append SingleStore specific params
		if cfg.DB_TYPE == "singlestore" {
			// interpolate params reduce the roundtrips to the DB by avoiding the prepare phase
			// it doesnt work with mysql json fields though
			// see: https://github.com/go-sql-driver/mysql/issues/819
			sqlConfig.InterpolateParams = true
		}

		if sqlConfig.Timeout == 0 {
			sqlConfig.Timeout = time.Second * 30
		}

		// INIT WORKSPACE DB

		workspaceDB, err = sql.Open(driverName, sqlConfig.FormatDSN())
		if err != nil {
			return nil, nil, err
		}

		// enable periodic recording of sql.DBStats for OpenCensus (open/idle/closed pool connections...)
		dbstatsCloser := ocsql.RecordStats(workspaceDB, 5*time.Second)
		defer dbstatsCloser()

		pingStart := time.Now()
		if err = workspaceDB.Ping(); err != nil {
			return nil, nil, err
		}
		log.WithField("component", "api").Infof("ping DB took %v", time.Since(pingStart))

		pingStart = time.Now()
		if err = workspaceDB.Ping(); err != nil {
			return nil, nil, err
		}
		log.WithField("component", "api").Infof("second ping DB took %v", time.Since(pingStart))

		workspaceDB.SetConnMaxLifetime(time.Second * 40)
		workspaceDB.SetMaxOpenConns(cfg.DB_MAX_OPEN_CONNS)
		workspaceDB.SetMaxIdleConns(cfg.DB_MAX_IDLE_CONNS)

		// INIT SYSTEM DB

		systemDB, err = sql.Open(driverName, sqlConfig.FormatDSN())
		if err != nil {
			return nil, nil, err
		}

		systemDB.SetConnMaxLifetime(time.Second * 40)
		systemDB.SetMaxOpenConns(cfg.DB_MAX_OPEN_CONNS)
		systemDB.SetMaxIdleConns(cfg.DB_MAX_IDLE_CONNS)
	}

	return repository.NewRepository(cfg, workspaceDB, systemDB), cfg, nil
}
