package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/iancoleman/strcase"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"github.com/teris-io/shortid"
)

func (svc *ServiceImpl) SegmentCreate(ctx context.Context, accountID string, segmentDTO *dto.Segment) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, segmentDTO.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "SegmentCreate")
	}

	// fetch existing segments
	existingSegments, err := svc.Repo.ListSegments(ctx, workspace.ID, false)

	if err != nil {
		return 500, err
	}

	// convert DTO to entity
	segment := &entity.Segment{
		Name:            segmentDTO.Name,
		Color:           segmentDTO.Color,
		ParentSegmentID: segmentDTO.ParentSegmentID,
		Tree:            segmentDTO.Tree,
		Timezone:        segmentDTO.Timezone,
		Status:          entity.SegmentStatusBuilding,
		Version:         1,
	}

	// generate id from name + shortid

	sid, err := shortid.New(1, shortid.DefaultABC, 777)
	if err != nil {
		return 500, eris.Wrap(err, "SegmentCreate shortid")
	}

	id, err := sid.Generate()
	if err != nil {
		return 500, eris.Wrap(err, "SegmentCreate shortid")
	}

	segment.ID = fmt.Sprintf("%v_%v", strings.ReplaceAll(strcase.ToSnake(strings.TrimSpace(segmentDTO.Name)), "_", ""), id)

	// validate segment
	if err := segment.Validate(existingSegments, entity.GenerateSchemas(workspace.InstalledApps)); err != nil {
		return 400, err
	}

	// preview segment
	_, segment.GeneratedSQL, segment.GeneratedArgs, err = svc.Repo.PreviewSegment(ctx, workspace.ID, segment.ParentSegmentID, &segment.Tree, segment.Timezone)

	if err != nil {
		return 500, err
	}

	if segment.GeneratedArgs == nil {
		segment.GeneratedArgs = []interface{}{}
	}

	// insert segment
	_, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (txCode int, txErr error) {

		if err := svc.Repo.InsertSegment(ctx, segment, tx); err != nil {
			return 500, err
		}

		// create a task to compute user segments
		state := entity.NewTaskState()
		state.Workers[0] = entity.TaskWorkerState{
			"segment_id":      segment.ID,
			"segment_version": float64(segment.Version), // float64 because of JSON
			"current_step":    entity.RecomputeSegmentStepMatchUsers,
		}

		taskExec := &entity.TaskExec{
			TaskID:          entity.TaskKindRecomputeSegment,
			Name:            fmt.Sprintf("Recompute segment %v, version %v", segment.Name, segment.Version),
			MultipleExecKey: entity.StringPtr(segment.ID),       // deduplicate tasks by segment ID
			OnMultipleExec:  entity.OnMultipleExecAbortExisting, // aborting existing task if segment is updated
			State:           state,
		}

		return svc.doTaskCreate(ctx, workspace.ID, taskExec)
	})

	if err != nil {
		return 500, err
	}

	return 200, nil
}

func (svc *ServiceImpl) SegmentUpdate(ctx context.Context, accountID string, segmentDTO *dto.Segment) (code int, err error) {
	// compare generated SQL and args with existing segment and increment version if changed
	// set status to building
	// create a task to compute user segment

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, segmentDTO.WorkspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, err
		}
		return 500, eris.Wrap(err, "SegmentCreate")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "SegmentCreate")
	}

	if !isAccount {
		return 400, eris.New("account is not part of the organization")
	}

	// reject ids "_all", "authenticated" and "anonymous"
	if segmentDTO.ID == "_all" || segmentDTO.ID == "authenticated" || segmentDTO.ID == "anonymous" {
		return 400, eris.New("SegmentCreate: invalid segment id")
	}

	// fetch existing segments
	existingSegments, err := svc.Repo.ListSegments(ctx, workspace.ID, false)

	if err != nil {
		return 500, err
	}

	// find existing segment
	existingSegment := &entity.Segment{}

	for _, seg := range existingSegments {
		if seg.ID == segmentDTO.ID {
			existingSegment = seg
			break
		}
	}

	if existingSegment.ID == "" {
		return 400, eris.New("SegmentCreate: segment not found")
	}

	// convert DTO to entity
	updatedSegment := &entity.Segment{
		ID:              existingSegment.ID,
		Name:            segmentDTO.Name,
		Color:           segmentDTO.Color,
		ParentSegmentID: segmentDTO.ParentSegmentID,
		Tree:            segmentDTO.Tree,
		Timezone:        segmentDTO.Timezone,
		Status:          entity.SegmentStatusBuilding,
		Version:         existingSegment.Version + 1,
	}

	// validate segment
	if err := updatedSegment.Validate(existingSegments, entity.GenerateSchemas(workspace.InstalledApps)); err != nil {
		return 400, err
	}

	// preview segment
	_, updatedSegment.GeneratedSQL, updatedSegment.GeneratedArgs, err = svc.Repo.PreviewSegment(ctx, workspace.ID, updatedSegment.ParentSegmentID, &updatedSegment.Tree, updatedSegment.Timezone)

	if err != nil {
		return 500, err
	}

	if updatedSegment.GeneratedArgs == nil {
		updatedSegment.GeneratedArgs = []interface{}{}
	}

	// compare generated SQL and args with existing segment to see if it changed

	hasChanges := false

	if existingSegment.GeneratedSQL != updatedSegment.GeneratedSQL {
		hasChanges = true
	}

	if len(existingSegment.GeneratedArgs) != len(updatedSegment.GeneratedArgs) {
		hasChanges = true
	}

	for i, arg := range existingSegment.GeneratedArgs {
		if arg != updatedSegment.GeneratedArgs[i] {
			hasChanges = true
			break
		}
	}

	if !hasChanges {
		return 200, nil
	}

	// update segment

	_, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (txCode int, txErr error) {

		if err := svc.Repo.UpdateSegment(ctx, updatedSegment, tx); err != nil {
			return 500, err
		}

		// create a task to compute user segments
		state := entity.NewTaskState()
		state.Workers[0] = entity.TaskWorkerState{
			"segment_id":      updatedSegment.ID,
			"segment_version": float64(updatedSegment.Version), // float64 because of JSON
			"current_step":    entity.RecomputeSegmentStepMatchUsers,
		}

		taskExec := &entity.TaskExec{
			TaskID:          entity.TaskKindRecomputeSegment,
			Name:            fmt.Sprintf("Recompute segment %v, version %v", updatedSegment.Name, updatedSegment.Version),
			MultipleExecKey: entity.StringPtr(updatedSegment.ID), // deduplicate tasks by segment ID
			OnMultipleExec:  entity.OnMultipleExecAbortExisting,  // aborting existing task if segment is updated
			State:           state,
		}

		return svc.doTaskCreate(ctx, workspace.ID, taskExec)
	})

	if err != nil {
		return 500, err
	}

	return 200, nil
}

func (svc *ServiceImpl) SegmentDelete(ctx context.Context, accountID string, deleteSegmentDTO *dto.DeleteSegment) (code int, err error) {

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, deleteSegmentDTO.WorkspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, err
		}
		return 500, eris.Wrap(err, "SegmentDelete")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "SegmentDelete")
	}

	if !isAccount {
		return 400, eris.New("account is not part of the organization")
	}

	// reject ids "_all", "authenticated" and "anonymous"
	if deleteSegmentDTO.ID == "_all" || deleteSegmentDTO.ID == "authenticated" || deleteSegmentDTO.ID == "anonymous" {
		return 400, eris.New("SegmentCreate: invalid segment id")
	}

	// mark status as deleted
	if err = svc.Repo.DeleteSegment(ctx, workspace.ID, deleteSegmentDTO.ID); err != nil {
		return 500, err
	}

	return 200, nil
}

func (svc *ServiceImpl) SegmentPreview(ctx context.Context, accountID string, params *dto.SegmentPreviewParams) (result *dto.SegmentPreviewResult, code int, err error) {

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, params.WorkspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "SegmentPreview")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return nil, code, eris.Wrap(err, "SegmentPreview")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	// validate segment filter
	if err := params.Tree.Validate(entity.GenerateSchemas(workspace.InstalledApps)); err != nil {
		return nil, 400, err
	}

	// fetch segment preview
	result = &dto.SegmentPreviewResult{}

	result.Count, result.SQL, result.Args, err = svc.Repo.PreviewSegment(ctx, workspace.ID, params.ParentSegmentID, params.Tree, params.Timezone)

	if err != nil {
		return nil, 500, err
	}

	return result, 200, nil
}

func (svc *ServiceImpl) SegmentList(ctx context.Context, accountID string, params *dto.SegmentListParams) (result *dto.SegmentListResult, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "SegmentList")
	}

	// fetch segments
	result = &dto.SegmentListResult{}

	result.Segments, err = svc.Repo.ListSegments(ctx, workspace.ID, params.WithUsersCount)

	if err != nil {
		return nil, 500, err
	}

	return result, 200, nil
}
