package services_test

import (
	"os"
	"testing"

	"github.com/jdaibello/fc-micro-encoder/application/services"
	"github.com/stretchr/testify/require"
)

func TestVideoServiceUpload(t *testing.T) {
	video, repo := Prepare()
	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo

	err := videoService.Download("micro-encoder")
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)

	err = videoService.Encode()
	require.Nil(t, err)

	videoUpload := services.NewVideoUpload()

	videoUpload.OutputBucket = "micro-encoder"
	videoUpload.VideoPath = os.Getenv("LOCAL_STORAGE_PATH") + "/" + video.ID

	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(50, doneUpload)

	result := <-doneUpload

	require.Equal(t, result, "Upload completed")

	err = videoService.Finish()
	require.Nil(t, err)
}
