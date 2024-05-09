package services

import (
	"errors"
	"os"
	"strconv"

	"github.com/jdaibello/fc-micro-encoder/application/repositories"
	"github.com/jdaibello/fc-micro-encoder/domain"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (j *JobService) Start() error {
	err := j.ChangeJobStatus("DOWNLOADING")

	if err != nil {
		return j.FailJob(err)
	}

	err = j.VideoService.Download(os.Getenv("INPUT_BUCKET_NAME"))

	if err != nil {
		return j.FailJob(err)
	}

	err = j.ChangeJobStatus("FRAGMENTING")

	if err != nil {
		return j.FailJob(err)
	}

	err = j.VideoService.Fragment()

	if err != nil {
		return j.FailJob(err)
	}

	err = j.ChangeJobStatus("ENCODING")

	if err != nil {
		return j.FailJob(err)
	}

	err = j.VideoService.Encode()

	if err != nil {
		return j.FailJob(err)
	}

	err = j.PerformUpload()

	if err != nil {
		return j.FailJob(err)
	}

	err = j.ChangeJobStatus("FINISHING")

	if err != nil {
		return j.FailJob(err)
	}

	err = j.VideoService.Finish()

	if err != nil {
		return j.FailJob(err)
	}

	err = j.ChangeJobStatus("COMPLETED")

	if err != nil {
		return j.FailJob(err)
	}

	return nil
}

func (j *JobService) PerformUpload() error {
	err := j.ChangeJobStatus("UPLOADING")

	if err != nil {
		return j.FailJob(err)
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("OUTPUT_BUCKET_NAME")
	videoUpload.VideoPath = os.Getenv("LOCAL_STORAGE_PATH") + "/" + j.VideoService.Video.ID

	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOAD"))
	doneUpload := make(chan string)

	go videoUpload.ProcessUpload(concurrency, doneUpload)

	var uploadResult = <-doneUpload

	if uploadResult != "Upload completed" {
		return j.FailJob(errors.New(uploadResult))
	}

	return err
}

func (j *JobService) ChangeJobStatus(status string) error {
	var err error

	j.Job.Status = status
	j.Job, err = j.JobRepository.Update(j.Job)

	if err != nil {
		return j.FailJob(err)
	}

	return nil
}

func (j *JobService) FailJob(error error) error {
	j.Job.Status = "FAILED"
	j.Job.Error = error.Error()

	_, err := j.JobRepository.Update(j.Job)

	if err != nil {
		return err
	}

	return error
}
