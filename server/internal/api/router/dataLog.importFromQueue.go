package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/dto"
	"go.opencensus.io/plugin/ochttp"
)

// Receives a DataLogImportFromQueue payload from the Queue and imports the data
// non-200 status will provoque a retry from the Queue
// unrecoverable errors returns a 200 status to drop the data and avoid retries

func (api *API) DataLogImportFromQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/dataLog.importFromQueue")

	// Read body
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "cannot read body", http.StatusUnprocessableEntity)
		return
	}

	dataLogInQueue := &dto.DataLogInQueue{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, dataLogInQueue); err != nil {
		log.Printf("cannot unmarshal data import from queue: %v", err)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(err.Error()))
		return
	}

	// in dev, we don't process it right away, but store it in-memory
	// we can then execute them sequentially with the endpoint /dev.execDataLogImportFromQueue
	if api.Config.ENV == entity.ENV_DEV {
		api.Svc.DevAddDataImportToQueue(dataLogInQueue)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data import added in-memory queue"))
		return
	}

	result := api.Svc.DataLogImportFromQueue(ctx, dataLogInQueue)

	if result.HasError {
		log.Printf("data log import has error: %+v\n", result)
	}

	// return code 500 to retry
	if result.QueueShouldRetry {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(result.Error))
		return
	}

	ReturnJSON(w, http.StatusOK, result)
}
