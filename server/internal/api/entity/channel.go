package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
)

var (
	OriginMatchOperatorEqual string = "equals"
	ChannelNotMapped         string = "not-mapped"

	ErrChannelIDRequired              = eris.New("channel id is required")
	ErrChannelIDInvalid               = eris.New("channel id is not valid")
	ErrChannelNameRequired            = eris.New("channel name is required")
	ErrChannelOriginsRequired         = eris.New("channel origins is required")
	ErrChannelOriginMathTypeInvalid   = eris.New("channel origin matchType is not valid")
	ErrChannelOriginValueInvalid      = eris.New("channel origin is not valid")
	ErrChannelOriginAlreadyMapped     = eris.New("channel origin is already mapped")
	ErrChannelVoucheCodeAlreadyMapped = eris.New("channel voucherCode is already mapped")
	ErrChannelVoucheCodeOriginInvalid = eris.New("channel voucherCode origin is not valid")
)

type Channel struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Origins      []*ChannelOrigin `json:"origins"`
	VoucherCodes []*VoucherCode   `json:"voucher_codes"`
	GroupID      string           `json:"group_id"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

func (ch *Channel) EnsureID(allChannels Channels) {

	// check channel id availability
	existingIDCount := 0

	for _, x := range allChannels {
		if strings.HasPrefix(ch.ID, x.ID) {
			existingIDCount += 1
		}
	}

	// increment suffix if id already exists
	if existingIDCount > 0 {
		existingIDCount += 1
		ch.ID = fmt.Sprintf("%v-%v", ch.ID, existingIDCount)
	}
}

func (ch *Channel) Validate(allChannels Channels, groups ChannelGroups) error {

	// sanitize
	ch.ID = strings.TrimSpace(ch.ID)
	ch.Name = strings.TrimSpace(ch.Name)
	ch.GroupID = strings.TrimSpace(ch.GroupID)

	if ch.ID == "" {
		return ErrChannelIDRequired
	}
	if ch.Name == "" {
		return ErrChannelNameRequired
	}

	groupExists := false
	for _, g := range groups {
		if g.ID == ch.GroupID {
			groupExists = true
		}
	}
	if !groupExists {
		return ErrChannelGroupIDInvalid
	}

	if ch.Origins == nil || len(ch.Origins) == 0 {
		return ErrChannelOriginsRequired
	}

	for _, origin := range ch.Origins {
		if err := origin.Validate(ch, allChannels); err != nil {
			return err
		}
	}

	if ch.VoucherCodes == nil {
		ch.VoucherCodes = []*VoucherCode{}
	}

	for _, code := range ch.VoucherCodes {
		code.Code = strings.TrimSpace(code.Code)
		code.OriginID = strings.TrimSpace(code.OriginID)
		if code.SetUTMCampaign != nil {
			trimedCampaign := strings.TrimSpace(*code.SetUTMCampaign)
			code.SetUTMCampaign = &trimedCampaign
		}
		if code.SetUTMContent != nil {
			trimedContent := strings.TrimSpace(*code.SetUTMContent)
			code.SetUTMContent = &trimedContent
		}
		if code.Description != nil {
			trimedDesc := strings.TrimSpace(*code.Description)
			code.Description = &trimedDesc
		}

		// verify that code isn't already mapped in another channel
		codeAlreadyMapped := false
		for _, x := range allChannels {
			for _, xx := range x.VoucherCodes {
				if xx.Code == code.Code && x.ID != ch.ID {
					codeAlreadyMapped = true
				}
			}
		}

		if codeAlreadyMapped {
			return ErrChannelVoucheCodeAlreadyMapped
		}

		// verify that origin exists in this channel
		originExists := false
		for _, origin := range ch.Origins {
			if origin.ID == code.OriginID {
				originExists = true
			}
		}

		if !originExists {
			return ErrChannelVoucheCodeOriginInvalid
		}
	}

	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = time.Now().UTC()
	}
	if ch.UpdatedAt.IsZero() {
		ch.UpdatedAt = time.Now().UTC()
	}

	return nil
}

type Channels []*Channel

func (x *Channels) Scan(val interface{}) error {

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

func (x Channels) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type VoucherCode struct {
	Code           string  `json:"code"`
	OriginID       string  `json:"origin_id"`
	SetUTMCampaign *string `json:"set_utm_campaign,omitempty"` // utm_campaign
	SetUTMContent  *string `json:"set_utm_content,omitempty"`  // utm_content
	Description    *string `json:"description,omitempty"`
}

type ChannelOrigin struct {
	ID            string  `json:"id"` // source / medium( / campaign)
	MatchOperator string  `json:"match_operator"`
	UTMSource     string  `json:"utm_source"`
	UTMMedium     string  `json:"utm_medium"`
	UTMCampaign   *string `json:"utm_campaign,omitempty"`
}

func (origin *ChannelOrigin) Validate(forChannel *Channel, allChannels []*Channel) error {

	origin.MatchOperator = strings.TrimSpace(origin.MatchOperator)
	origin.UTMSource = strings.TrimSpace(origin.UTMSource)
	origin.UTMMedium = strings.TrimSpace(origin.UTMMedium)

	// ensure ID
	origin.ID = fmt.Sprintf("%v / %v", origin.UTMSource, origin.UTMMedium)

	if origin.UTMCampaign != nil {
		trimedCampaign := strings.TrimSpace(*origin.UTMCampaign)
		origin.UTMCampaign = &trimedCampaign

		// append campaign to ID
		origin.ID = fmt.Sprintf("%v / %v", origin.ID, *origin.UTMCampaign)
	}

	if origin.MatchOperator != OriginMatchOperatorEqual {
		return ErrChannelOriginMathTypeInvalid
	}

	// verify that origin is not already mapped
	if found, _ := FindChannelFromOrigin(allChannels, origin.UTMSource, origin.UTMMedium, origin.UTMCampaign); found != nil && found.ID != forChannel.ID {
		// log.Printf("forChannel %+v\n", forChannel)
		// log.Printf("exists in %+v\n", found)
		campaign := ""
		if origin.UTMCampaign != nil {
			campaign = "/ " + *origin.UTMCampaign
		}
		return eris.Wrapf(ErrChannelOriginAlreadyMapped, "%v / %v %v exists in channel %v", origin.UTMSource, origin.UTMMedium, campaign, found.Name)
	}

	return nil
}

var DefaultChannels = []*Channel{
	{ID: "direct", Name: "Direct traffic", GroupID: "direct", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "direct / none", MatchOperator: OriginMatchOperatorEqual, UTMSource: "direct", UTMMedium: "none"},
		{ID: "www.google.com / brand-sem", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.google.com", UTMMedium: "brand-sem"},
		{ID: "www.google.com / brand-seo", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.google.com", UTMMedium: "brand-seo"},
		{ID: "www.bing.com / brand-sem", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.bing.com", UTMMedium: "brand-sem"},
		{ID: "www.bing.com / brand-seo", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.bing.com", UTMMedium: "brand-seo"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "facebook-organic", Name: "Facebook organic", GroupID: "social-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "facebook.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "facebook.com", UTMMedium: "referral"},
		{ID: "www.facebook.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.facebook.com", UTMMedium: "referral"},
		{ID: "m.facebook.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "m.facebook.com", UTMMedium: "referral"},
		{ID: "l.facebook.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "l.facebook.com", UTMMedium: "referral"},
		{ID: "lm.facebook.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "lm.facebook.com", UTMMedium: "referral"},
		{ID: "web.facebook.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "web.facebook.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "facebook-ads", Name: "Facebook Ads", GroupID: "social-paid", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.facebook.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.facebook.com", UTMMedium: "cpc"},
		{ID: "www.facebook.com / cpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.facebook.com", UTMMedium: "cpm"},
		{ID: "www.facebook.com / ocpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.facebook.com", UTMMedium: "ocpm"},
		{ID: "m.facebook.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "m.facebook.com", UTMMedium: "cpc"},
		{ID: "m.facebook.com / cpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "m.facebook.com", UTMMedium: "cpm"},
		{ID: "m.facebook.com / ocpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "m.facebook.com", UTMMedium: "ocpm"},
		{ID: "l.facebook.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "l.facebook.com", UTMMedium: "cpc"},
		{ID: "l.facebook.com / cpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "l.facebook.com", UTMMedium: "cpm"},
		{ID: "l.facebook.com / ocpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "l.facebook.com", UTMMedium: "ocpm"},
		{ID: "web.facebook.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "web.facebook.com", UTMMedium: "cpc"},
		{ID: "web.facebook.com / cpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "web.facebook.com", UTMMedium: "cpm"},
		{ID: "web.facebook.com / ocpm", MatchOperator: OriginMatchOperatorEqual, UTMSource: "web.facebook.com", UTMMedium: "ocpm"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "linkedin-organic", Name: "LinkedIn organic", GroupID: "social-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.linkedin.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.linkedin.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "linkedin-paid", Name: "LinkedIn Ads", GroupID: "social-paid", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.linkedin.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.linkedin.com", UTMMedium: "cpc"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "twitter", Name: "Twitter", GroupID: "social-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "twitter.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "twitter.com", UTMMedium: "referral"},
		{ID: "t.co / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "t.co", UTMMedium: "referral"},
		{ID: "tweetdeck.twitter.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "tweetdeck.twitter.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "youtube", Name: "Youtube", GroupID: "social-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.youtube.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.youtube.com", UTMMedium: "referral"},
		{ID: "m.youtube.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "m.youtube.com", UTMMedium: "referral"},
		{ID: "youtube.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "youtube.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "pinterest", Name: "Pinterest", GroupID: "social-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.pinterest.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.pinterest.com", UTMMedium: "referral"},
		{ID: "pinterest.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "pinterest.com", UTMMedium: "referral"},
		{ID: "com.pinterest / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "com.pinterest", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "email", Name: "Emails", GroupID: "email-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "mail.google.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "mail.google.com", UTMMedium: "referral"},
		{ID: "m.mg.mail.yahoo.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "m.mg.mail.yahoo.com", UTMMedium: "referral"},
		{ID: "mail.yahoo.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "mail.yahoo.com", UTMMedium: "referral"},
		{ID: "mail.aol.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "mail.aol.com", UTMMedium: "referral"},
		{ID: "outlook.live.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "outlook.live.com", UTMMedium: "referral"},
		{ID: "mail.live.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "mail.live.com", UTMMedium: "referral"},
		{ID: "go.mail.ru.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "go.mail.ru.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "vk", Name: "VK", GroupID: "social-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "vk.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "vk.com", UTMMedium: "referral"},
		{ID: "m.vk.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "m.vk.com", UTMMedium: "referral"},
		{ID: "new.vk.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "new.vk.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "google-organic", Name: "Google SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.google.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.google.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "google-ads", Name: "Google Ads", GroupID: "search-paid", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.google.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.google.com", UTMMedium: "cpc"},
		{ID: "google.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "google.com", UTMMedium: "cpc"},
		{ID: "com.google.android.googlequicksearchbox / ads", MatchOperator: OriginMatchOperatorEqual, UTMSource: "com.google.android.googlequicksearchbox", UTMMedium: "ads"},
		{ID: "com.google.android.gm / ads", MatchOperator: OriginMatchOperatorEqual, UTMSource: "com.google.android.gm", UTMMedium: "ads"},
		{ID: "googleads.g.doubleclick.net / ads", MatchOperator: OriginMatchOperatorEqual, UTMSource: "googleads.g.doubleclick.net", UTMMedium: "ads"},
		{ID: "googleads.g.doubleclick.net / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "googleads.g.doubleclick.net", UTMMedium: "referral"},
		{ID: "tpc.googlesyndication.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "tpc.googlesyndication.com", UTMMedium: "referral"},
		{ID: "tpc.googlesyndication.com / ads", MatchOperator: OriginMatchOperatorEqual, UTMSource: "tpc.googlesyndication.com", UTMMedium: "ads"},
		{ID: "www.googleadservices.com / ads", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.googleadservices.com", UTMMedium: "ads"},
		{ID: "googleadservices.com / ads", MatchOperator: OriginMatchOperatorEqual, UTMSource: "googleadservices.com", UTMMedium: "ads"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "google-images", Name: "Google images", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "images.google.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "images.google.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "bing-organic", Name: "Bing SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.bing.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.bing.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "bing-paid", Name: "Bing Ads", GroupID: "search-paid", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.bing.com / cpc", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.bing.com", UTMMedium: "cpc"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "instagram", Name: "Instagram organic", GroupID: "social-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.instagram.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.instagram.com", UTMMedium: "referral"},
		{ID: "instagram.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "instagram.com", UTMMedium: "referral"},
		{ID: "l.instagram.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "l.instagram.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "yahoo-organic", Name: "Yahoo SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "search.yahoo.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "search.yahoo.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "baidu-organic", Name: "Baidu SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.baidu.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.baidu.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "yandex-organic", Name: "Yandex SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "yandex.ru / organic", MatchOperator: OriginMatchOperatorEqual, UTMSource: "yandex.ru", UTMMedium: "organic"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "aol-organic", Name: "AOL SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "search.aol.com / organic", MatchOperator: OriginMatchOperatorEqual, UTMSource: "search.aol.com", UTMMedium: "organic"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "qwant-organic", Name: "Qwant SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.qwant.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.qwant.com", UTMMedium: "referral"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "ask-organic", Name: "Ask SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "www.ask.com / organic", MatchOperator: OriginMatchOperatorEqual, UTMSource: "www.ask.com", UTMMedium: "organic"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	{ID: "duckduckgo-organic", Name: "Duckduckgo SEO", GroupID: "search-organic", VoucherCodes: []*VoucherCode{}, Origins: []*ChannelOrigin{
		{ID: "duckduckgo.com / referral", MatchOperator: OriginMatchOperatorEqual, UTMSource: "duckduckgo.com", UTMMedium: "referral"},
		{ID: "duckduckgo / organic", MatchOperator: OriginMatchOperatorEqual, UTMSource: "duckduckgo", UTMMedium: "organic"},
	}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
}

func ExtractSourceMediumCampaignFromOrigin(origin string) (source string, medium string, campaign string, err error) {
	source = ""
	medium = ""
	campaign = ""

	parts := strings.Split(origin, " / ")

	// source medium path should contain at least a " / "
	if len(parts) < 2 {
		return "", "", "", ErrChannelOriginValueInvalid
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", "", ErrChannelOriginValueInvalid
	}
	source = parts[0]
	medium = parts[1]

	// third part is an optionnal campaign
	if len(parts) == 3 {
		if parts[2] == "" {
			return "", "", "", ErrChannelOriginValueInvalid
		}

		campaign = parts[2]
	}

	return source, medium, campaign, nil
}

func FindChannelFromOrigin(channels []*Channel, source string, medium string, campaignName *string) (*Channel, *ChannelOrigin) {

	// find mapped campaign first
	if campaignName != nil && *campaignName != "" {

		for _, ch := range channels {
			for _, origin := range ch.Origins {
				// source medium campaign matches
				if origin.UTMCampaign != nil && *origin.UTMCampaign == *campaignName && origin.UTMSource == source && origin.UTMMedium == medium {
					return ch, origin
				}
			}
		}
	}

	for _, ch := range channels {
		for _, origin := range ch.Origins {
			// source medium matches
			if origin.UTMCampaign == nil && origin.UTMSource == source && origin.UTMMedium == medium {
				return ch, origin
			}
		}
	}

	return nil, nil
}
