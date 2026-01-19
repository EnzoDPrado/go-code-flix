package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"errors"
	"os"
	"strconv"
)

type JobService struct {
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (j *JobService) Start(job *domain.Job, video *domain.Video) error {
	err := j.changeJobStatus(job, "DOWNLOADING")

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.VideoService.Download(video, os.Getenv("inputBucketName"))

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.changeJobStatus(job, "FRAGMENTING")

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.VideoService.Fragment(video)

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.changeJobStatus(job, "ENCODING")

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.VideoService.Encode(video)

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.changeJobStatus(job, "UPLOADING")

	err = j.performUpload(job, video)

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.changeJobStatus(job, "FINISHING")

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.VideoService.Finish(video)

	if err != nil {
		return j.failJob(job, err)
	}

	err = j.changeJobStatus(job, "COMPLETED")

	if err != nil {
		return j.failJob(job, err)
	}

	return nil
}

func (j *JobService) performUpload(job *domain.Job, video *domain.Video) error {
	err := j.changeJobStatus(job, "UPLOADING")

	if err != nil {
		return j.failJob(job, err)
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("outputBucketName")
	videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + video.ID
	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOAD"))

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(concurrency, doneUpload)

	var uploadResult string
	uploadResult = <-doneUpload

	if uploadResult != "Upload completed" {
		return j.failJob(job, errors.New(uploadResult))
	}

	return err
}

func (j *JobService) changeJobStatus(job *domain.Job, status string) error {
	var err error

	job.Status = status
	job, err = j.JobRepository.Update(job)

	if err != nil {
		return j.failJob(job, err)
	}

	return nil
}

func (j *JobService) failJob(job *domain.Job, error error) error {

	job.Status = "FAILED"
	job.Error = error.Error()

	_, err := j.JobRepository.Update(job)

	if err != nil {
		return err
	}

	return error
}
