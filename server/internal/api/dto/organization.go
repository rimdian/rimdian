package dto

import (
	"time"
)

type OrganizationListResult struct {
	Organizations []*OrganizationResult `json:"organizations"`
}

type OrganizationResult struct {
	ID                      string    `json:"id"`
	Name                    string    `json:"name"`
	Currency                string    `json:"currency"`
	DataProtectionOfficerID string    `db:"dpo_id" json:"dpo_id"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
	ImOwner                 bool      `json:"im_owner"`
}

type OrganizationProfileResult struct {
	Organization *OrganizationResult `json:"organization"`
}

type OrganizationProfile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type OrganizationCreate struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	Currency                string `json:"currency"`
	DataProtectionOfficerID string `json:"dpo_id"`
}
