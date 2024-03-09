package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/tidwall/sjson"
)

var (
	seededRand      = rand.New(rand.NewSource(time.Now().UnixNano()))
	maleFaceCount   = 0 // keep track of how many females are generated today
	femaleFaceCount = 0 // keep track of how many females are generated today
	lastTimezone    = "Europe/Paris"

	direct        = entity.Channel{}
	googleAds     = entity.Channel{}
	googleOrganic = entity.Channel{}
	fbOrganic     = entity.Channel{}
	gizmodo       = entity.Channel{}
	engadget      = entity.Channel{}
	techcrunch    = entity.Channel{}
	retailmenot   = entity.Channel{}
	adroll        = entity.Channel{}
)

func randomInt(seededRandom *rand.Rand, min, max int) int {
	return seededRandom.Intn(max-min) + min
}

func myUuid() string {
	id, _ := uuid.NewRandom()
	return id.String()
}

// generates fixtures for the current day
func generateOrderDemoFixtures(ctx context.Context, workspace *entity.Workspace, currentDay int, totalDays int) (items []string, err error) {
	// init
	items = []string{}

	// compute how many daily new users to generate a growing user base
	totalUsersAtTheEnd := 1500
	dailyIncrement := int(math.Ceil(float64(totalUsersAtTheEnd) / float64(totalDays)))

	// extract web domain from workspace.Domains
	var webDomain *entity.Domain
	for _, domain := range workspace.Domains {
		if domain.Type == entity.DomainWeb {
			webDomain = domain
			break
		}
	}

	// extract web channels
	for _, channel := range workspace.Channels {
		if channel.ID == "direct" {
			direct = *channel
		}
		if channel.ID == "google-ads" {
			googleAds = *channel
		}
		if channel.ID == "google-organic" {
			googleOrganic = *channel
		}
		if channel.ID == "facebook-organic" {
			fbOrganic = *channel
		}
		if channel.ID == "gizmodo" {
			gizmodo = *channel
		}
		if channel.ID == "engadget" {
			engadget = *channel
		}
		if channel.ID == "techcrunch" {
			techcrunch = *channel
		}
		if channel.ID == "retailmenot" {
			retailmenot = *channel
		}
		if channel.ID == "adroll" {
			adroll = *channel
		}
	}

	// generate daily users
	// by default they are anonymous, and some of them will signup to order

	for i := 1; i <= dailyIncrement; i++ {

		tzLocation, errLoc := time.LoadLocation(lastTimezone)
		if errLoc != nil {
			err = errLoc
			return
		}

		date := time.Now().AddDate(0, 0, -totalDays-1).AddDate(0, 0, currentDay)
		date = time.Date(date.Year(), date.Month(), date.Day(), randomInt(seededRand, 9, 22), randomInt(seededRand, 0, 59), randomInt(seededRand, 0, 59), 0, tzLocation)

		if date.After(time.Now()) {
			continue
		}

		authUser := generateAuthUser(fmt.Sprintf("auth-%v-%v", currentDay, i), date)

		// anon user might become authenticated if the scenario converts
		anonUser := &entity.User{
			ExternalID:      fmt.Sprintf("anon-%v-%v", currentDay, i),
			IsAuthenticated: false,
			CreatedAt:       authUser.CreatedAt,
			Timezone:        authUser.Timezone,
			Language:        authUser.Language,
			Country:         authUser.Country,
		}

		anonUserJSON := fmt.Sprintf(`{
				"external_id":"%v",
				"is_authenticated":false,
				"created_at":"%v",
				"timezone":"%v",
				"language":"%v",
				"country":"%v"
			}`,
			anonUser.ExternalID,
			anonUser.CreatedAt.Format(time.RFC3339),
			*anonUser.Timezone,
			*anonUser.Language,
			*anonUser.Country,
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"user",
			"user": %v
			}`,
			anonUserJSON,
		))

		currentTime := date

		// generate a DataImportItemDevice, 55% mobile, 40% desktop, 15% tablet

		deviceType := "mobile"
		if seededRand.Intn(100) <= 40 {
			deviceType = "desktop"
		} else if seededRand.Intn(100) <= 15 {
			deviceType = "tablet"
		}

		session1_device_ctx := generateDeviceCtx(deviceType, *anonUser.Language, currentTime)

		// by default people will buy an iphone
		scenario := entity.IphoneSE

		// 40% iphone 13
		if seededRand.Intn(100) <= 40 {
			scenario = entity.Iphone13
		} else if seededRand.Intn(100) <= 30 {
			// 30% will buy an ipad
			scenario = entity.IpadAir

			// 40% ipad pro
			if seededRand.Intn(100) <= 40 {
				scenario = entity.IpadPro
			}
		} else if seededRand.Intn(100) <= 20 {
			// 20% will buy a mac
			scenario = entity.MacBookAir

			// 40% macbook pro
			if seededRand.Intn(100) <= 40 {
				scenario = entity.MacBookPro
			}
		}

		// generate the 1st session and generate its origin
		generateSessionOrigin(1, &scenario.Session1)

		session1_ctx := generateSessionCtx(scenario.Session1, currentTime, session1_device_ctx.ID, webDomain.ID, *anonUser.Timezone)

		// generate the first DataImportItemPageview
		session1_pageview1 := generatePageviewItem(session1_device_ctx, session1_ctx, anonUser, scenario.Session1.Page1, currentTime)

		session1_ctx.PageviewsCount = entity.Int64Ptr(1)

		pageview1_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session1_pageview1.ExternalID,
			session1_pageview1.DomainID,
			session1_pageview1.PageID,
			session1_pageview1.Title,
			session1_pageview1.CreatedAt.Format(time.RFC3339),
		)

		session1_json := fmt.Sprintf(`{
				"external_id":"%v",
				"created_at":"%v",
				"domain_id":"%v",
				"timezone":"%v",
				"device_external_id":"%v",
				"duration": %v,
				"pageviews_count":%v,
				"utm_source":"%v",
				"utm_medium":"%v"
			}`,
			session1_ctx.ExternalID,
			session1_ctx.CreatedAt.Format(time.RFC3339),
			session1_ctx.DomainID,
			*session1_ctx.Timezone,
			session1_device_ctx.ExternalID,
			session1_ctx.Duration.Int64,
			*session1_ctx.PageviewsCount,
			session1_ctx.UTMSource.String,
			session1_ctx.UTMMedium.String,
		)

		if session1_ctx.Referrer != nil {
			if session1_json, err = sjson.Set(session1_json, "referrer", session1_ctx.Referrer.String); err != nil {
				log.Printf("error setting referrer: %v", err)
				continue
			}
		}

		if session1_ctx.LandingPage != nil {
			if session1_json, err = sjson.Set(session1_json, "landing_page", session1_ctx.LandingPage.String); err != nil {
				log.Printf("error setting landing_page: %v", err)
				continue
			}
		}

		if session1_ctx.UTMCampaign != nil {
			if session1_json, err = sjson.Set(session1_json, "utm_campaign", session1_ctx.UTMCampaign.String); err != nil {
				log.Printf("error setting utm_campaign: %v", err)
				continue
			}
		}

		if session1_ctx.UTMContent != nil {
			if session1_json, err = sjson.Set(session1_json, "utm_content", session1_ctx.UTMContent.String); err != nil {
				log.Printf("error setting utm_content: %v", err)
				continue
			}
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"device": {
				"external_id":"%v",
				"created_at":"%v",
				"device_type":"%v",
				"user_agent":"%v",
				"browser":"%v",
				"browser_version":"%v",
				"browser_version_major":"%v",
				"os":"%v",
				"resolution":"%v",
				"language":"%v",
				"ad_blocker":%v
			},
			"user": %v
		}`,
			pageview1_json,
			session1_json,
			session1_device_ctx.ExternalID,
			session1_device_ctx.CreatedAt.Format(time.RFC3339),
			session1_device_ctx.DeviceType.String,
			session1_device_ctx.UserAgent.String,
			session1_device_ctx.Browser.String,
			session1_device_ctx.BrowserVersion.String,
			session1_device_ctx.BrowserVersionMajor.String,
			session1_device_ctx.OS.String,
			session1_device_ctx.Resolution.String,
			session1_device_ctx.Language.String,
			session1_device_ctx.AdBlocker.Bool,
			anonUserJSON,
		))

		// otherwise set pageview duration over 15 secs
		duration := entity.Int64Ptr(int64(randomInt(seededRand, 15, 60)))
		bounce := false

		// 20% will bounce, set pageview duration under 20 secs
		if x := randomInt(seededRand, 1, 100); x <= 20 {
			bounce = true
			duration = entity.Int64Ptr(int64(randomInt(seededRand, 3, 14)))
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		addSessionDuration(session1_ctx, duration)

		if session1_json, err = sjson.Set(session1_json, "duration", *session1_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}
		if pageview1_json, err = sjson.Set(pageview1_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}
		if pageview1_json, err = sjson.Set(pageview1_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			pageview1_json,
			session1_json,
			anonUserJSON,
		))

		if bounce {
			continue
		}

		// generate the second DataImportItemPageview
		session1_pageview2 := generatePageviewItem(session1_device_ctx, session1_ctx, anonUser, scenario.Session1.Page2, currentTime)

		session1_ctx.PageviewsCount = entity.Int64Ptr(2)

		if session1_json, err = sjson.Set(session1_json, "pageviews_count", session1_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		pageview2_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session1_pageview2.ExternalID,
			session1_pageview2.DomainID,
			session1_pageview2.PageID,
			session1_pageview2.Title,
			session1_pageview2.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			pageview2_json,
			session1_json,
			anonUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session1_ctx, duration)

		if session1_json, err = sjson.Set(session1_json, "duration", *session1_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if pageview2_json, err = sjson.Set(pageview2_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}
		if pageview2_json, err = sjson.Set(pageview2_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			pageview2_json,
			session1_json,
			anonUserJSON,
		))

		// 20% won't look at 3rd page
		if seededRand.Intn(100) <= 20 {
			continue
		}

		// generate the third DataImportItemPageview
		session1_pageview3 := generatePageviewItem(session1_device_ctx, session1_ctx, anonUser, scenario.Session1.Page3, currentTime)

		session1_ctx.PageviewsCount = entity.Int64Ptr(3)

		if session1_json, err = sjson.Set(session1_json, "pageviews_count", session1_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		pageview3_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session1_pageview3.ExternalID,
			session1_pageview3.DomainID,
			session1_pageview3.PageID,
			session1_pageview3.Title,
			session1_pageview3.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			pageview3_json,
			session1_json,
			anonUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session1_ctx, duration)

		if session1_json, err = sjson.Set(session1_json, "duration", *session1_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if pageview3_json, err = sjson.Set(pageview3_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}
		if pageview3_json, err = sjson.Set(pageview3_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			pageview3_json,
			session1_json,
			anonUserJSON,
		))

		// 20% will never come back
		if x := randomInt(seededRand, 1, 100); x <= 20 {
			continue
		}

		// generate second session in few days after 1st session (if not in the future)
		currentTime = currentTime.Add(time.Duration(randomInt(seededRand, 1, 15)) * 24 * time.Hour)

		if currentTime.After(time.Now()) {
			continue
		}

		// SECOND SESSION, new device

		deviceType = "mobile"
		if seededRand.Intn(100) <= 40 {
			deviceType = "desktop"
		} else if seededRand.Intn(100) <= 15 {
			deviceType = "tablet"
		}

		session2_device_ctx := generateDeviceCtx(deviceType, *anonUser.Language, currentTime)

		// generate the second DataImportItemSession
		generateSessionOrigin(2, &scenario.Session2)
		session2_ctx := generateSessionCtx(scenario.Session2, currentTime, session2_device_ctx.ID, webDomain.ID, *anonUser.Timezone)

		// generate the first pageview for the second session
		session2_pageview1 := generatePageviewItem(session2_device_ctx, session2_ctx, anonUser, scenario.Session2.Page1, currentTime)

		session2_ctx.PageviewsCount = entity.Int64Ptr(1)

		session2_json := fmt.Sprintf(`{
				"external_id":"%v",
				"created_at":"%v",
				"domain_id":"%v",
				"timezone":"%v",
				"device_external_id":"%v",
				"duration": %v,
				"pageviews_count":%v,
				"utm_source":"%v",
				"utm_medium":"%v"
			}`,
			session2_ctx.ExternalID,
			session2_ctx.CreatedAt.Format(time.RFC3339),
			session2_ctx.DomainID,
			*session2_ctx.Timezone,
			session2_device_ctx.ExternalID,
			// session2_ctx.LandingPage.String,
			// session2_ctx.Referrer.String,
			session2_ctx.Duration.Int64,
			*session2_ctx.PageviewsCount,
			session2_ctx.UTMSource.String,
			session2_ctx.UTMMedium.String,
			// session2_ctx.UTMCampaign.String,
			// session2_ctx.UTMContent.String,
		)

		if session2_ctx.Referrer != nil {
			if session2_json, err = sjson.Set(session2_json, "referrer", session2_ctx.Referrer.String); err != nil {
				log.Printf("error setting referrer: %v", err)
				continue
			}
		}

		if session2_ctx.LandingPage != nil {
			if session2_json, err = sjson.Set(session2_json, "landing_page", session2_ctx.LandingPage.String); err != nil {
				log.Printf("error setting landing_page: %v", err)
				continue
			}
		}

		if session2_ctx.UTMCampaign != nil {
			if session2_json, err = sjson.Set(session2_json, "utm_campaign", session2_ctx.UTMCampaign.String); err != nil {
				log.Printf("error setting utm_campaign: %v", err)
				continue
			}
		}

		if session2_ctx.UTMContent != nil {
			if session2_json, err = sjson.Set(session2_json, "utm_content", session2_ctx.UTMContent.String); err != nil {
				log.Printf("error setting utm_content: %v", err)
				continue
			}
		}

		session2_pageview1_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session2_pageview1.ExternalID,
			session2_pageview1.DomainID,
			session2_pageview1.PageID,
			session2_pageview1.Title,
			session2_pageview1.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"device": {
				"external_id":"%v",
				"created_at":"%v",
				"device_type":"%v",
				"user_agent":"%v",
				"browser":"%v",
				"browser_version":"%v",
				"browser_version_major":"%v",
				"os":"%v",
				"resolution":"%v",
				"language":"%v",
				"ad_blocker":%v
			},
			"user": %v
		}`,
			session2_pageview1_json,
			session2_json,
			session2_device_ctx.ExternalID,
			session2_device_ctx.CreatedAt.Format(time.RFC3339),
			session2_device_ctx.DeviceType.String,
			session2_device_ctx.UserAgent.String,
			session2_device_ctx.Browser.String,
			session2_device_ctx.BrowserVersion.String,
			session2_device_ctx.BrowserVersionMajor.String,
			session2_device_ctx.OS.String,
			session2_device_ctx.Resolution.String,
			session2_device_ctx.Language.String,
			session2_device_ctx.AdBlocker.Bool,
			anonUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session2_ctx, duration)

		if session2_json, err = sjson.Set(session2_json, "duration", *session2_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if session2_pageview1_json, err = sjson.Set(session2_pageview1_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session2_pageview1_json, err = sjson.Set(session2_pageview1_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session2_pageview1_json,
			session2_json,
			anonUserJSON,
		))

		session2_pageview2 := generatePageviewItem(session2_device_ctx, session2_ctx, anonUser, scenario.Session2.Page2, currentTime)

		session2_ctx.PageviewsCount = entity.Int64Ptr(2)

		if session2_json, err = sjson.Set(session2_json, "pageviews_count", session2_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		session2_pageview2_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session2_pageview2.ExternalID,
			session2_pageview2.DomainID,
			session2_pageview2.PageID,
			session2_pageview2.Title,
			session2_pageview2.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session2_pageview2_json,
			session2_json,
			anonUserJSON,
		))

		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session2_ctx, duration)

		if session2_json, err = sjson.Set(session2_json, "duration", *session2_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if session2_pageview2_json, err = sjson.Set(session2_pageview2_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session2_pageview2_json, err = sjson.Set(session2_pageview2_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session2_pageview2_json,
			session2_json,
			anonUserJSON,
		))

		// 20% won't look at 3rd page
		if seededRand.Intn(100) <= 20 {
			continue
		}

		// generate the third DataImportItemPageview
		session2_pageview3 := generatePageviewItem(session2_device_ctx, session2_ctx, anonUser, scenario.Session2.Page3, currentTime)

		session2_ctx.PageviewsCount = entity.Int64Ptr(3)

		if session2_json, err = sjson.Set(session2_json, "pageviews_count", session2_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		session2_pageview3_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session2_pageview3.ExternalID,
			session2_pageview3.DomainID,
			session2_pageview3.PageID,
			session2_pageview3.Title,
			session2_pageview3.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session2_pageview3_json,
			session2_json,
			anonUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session2_ctx, duration)

		if session2_json, err = sjson.Set(session2_json, "duration", *session2_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if session2_pageview3_json, err = sjson.Set(session2_pageview3_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session2_pageview3_json, err = sjson.Set(session2_pageview3_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session2_pageview3_json,
			session2_json,
			anonUserJSON,
		))

		// 40% will never come back
		if x := randomInt(seededRand, 1, 100); x <= 40 {
			continue
		}

		// generate third session in few days after 2nd session
		currentTime = currentTime.Add(time.Duration(randomInt(seededRand, 1, 8)) * 24 * time.Hour)

		if currentTime.After(time.Now()) {
			continue
		}

		// THIRD SESSION, new device
		deviceType = "mobile"
		if seededRand.Intn(100) <= 70 {
			// 70% buy on laptop
			deviceType = "desktop"
		} else if seededRand.Intn(100) <= 15 {
			deviceType = "tablet"
		}

		session3_device_ctx := generateDeviceCtx(deviceType, *anonUser.Language, currentTime)

		// generate the third DataImportItemSession
		generateSessionOrigin(3, &scenario.Session3)
		session3_ctx := generateSessionCtx(scenario.Session3, currentTime, session3_device_ctx.ID, webDomain.ID, *anonUser.Timezone)

		// generate the first pageview for the third session
		session3_pageview1 := generatePageviewItem(session3_device_ctx, session3_ctx, anonUser, scenario.Session3.Page1, currentTime)

		session3_ctx.PageviewsCount = entity.Int64Ptr(1)

		session3_json := fmt.Sprintf(`{
				"external_id":"%v",
				"created_at":"%v",
				"domain_id":"%v",
				"timezone":"%v",
				"device_external_id":"%v",
				"duration": %v,
				"pageviews_count":%v,
				"utm_source":"%v",
				"utm_medium":"%v"
			}`,
			session3_ctx.ExternalID,
			session3_ctx.CreatedAt.Format(time.RFC3339),
			session3_ctx.DomainID,
			*session3_ctx.Timezone,
			session3_device_ctx.ExternalID,
			// session3_ctx.LandingPage.String,
			// session3_ctx.Referrer.String,
			session3_ctx.Duration.Int64,
			*session3_ctx.PageviewsCount,
			session3_ctx.UTMSource.String,
			session3_ctx.UTMMedium.String,
			// session3_ctx.UTMCampaign.String,
			// session3_ctx.UTMContent.String,
		)

		if session3_ctx.Referrer != nil {
			if session3_json, err = sjson.Set(session3_json, "referrer", session3_ctx.Referrer.String); err != nil {
				log.Printf("error setting referrer: %v", err)
				continue
			}
		}

		if session3_ctx.LandingPage != nil {
			if session3_json, err = sjson.Set(session3_json, "landing_page", session3_ctx.LandingPage.String); err != nil {
				log.Printf("error setting landing_page: %v", err)
				continue
			}
		}

		if session3_ctx.UTMCampaign != nil {
			if session3_json, err = sjson.Set(session3_json, "utm_campaign", session3_ctx.UTMCampaign.String); err != nil {
				log.Printf("error setting utm_campaign: %v", err)
				continue
			}
		}

		if session3_ctx.UTMContent != nil {
			if session3_json, err = sjson.Set(session3_json, "utm_content", session3_ctx.UTMContent.String); err != nil {
				log.Printf("error setting utm_content: %v", err)
				continue
			}
		}

		session3_pageview1_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session3_pageview1.ExternalID,
			session3_pageview1.DomainID,
			session3_pageview1.PageID,
			session3_pageview1.Title,
			session3_pageview1.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"device": {
				"external_id":"%v",
				"created_at":"%v",
				"device_type":"%v",
				"user_agent":"%v",
				"browser":"%v",
				"browser_version":"%v",
				"browser_version_major":"%v",
				"os":"%v",
				"resolution":"%v",
				"language":"%v",
				"ad_blocker":%v
			},
			"user": %v
		}`,
			session3_pageview1_json,
			session3_json,
			session3_device_ctx.ExternalID,
			session3_device_ctx.CreatedAt.Format(time.RFC3339),
			session3_device_ctx.DeviceType.String,
			session3_device_ctx.UserAgent.String,
			session3_device_ctx.Browser.String,
			session3_device_ctx.BrowserVersion.String,
			session3_device_ctx.BrowserVersionMajor.String,
			session3_device_ctx.OS.String,
			session3_device_ctx.Resolution.String,
			session3_device_ctx.Language.String,
			session3_device_ctx.AdBlocker.Bool,
			anonUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session3_ctx, duration)

		if session3_json, err = sjson.Set(session3_json, "duration", *session3_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if session3_pageview1_json, err = sjson.Set(session3_pageview1_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session3_pageview1_json, err = sjson.Set(session3_pageview1_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session3_pageview1_json,
			session3_json,
			anonUserJSON,
		))

		// 50% will not add products to cart and leave
		if x := randomInt(seededRand, 1, 100); x <= 50 {
			continue
		}

		// generate the second pageview for the third session
		session3_pageview2 := generatePageviewItem(session3_device_ctx, session3_ctx, anonUser, scenario.Session3.Page2, currentTime)

		session3_ctx.PageviewsCount = entity.Int64Ptr(2)

		if session3_json, err = sjson.Set(session3_json, "pageviews_count", session3_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		session3_pageview2_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session3_pageview2.ExternalID,
			session3_pageview2.DomainID,
			session3_pageview2.PageID,
			session3_pageview2.Title,
			session3_pageview2.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session3_pageview2_json,
			session3_json,
			anonUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session3_ctx, duration)

		if session3_json, err = sjson.Set(session3_json, "duration", *session3_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if session3_pageview2_json, err = sjson.Set(session3_pageview2_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session3_pageview2_json, err = sjson.Set(session3_pageview2_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
		}`,
			session3_pageview2_json,
			session3_json,
			anonUserJSON,
		))

		// generate the first DataImportItemCart
		session3_cart1 := generateCart(session3_ctx, anonUser, scenario.Cart, currentTime)

		cartItems_json := "[]"

		for _, item := range session3_cart1.Items {
			if cartItems_json, err = sjson.SetRaw(cartItems_json, "-1", fmt.Sprintf(`{
				"external_id":"%v",
				"product_external_id":"%v",
				"sku":"%v",
				"name":"%v",
				"brand":"%v",
				"category":"%v",
				"variant_external_id":"%v",
				"variant_title":"%v",
				"quantity":%v,
				"price":%v,
				"image_url":"%v"
			}`,
				item.ExternalID,
				item.ProductExternalID,
				item.SKU.String,
				item.Name,
				item.Brand.String,
				item.Category.String,
				item.VariantExternalID.String,
				item.VariantTitle.String,
				item.Quantity,
				item.Price,
				item.ImageURL.String,
			)); err != nil {
				log.Printf("error setting cart item: %v", err)
				continue
			}
		}

		session3_cart1_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"created_at":"%v",
				"currency":"%v",
				"items": %v
			}`,
			session3_cart1.ExternalID,
			session3_cart1.DomainID,
			session3_cart1.CreatedAt.Format(time.RFC3339),
			workspace.Currency,
			cartItems_json,
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"cart",
			"cart": %v,
			"session": %v,
			"user": %v
		}`,
			session3_cart1_json,
			session3_json,
			anonUserJSON,
		))

		// 10% will abandon the cart
		if x := randomInt(seededRand, 1, 100); x <= 10 {
			continue
		}

		// checkout pageview

		session3_checkout_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			myUuid(),
			session3_cart1.DomainID,
			"https://apple.com/checkout",
			"Checkout - Apple",
			currentTime.Format(time.RFC3339),
		)

		session3_ctx.PageviewsCount = entity.Int64Ptr(3)

		if session3_json, err = sjson.Set(session3_json, "pageviews_count", session3_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
		}`,
			session3_checkout_json,
			session3_json,
			anonUserJSON,
		))

		// set pageview duration over 30 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 30, 120)))
		addSessionDuration(session3_ctx, duration)

		if session3_json, err = sjson.Set(session3_json, "duration", *session3_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)
		if currentTime.After(time.Now()) {
			continue
		}

		if session3_checkout_json, err = sjson.Set(session3_checkout_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session3_checkout_json, err = sjson.Set(session3_checkout_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
		}`,
			session3_checkout_json,
			session3_json,
			anonUserJSON,
		))

		// import authenticated user
		authUserCreatedAt := currentTime // clone time
		authUser.CreatedAt = authUserCreatedAt

		authUser.UpdatedAt = &authUserCreatedAt
		authUser.SignedUpAt = &authUserCreatedAt
		authUser.ConsentAll = &entity.NullableBool{IsNull: false, Bool: true}

		authUserJSON := fmt.Sprintf(`{
			"external_id":"%v",	
			"is_authenticated":true,
			"created_at":"%v",
			"signed_up_at":"%v",
			"updated_at":"%v",
			"consent_all":true,
			"email":"%v",
			"first_name":"%v",
			"last_name":"%v",
			"gender": "%v",
			"country":"%v",
			"timezone":"%v",
			"language":"%v",
			"latitude":%v,
			"longitude":%v,
			"birthday":"%v",
			"photo_url":"%v"
		}`,
			authUser.ExternalID,
			authUser.CreatedAt.Format(time.RFC3339),
			authUser.SignedUpAt.Format(time.RFC3339),
			authUser.UpdatedAt.Format(time.RFC3339),
			authUser.Email.String,
			authUser.FirstName.String,
			authUser.LastName.String,
			authUser.Gender.String,
			*authUser.Country,
			*authUser.Timezone,
			*authUser.Language,
			authUser.Latitude.Float64,
			authUser.Longitude.Float64,
			authUser.Birthday.String,
			authUser.PhotoURL.String,
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"user",
			"user": %v
		}`,
			authUserJSON,
		))

		// generate a user_alias with the anonUser and authUser

		items = append(items, fmt.Sprintf(`{
			"kind":"user_alias",
			"user_alias": {
				"from_user_external_id":"%v",
				"to_user_external_id":"%v",
				"to_user_is_authenticated": true,
				"to_user_created_at":"%v"
			}
		}`,
			anonUser.ExternalID,
			authUser.ExternalID,
			authUser.CreatedAt.Format(time.RFC3339),
		))

		// place order

		subtotal := session3_cart1.Items[0].Price

		// shipping is a randomint64 from 5 to 30
		shipping := int64(randomInt(seededRand, 5, 30))

		// tax is a 20% of the subtotal
		tax := subtotal * 20 / 100

		// total price is subtotal + shipping + tax
		total := subtotal + shipping + tax

		orderExternalID := myUuid()
		order_items_json := "[]"

		for _, item := range session3_cart1.Items {
			if order_items_json, err = sjson.SetRaw(order_items_json, "-1", fmt.Sprintf(`{
				"order_external_id":"%v",
				"quantity":%v,
				"external_id":"%v",	
				"quantity":%v,
				"product_external_id":"%v",
				"sku":"%v",
				"name":"%v",
				"brand":"%v",
				"category":"%v",
				"variant_external_id":"%v",
				"variant_title":"%v",
				"price":%v,
				"image_url":"%v"
			}`,
				orderExternalID,
				item.Quantity,
				item.ExternalID,
				item.Quantity,
				item.ProductExternalID,
				item.SKU.String,
				item.Name,
				item.Brand.String,
				item.Category.String,
				item.VariantExternalID.String,
				item.VariantTitle.String,
				item.Price,
				item.ImageURL.String,
			)); err != nil {
				log.Printf("error setting order_items_json: %v", err)
				continue
			}
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"order",
			"order": {
				"external_id":"%v",
				"session_external_id":"%v",
				"created_at":"%v",
				"currency":"%v",
				"subtotal_price":%v,
				"total_price":%v,
				"items": %v
			},
			"session": %v,
			"user": %v
		}`,
			orderExternalID,
			session3_ctx.ExternalID,
			currentTime.Format(time.RFC3339),
			workspace.Currency,
			subtotal,
			total,
			order_items_json,
			session3_json,
			authUserJSON,
		))

		// SECOND ORDER - same as session 3

		// by default people will buy an iphone
		scenario2 := entity.IphoneSE

		// 40% iphone 13
		if seededRand.Intn(100) <= 40 {
			scenario2 = entity.Iphone13
		} else if seededRand.Intn(100) <= 30 {
			// 30% will buy an ipad
			scenario2 = entity.IpadAir

			// 40% ipad pro
			if seededRand.Intn(100) <= 40 {
				scenario2 = entity.IpadPro
			}
		} else if seededRand.Intn(100) <= 20 {
			// 20% will buy a mac
			scenario2 = entity.MacBookAir

			// 40% macbook pro
			if seededRand.Intn(100) <= 40 {
				scenario2 = entity.MacBookPro
			}
		}

		// 40% will never come back
		if x := randomInt(seededRand, 1, 100); x <= 40 {
			continue
		}

		// generate fourth session in few weeks after 2nd session
		currentTime = currentTime.Add(time.Duration(randomInt(seededRand, 15, 25)) * 24 * time.Hour)

		if currentTime.After(time.Now()) {
			continue
		}

		// 4th SESSION, new device
		deviceType = "mobile"
		if seededRand.Intn(100) <= 70 {
			// 70% buy on laptop
			deviceType = "desktop"
		} else if seededRand.Intn(100) <= 15 {
			deviceType = "tablet"
		}

		session4_device_ctx := generateDeviceCtx(deviceType, *authUser.Language, currentTime)

		// generate the fourth DataImportItemSession
		generateSessionOrigin(3, &scenario2.Session3)
		session4_ctx := generateSessionCtx(scenario2.Session3, currentTime, session4_device_ctx.ID, webDomain.ID, *authUser.Timezone)

		// generate the first pageview for the third session
		session4_pageview1 := generatePageviewItem(session4_device_ctx, session4_ctx, authUser, scenario2.Session3.Page1, currentTime)

		session4_ctx.PageviewsCount = entity.Int64Ptr(1)

		session4_json := fmt.Sprintf(`{
				"external_id":"%v",
				"created_at":"%v",
				"domain_id":"%v",
				"timezone":"%v",
				"device_external_id":"%v",
				"duration": %v,
				"pageviews_count":%v,
				"utm_source":"%v",
				"utm_medium":"%v"
			}`,
			session4_ctx.ExternalID,
			session4_ctx.CreatedAt.Format(time.RFC3339),
			session4_ctx.DomainID,
			*session4_ctx.Timezone,
			session4_device_ctx.ExternalID,
			// session4_ctx.LandingPage.String,
			// session4_ctx.Referrer.String,
			session4_ctx.Duration.Int64,
			*session4_ctx.PageviewsCount,
			session4_ctx.UTMSource.String,
			session4_ctx.UTMMedium.String,
			// session4_ctx.UTMCampaign.String,
			// session4_ctx.UTMContent.String,
		)

		if session4_ctx.Referrer != nil {
			if session4_json, err = sjson.Set(session4_json, "referrer", session4_ctx.Referrer.String); err != nil {
				log.Printf("error setting referrer: %v", err)
				continue
			}
		}

		if session4_ctx.LandingPage != nil {
			if session4_json, err = sjson.Set(session4_json, "landing_page", session4_ctx.LandingPage.String); err != nil {
				log.Printf("error setting landing_page: %v", err)
				continue
			}
		}

		if session4_ctx.UTMCampaign != nil {
			if session4_json, err = sjson.Set(session4_json, "utm_campaign", session4_ctx.UTMCampaign.String); err != nil {
				log.Printf("error setting utm_campaign: %v", err)
				continue
			}
		}

		if session4_ctx.UTMContent != nil {
			if session4_json, err = sjson.Set(session4_json, "utm_content", session4_ctx.UTMContent.String); err != nil {
				log.Printf("error setting utm_content: %v", err)
				continue
			}
		}

		session4_pageview1_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session4_pageview1.ExternalID,
			session4_pageview1.DomainID,
			session4_pageview1.PageID,
			session4_pageview1.Title,
			session4_pageview1.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"device": {
				"external_id":"%v",
				"created_at":"%v",
				"device_type":"%v",
				"user_agent":"%v",
				"browser":"%v",
				"browser_version":"%v",
				"browser_version_major":"%v",
				"os":"%v",
				"resolution":"%v",
				"language":"%v",
				"ad_blocker":%v
			},
			"user": %v
		}`,
			session4_pageview1_json,
			session4_json,
			session4_device_ctx.ExternalID,
			session4_device_ctx.CreatedAt.Format(time.RFC3339),
			session4_device_ctx.DeviceType.String,
			session4_device_ctx.UserAgent.String,
			session4_device_ctx.Browser.String,
			session4_device_ctx.BrowserVersion.String,
			session4_device_ctx.BrowserVersionMajor.String,
			session4_device_ctx.OS.String,
			session4_device_ctx.Resolution.String,
			session4_device_ctx.Language.String,
			session4_device_ctx.AdBlocker.Bool,
			authUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session4_ctx, duration)

		if session4_json, err = sjson.Set(session4_json, "duration", *session4_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if session4_pageview1_json, err = sjson.Set(session4_pageview1_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session4_pageview1_json, err = sjson.Set(session4_pageview1_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session4_pageview1_json,
			session4_json,
			authUserJSON,
		))

		// 50% will not add products to cart and leave
		if x := randomInt(seededRand, 1, 100); x <= 50 {
			continue
		}

		// generate the second pageview for the third session
		session4_pageview2 := generatePageviewItem(session4_device_ctx, session4_ctx, authUser, scenario2.Session3.Page2, currentTime)

		session4_ctx.PageviewsCount = entity.Int64Ptr(2)

		if session4_json, err = sjson.Set(session4_json, "pageviews_count", session4_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		session4_pageview2_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			session4_pageview2.ExternalID,
			session4_pageview2.DomainID,
			session4_pageview2.PageID,
			session4_pageview2.Title,
			session4_pageview2.CreatedAt.Format(time.RFC3339),
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session4_pageview2_json,
			session4_json,
			authUserJSON,
		))

		// set pageview duration over 15 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 16, 60)))
		addSessionDuration(session4_ctx, duration)

		if session4_json, err = sjson.Set(session4_json, "duration", *session4_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)

		if currentTime.After(time.Now()) {
			continue
		}

		if session4_pageview2_json, err = sjson.Set(session4_pageview2_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session4_pageview2_json, err = sjson.Set(session4_pageview2_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session4_pageview2_json,
			session4_json,
			authUserJSON,
		))

		// generate the first DataImportItemCart
		session4_cart1 := generateCart(session4_ctx, authUser, scenario2.Cart, currentTime)

		cartItems_json = "[]"

		for _, item := range session4_cart1.Items {
			if cartItems_json, err = sjson.SetRaw(cartItems_json, "-1", fmt.Sprintf(`{
				"external_id":"%v",
				"product_external_id":"%v",
				"sku":"%v",
				"name":"%v",
				"brand":"%v",
				"category":"%v",
				"variant_external_id":"%v",
				"variant_title":"%v",
				"quantity":%v,
				"price":%v,
				"image_url":"%v"
			}`,
				item.ExternalID,
				item.ProductExternalID,
				item.SKU.String,
				item.Name,
				item.Brand.String,
				item.Category.String,
				item.VariantExternalID.String,
				item.VariantTitle.String,
				item.Quantity,
				item.Price,
				item.ImageURL.String,
			)); err != nil {
				log.Printf("error setting cart item: %v", err)
				continue
			}
		}

		session4_cart1_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"created_at":"%v",
				"currency":"%v",
				"items": %v
			}`,
			session4_cart1.ExternalID,
			session4_cart1.DomainID,
			session4_cart1.CreatedAt.Format(time.RFC3339),
			workspace.Currency,
			cartItems_json,
		)

		items = append(items, fmt.Sprintf(`{
			"kind":"cart",
			"cart": %v,
			"session": %v,
			"user": %v
		}`,
			session4_cart1_json,
			session4_json,
			authUserJSON,
		))

		// 10% will abandon the cart
		if x := randomInt(seededRand, 1, 100); x <= 10 {
			continue
		}

		// checkout page
		session4_checkout_json := fmt.Sprintf(`{
				"external_id":"%v",
				"domain_id":"%v",
				"page_id":"%v",
				"title":"%v",
				"created_at":"%v"
			}`,
			myUuid(),
			session4_ctx.DomainID,
			"https://apple.com/checkout",
			"Checkout - Apple",
			currentTime.Format(time.RFC3339),
		)

		session4_ctx.PageviewsCount = entity.Int64Ptr(3)

		if session4_json, err = sjson.Set(session4_json, "pageviews_count", session4_ctx.PageviewsCount); err != nil {
			log.Printf("error setting pageviews_count: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
			}`,
			session4_checkout_json,
			session4_json,
			authUserJSON,
		))

		// set pageview duration over 30 secs
		duration = entity.Int64Ptr(int64(randomInt(seededRand, 30, 120)))
		addSessionDuration(session4_ctx, duration)

		if session4_json, err = sjson.Set(session4_json, "duration", *session4_ctx.Duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		// add elapsed time on the page
		currentTime = currentTime.Add(time.Duration(*duration) * time.Second)
		if currentTime.After(time.Now()) {
			continue
		}

		if session4_checkout_json, err = sjson.Set(session4_checkout_json, "duration", duration); err != nil {
			log.Printf("error setting duration: %v", err)
			continue
		}

		if session4_checkout_json, err = sjson.Set(session4_checkout_json, "updated_at", currentTime.Format(time.RFC3339)); err != nil {
			log.Printf("error setting updated_at: %v", err)
			continue
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"pageview",
			"pageview": %v,
			"session": %v,
			"user": %v
		}`,
			session4_checkout_json,
			session4_json,
			authUserJSON,
		))

		// place order

		subtotal = session4_cart1.Items[0].Price

		// shipping is a randomint64 from 5 to 30
		shipping = int64(randomInt(seededRand, 5, 30))

		// tax is a 20% of the subtotal
		tax = subtotal * 20 / 100

		// total price is subtotal + shipping + tax
		total = subtotal + shipping + tax

		order2ExternalID := myUuid()

		order_items_json = "[]"

		for _, item := range session4_cart1.Items {
			if order_items_json, err = sjson.SetRaw(order_items_json, "-1", fmt.Sprintf(`{
				"order_external_id":"%v",
				"external_id":"%v",	
				"quantity":%v,
				"product_external_id":"%v",
				"sku":"%v",
				"name":"%v",
				"brand":"%v",
				"category":"%v",
				"variant_external_id":"%v",
				"variant_title":"%v",
				"price":%v,
				"image_url":"%v"
			}`,
				order2ExternalID,
				item.ExternalID,
				item.Quantity,
				item.ProductExternalID,
				item.SKU.String,
				item.Name,
				item.Brand.String,
				item.Category.String,
				item.VariantExternalID.String,
				item.VariantTitle.String,
				item.Price,
				item.ImageURL.String,
			)); err != nil {
				log.Printf("error setting order_items_json: %v", err)
				continue
			}
		}

		items = append(items, fmt.Sprintf(`{
			"kind":"order",
			"order": {
				"external_id":"%v",
				"session_external_id":"%v",
				"created_at":"%v",
				"currency":"%v",
				"subtotal_price":%v,
				"total_price":%v,
				"items": %v
			},
			"session": %v,
			"user": %v
		}`,
			order2ExternalID,
			session4_ctx.ExternalID,
			currentTime.Format(time.RFC3339),
			workspace.Currency,
			subtotal,
			total,
			order_items_json,
			session4_json,
			authUserJSON,
		))

	}

	return
}

func generateAuthUser(externalID string, createdAt time.Time) (userItem *entity.User) {

	isMale := true

	// 56% of females
	if x := randomInt(seededRand, 0, 100); x > 55 {
		isMale = false
	}

	birthDate := randomInt(seededRand, 1, 31)
	birthMonth := randomInt(seededRand, 1, 12)
	birthYear := randomInt(seededRand, 1975, 2000)

	birthday := time.Date(birthYear, time.Month(birthMonth), birthDate, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	// select a male or female profile in the fake users list
	if maleFaceCount >= 999 {
		maleFaceCount = 0
	} else {
		maleFaceCount++
	}

	fakeUser := entity.FakeMales[maleFaceCount]

	if !isMale {
		if femaleFaceCount >= 999 {
			femaleFaceCount = 0
		} else {
			femaleFaceCount++
		}

		fakeUser = entity.FakeFemales[femaleFaceCount]
	}

	var timezone, lang string
	var latitude, longitude float64

	switch fakeUser.Nat {
	case "AU":
		timezone = "Australia/Sydney"
		lang = "en"
		latitude = -33.836315
		longitude = 151.040380
	case "BR":
		timezone = "America/Sao_Paulo"
		lang = "es"
		latitude = -23.550519
		longitude = -46.633309
	case "CA":
		timezone = "America/Toronto"
		lang = "en"
		latitude = 43.851620
		longitude = -79.487554
	case "CH":
		timezone = "Europe/Zurich"
		lang = "fr"
		latitude = 47.376886
		longitude = 8.541694
	case "DE":
		timezone = "Europe/Berlin"
		lang = "de"
		latitude = 52.520006
		longitude = 13.404953
	case "DK":
		timezone = "Europe/Copenhagen"
		lang = "da"
		latitude = 55.710153
		longitude = 12.359596
	case "ES":
		timezone = "Europe/Madrid"
		lang = "es"
		latitude = 40.416775
		longitude = -3.703790
	case "FI":
		timezone = "Europe/Helsinki"
		lang = "fi"
		latitude = 60.169855
		longitude = 24.938379
	case "FR":
		timezone = "Europe/Paris"
		lang = "fr"
		latitude = 48.856614
		longitude = 2.352221
	// DISABLE: default
	// case "GB":
	//  timezone = "Europe/London"
	//  lang = "en"
	//       latitude = 51.5073509
	//       longitude = -0.12775829999998223
	case "IE":
		timezone = "Europe/Dublin"
		lang = "en"
		latitude = 53.353084
		longitude = -6.364679
	case "IR":
		timezone = "Europe/London"
		lang = "en"
		latitude = 54.559079
		longitude = -5.952092
	case "NL":
		timezone = "Europe/Amsterdam"
		lang = "nl"
		latitude = 52.336664
		longitude = 4.862208
	case "NZ":
		timezone = "Pacific/Auckland"
		lang = "en"
		latitude = -36.858898
		longitude = 174.760928
	case "TR":
		timezone = "Europe/Istanbul"
		lang = "tr"
		latitude = 41.008237
		longitude = 28.973895
	case "US":
		timezone = "America/New_York"
		lang = "en"
		latitude = 40.720842
		longitude = -73.994986
	default:
		timezone = "Europe/London"
		lang = "en"
		latitude = 51.507350
		longitude = -0.1277582
	}
	lastTimezone = timezone // location used by future unkown users

	gender := "male"

	if !isMale {
		gender = "female"
	}

	email := externalID + "@example.com"

	userItem = &entity.User{
		ID:              entity.ComputeUserID(externalID),
		ExternalID:      externalID,
		IsAuthenticated: true,
		CreatedAt:       createdAt,
		SignedUpAt:      &createdAt,
		Timezone:        &timezone,
		Language:        &lang,
		Country:         &fakeUser.Nat,
		Gender:          entity.NewNullableString(&gender),
		Latitude:        entity.NewNullableFloat64(&latitude),
		Longitude:       entity.NewNullableFloat64(&longitude),
		Birthday:        entity.NewNullableString(&birthday),
		FirstName:       entity.NewNullableString(&fakeUser.Name.First),
		LastName:        entity.NewNullableString(&fakeUser.Name.Last),
		PhotoURL:        entity.NewNullableString(&fakeUser.Picture.Medium),
		Email:           entity.NewNullableString(&email),
	}

	return
}

func generateDeviceCtx(deviceType string, language string, createdAt time.Time) (deviceCtx *entity.Device) {
	chromeDesktop := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36"
	safariDesktop := "Mozilla/5.0 (Macintosh; Intel Mac OS X 12_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Safari/605.1.15"
	firefoxDesktop := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:101.0) Gecko/20100101 Firefox/101.0"
	safariIOS := "Mozilla/5.0 (iPhone; CPU iPhone OS 15_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1"
	chromeAndroid := "Mozilla/5.0 (Linux; Android 12) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.99 Mobile Safari/537.36"

	// init desktop
	userAgent := chromeDesktop
	resolution := "1920x1080"
	browser := ""
	browserVersion := ""
	browserVersionMajor := ""
	os := ""

	if x := randomInt(seededRand, 1, 100); x <= 25 {
		resolution = "1366x768"
	}
	if x := randomInt(seededRand, 1, 100); x <= 11 {
		resolution = "1536x864"
	}
	if x := randomInt(seededRand, 1, 100); x <= 7 {
		resolution = "1440x900"
	}
	if x := randomInt(seededRand, 1, 100); x <= 7 {
		resolution = "1280x720"
	}

	// 20% are firefox
	if x := randomInt(seededRand, 1, 100); x <= 20 {
		userAgent = firefoxDesktop
		browser = "Firefox"
		browserVersion = "101.0"
		browserVersionMajor = "101"
		os = "Mac OS X"
	}
	// 30% are safari
	if x := randomInt(seededRand, 1, 100); x <= 30 {
		userAgent = safariDesktop
		browser = "Safari"
		browserVersion = "15.4"
		browserVersionMajor = "15"
		os = "Mac OS X"
	}

	// for mobile devices
	if deviceType == "mobile" || deviceType == "tablet" {
		userAgent = chromeAndroid
		browser = "Chrome"
		browserVersion = "102.0.5005.99"
		browserVersionMajor = "102"
		os = "Android"

		// 35% are ios
		if x := randomInt(seededRand, 1, 100); x <= 35 {
			userAgent = safariIOS
			browser = "Safari"
			browserVersion = "15.4"
			browserVersionMajor = "15"
			os = "iOS"
		}
	}

	externalID := myUuid()
	id := entity.ComputeCartID(externalID)

	deviceCtx = &entity.Device{
		ID:                  id,
		ExternalID:          externalID,
		CreatedAt:           createdAt,
		UpdatedAt:           &createdAt,
		DeviceType:          entity.NewNullableString(&deviceType),
		UserAgent:           entity.NewNullableString(&userAgent),
		Resolution:          entity.NewNullableString(&resolution),
		Language:            entity.NewNullableString(&language),
		Browser:             entity.NewNullableString(&browser),
		BrowserVersion:      entity.NewNullableString(&browserVersion),
		BrowserVersionMajor: entity.NewNullableString(&browserVersionMajor),
		OS:                  entity.NewNullableString(&os),
		AdBlocker:           entity.NewNullableBool(entity.BoolPtr(false)),
	}

	return deviceCtx
}

func generateSessionCtx(sessionData entity.DemoSession, createdAt time.Time, deviceID string, domainID string, timezone string) (sessionItem *entity.Session) {

	extId := myUuid()
	id := entity.ComputeSessionID(extId)

	sessionItem = &entity.Session{
		ID:         id,
		ExternalID: extId,
		CreatedAt:  createdAt,
		DomainID:   domainID,
		Timezone:   &timezone,
		DeviceID:   &deviceID,

		UTMSource:      &entity.NullableString{IsNull: false, String: sessionData.UTMSource},
		UTMMedium:      &entity.NullableString{IsNull: false, String: sessionData.UTMMedium},
		PageviewsCount: entity.Int64Ptr(0),
		Duration:       &entity.NullableInt64{IsNull: true, Int64: 0},
	}

	if sessionData.Referrer != "" {
		sessionItem.Referrer = &entity.NullableString{IsNull: false, String: sessionData.Referrer}
	}

	if sessionData.LandingPage != "" {
		sessionItem.LandingPage = &entity.NullableString{IsNull: false, String: sessionData.LandingPage}
	}

	if sessionData.UTMCampaign != "" {
		sessionItem.UTMCampaign = &entity.NullableString{IsNull: false, String: sessionData.UTMCampaign}
	}
	if sessionData.UTMContent != "" {
		sessionItem.UTMContent = &entity.NullableString{IsNull: false, String: sessionData.UTMContent}
	}

	return
}

func generatePageviewItem(deviceCtx *entity.Device, sessionCtx *entity.Session, anonUser *entity.User, pageData entity.DemoPage, currentTime time.Time) (pageview *entity.Pageview) {
	extID := myUuid()
	id := entity.ComputePageviewID(extID)
	pageview = &entity.Pageview{
		ID:         id,
		ExternalID: extID,
		DomainID:   sessionCtx.DomainID,
		PageID:     pageData.PageID,
		Title:      pageData.Title,
		CreatedAt:  currentTime,
	}
	return
}

func generateSessionOrigin(sessionIndex int, session *entity.DemoSession) {

	sessionOrigin := googleOrganic
	utmCampaign := ""
	utmContent := ""
	session.SetOrigin("https://www.google.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)

	// 50% adwords
	if seededRand.Intn(100) <= 50 {
		sessionOrigin = googleAds
		utmCampaign := "Black friday"
		if seededRand.Intn(100) <= 30 {
			utmCampaign = "Chrismas"
		}
		if seededRand.Intn(100) <= 20 {
			utmCampaign = "Easter"
		}
		utmContent := fmt.Sprintf("Ad %v", seededRand.Intn(8))
		session.SetOrigin("https://www.google.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	} else if seededRand.Intn(100) <= 20 {
		// 20% fb organic
		sessionOrigin = fbOrganic
		session.SetOrigin("https://www.facebook.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	} else if seededRand.Intn(100) <= 10 {
		// 10% engadget
		sessionOrigin = engadget
		utmCampaign = "Engadget partnership"
		utmContent = "Home splash"
		session.SetOrigin("https://www.engadget.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	} else if seededRand.Intn(100) <= 8 {
		// 8% gizmodo
		sessionOrigin = gizmodo
		utmCampaign = "Gizmodo partnership"
		utmContent = "Search bar"
		session.SetOrigin("https://www.gizmodo.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	} else if seededRand.Intn(100) <= 5 {
		// 5% techcrunch
		sessionOrigin = techcrunch
		utmCampaign = "Techcrunch partnership"
		utmContent = "Side bar"
		session.SetOrigin("https://www.techcrunch.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	}

	if sessionIndex == 1 {
		return
	}

	// has 40% of direct traffic for the second pageview
	if seededRand.Intn(100) <= 50 {
		sessionOrigin = direct
		utmCampaign := ""
		utmContent := ""
		session.SetOrigin("https://www.apple.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	}

	if seededRand.Intn(100) <= 20 {
		// has 20% retargeting on second session
		sessionOrigin = adroll
		utmCampaign := "Black friday"
		if seededRand.Intn(100) <= 30 {
			utmCampaign = "Chrismas"
		}
		if seededRand.Intn(100) <= 20 {
			utmCampaign = "Easter"
		}
		utmContent := fmt.Sprintf("Ad %v", seededRand.Intn(8))
		session.SetOrigin("https://www.adroll.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	}

	if sessionIndex != 3 {
		return
	}

	// has 20% coupon on last session
	if seededRand.Intn(100) <= 50 {
		sessionOrigin = retailmenot
		utmCampaign = "Black friday"
		utmContent = "voucher 5%"
		session.SetOrigin("https://www.retailmenot.com", sessionOrigin.Origins[0].UTMSource, sessionOrigin.Origins[0].UTMMedium, utmCampaign, utmContent)
	}
}

func generateCart(sessionCtx *entity.Session, anonUser *entity.User, cartData entity.Cart, currentTime time.Time) (cart *entity.Cart) {

	extID := myUuid()
	id := entity.ComputeCartID(extID)

	cart = &entity.Cart{
		ID:         id,
		UserID:     anonUser.ID,
		ExternalID: extID,
		DomainID:   sessionCtx.DomainID,
		SessionID:  entity.StringPtr(entity.ComputeSessionID(sessionCtx.ExternalID)),
		Items:      entity.CartItems{},
		CreatedAt:  currentTime,
		UpdatedAt:  entity.TimePtr(currentTime),
	}

	for _, item := range cartData.Items {
		cart.Items = append(cart.Items, &entity.CartItem{
			CartID:            cart.ID,
			UserID:            anonUser.ID,
			ExternalID:        myUuid(),
			ProductExternalID: item.ProductExternalID,
			Name:              item.Name,
			Brand:             item.Brand,
			SKU:               item.SKU,
			Category:          item.Category,
			VariantExternalID: item.VariantExternalID,
			VariantTitle:      item.VariantTitle,
			ImageURL:          item.ImageURL,
			Quantity:          1,
			Price:             item.Price,
			CreatedAt:         currentTime,
			UpdatedAt:         entity.TimePtr(currentTime),
		})
	}

	return
}

func addSessionDuration(sessionCtx *entity.Session, duration *int64) {
	if duration == nil || sessionCtx == nil {
		log.Fatal("addSessionDuration has empty session or duration")
	}

	sessionCtx.Duration.IsNull = false
	sessionCtx.Duration.Int64 = sessionCtx.Duration.Int64 + *duration
}
