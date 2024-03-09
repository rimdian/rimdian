package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func ComputeSegmentsForGivenUsers(ctx context.Context, pipe Pipeline) {

	spanCtx, span := trace.StartSpan(ctx, "ComputeSegmentsForGivenUsers")
	defer span.End()

	// abort if no user ids found
	if len(pipe.GetUserIDs()) == 0 {
		return
	}

	// fetch segments
	segments, err := pipe.Repo().ListSegments(spanCtx, pipe.GetWorkspace().ID, false)

	if err != nil {
		// err = eris.Wrap(err, "ComputeUserSegments")
		pipe.SetError("server", err.Error(), true)
		return
	}

	now := time.Now()

	// fetch users
	users, _, _, err := pipe.Repo().ListUsers(spanCtx, pipe.GetWorkspace(), &dto.UserListParams{
		WorkspaceID: pipe.GetWorkspace().ID,
		Limit:       100,
		UserIDs:     entity.StringPtr(strings.Join(pipe.GetUserIDs(), ",")),
	})

	if err != nil {
		pipe.SetError("server", err.Error(), true)
		return
	}

	// only keep not merged users
	activeUsers := []*entity.User{}
	activeUserIDs := []string{}
	for _, user := range users {
		if !user.IsMerged {
			activeUsers = append(activeUsers, user)
			activeUserIDs = append(activeUserIDs, user.ID)
		}
	}
	users = activeUsers

	// abort if no users
	if len(users) == 0 {
		return
	}

	// fetch user_segments
	userSegments, err := pipe.Repo().ListUserSegments(spanCtx, pipe.GetWorkspace().ID, activeUserIDs, nil)

	if err != nil {
		// err = eris.Wrap(err, "ComputeUserSegments")
		pipe.SetError("server", err.Error(), true)
		return
	}

	_, err = pipe.Repo().RunInTransactionForWorkspace(spanCtx, pipe.GetWorkspace().ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		for _, segment := range segments {

			// skip if segment is not active, or "_all"
			if segment.Status != entity.SegmentStatusActive || segment.ID == entity.SegmentAllUsersID {
				continue
			}

			matchingUserIDs := []*string{}

			switch segment.ID {
			case "anonymous":
				for _, user := range users {
					if !user.IsAuthenticated {
						matchingUserIDs = append(matchingUserIDs, &user.ID)
					}
				}
			case "authenticated":
				for _, user := range users {
					if user.IsAuthenticated {
						matchingUserIDs = append(matchingUserIDs, &user.ID)
					}
				}
			default:

				// segment with conditions
				matchingUserIDs, err = pipe.Repo().MatchSegmentUsers(ctx, pipe.GetWorkspace().ID, segment, activeUserIDs)

				if err != nil {
					return 500, err
				}
			}

			for _, userID := range activeUserIDs {

				// check if user is matching the segment
				isMatchingSegment := false
				for _, matchingUserID := range matchingUserIDs {
					if userID == *matchingUserID {
						isMatchingSegment = true
						break
					}
				}

				// check if user is already in the segment
				isAlreadyInSegment := false
				for _, us := range userSegments {
					if us.UserID == userID && us.SegmentID == segment.ID {
						isAlreadyInSegment = true
						break
					}
				}

				if isMatchingSegment {
					// if user is matching and not already in the segment, insert it
					if !isAlreadyInSegment {

						// insert new user_segment
						newUserSegment := entity.NewUserSegment(userID, segment.ID)
						if err = pipe.Repo().InsertUserSegment(ctx, newUserSegment, tx); err != nil {
							return 500, err
						}

						if err = pipe.InsertChildDataLog(spanCtx, "segment", "enter", userID, segment.ID, segment.ID, entity.UpdatedFields{}, now, tx); err != nil {
							return 500, err
						}
					}
				} else {
					// if user is not matching and already in the segment, delete it
					if isAlreadyInSegment {

						// delete user_segment
						if err = pipe.Repo().DeleteUserSegment(ctx, userID, segment.ID, tx); err != nil {
							return 500, err
						}

						if err = pipe.InsertChildDataLog(spanCtx, "segment", "exit", userID, segment.ID, segment.ID, entity.UpdatedFields{}, now, tx); err != nil {
							return 500, err
						}
					}
				}
			}
		}

		return 200, nil
	})

	if err != nil {
		pipe.SetError("server", err.Error(), true)
		return
	}

	return
}
