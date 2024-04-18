//go:generate moq -out taskorchestrator_client_mock.go . Client
package taskorchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rotisserie/eris"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrJsonDataImportTask = eris.New("cannot json data task")

	ENV_DEV  = "development"
	ENV_TEST = "test"
	ENV_PROD = "production"
)

type TaskRequest struct {
	QueueLocation     string
	QueueName         string
	UniqueID          *string // deduplication key
	PostEndpoint      string
	Payload           interface{}
	ScheduleTime      *time.Time
	TaskTimeoutInSecs *int64
}

func (t *TaskRequest) Validate() error {
	if t.QueueLocation == "" {
		return eris.New("task queue location is required")
	}
	if t.QueueName == "" {
		return eris.New("task queue name is required")
	}
	if t.UniqueID != nil && *t.UniqueID == "" {
		return eris.New("task queue unique id is required")
	}
	if !govalidator.IsURL(t.PostEndpoint) {
		return eris.New("task queue post endpoint is not valid URL")
	}
	if t.Payload == nil {
		return eris.New("task queue payload is required")
	}
	return nil
}

type Client interface {
	EnsureQueue(ctx context.Context, queueLocation string, queueName string, maxConcurrent int32) error
	PostRequest(ctx context.Context, taskRequest *TaskRequest) error
	GetHistoricalQueueNameForWorkspace(workspaceID string) string
	GetLiveQueueNameForWorkspace(workspaceID string) string
	GetTransactionalMessageQueueNameForWorkspace(workspaceID string) string
	GetMarketingMessageQueueNameForWorkspace(workspaceID string) string
	GetTaskRunningJob(ctx context.Context, queueLocation string, queueName string, taskID string) (jobInfo *dto.TaskExecJobInfoInfo, err error)
	DeleteTask(ctx context.Context, queueLocation string, queueName string, taskID string) error
}

type ClientImpl struct {
	GcloudProject string
	Env           string
	CloudTask     *cloudtasks.Client
}

func NewClient(gcloudProject string, env string, cloudClient *cloudtasks.Client) Client {
	return &ClientImpl{
		GcloudProject: gcloudProject,
		Env:           env,
		CloudTask:     cloudClient,
	}
}

func (client *ClientImpl) DeleteTask(ctx context.Context, queueLocation string, queueName string, taskID string) (err error) {

	// See https://pkg.go.dev/cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb#DeleteTaskRequest.
	if err = client.CloudTask.DeleteTask(ctx, &cloudtaskspb.DeleteTaskRequest{
		// projects/PROJECT_ID/locations/LOCATION_ID/queues/QUEUE_ID/tasks/TASK_ID
		Name: fmt.Sprintf("projects/%s/locations/%s/queues/%s/tasks/%s", client.GcloudProject, queueLocation, queueName, taskID),
	}); err != nil {
		// should error if task not found
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		log.Printf("cloudtasks.DeleteTask error: %v", err)
		return eris.Wrap(err, "DeleteTask")
	}

	return nil
}

func (client *ClientImpl) GetTaskRunningJob(ctx context.Context, queueLocation string, queueName string, taskID string) (jobInfo *dto.TaskExecJobInfoInfo, err error) {

	// projects/PROJECT_ID/locations/LOCATION_ID/queues/QUEUE_ID/tasks/TASK_ID
	name := fmt.Sprintf("projects/%s/locations/%s/queues/%s/tasks/%s", client.GcloudProject, queueLocation, queueName, taskID)

	// See https://pkg.go.dev/cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb#GetTaskRequest.
	task, err := client.CloudTask.GetTask(ctx, &cloudtaskspb.GetTaskRequest{
		Name: name,
	})
	if err != nil {
		return nil, eris.Wrap(err, "GetTaskRunningJob")
	}

	jobInfo = &dto.TaskExecJobInfoInfo{
		ID:            task.GetName(),
		DispatchCount: task.GetDispatchCount(),
		ResponseCount: task.GetResponseCount(),
		// DispatchCount int32               `json:"dispatch_count"`
		// ResponseCount int32               `json:"response_count"`
		// FirstAttempt  *TaskRunningJobAttempt `json:"first_attempt,omitempty"`
		// LastAttempt   *TaskRunningJobAttempt `json:"last_attempt,omitempty"`
	}

	if task.GetCreateTime() != nil {
		t := task.GetCreateTime().AsTime()
		jobInfo.CreateTime = &t
	}

	if task.GetScheduleTime() != nil {
		t := task.GetScheduleTime().AsTime()
		jobInfo.ScheduleTime = &t
	}

	if task.GetFirstAttempt() != nil {
		jobInfo.FirstAttempt = &dto.TaskExecJobInfoAttempt{}
		if task.GetFirstAttempt().GetDispatchTime() != nil {
			t := task.GetFirstAttempt().GetDispatchTime().AsTime()
			jobInfo.FirstAttempt.DispatchTime = &t
		}
		if task.GetFirstAttempt().GetResponseTime() != nil {
			t := task.GetFirstAttempt().GetResponseTime().AsTime()
			jobInfo.FirstAttempt.ResponseTime = &t
		}
		if task.GetFirstAttempt().GetResponseStatus() != nil {
			code := task.GetFirstAttempt().GetResponseStatus().GetCode()
			jobInfo.FirstAttempt.ResponseCode = &code

			message := task.GetFirstAttempt().GetResponseStatus().GetMessage()
			jobInfo.FirstAttempt.ReponseMessage = &message
		}
	}

	if task.GetLastAttempt() != nil {
		jobInfo.LastAttempt = &dto.TaskExecJobInfoAttempt{}
		if task.GetLastAttempt().GetDispatchTime() != nil {
			t := task.GetLastAttempt().GetDispatchTime().AsTime()
			jobInfo.LastAttempt.DispatchTime = &t
		}
		if task.GetLastAttempt().GetResponseTime() != nil {
			t := task.GetLastAttempt().GetResponseTime().AsTime()
			jobInfo.LastAttempt.ResponseTime = &t
		}
		if task.GetLastAttempt().GetResponseStatus() != nil {
			code := task.GetLastAttempt().GetResponseStatus().GetCode()
			jobInfo.LastAttempt.ResponseCode = &code

			message := task.GetLastAttempt().GetResponseStatus().GetMessage()
			jobInfo.LastAttempt.ReponseMessage = &message
		}
	}

	return jobInfo, err
}

func (client *ClientImpl) GetHistoricalQueueNameForWorkspace(workspaceID string) string {
	return strings.ReplaceAll(workspaceID+"-data-imports", "_", "-")
}

func (client *ClientImpl) GetLiveQueueNameForWorkspace(workspaceID string) string {
	return strings.ReplaceAll(workspaceID+"-data-imports-live", "_", "-")
}

func (client *ClientImpl) GetTransactionalMessageQueueNameForWorkspace(workspaceID string) string {
	return strings.ReplaceAll(workspaceID+"-messages-transactional", "_", "-")
}

func (client *ClientImpl) GetMarketingMessageQueueNameForWorkspace(workspaceID string) string {
	return strings.ReplaceAll(workspaceID+"-messages-marketing", "_", "-")
}

func (client *ClientImpl) EnsureQueue(ctx context.Context, queueLocation string, queueName string, maxConcurrentDispatches int32) error {

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", client.GcloudProject, queueLocation, queueName)

	q, err := client.CloudTask.GetQueue(ctx, &cloudtaskspb.GetQueueRequest{
		Name: queuePath,
	})

	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return eris.Wrap(err, "EnsureQueue")
	}

	if q == nil {
		_, err = client.CloudTask.CreateQueue(ctx, &cloudtaskspb.CreateQueueRequest{
			Parent: fmt.Sprintf("projects/%s/locations/%s", client.GcloudProject, queueLocation),
			Queue: &cloudtaskspb.Queue{
				Name: queuePath,
				RateLimits: &cloudtaskspb.RateLimits{
					MaxDispatchesPerSecond:  200,
					MaxBurstSize:            200,
					MaxConcurrentDispatches: maxConcurrentDispatches,
				},
				RetryConfig: &cloudtaskspb.RetryConfig{
					MaxAttempts: 100,
					MaxRetryDuration: &durationpb.Duration{
						Seconds: 0,
						// Seconds: 60 * 60 * 24 * 7,
					},
					MinBackoff: &durationpb.Duration{
						Seconds: 5,
					},
					MaxBackoff: &durationpb.Duration{
						Seconds: 3600,
					},
					MaxDoublings: 16,
				},
			},
		})
		if err != nil {
			return eris.Wrapf(err, "cannot create queue %s", queuePath)
		}
	}

	return nil
}

func (client *ClientImpl) PostRequest(ctx context.Context, taskRequest *TaskRequest) error {

	if err := taskRequest.Validate(); err != nil {
		return err
	}

	// skip in dev
	if client.Env == ENV_DEV {
		return nil
	}

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", client.GcloudProject, taskRequest.QueueLocation, taskRequest.QueueName)

	jsonData, err := json.Marshal(taskRequest.Payload)

	if err != nil {
		log.Println(err)
		return eris.Wrap(ErrJsonDataImportTask, "EnqueueJob")
	}

	req := &cloudtaskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &cloudtaskspb.Task{
			// Name: ,
			MessageType: &cloudtaskspb.Task_HttpRequest{
				HttpRequest: &cloudtaskspb.HttpRequest{
					Url:        taskRequest.PostEndpoint,
					HttpMethod: cloudtaskspb.HttpMethod_POST,
					Body:       jsonData,
				},
			},
		},
	}

	// use a specific task name to retrieve its status later
	if taskRequest.UniqueID != nil && *taskRequest.UniqueID != "" {
		req.Task.Name = fmt.Sprintf("%v/tasks/%v", queuePath, *taskRequest.UniqueID)
	}

	// retry task if not returned after deadline
	if taskRequest.TaskTimeoutInSecs != nil {
		req.Task.DispatchDeadline = &durationpb.Duration{
			Seconds: *taskRequest.TaskTimeoutInSecs,
		}
	}

	// deliver in the future
	if taskRequest.ScheduleTime != nil {
		req.Task.ScheduleTime = &timestamppb.Timestamp{
			Seconds: taskRequest.ScheduleTime.Unix(),
		}
	}

	if _, err := client.CloudTask.CreateTask(ctx, req); err != nil {
		log.Printf("cloudtasks.CreateTask error: %v, %v", err, req.Task.Name)
		return eris.Wrap(err, "EnqueueJob")
	}

	return nil
}
