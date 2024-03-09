package dto

import (
	"github.com/rimdian/rimdian/internal/api/entity"
)

// create / update domain "Data Transfer Object"
type Domain struct {
	ID              string               `json:"id"`
	WorkspaceID     string               `json:"workspace_id"`
	Type            string               `json:"type"` // web / app / marketplace
	Name            string               `json:"name"`
	Hosts           []*entity.DomainHost `json:"hosts"`
	ParamsWhitelist []string             `json:"params_whitelist"`
	// BrandKeywordsAsDirect bool                 `json:"brandKeywordsAsDirect"`
	// BrandKeywords         []string             `json:"brandKeywords"`
	// HomepagePaths         []string             `json:"homepagePaths"`
}

type DomainDelete struct {
	ID                string `json:"id"`
	WorkspaceID       string `json:"workspace_id"`
	MigrateToDomainID string `json:"migrate_to_domain_id"`
}
