package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

var (
	SendMessageTimeoutInSecs  int64  = 25
	SendMessageEndpoint       string = "/api/message.send"
	TransactionalMessageQueue        = "messages_transactional"
	MarketingMessageQueue            = "messages_marketing"
)

func (pipe *DataLogPipeline) StepExecuteSpecialAction(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepExecuteSpecialAction")
	defer span.End()

	if !pipe.DataLog.IsSpecialAction() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointSpecialActionExecuted
		pipe.ProcessNextStep(spanCtx)
		return
	}

	// don't acquire user lock for sending messages

	switch pipe.DataLog.Kind {
	case entity.ItemKindMessage:
		// only send messages on create
		if pipe.DataLog.Action == "create" {
			EnqueueMessage(spanCtx, pipe)
		}
	default:
		pipe.SetError("server", fmt.Sprintf("unknown special action: %v", pipe.DataLog.Kind), true)
		return
	}

	// set status
	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointSpecialActionExecuted
	}
}

func EnqueueMessage(ctx context.Context, pipe *DataLogPipeline) {

	spanCtx, span := trace.StartSpan(ctx, "EnqueueMessage")
	defer span.End()

	if pipe.DataLog.UpsertedMessage == nil {
		pipe.SetError("server", "EnqueueMessage: no message object", true)
		return
	}

	// inbound messages should not be sent :)
	if pipe.DataLog.UpsertedMessage.IsInbound {
		return
	}

	if pipe.DataLog.UpsertedMessage.Channel != "email" {
		pipe.SetError("server", fmt.Sprintf("EnqueueMessage: channel not implemented: %v", pipe.DataLog.UpsertedMessage.Channel), true)
		return
	}

	if pipe.DataLog.UpsertedMessage.MessageTemplate == nil {
		pipe.SetError("server", "EnqueueMessage: no message template object", true)
		return
	}

	if pipe.DataLog.UpsertedUser.Email == nil || pipe.DataLog.UpsertedUser.Email.String == "" {
		pipe.SetError("user", "EnqueueMessage: user has no email address", false)
		return
	}

	// shortcut
	template := pipe.DataLog.UpsertedMessage.MessageTemplate

	// data payload
	jsonDataBytes, err := json.Marshal(pipe.DataLog.UpsertedMessage.Data)

	if err != nil {
		pipe.SetError("server", fmt.Sprintf("send message json err %v", err), false)
		return
	}

	jsonData := string(jsonDataBytes)

	// build content
	subject, err := CompileNunjucksTemplate(template.Email.Subject, jsonData)

	if err != nil {
		pipe.SetError("server", fmt.Sprintf("nunjucks subject err %v", err), false)
		return
	}

	html, err := CompileNunjucksTemplate(template.Email.Content, jsonData)

	if err != nil {
		pipe.SetError("server", fmt.Sprintf("nunjucks html err %v", err), false)
		return
	}

	var text string
	if template.Email.Text != nil {
		text, err = CompileNunjucksTemplate(*template.Email.Text, jsonData)
		if err != nil {
			pipe.SetError("server", fmt.Sprintf("nunjucks text err %v", err), false)
			return
		}
	}

	log.Printf("subject: %v", subject)
	log.Printf("html: %v", html)
	log.Printf("text: %v", text)

	queueName := MarketingMessageQueue

	if pipe.DataLog.UpsertedMessage.IsTransactional != nil && *pipe.DataLog.UpsertedMessage.IsTransactional {
		queueName = TransactionalMessageQueue
	}

	payload := dto.SendMessage{
		WorkspaceID:         pipe.Workspace.ID,
		MessageID:           pipe.DataLog.UpsertedMessage.ID,
		MessageExternalID:   pipe.DataLog.UpsertedMessage.ExternalID,
		UserID:              pipe.DataLog.UpsertedUser.ID,
		UserExternalID:      pipe.DataLog.UpsertedUser.ExternalID,
		UserIsAuthenticated: pipe.DataLog.UpsertedUser.IsAuthenticated,
		Channel:             pipe.DataLog.UpsertedMessage.Channel,
		ScheduledAt:         pipe.DataLog.UpsertedMessage.ScheduledAt,
		Email: &dto.SendMessageEmail{
			Subject:         subject,
			HTML:            html,
			Text:            text,
			IsTransactional: *pipe.DataLog.UpsertedMessage.IsTransactional,
			// TODO:
			Provider: "sparkpost",
			SparkPostCrendentials: &dto.SparkPostCrendentials{
				EncryptedApiKey: "TODO",
			},
		},
	}

	// enqueue email
	job := &taskorchestrator.TaskRequest{
		QueueLocation:     pipe.Config.TASK_QUEUE_LOCATION,
		QueueName:         queueName,
		PostEndpoint:      pipe.Config.API_ENDPOINT + SendMessageEndpoint + "?workspace_id=" + pipe.Workspace.ID,
		TaskTimeoutInSecs: &TaskTimeoutInSecs,
		Payload:           payload,
	}

	if payload.ScheduledAt != nil {
		job.ScheduleTime = payload.ScheduledAt
	}

	if err := pipe.TaskOrchestrator.PostRequest(spanCtx, job); err != nil {
		pipe.SetError("server", fmt.Sprintf("enqueue message err %v", err), true)
		return
	}
}

func CompileNunjucksTemplate(templateString string, jsonData string) (result string, err error) {

	templateStringB64 := base64.StdEncoding.EncodeToString([]byte(templateString))
	dataB64 := base64.StdEncoding.EncodeToString([]byte(jsonData))

	dir, err := GetNodeJSDir()
	if err != nil {
		return "", eris.Wrap(err, "CompileNunjucksTemplate")
	}

	scriptPath := dir + "nunjucks.js"

	// call nodejs cmd
	output, err := exec.Command("node", scriptPath, templateStringB64, dataB64).Output()

	log.Printf("output: %v", string(output))
	log.Printf("err: %v", err)

	if err != nil {
		return "", eris.Wrap(err, "CompileNunjucksTemplate")
	}

	return string(output), nil
}

func GenerateDoubleOptInLink(endpoint string, secretKey string, list *entity.SubscriptionList, user *entity.User) (token string, err error) {

	// create a token with custom claims
	pasetoToken := paseto.NewToken()
	pasetoToken.SetAudience(endpoint)
	pasetoToken.SetIssuedAt(time.Now())
	pasetoToken.SetString("lid", list.ID)
	pasetoToken.SetString("lname", list.Name)
	pasetoToken.SetString("email", user.Email.String)

	if user.IsAuthenticated {
		pasetoToken.SetString("auth_uid", user.ExternalID)
	} else {
		pasetoToken.SetString("anon_uid", user.ExternalID)
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(secretKey))
	if err != nil {
		return "", eris.Wrap(err, "GenerateDoubleOptInLink V4SymmetricKeyFromBytes")
	}

	return endpoint + entity.DoubleOptInPath + "?token=" + pasetoToken.V4Encrypt(key, nil), nil
}

func GenerateEmailUnsubscribeLink(endpoint string, secretKey string, list *entity.SubscriptionList, user *entity.User) (token string, err error) {

	// create a token with custom claims
	pasetoToken := paseto.NewToken()
	pasetoToken.SetAudience(endpoint)
	pasetoToken.SetIssuedAt(time.Now())
	pasetoToken.SetString("lid", list.ID)
	pasetoToken.SetString("lname", list.Name)
	pasetoToken.SetString("email", user.Email.String)

	if user.IsAuthenticated {
		pasetoToken.SetString("auth_uid", user.ExternalID)
	} else {
		pasetoToken.SetString("anon_uid", user.ExternalID)
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(secretKey))
	if err != nil {
		return "", eris.Wrap(err, "GenerateEmailUnsubscribeLink V4SymmetricKeyFromBytes")
	}

	return endpoint + entity.UnsubscribeEmailPath + "?token=" + pasetoToken.V4Encrypt(key, nil), nil
}
