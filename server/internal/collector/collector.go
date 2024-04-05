package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	logger "github.com/chi-middleware/logrus-logger"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	tld "github.com/jpillora/go-tld"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rimdian/rimdian/internal/common/cors"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

type Collector struct {
	Config                 *Config
	TaskOrchestratorClient taskorchestrator.Client
	NetClient              httpClient.HTTPClient
	Handler                http.Handler
	Logger                 *logrus.Logger
}

var (
	DataImportEndpointPath = "/api/dataLog.importFromQueue"

	ErrEmptyBody          = eris.New("empty body")
	ErrJsonDataImportTask = eris.New("cannot json data task")

	ResponseQueued       string = "queued"
	QueueAsyncLive       string = "async-live"
	QueueAsyncHistorical string = "async-historical"
	QueueSync            string = "sync"
)

func NewCollector(cfg *Config, taskClient taskorchestrator.Client, netClient httpClient.HTTPClient, log *logrus.Logger) *Collector {

	r := chi.NewRouter()

	collector := &Collector{
		Config:                 cfg,
		TaskOrchestratorClient: taskClient,
		NetClient:              netClient,
		Handler:                r,
		Logger:                 log,
	}

	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(logger.Logger("collector", log))

	r.Use(cors.Middleware)

	if cfg.ENV != entity.ENV_DEV {
		r.Use(middleware.Recoverer)
	}

	// middleware that reads an eventual paseto token and sets context
	r.Use(auth.MiddlewarePasetoExtractor(collector.Logger, cfg.API_ENDPOINT, cfg.SECRET_KEY))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Data collector running"))
	})

	// receives a data payload to push into the Task Queue
	r.Post("/data", collector.AsyncHistorical) // historical
	r.Post("/live", collector.AsyncLive)       // JS SDK
	r.Post("/bypass", collector.Sync)          // bypass task queue and hit API directly
	r.Post("/sync", collector.Sync)            // bypass task queue and hit API directly

	// message tracking
	r.Get("/double-opt-in", collector.DoubleOptIn)
	r.Get("/unsubscribe-email", collector.UnsubscribeEmail)
	r.Get("/open-email", collector.OpenEmail)

	return collector
}

func (collector *Collector) AsyncLive(w http.ResponseWriter, r *http.Request) {
	collector.ForwardData(QueueAsyncLive, w, r)
}

func (collector *Collector) AsyncHistorical(w http.ResponseWriter, r *http.Request) {
	collector.ForwardData(QueueAsyncHistorical, w, r)
}

// sync will bypass the tasks queue and hit the API directly
func (collector *Collector) Sync(w http.ResponseWriter, r *http.Request) {
	collector.ForwardData(QueueSync, w, r)
}

// email double opt-in link
func (collector *Collector) DoubleOptIn(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	// 1. verify token
	// 2. posts a message datalog to confirm email
	// 3. show success message

	w.Write([]byte("TODO"))
}

// email unsubscribe link
func (collector *Collector) UnsubscribeEmail(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	// 1. verify token
	// 2. posts a message datalog to unsubscribe email
	// 3. show success message

	w.Write([]byte("TODO"))
}

// email open tracking
func (collector *Collector) OpenEmail(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	// 1. verify token
	// 2. posts a message datalog to unsubscribe email
	// 3. show success message

	w.Write([]byte("TODO"))
}

func (collector *Collector) ForwardData(mode string, w http.ResponseWriter, r *http.Request) {

	rows, code, err := dto.NewDataLogInQueueFromRequest(r, collector.Config.SECRET_KEY)

	if err != nil {
		log.Printf("ForwardData err  %+v\n", err)
		http.Error(w, err.Error(), code)
		return
	}

	// log.Printf("data %+v\n", data)

	// Bypass task queue and import data directly into the API
	// used in dev to get eventual errors right away

	if mode == QueueSync {

		if len(rows) > 1 {
			log.Printf("Sync mode only supports one row at a time")
			http.Error(w, "Sync mode only supports one row at a time", 400)
			return
		}

		jsonData, err := json.Marshal(rows[0])

		if err != nil {
			log.Println(err)
			http.Error(w, "cannot json data task", http.StatusInternalServerError)
			return
		}

		endpoint := collector.Config.API_ENDPOINT + DataImportEndpointPath

		req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		// Forward known headers
		if r.Header.Get("Authorization") != "" {
			req.Header.Set("Authorization", r.Header.Get("Authorization"))
		}
		if r.Header.Get(dto.HeaderSignature) != "" {
			req.Header.Set(dto.HeaderSignature, r.Header.Get(dto.HeaderSignature))
		}
		if r.Header.Get(dto.HeaderOrigin) != "" {
			req.Header.Set(dto.HeaderOrigin, r.Header.Get(dto.HeaderOrigin))
		}
		if r.Header.Get(dto.HeaderOriginID) != "" {
			req.Header.Set(dto.HeaderOriginID, r.Header.Get(dto.HeaderOriginID))
		}
		if r.Header.Get(dto.HeaderReplayID) != "" {
			req.Header.Set(dto.HeaderReplayID, r.Header.Get(dto.HeaderReplayID))
		}

		resp, err := collector.NetClient.Do(req)

		if err != nil {
			log.Printf("Process batch sync error: %v", err)
			http.Error(w, "Process batch sync error", 500)
			return
		}

		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			log.Printf("post returned %v: %v", resp.StatusCode, string(b))
		}

		// return actual API data import response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(b)

		return
	}

	workspaceID := rows[0].Context.WorkspaceID

	// live queue for JS SDK by default
	queueName := collector.TaskOrchestratorClient.GetLiveQueueNameForWorkspace(workspaceID)

	// historical queue for historical data
	if mode == QueueAsyncHistorical {
		queueName = collector.TaskOrchestratorClient.GetHistoricalQueueNameForWorkspace(workspaceID)
	}

	// enqueue tasks, in parallel with wait group
	// if any task fails, return error
	wg := &sync.WaitGroup{}
	var taskError error

	for _, row := range rows {

		wg.Add(1)

		rowCopy := row
		go func(row *dto.DataLogInQueue) {
			defer wg.Done()

			cloudTask := &taskorchestrator.TaskRequest{
				QueueLocation: collector.Config.TASK_QUEUE_LOCATION,
				QueueName:     queueName,
				PostEndpoint:  fmt.Sprintf("%v%v", collector.Config.API_ENDPOINT, DataImportEndpointPath),
				Payload:       row,
			}

			// avoid duplicates in the queue, except for replays
			if !row.IsReplay {
				cloudTask.DeduplicationKey = &row.ID
			}

			// 28 secs max (graceful shutdown is at 30 secs)
			retry := retrier.New(retrier.ConstantBackoff(14, 2*time.Second), nil)

			errRetry := retry.Run(func() error {
				err := collector.TaskOrchestratorClient.PostRequest(context.Background(), cloudTask)
				// ignore error if it contains "AlreadyExists", its a duplicated task
				if err != nil && strings.Contains(err.Error(), "AlreadyExists") {
					return nil
				}
				return err
			})

			if errRetry != nil {
				taskError = errRetry
			}
		}(rowCopy)
	}

	wg.Wait()

	if taskError != nil {
		log.Printf("Task error: %v", taskError)
		http.Error(w, "Task error", 500)
		return
	}

	// read cookies and rewrite them from server for 12 months
	// except the session cookie, that should expire
	// don't do it on Safari anymore, they restrict it

	if !strings.Contains(r.UserAgent(), "Safari") {
		if u, err := tld.Parse(r.Header.Get("Referer")); err == nil {
			cookieNames := []string{
				"user",
				"device",
			}

			for _, cookieName := range cookieNames {
				if cook, err := r.Cookie("_rmd_" + cookieName); err == nil {
					cook.Domain = fmt.Sprintf(".%v.%v", u.Domain, u.TLD)
					cook.Expires = time.Now().AddDate(0, 12, 0)
					cook.Secure = true
					cook.Path = "/"
					cookStr := strings.ReplaceAll(cook.String(), "Domain=", "Domain=.")
					// log.Printf("cookie: %v", cookStr)
					w.Header().Add("Set-Cookie", cookStr)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ResponseQueued))
}
