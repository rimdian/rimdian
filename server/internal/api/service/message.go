package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rimdian/rimdian/internal/api/dto"
	common "github.com/rimdian/rimdian/internal/common/dto"
	mail "github.com/xhit/go-simple-mail/v2"
	"go.opencensus.io/trace"
)

func (svc *ServiceImpl) MessageSend(ctx context.Context, data *dto.SendMessage) (result *common.ResponseForTaskQueue) {

	spanCtx, span := trace.StartSpan(ctx, "MessageSend")
	defer span.End()

	// data := dto.SendMessage{
	// 	WorkspaceID:         pipe.Workspace.ID,
	// 	MessageID:           pipe.DataLog.UpsertedMessage.ID,
	// 	MessageExternalID:   pipe.DataLog.UpsertedMessage.ExternalID,
	// 	UserID:              pipe.DataLog.UpsertedUser.ID,
	// 	UserExternalID:      pipe.DataLog.UpsertedUser.ExternalID,
	// 	UserIsAuthenticated: pipe.DataLog.UpsertedUser.IsAuthenticated,
	// 	Channel:             pipe.DataLog.UpsertedMessage.Channel,
	// 	ScheduledAt:         pipe.DataLog.UpsertedMessage.ScheduledAt,
	// 	Email: &dto.SendMessageEmail{
	// 		FromAdrress:     template.Email.FromAdrress,
	// 		FromName:        template.Email.FromName,
	// 		ReplyTo:         template.Email.ReplyTo,
	// 		ToAdrress:       pipe.DataLog.UpsertedUser.Email.String,
	// 		Subject:         subject,
	// 		HTML:            html,
	// 		Text:            text,
	// 		IsTransactional: isTransactional,
	// 		EmailProvider:   emailProvider,
	// 	},
	// }

	if data.Channel == "email" {

		// decrypt credentials
		if err := data.Email.EmailProvider.Decrypt(svc.Config.SECRET_KEY); err != nil {
			return &common.ResponseForTaskQueue{
				HasError:         true,
				QueueShouldRetry: false,
				Error:            "decrypt error",
			}
		}

		if data.Email.EmailProvider.Provider == "sparkpost" {
			result = svc.SendEmailWithSparkpost(spanCtx, data)
		} else if data.Email.EmailProvider.Provider == "smtp" {
			result = svc.SendEmailWithSMTP(spanCtx, data)
		} else {
			result = &common.ResponseForTaskQueue{
				HasError:         true,
				QueueShouldRetry: false,
				Error:            "no email provider",
			}
		}
	} else {
		result = &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: false,
			Error:            "no channel implemented",
		}
	}

	return
}

func (svc *ServiceImpl) SendEmailWithSparkpost(ctx context.Context, data *dto.SendMessage) (result *common.ResponseForTaskQueue) {

	spanCtx, span := trace.StartSpan(ctx, "SendEmailWithSparkpost")
	defer span.End()

	msg := dto.SparkPostMessage{
		Options: dto.SparkPostOptions{
			Transactional: data.Email.IsTransactional,
			// tracking is managed by Rimdian
			ClickTracking: false,
			InlineCSS:     false,
			OpenTracking:  false,
		},
		Recipients: []dto.SparkPostRecipient{
			{
				Address: dto.SparkPostAddress{
					Email: data.Email.ToAdrress,
				},
			},
		},
		Content: dto.SparkPostcontent{
			From: dto.SparkPostAddress{
				Email: data.Email.FromAdrress,
				Name:  &data.Email.FromName,
			},
			Subject: data.Email.Subject,
		},
	}

	if data.Email.HTML != "" {
		msg.Content.HTML = &data.Email.HTML
	}

	if data.Email.Text != "" {
		msg.Content.Text = &data.Email.Text
	}

	// marshal
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: false,
			Error:            fmt.Sprintf("SendEmailWithSparkpost error: %v", err),
		}
	}

	req, _ := http.NewRequestWithContext(spanCtx, "POST", *data.Email.EmailProvider.Host+"/api/v1/transmissions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", *data.Email.EmailProvider.Password)

	res, err := svc.NetClient.Do(req)

	if err != nil {
		return &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: true,
			Error:            fmt.Sprintf("SendEmailWithSparkpost error: %v", err),
		}
	}

	defer res.Body.Close()

	// read body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: true,
			Error:            fmt.Sprintf("SendEmailWithSparkpost error: %v", err),
		}
	}

	response := string(body)

	if res.StatusCode < 300 {
		return &common.ResponseForTaskQueue{
			HasError:         false,
			QueueShouldRetry: false,
			Error:            "",
		}
	} else if res.StatusCode == 420 {
		// 420 is a rate limit error
		return &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: true,
			Error:            fmt.Sprintf("SendEmailWithSparkpost error: %v", response),
		}
	} else {
		return &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: false,
			Error:            fmt.Sprintf("SendEmailWithSparkpost error: %v", response),
		}
	}
}

func (svc *ServiceImpl) SendEmailWithSMTP(ctx context.Context, data *dto.SendMessage) (result *common.ResponseForTaskQueue) {

	// spanCtx, span := trace.StartSpan(ctx, "SendEmailWithSMTP")
	// defer span.End()

	server := mail.NewSMTPClient()
	server.Host = *data.Email.EmailProvider.Host
	server.Port = *data.Email.EmailProvider.Port
	server.Username = *data.Email.EmailProvider.Username
	server.Password = *data.Email.EmailProvider.Password

	switch *data.Email.EmailProvider.Encryption {
	case "SSL":
		server.Encryption = mail.EncryptionSSLTLS
	case "STARTTLS":
		server.Encryption = mail.EncryptionSTARTTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	server.Authentication = mail.AuthPlain
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	smtpClient, err := server.Connect()

	if err != nil {
		return &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: false,
			Error:            fmt.Sprintf("SendEmailWithSMTP error: %v", err),
		}
	}

	from := fmt.Sprintf("%s <%s>", data.Email.FromName, data.Email.FromAdrress)

	email := mail.NewMSG()
	email.SetFrom(from).SetSubject(data.Email.Subject).AddTo(data.Email.ToAdrress)

	if data.Email.ReplyTo != nil {
		email.SetReplyTo(*data.Email.ReplyTo)
	}

	if data.Email.HTML != "" {
		email.SetBody(mail.TextHTML, data.Email.HTML)
	}

	if data.Email.Text != "" {
		email.AddAlternative(mail.TextPlain, data.Email.Text)
	}

	if err := email.Send(smtpClient); err != nil {
		return &common.ResponseForTaskQueue{
			HasError:         true,
			QueueShouldRetry: false,
			Error:            fmt.Sprintf("SendEmailWithSMTP error: %v", err),
		}
	}

	return
}
