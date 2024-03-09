package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	DomainWeb         string = "web"
	DomainApp         string = "app"
	DomainMarketplace string = "marketplace"
	DomainTelephone   string = "telephone"
	DomainRetail      string = "retail"

	ErrDomainIDRequired   = eris.New("domain id is required")
	ErrDomainNameRequired = eris.New("domain name is required")
	ErrDomainTypeInvalid  = eris.New("domain type is not valid")
	ErrDomainHostRequired = eris.New("domain host is required")
	ErrDomainHostInvalid  = eris.New("domain host is not valid")
)

type DomainHost struct {
	Host       string  `json:"host"`
	PathPrefix *string `json:"path_prefix,omitempty"`
}

func (h *DomainHost) Validate() error {
	if !govalidator.IsHost(h.Host) {
		return eris.Wrapf(ErrDomainHostInvalid, "got: %v", h.Host)
	}

	if h.PathPrefix != nil {
		trimed := strings.TrimSpace(*h.PathPrefix)
		h.PathPrefix = &trimed
	}

	return nil
}
func (x *DomainHost) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), &x)
}

func (x DomainHost) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type Domain struct {
	ID              string        `json:"id"`
	Type            string        `json:"type"` // web / app / marketplace
	Name            string        `json:"name"`
	Hosts           []*DomainHost `json:"hosts"`
	ParamsWhitelist []string      `json:"params_whitelist"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	DeletedAt       *time.Time    `json:"deleted_at"`
}

func (dom *Domain) Validate() error {

	// sanitize
	dom.ID = strings.TrimSpace(dom.ID)
	dom.Name = strings.TrimSpace(dom.Name)

	if dom.ID == "" {
		return ErrDomainIDRequired
	}
	if !govalidator.IsIn(dom.Type, DomainWeb, DomainApp, DomainMarketplace, DomainTelephone, DomainRetail) {
		return ErrDomainTypeInvalid
	}
	if dom.Name == "" {
		return ErrDomainNameRequired
	}

	// web fields
	if dom.Type == DomainWeb {
		if len(dom.Hosts) == 0 {
			return ErrDomainHostRequired
		}

		for _, host := range dom.Hosts {
			if err := host.Validate(); err != nil {
				return err
			}
		}

		if dom.ParamsWhitelist == nil {
			dom.ParamsWhitelist = []string{}
		}

		for k, v := range dom.ParamsWhitelist {
			dom.ParamsWhitelist[k] = strings.TrimSpace(v)
		}
	}

	if dom.CreatedAt.IsZero() {
		dom.CreatedAt = time.Now().UTC()
	}
	if dom.UpdatedAt.IsZero() {
		dom.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (d *Domain) MarshalJSON() ([]byte, error) {

	type Alias Domain

	// make sure Hosts / ParamsWhitelist / HomepagePaths is not nil

	if d.Hosts == nil {
		d.Hosts = []*DomainHost{}
	}

	if d.ParamsWhitelist == nil {
		d.ParamsWhitelist = []string{}
	}

	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	})
}

type Domains []*Domain

func (x *Domains) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x Domains) Value() (driver.Value, error) {
	return json.Marshal(x)
}
