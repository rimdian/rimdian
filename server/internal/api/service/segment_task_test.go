package service

// import (
// 	"context"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/rimdian/rimdian/internal/api/entity"
// 	"github.com/rimdian/rimdian/internal/api/repository"
// )

// func TestServiceImpl_UserSegment(t *testing.T) {

// 	cfgSecretKey := "12345678901234567890123456789012"
// 	orgID := "testing"
// 	workspaceID := fmt.Sprintf("%v_%v", orgID, "demoecommerce")

// 	demoWorkspace, err := entity.GenerateDemoWorkspace(workspaceID, entity.WorkspaceDemoOrder, orgID, cfgSecretKey)

// 	if err != nil {
// 		t.Fatalf("generate demo workspace err %v", err)
// 	}

// 	t.Run("should match & add users to a segment", func(t *testing.T) {

// 		// create a context with a 20s timeout
// 		ctxWithTimeout := context.Background()
// 		ctxWithTimeout, cancel := context.WithTimeout(ctxWithTimeout, 20*time.Second)
// 		defer cancel()

// 		authSegment := entity.DefaultAuthenticatedSegment
// 		workerID := 0

// 		state := entity.NewTaskState()
// 		state.Workers[0] = entity.TaskWorkerState{
// 			"segment_id":      authSegment.ID,
// 			"segment_version": float64(authSegment.Version), // float64 because of JSON
// 			"current_step":    entity.RecomputeSegmentStepMatchUsers,
// 		}

// 		taskExec := &entity.TaskExec{
// 			ID:              "task-exec-id",
// 			Kind:            entity.TaskKindRecomputeSegment,
// 			Name:            "recompute segment",
// 			MultipleExecKey: entity.StringPtr("segment_id_version_1"),
// 			OnMultipleExec:  entity.OnMultipleExecAbortExisting,
// 			State:           state,
// 		}

// 		repoMock := &repository.RepositoryMock{
// 			GetSegmentFunc: func(ctx context.Context, workspaceID, segmentID string) (*entity.Segment, error) {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segmentID) {
// 					return nil, fmt.Errorf("want GetSegment args to be %v, %v, got %v, %v", demoWorkspace.ID, authSegment.ID, workspaceID, segmentID)
// 				}
// 				return &authSegment, nil
// 			},
// 			ClearUserSegmentQueueFunc: func(ctx context.Context, workspaceID, segmentID string, segmentVersion int) error {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segmentID) || (authSegment.Version != segmentVersion) {
// 					return fmt.Errorf("want ClearUserSegmentQueue args to be %v, %v, %v, got %v, %v, %v", demoWorkspace.ID, segmentID, segmentVersion, workspaceID, segmentID, segmentVersion)
// 				}
// 				return nil
// 			},
// 			EnqueueMatchingSegmentUsersFunc: func(ctx context.Context, workspaceID string, segment *entity.Segment) (int, int, error) {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segment.ID) || (authSegment.Version != segment.Version) {
// 					return 0, 0, fmt.Errorf("want EnqueueMatchingSegmentUsers args to be %v, %v, %v, got %v, %v, %v", demoWorkspace.ID, segment.ID, segment.Version, workspaceID, segment.ID, segment.Version)
// 				}
// 				return 1, 0, nil
// 			},
// 			EnterUserSegmentFromQueueFunc: func(ctx context.Context, workspaceID, segmentID string, segmentVersion int) error {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segmentID) || (authSegment.Version != segmentVersion) {
// 					return fmt.Errorf("want EnterUserSegmentFromQueue args to be %v, %v, %v, got %v, %v, %v", demoWorkspace.ID, segmentID, segmentVersion, workspaceID, segmentID, segmentVersion)
// 				}
// 				return nil
// 			},
// 			ExitUserSegmentFromQueueFunc: func(ctx context.Context, workspaceID, segmentID string, segmentVersion int) error {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segmentID) || (authSegment.Version != segmentVersion) {
// 					return fmt.Errorf("want ExitUserSegmentFromQueue args to be %v, %v, %v, got %v, %v, %v", demoWorkspace.ID, segmentID, segmentVersion, workspaceID, segmentID, segmentVersion)
// 				}
// 				return nil
// 			},
// 			ActivateSegmentFunc: func(ctx context.Context, workspaceID, segmentID string, segmentVersion int) (bool, error) {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segmentID) || (authSegment.Version != segmentVersion) {
// 					return false, fmt.Errorf("want ActivateSegment args to be %v, %v, %v, got %v, %v, %v", demoWorkspace.ID, segmentID, segmentVersion, workspaceID, segmentID, segmentVersion)
// 				}
// 				return true, nil
// 			},
// 			GetUserSegmentQueueRowsForWorkerFunc: func(ctx context.Context, workspaceID, segmentID string, segmentVersion, workerID, limit int) ([]*entity.UserSegmentQueue, error) {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segmentID) || (authSegment.Version != segmentVersion) {
// 					return nil, fmt.Errorf("want GetUserSegmentQueueRowsForWorker args to be %v, %v, %v, got %v, %v, %v", demoWorkspace.ID, segmentID, segmentVersion, workspaceID, segmentID, segmentVersion)
// 				}
// 				return []*entity.UserSegmentQueue{
// 					{
// 						UserID:         "user-id",
// 						Enters:         1,
// 						SegmentVersion: authSegment.Version,
// 					},
// 				}, nil
// 			},
// 			InsertSegmentItemTimelinesFunc: func(ctx context.Context, workspaceID, segmentID string, segmentVersion int, taskID string, isEnter bool, createdAt time.Time) error {
// 				if (workspaceID != demoWorkspace.ID) || (authSegment.ID != segmentID) || (authSegment.Version != segmentVersion) {
// 					return fmt.Errorf("want InsertSegmentItemTimelines args to be %v, %v, %v, got %v, %v, %v", demoWorkspace.ID, segmentID, segmentVersion, workspaceID, segmentID, segmentVersion)
// 				}
// 				return nil
// 			},
// 			// RunInTransactionForWorkspaceFunc: func(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (int, error) {
// 			// 	return f(ctx, nil)
// 			// },
// 			// DeleteUserSegmentQueueRowFunc: func(ctx context.Context, workspaceID, segmentID string, segmentVersion int, userID string, tx *sql.Tx) error {
// 			// 	return nil
// 			// },
// 		}

// 		svc := &ServiceImpl{
// 			Config: &entity.Config{SECRET_KEY: "12345678901234567890123456789012"},
// 			Repo:   repoMock,
// 		}

// 		_, code, err := svc.TaskRecomputeSegment(ctxWithTimeout, demoWorkspace, taskExec, workerID)

// 		if err != nil {
// 			t.Fatalf("should not fail on TaskRecomputeSegment, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.GetSegmentCalls()) != 1 {
// 			t.Errorf("should call GetSegment once, did: %+v\n", len(repoMock.GetSegmentCalls()))
// 		}
// 		if len(repoMock.ClearUserSegmentQueueCalls()) != 1 {
// 			t.Errorf("should call ClearUserSegmentQueue once, did: %+v\n", len(repoMock.ClearUserSegmentQueueCalls()))
// 		}
// 		if len(repoMock.EnqueueMatchingSegmentUsersCalls()) != 1 {
// 			t.Errorf("should call EnqueueMatchingSegmentUsers once, did: %+v\n", len(repoMock.EnqueueMatchingSegmentUsersCalls()))
// 		}

// 		// step enter users
// 		state.Workers[0]["current_step"] = entity.RecomputeSegmentStepEnterUsers
// 		taskExec.State = state

// 		_, code, err = svc.TaskRecomputeSegment(ctxWithTimeout, demoWorkspace, taskExec, workerID)

// 		if err != nil {
// 			t.Fatalf("should not fail on TaskRecomputeSegment, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.EnterUserSegmentFromQueueCalls()) != 1 {
// 			t.Errorf("should call EnterUserSegmentFromQueue once, did: %+v\n", len(repoMock.EnterUserSegmentFromQueueCalls()))
// 		}

// 		// step exit users
// 		state.Workers[0]["current_step"] = entity.RecomputeSegmentStepExitUsers
// 		taskExec.State = state

// 		_, code, err = svc.TaskRecomputeSegment(ctxWithTimeout, demoWorkspace, taskExec, workerID)

// 		if err != nil {
// 			t.Fatalf("should not fail on TaskRecomputeSegment, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.ExitUserSegmentFromQueueCalls()) != 1 {
// 			t.Errorf("should call ExitUserSegmentFromQueue once, did: %+v\n", len(repoMock.ExitUserSegmentFromQueueCalls()))
// 		}

// 		if len(repoMock.ActivateSegmentCalls()) != 1 {
// 			t.Errorf("should call ActivateSegment once, did: %+v\n", len(repoMock.ActivateSegmentCalls()))
// 		}

// 		// step enter item_timelines

// 		state.Workers[0]["current_step"] = entity.RecomputeSegmentStepEnterDataLogs
// 		taskExec.State = state

// 		_, code, err = svc.TaskRecomputeSegment(ctxWithTimeout, demoWorkspace, taskExec, workerID)

// 		if err != nil {
// 			t.Fatalf("should not fail on TaskRecomputeSegment, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.InsertSegmentItemTimelinesCalls()) != 1 {
// 			t.Errorf("should call InsertSegmentItemTimelines once, did: %+v\n", len(repoMock.InsertSegmentItemTimelinesCalls()))
// 		}

// 		// step exit item_timelines

// 		state.Workers[0]["current_step"] = entity.RecomputeSegmentStepEnterDataLogs
// 		taskExec.State = state

// 		_, code, err = svc.TaskRecomputeSegment(ctxWithTimeout, demoWorkspace, taskExec, workerID)

// 		if err != nil {
// 			t.Fatalf("should not fail on TaskRecomputeSegment, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.InsertSegmentItemTimelinesCalls()) != 2 {
// 			t.Errorf("should call InsertSegmentItemTimelines twice, did: %+v\n", len(repoMock.InsertSegmentItemTimelinesCalls()))
// 		}

// 		// step timeline_automations

// 		state.Workers[0]["current_step"] = entity.RecomputeSegmentStepTimelineAutomations
// 		taskExec.State = state

// 		result, code, err := svc.TaskRecomputeSegment(ctxWithTimeout, demoWorkspace, taskExec, workerID)

// 		if err != nil {
// 			t.Fatalf("should not fail on TaskRecomputeSegment, got: %v, %v", code, err)
// 		}

// 		// if len(repoMock.GetUserSegmentQueueRowsForWorkerCalls()) != 1 {
// 		// 	t.Errorf("should call GetUserSegmentQueueRowsForWorker once, did: %+v\n", len(repoMock.GetUserSegmentQueueRowsForWorkerCalls()))
// 		// }

// 		// if len(repoMock.RunInTransactionForWorkspaceCalls()) != 1 {
// 		// 	t.Errorf("should call RunInTransactionForWorkspace once, did: %+v\n", len(repoMock.RunInTransactionForWorkspaceCalls()))
// 		// }

// 		// if len(repoMock.DeleteUserSegmentQueueRowCalls()) != 1 {
// 		// 	t.Errorf("should call DeleteUserSegmentQueueRow once, did: %+v\n", len(repoMock.DeleteUserSegmentQueueRowCalls()))
// 		// }

// 		// if result.UpdatedWorkerState["timeline_automations_processed"] != float64(1) {
// 		// 	t.Errorf("want timeline_automations_processed to be 1, got %+v\n", result.UpdatedWorkerState["timeline_automations_processed"])
// 		// }

// 		if result.IsDone != true {
// 			t.Errorf("want IsDone to be true, got %+v\n", result.IsDone)
// 		}

// 		if result.UpdatedWorkerState["current_step"] != entity.RecomputeSegmentStepDone {
// 			t.Errorf("want current_step to be %v, got %+v\n", entity.RecomputeSegmentStepDone, result.UpdatedWorkerState["current_step"])
// 		}
// 	})

// 	t.Run("should match existing users to a segment and do nothing", func(t *testing.T) {
// 	})

// 	t.Run("should not match existing users and exit from a segment", func(t *testing.T) {
// 	})

// 	t.Run("should not match users and do nothing", func(t *testing.T) {
// 	})

// }
