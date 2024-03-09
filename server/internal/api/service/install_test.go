package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
)

func TestServiceImpl_InstallOrVerifyServer(t *testing.T) {
	type fields struct {
		Config *entity.Config
		Repo   repository.Repository
	}
	type args struct {
		ctx context.Context
	}

	installed := false
	orgID := "acme"

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantSuccess bool
		wantErr     bool
	}{
		{
			name: "should skip when DB maintenance is true",
			fields: fields{
				Config: &entity.Config{
					DB_MAINTENANCE:    true,
					ORGANIZATION_ID:   orgID,
					ORGANIZATION_NAME: orgID,
				},
				Repo: nil,
			},
			args:        args{ctx: context.Background()},
			wantSuccess: false,
			wantErr:     false,
		},
		{
			name: "should install on first pass",
			fields: fields{
				Config: &entity.Config{
					DB_MAINTENANCE:    false,
					ORGANIZATION_ID:   orgID,
					ORGANIZATION_NAME: orgID,
				},
				Repo: &repository.RepositoryMock{
					GetSystemConnectionFunc: func(ctx context.Context) (*sql.Conn, error) {
						return nil, nil
					},
					GetSettingsFunc: func(ctx context.Context) (*entity.Settings, error) {

						// on first pass settings dont exist
						// after install settings are available
						if installed {
							return &entity.Settings{
								InstalledVersion: common.Version,
							}, nil
						}

						installed = true
						return nil, entity.ErrSettingsTableNotFound
					},
					InstallFunc: func(ctx context.Context, rootAccount *entity.Account, defaultOrg *entity.Organization) error {
						return nil
					},
				},
			},
			args:        args{ctx: context.Background()},
			wantSuccess: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &ServiceImpl{
				Config: tt.fields.Config,
				Repo:   tt.fields.Repo,
			}
			success, err := svc.InstallOrVerifyServer(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceImpl.InstallOrVerifyServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if success != tt.wantSuccess {
				t.Errorf("ServiceImpl.InstallOrVerifyServer() = %v, want %v", success, tt.wantSuccess)
			}
		})
	}
}
