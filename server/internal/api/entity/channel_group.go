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
	ReservedChannelGroupIDs []string = []string{"direct", "not-mapped"}

	ErrChannelGroupIDInvalid        = eris.New("channel groupId is not valid")
	ErrChannelGroupIDRequired       = eris.New("channel group id is required")
	ErrChannelGroupIDReserved       = eris.New("channel group id is reserved")
	ErrChannelGroupNameRequired     = eris.New("channel group name is required")
	ErrChannelGroupColorRequired    = eris.New("channel group color is required")
	ErrChannelGroupStillHasChannels = eris.New("channel group still has channels attached to it")
)

type ChannelGroup struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (gr *ChannelGroup) Validate(groups ChannelGroups) error {

	// sanitize
	gr.ID = strings.TrimSpace(gr.ID)
	gr.Name = strings.TrimSpace(gr.Name)
	gr.Color = strings.TrimSpace(gr.Color)

	if gr.ID == "" {
		return ErrChannelGroupIDRequired
	}
	if gr.Name == "" {
		return ErrChannelGroupNameRequired
	}
	if gr.Color == "" {
		return ErrChannelGroupColorRequired
	}

	if govalidator.IsIn(gr.ID, ReservedChannelGroupIDs...) {
		return ErrChannelGroupIDReserved
	}

	return nil
}

type ChannelGroups []*ChannelGroup

func (x *ChannelGroups) Scan(val interface{}) error {

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

func (x ChannelGroups) Value() (driver.Value, error) {
	return json.Marshal(x)
}

var DefaultChannelGroups = []*ChannelGroup{
	{ID: "not-mapped", Name: "Not mapped", Color: "black"},
	{ID: "direct", Name: "Direct", Color: "gold"},
	// {ID: "conversion-rules", Name: "Conversion rules", Color: "gold"},
	{ID: "email-organic", Name: "Email organic", Color: "gold"},
	{ID: "search-organic", Name: "Search organic", Color: "green"},
	{ID: "social-organic", Name: "Social organic", Color: "green"},
	{ID: "video-organic", Name: "Video organic", Color: "green"},
	{ID: "content-marketing", Name: "Content marketing", Color: "green"},
	{ID: "social-paid", Name: "Social paid", Color: "blue"},
	{ID: "search-paid", Name: "Search paid", Color: "blue"},
	{ID: "shopping-engine", Name: "Shopping engine", Color: "blue"},
	{ID: "video-paid", Name: "Video paid", Color: "blue"},
	{ID: "email-acquisition", Name: "Email acquisition", Color: "blue"},
	{ID: "referral", Name: "Referral", Color: "purple"},
	{ID: "email-retention", Name: "Email retention", Color: "magenta"},
	{ID: "phone", Name: "Phone", Color: "magenta"},
	{ID: "retargeting", Name: "Retargeting", Color: "cyan"},
	{ID: "display-banner", Name: "Display banner", Color: "cyan"},
	{ID: "display-video", Name: "Display video", Color: "cyan"},
	{ID: "print", Name: "Print", Color: "cyan"},
	{ID: "tv", Name: "TV", Color: "cyan"},
	{ID: "radio", Name: "Radio", Color: "cyan"},
	{ID: "cashback", Name: "Cashback", Color: "orange"},
	{ID: "coupon", Name: "Coupon", Color: "orange"},
	{ID: "in-store", Name: "In-store", Color: "red"},
	{ID: "other", Name: "Other advertising", Color: ""},
}
