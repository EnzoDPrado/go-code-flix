package services_test

import (
	"encoder/application/services"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func TestVideoServiceUpload(t *testing.T) {
	video, repo := prepare()

	videoService := services.NewVideoService()
	videoService.VideoRepository = repo

	err := videoService.Download(video, "bucket-study-go")
	require.Nil(t, err)

	err = videoService.Fragment(video)
	require.Nil(t, err)

	err = videoService.Encode(video)
	require.Nil(t, err)

	videoUpload := services.NewVideoUpload()
	videoUpload.OutputBucket = "bucket-study-go"
	videoUpload.VideoPath = os.Getenv("localStoragePath") + "/" + video.ID

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(50, doneUpload)

	result := <-doneUpload
	require.Equal(t, result, "Upload completed")

	err = videoService.Finish(video)
	require.Nil(t, err)

}
