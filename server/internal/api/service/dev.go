package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/eapache/go-resiliency/semaphore"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
)

var ErrDevOnly = eris.New("dev env only")

// clean and reinstall tables
func (svc *ServiceImpl) DevResetDB(ctx context.Context) (err error) {
	if svc.Config.ENV != entity.ENV_DEV {
		return ErrDevOnly
	}

	rootAccount, err := entity.GenerateRootAccount(svc.Config)

	if err != nil {
		return eris.Wrap(err, "DevResetDB")
	}

	defaultOrganization, err := entity.GenerateDefaultOrganization(svc.Config.ORGANIZATION_ID, svc.Config.ORGANIZATION_NAME)

	if err != nil {
		return eris.Wrap(err, "DevResetDB")
	}

	return svc.Repo.DevResetDB(ctx, rootAccount, defaultOrganization)
}

// fetch a not-yet-done task and execute its workers in parallel
func (svc *ServiceImpl) DevExecTaskWithWorkers(ctx context.Context, workspaceID string) (code int, err error) {
	if svc.Config.ENV != entity.ENV_DEV {
		return 400, ErrDevOnly
	}

	workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, eris.Wrapf(err, "workspace %v not found", workspaceID)
		}
		return 500, eris.Wrap(err, "DevExecTaskWithWorkers")
	}

	// fetch taskExec
	taskExec, err := svc.Repo.GetTaskExec(ctx, workspace.ID, entity.TaskExecIDDev)

	if err != nil {
		if eris.Is(err, entity.ErrTaskExecNotFound) {
			return 400, eris.Wrap(err, "DevExecTaskWithWorkers")
		}
		return 500, err
	}

	// launch workers in parallel and wait for all to complete
	var wg sync.WaitGroup

	for workerID := range taskExec.State.Workers {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			// run worker
			svc.TaskExecDo(ctx, workspace.ID, &dto.TaskExecRequestPayload{
				TaskExecID: taskExec.ID,
				WorkerID:   id,
				JobID:      entity.TaskExecJobIDDev,
			})
		}(workerID)
	}

	wg.Wait()
	return 200, nil
}

// fetch a not-yet-done dataLog and process it
func (svc *ServiceImpl) DevExecDataImportFromQueue(ctx context.Context, concurrency int) (code int, err error) {
	if svc.Config.ENV != entity.ENV_DEV {
		return 400, ErrDevOnly
	}

	log.Printf("data import with concurrency %v", concurrency)

	if concurrency <= 1 {
		// fetch data import to process
		dataLogInQueue := svc.DevDataImportQueue.GetOne()
		if dataLogInQueue == nil {
			return 400, eris.New("no more data imports to process")
		}
		result := svc.DataLogImportFromQueue(ctx, dataLogInQueue)
		if result.HasError {
			log.Printf("DataImportFromQueue error: %v\n", result.Error)
			return 500, eris.Errorf("DataImportFromQueue error: %v", result.Error)
		}
		log.Printf("DataImportFromQueue result: %+v\n", result)
		return 200, nil
	}

	// limit execution to 25secs
	secs := 25
	multiplier := time.Duration(secs)
	duration := time.Second

	bgCtx := context.Background()
	ctx, cancel := context.WithDeadline(bgCtx, time.Now().Add(multiplier*duration))

	defer cancel()

	// max tickets in parallel
	sem := semaphore.New(concurrency, 6*time.Second)

	var maxTimeTaken time.Duration
	var wg sync.WaitGroup

	shouldContinue := true

	for shouldContinue {

		if err := sem.Acquire(); err != nil {
			continue
		}

		// tickets can wait in the Acquire() before the shouldContinue changed to false
		// check if we can still continue
		if !shouldContinue {
			log.Println("abort, shouldContinue == false")
			continue
		}

		wg.Add(1)

		go func() {
			defer wg.Done()
			defer sem.Release()

			startTimer := time.Now()

			// fetch data import to process
			dataLogInQueue := svc.DevDataImportQueue.GetOne()
			if dataLogInQueue == nil {
				time.Sleep(3 * time.Second) // no data wait a bit before checking again
				return
			}
			log.Printf("remaining: %v, processing data_log %v", len(svc.DevDataImportQueue.List), dataLogInQueue.ID)

			result := svc.DataLogImportFromQueue(ctx, dataLogInQueue)

			// requeue internal errors
			if result.HasError && result.QueueShouldRetry {
				dataLogInQueue.IsReplay = true
				svc.DevDataImportQueue.Add(dataLogInQueue)
				time.Sleep(3 * time.Second) // wait a bit before checking again
			}

			// log.Printf("data import success: %v, will retry: %v, errors: %+v\n", result.Success, result.WillRetry, result.Errors)

			timeTaken := time.Since(startTimer)
			if timeTaken > maxTimeTaken {
				maxTimeTaken = timeTaken
			}
		}()

		// still has enough time?
		if deadline, _ := ctx.Deadline(); time.Until(deadline) < maxTimeTaken+3 {
			shouldContinue = false
		}
	}

	wg.Wait()

	return 200, nil
}

// add async data import to the in-memory queue
func (svc *ServiceImpl) DevAddDataImportToQueue(dataLogInQUeue *common.DataLogInQueue) {
	svc.DevDataImportQueue.Add(dataLogInQUeue)
}
