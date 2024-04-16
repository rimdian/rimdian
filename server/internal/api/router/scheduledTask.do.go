package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/plugin/ochttp"
)

// Receives a task job from the Queue
// non-200 status will provoque a retry from the Queue
// unrecoverable errors returns a 200 status to drop the data and avoid retries

func (api *API) ScheduledTaskDo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/scheduledTask.do")

	// Read body

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "cannot read body", http.StatusUnprocessableEntity)
		return
	}

	payload := &entity.ScheduledTask{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, payload); err != nil {
		w.WriteHeader(http.StatusOK)
		log.Printf("unmarshal scheduled task payload err: %v", err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	result := api.Svc.ScheduledTaskDo(ctx, payload)

	if result.HasError {
		log.Printf("scheduled task has error: %+v\n", result)
	}

	// return code 500 to retry
	if result.QueueShouldRetry {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(result.Error))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
