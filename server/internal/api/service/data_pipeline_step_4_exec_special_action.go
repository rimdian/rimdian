package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"github.com/tidwall/sjson"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) StepExecuteSpecialAction(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepExecuteSpecialAction")
	defer span.End()

	if !pipe.DataLog.IsSpecialAction() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointSpecialActionExecuted
		pipe.ProcessNextStep(spanCtx)
		return
	}

	// don't lock users for sending messages

	switch pipe.DataLog.Kind {
	case entity.ItemKindMessage:
		SendMessage(spanCtx, pipe)
	default:
		pipe.SetError("server", fmt.Sprintf("unknown special action: %v", pipe.DataLog.Kind), true)
		return
	}

	// set status
	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointSpecialActionExecuted
	}
}

func SendMessage(ctx context.Context, pipe *DataLogPipeline) {

	// spanCtx, span := trace.StartSpan(ctx, "SendMessage")
	// defer span.End()

	if pipe.DataLog.UpsertedMessage == nil {
		pipe.SetError("server", "SendMessage: no message object", true)
		return
	}

	// inbound messages should not be sent :)
	if pipe.DataLog.UpsertedMessage.IsInbound {
		return
	}

	if pipe.DataLog.UpsertedMessage.Channel != "email" {
		pipe.SetError("server", fmt.Sprintf("SendMessage: channel not implemented: %v", pipe.DataLog.UpsertedMessage.Channel), true)
		return
	}

	if pipe.DataLog.UpsertedMessage.MessageTemplate == nil {
		pipe.SetError("server", "SendMessage: no message template object", true)
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

	// attach user data
	jsonUser, err := json.Marshal(pipe.DataLog.UpsertedUser)
	if err != nil {
		pipe.SetError("server", fmt.Sprintf("send message json err %v", err), false)
		return
	}

	jsonData := string(jsonDataBytes)

	if jsonData, err = sjson.SetRaw(jsonData, "user", string(jsonUser)); err != nil {
		pipe.SetError("server", fmt.Sprintf("send message json err %v", err), false)
		return
	}

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

	// inject double opt-in / unsubscribe link
	if pipe.DataLog.UpsertedMessage.SubscriptionList != nil {

		// check if template has DoubleOptInKeyword
		if strings.Contains(html, entity.DoubleOptInKeyword) {
			doubleOptInLink := "TODO"
			html = strings.ReplaceAll(html, entity.DoubleOptInKeyword, doubleOptInLink)
			text = strings.ReplaceAll(text, entity.DoubleOptInKeyword, doubleOptInLink)
		}

		if strings.Contains(template.Email.Content, entity.UnsubscribeKeyword) {
			unsubLink := "TODO"
			html = strings.ReplaceAll(html, entity.UnsubscribeKeyword, unsubLink)
			text = strings.ReplaceAll(text, entity.UnsubscribeKeyword, unsubLink)
		}
	}

	log.Printf("subject: %v", subject)
	log.Printf("html: %v", html)
	log.Printf("text: %v", text)

	// send email
	// TODO
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
