package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
)

func TestServiceImpl_DataLog(t *testing.T) {

	cfgSecretKey := "12345678901234567890123456789012"

	orgID := "testing"
	workspaceID := fmt.Sprintf("%v_%v", orgID, "demoecommerce")

	demoWorkspace, err := entity.GenerateDemoWorkspace(workspaceID, entity.WorkspaceDemoOrder, orgID, cfgSecretKey)

	if err != nil {
		t.Fatalf("generate demo workspace err %v", err)
	}

	var webHost string

	for _, dom := range demoWorkspace.Domains {
		if dom.Type == entity.DomainWeb {
			webHost = dom.Hosts[0].Host
		}
	}

	t.Run("should reject invalid workspace", func(t *testing.T) {

		repoMock := &repository.RepositoryMock{
			GetWorkspaceFunc: func(ctx context.Context, workspaceID string) (*entity.Workspace, error) {
				return nil, sql.ErrNoRows
			},
		}

		svc := &ServiceImpl{
			Config: &entity.Config{SECRET_KEY: cfgSecretKey},
			Repo:   repoMock,
			Mailer: nil,
		}

		dataLogInQueue := &commonDTO.DataLogInQueue{
			Origin:   commonDTO.DataLogOriginClient,
			OriginID: webHost,
			Context: commonDTO.DataLogContext{
				WorkspaceID: "invalid",
				HeadersAndParams: commonDTO.MapOfStrings{
					"Origin": "www.apple.com",
				},
			},
			Item: "{}",
		}

		dataLogInQueue.ID = commonDTO.ComputeDataLogID(svc.Config.SECRET_KEY, dataLogInQueue.Origin, dataLogInQueue.Item)

		result := svc.DataLogImportFromQueue(context.Background(), dataLogInQueue)

		if !result.HasError {
			t.Fatalf("expected error")
		}

		if result.QueueShouldRetry {
			t.Fatalf("expected no retry %+v", result)
		}

		if result.Error != "DataLogImportFromQueue: workspace not found: invalid" {
			t.Fatalf("expected error")
		}
	})
}
