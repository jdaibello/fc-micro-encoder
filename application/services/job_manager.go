package services

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/jdaibello/fc-micro-encoder/application/repositories"
	"github.com/jdaibello/fc-micro-encoder/domain"
	"github.com/jdaibello/fc-micro-encoder/framework/queue"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type JobManager struct {
	Db               *gorm.DB
	Domain           domain.Job
	MessageChannel   chan amqp.Delivery
	JobReturnChannel chan JobWorkerResult
	RabbitMQ         *queue.RabbitMQ
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewJobManage(db *gorm.DB, rabbitMQ *queue.RabbitMQ, jobReturnChannel chan JobWorkerResult, messageChannel chan amqp.Delivery) *JobManager {
	return &JobManager{
		Db:               db,
		Domain:           domain.Job{},
		MessageChannel:   messageChannel,
		JobReturnChannel: jobReturnChannel,
		RabbitMQ:         rabbitMQ,
	}
}

func (j *JobManager) Start(ch *amqp.Channel) {
	videoService := NewVideoService()

	videoService.VideoRepository = repositories.VideoRepositoryDb{Db: j.Db}

	jobService := JobService{
		JobRepository: repositories.JobRepositoryDb{Db: j.Db},
		VideoService:  videoService,
	}

	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKERS"))

	if err != nil {
		log.Fatalf("error loading var: CONCURRENCY_WORKERS")
	}

	for processQuantity := 0; processQuantity < concurrency; processQuantity++ {
		go JobWorker(j.MessageChannel, j.JobReturnChannel, jobService, j.Domain, processQuantity)
	}

	for jobResult := range j.JobReturnChannel {
		if jobResult.Error != nil {
			err = j.CheckParseErrors(jobResult)
		} else {
			err = j.NotifySuccess(jobResult, ch)
		}

		if err != nil {
			jobResult.Message.Reject(false)
		}
	}
}

func (j *JobManager) NotifySuccess(jobResult JobWorkerResult, ch *amqp.Channel) error {
	jobJson, err := json.Marshal(jobResult.Job)

	if err != nil {
		return err
	}

	err = j.Notify(jobJson)

	if err != nil {
		return err
	}

	err = jobResult.Message.Ack(false)

	if err != nil {
		return err
	}

	return nil
}

func (j *JobManager) CheckParseErrors(jobResult JobWorkerResult) error {
	if jobResult.Job.ID != "" {
		log.Printf("MessageID #{jobResult.Message.DeliveryTag}. Error parsing job: #{jobResult.Job.ID}")
	} else {
		log.Printf("MessageID #{jobResult.Message.DeliveryTag}. Error parsing message: #{jobResult.Error}")
	}

	errorMessage := JobNotificationError{
		Message: string(jobResult.Message.Body),
		Error:   jobResult.Error.Error(),
	}

	jobJson, err := json.Marshal(errorMessage)

	if err != nil {
		return err
	}

	err = j.Notify(jobJson)

	if err != nil {
		return err
	}

	err = jobResult.Message.Reject(false)

	if err != nil {
		return err
	}

	return nil
}

func (j *JobManager) Notify(jobJson []byte) error {
	err := j.RabbitMQ.Notify(string(jobJson), "application/json", os.Getenv("RABBITMQ_NOTIFICATION_EX"), os.Getenv("RABBITMQ_NOTIFICATION_ROUTING_KEY"))

	if err != nil {
		return err
	}

	return nil
}
