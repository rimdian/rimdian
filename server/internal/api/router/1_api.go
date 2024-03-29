package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	logger "github.com/chi-middleware/logrus-logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/service"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rimdian/rimdian/internal/common/cors"
	"github.com/rimdian/rimdian/internal/common/utils"
	"github.com/sirupsen/logrus"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

type API struct {
	Config  *entity.Config
	Svc     service.Service
	Logger  *logrus.Logger
	Handler http.Handler
}

func (api *API) ReturnJSONError(w http.ResponseWriter, code int, err error) {
	utils.ReturnJSONError(w, api.Logger, code, err)
}

func ReturnJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %s	", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error encoding JSON: %s", err)))

		return
	}
}

func NewAPI(cfg *entity.Config, svc service.Service, log *logrus.Logger) *API {

	r := chi.NewRouter()

	api := &API{
		Config:  cfg,
		Svc:     svc,
		Logger:  log,
		Handler: r,
	}

	r.Use(middleware.RealIP)
	r.Use(logger.Logger("api", log))
	r.Use(cors.Middleware)

	if cfg.ENV != entity.ENV_DEV {
		r.Use(middleware.Recoverer)
	}

	if cfg.SENTRY_DSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SENTRY_DSN,
			Debug:       cfg.ENV == entity.ENV_DEV,
			ServerName:  fmt.Sprintf("%s-%s", cfg.ENV, cfg.API_ENDPOINT),
			Release:     fmt.Sprintf("%v", common.Version),
			Environment: cfg.ENV,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
		// If using middleware.Recoverer the Sentry middleware must come afterwards
		// and Repanic must be set to true.
		r.Use(sentryhttp.New(sentryhttp.Options{
			Repanic: cfg.ENV != entity.ENV_DEV,
		}).Handle)
	}

	// middleware that reads an eventual paseto token and sets context
	r.Use(auth.MiddlewarePasetoExtractor(api.Logger, cfg.API_ENDPOINT, cfg.SECRET_KEY))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API running"))
	})

	// PUBLIC routes
	r.Post("/api/dataLog.importFromQueue", api.DataLogImportFromQueue) // receives a data_log item from the task queue
	r.Post("/api/dataLog.reprocessUntil", api.DataLogReprocessUntil)   // reprocess "pending" data_logs until a given date
	r.Post("/api/taskExec.do", api.TaskExecDo)                         // receives a job from the task queue
	r.Post("/api/task.wakeUpCron", api.TaskWakeUpCron)                 // cron to wake up the server and trigger cron tasks
	r.Post("/api/account.login", api.AccountLogin)
	r.Post("/api/account.resetPassword", api.AccountResetPassword)
	r.Post("/api/account.consumeResetPassword", api.AccountConsumeResetPassword)
	r.Post("/api/organizationInvitation.consume", api.OrganizationInvitationConsume)
	r.Post("/api/organizationInvitation.read", api.OrganizationInvitationRead)
	r.Post("/api/app.getFromToken", api.AppFromToken) // get app from iframe ?token=xxx (for apps)

	// ADMIN routes protected with token
	r.Group(func(r chi.Router) {
		r.Use(auth.MiddlewarePasetoRequired(api.Logger, cfg.SECRET_KEY))
		r.Post("/api/account.refreshAccessToken", api.AccountRefreshAccessToken)
		r.Post("/api/account.setProfile", api.AccountSetProfile)
		r.Post("/api/account.logout", api.AccountLogout)

		r.Get("/api/organization.list", api.OrganizationList)
		r.Post("/api/organization.create", api.OrganizationCreate)
		r.Post("/api/organization.setProfile", api.OrganizationSetProfile)

		r.Post("/api/organizationAccount.createServiceAccount", api.OrganizationAccountCreateServiceAccount)
		r.Post("/api/organizationAccount.transferOwnership", api.OrganizationAccountTransferOwnership)
		r.Post("/api/organizationAccount.deactivate", api.OrganizationAccountDeactivate)
		r.Get("/api/organizationAccount.list", api.OrganizationAccountList)

		r.Post("/api/organizationInvitation.create", api.OrganizationInvitationCreate)
		r.Get("/api/organizationInvitation.list", api.OrganizationInvitationList)
		r.Post("/api/organizationInvitation.cancel", api.OrganizationInvitationCancel)
		// r.Post("/api/organizationInvitation.resend", api.OrganizationInvitationResend)

		r.Post("/api/workspace.create", api.WorkspaceCreate)
		r.Post("/api/workspace.update", api.WorkspaceUpdate)
		r.Post("/api/workspace.createOrResetDemo", api.WorkspaceCreateOrResetDemo)
		r.Get("/api/workspace.showTables", api.WorkspaceShowTables)
		r.Get("/api/workspace.list", api.WorkspaceList)
		r.Get("/api/workspace.show", api.WorkspaceShow)
		r.Post("/api/workspace.getSecretKey", api.WorkspaceGetSecretKey) // POST for XSS

		r.Post("/api/domain.upsert", api.DomainUpsert)
		r.Post("/api/domain.delete", api.DomainDelete)

		r.Post("/api/channel.create", api.ChannelCreate)
		r.Post("/api/channel.update", api.ChannelUpdate)
		r.Post("/api/channel.delete", api.ChannelDelete)
		r.Post("/api/channelGroup.upsert", api.ChannelGroupUpsert)
		r.Post("/api/channelGroup.delete", api.ChannelGroupDelete)

		r.Get("/api/task.list", api.TaskList)
		r.Post("/api/task.run", api.TaskRun)
		r.Get("/api/taskExec.list", api.TaskExecList)
		r.Post("/api/taskExec.abort", api.TaskExecAbort)
		r.Post("/api/taskExec.create", api.TaskExecCreate)
		r.Get("/api/taskExec.jobs", api.TaskExecJobs)
		r.Get("/api/taskExec.jobInfo", api.TaskExecJobInfo)

		r.Get("/api/dataLog.list", api.DataLogList)
		r.Post("/api/dataLog.reprocessOne", api.DataLogReprocessOne)

		r.Post("/api/dataHook.update", api.DataHookUpdate)

		r.Get("/api/user.list", api.UserList)
		r.Get("/api/user.show", api.UserShow)

		r.Get("/api/segment.list", api.SegmentList)
		r.Post("/api/segment.preview", api.SegmentPreview)
		r.Post("/api/segment.create", api.SegmentCreate)
		r.Post("/api/segment.update", api.SegmentUpdate)
		r.Post("/api/segment.delete", api.SegmentDelete)

		r.Get("/api/subscriptionList.list", api.SubscriptionListList)
		r.Post("/api/subscriptionList.create", api.SubscriptionListCreate)

		r.Get("/api/messageTemplate.list", api.MessageTemplateList)
		r.Post("/api/messageTemplate.upsert", api.MessageTemplateUpsert)

		r.Get("/api/app.list", api.AppList)
		r.Get("/api/app.get", api.AppGet)
		r.Post("/api/app.install", api.AppInstall)
		r.Post("/api/app.activate", api.AppActivate)
		r.Post("/api/app.stop", api.AppStop)
		r.Post("/api/app.delete", api.AppDelete)
		r.Post("/api/app.mutateState", api.AppMutateState)
		r.Post("/api/app.execQuery", api.AppExecQuery)

		r.Post("/api/db.select", api.DBSelect)
		r.Get("/api/cubejs.schemas", api.CubeJSSchemas)
	})

	// DEV routes
	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if cfg.ENV != entity.ENV_DEV {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
			})
		})
		r.Post("/api/dev.resetDB", api.DevResetDB)
		r.Get("/api/dev.execTaskWithWorkers", api.DevExecTaskWithWorkers)
		r.Get("/api/dev.execDataImportFromQueue", api.DevExecDataImportFromQueue)
	})

	consoleDir := "console"
	fs := http.FileServer(http.Dir(consoleDir))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {

		// generate frontend config
		if r.RequestURI == "/config.js" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")

			consoleConfig := struct {
				ENV                string `json:"ENV"`
				API_ENDPOINT       string `json:"API_ENDPOINT"`
				COLLECTOR_ENDPOINT string `json:"COLLECTOR_ENDPOINT"`
				CUBEJS_ENDPOINT    string `json:"CUBEJS_ENDPOINT"`
				MANAGED_RMD        bool   `json:"MANAGED_RMD"`
			}{
				ENV:                cfg.ENV,
				API_ENDPOINT:       cfg.API_ENDPOINT + "/api",
				COLLECTOR_ENDPOINT: cfg.COLLECTOR_ENDPOINT,
				CUBEJS_ENDPOINT:    cfg.CUBEJS_ENDPOINT,
				MANAGED_RMD:        cfg.MANAGED_RMD,
			}

			jsonConfig, _ := json.Marshal(consoleConfig)
			// log.Printf("window.Config = %v;", string(jsonConfig))

			w.Write([]byte(fmt.Sprintf("window.Config = %v;", string(jsonConfig))))
			return
		}

		if _, err := os.Stat(consoleDir + r.URL.Path); os.IsNotExist(err) {
			http.StripPrefix(r.URL.Path, fs).ServeHTTP(w, r)
		} else {
			fs.ServeHTTP(w, r)
		}
	})

	return api
}
