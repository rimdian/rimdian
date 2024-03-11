package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opencensus.io/trace"
)

// execute the on_validation hook
func (pipe *DataLogPipeline) StepPending(ctx context.Context) {

	_, span := trace.StartSpan(ctx, "StepPending")
	defer span.End()

	// BEFORE VALIDATION HOOKS TO IMPLEMENT
	var err error

	// check if has item.user
	hasUser := gjson.Get(pipe.DataLog.Item, "user")
	if hasUser.Exists() {

		// Extract Google Geo headers
		if _, ok := pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Country"]; ok {
			country := strings.ToUpper(pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Country"])
			if country != "" && govalidator.IsISO3166Alpha2(country) {
				// set user country in item json
				pipe.DataLog.Item, err = sjson.Set(pipe.DataLog.Item, "user.country", country)
				if err != nil {
					// log error and ignore
					pipe.Logger.Printf("Error setting user.country: %v", err.Error())
					return
				}
			}
		}

		if _, ok := pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Latlon"]; ok {
			latLon := pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Latlon"]

			if strings.Contains(latLon, ",") {

				// split LatLon
				parts := strings.Split(latLon, ",")

				latitude, latErr := strconv.ParseFloat(parts[0], 64)
				longitude, lonErr := strconv.ParseFloat(parts[1], 64)

				if latErr != nil {
					pipe.Logger.Printf("parse X-Client-Geo-Latlon latitude: %v", latErr.Error())
				} else if lonErr != nil {
					pipe.Logger.Printf("parse X-Client-Geo-Latlon longitude: %v", lonErr.Error())
				}

				// set user latitude in item json
				pipe.DataLog.Item, err = sjson.Set(pipe.DataLog.Item, "user.latitude", latitude)
				if err != nil {
					// log error and ignore
					pipe.Logger.Printf("Error setting user.latitude: %v", err.Error())
					return
				}
				pipe.DataLog.Item, err = sjson.Set(pipe.DataLog.Item, "user.longitude", longitude)
				if err != nil {
					// log error and ignore
					pipe.Logger.Printf("Error setting user.longitude: %v", err.Error())
					return
				}
			}
		}
	}

	pipe.DataLog.Checkpoint = entity.DataLogCheckpointHookOnValidationExecuted
}
