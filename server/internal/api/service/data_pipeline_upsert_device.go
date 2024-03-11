package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os/exec"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) ParseUserAgent(userAgent string) (result *entity.UserAgentResult, err error) {
	userAgentB64 := base64.StdEncoding.EncodeToString([]byte(userAgent))

	dir, err := GetNodeJSDir()
	if err != nil {
		return nil, eris.Wrap(err, "ParseUserAgent")
	}

	scriptPath := dir + "parse-ua.js"

	// call nodejs cmd
	output, err := exec.Command("node", scriptPath, userAgentB64).Output()

	// svc.Logger.Printf("output: %v", string(output))
	// svc.Logger.Printf("err: %v", err)
	if err != nil {
		return nil, eris.Wrap(err, "ParseUserAgent")
	}

	// unmarshal response
	result = &entity.UserAgentResult{}
	if err = json.Unmarshal([]byte(output), result); err != nil {
		return nil, eris.Wrap(err, "ParseUserAgent")
	}

	return result, nil
}

func (pipe *DataLogPipeline) UpsertDevice(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertDevice")
	defer span.End()

	// find eventual existing device
	var existingDevice *entity.Device
	updatedFields := []*entity.UpdatedField{}

	existingDevice, err = pipe.Repository.FindDeviceByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedDevice.ID, pipe.DataLog.UpsertedDevice.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "DeviceUpsert")
	}

	// insert new device
	if existingDevice == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedDevice.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertDevice(spanCtx, pipe.DataLog.UpsertedDevice, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, "device", "create", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedDevice.ID, pipe.DataLog.UpsertedDevice.ExternalID, updatedFields, *pipe.DataLog.UpsertedDevice.UpdatedAt, tx); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		return
	}

	// merge fields if device already exists
	updatedFields = pipe.DataLog.UpsertedDevice.MergeInto(existingDevice)
	pipe.DataLog.UpsertedDevice = existingDevice

	// abort if no fields were updated
	if len(updatedFields) == 0 {
		if !isChild {
			pipe.DataLog.Action = "noop"
		}
		return nil
	}

	if !isChild {
		pipe.DataLog.Action = "update"
		pipe.DataLog.UpdatedFields = updatedFields
	}

	// persist changes
	if err = pipe.Repository.UpdateDevice(spanCtx, pipe.DataLog.UpsertedDevice, tx); err != nil {
		return eris.Wrap(err, "DeviceUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, "device", "update", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedDevice.ID, pipe.DataLog.UpsertedDevice.ExternalID, updatedFields, *pipe.DataLog.UpsertedDevice.UpdatedAt, tx); err != nil {
			return err
		}
	}

	return nil
}
