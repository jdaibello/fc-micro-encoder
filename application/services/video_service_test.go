package services_test

import (
	"log"
	"testing"
	"time"

	"github.com/jdaibello/fc-micro-encoder/application/repositories"
	"github.com/jdaibello/fc-micro-encoder/application/services"
	"github.com/jdaibello/fc-micro-encoder/domain"
	"github.com/jdaibello/fc-micro-encoder/framework/database"
	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func Init() {
	err := godotenv.Load("../../.env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func Prepare() (*domain.Video, repositories.VideoRepositoryDb) {
	db := database.NewDbTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "paisagem.mp4"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepositoryDb{Db: db}

	return video, repo
}

func TestVideoServiceDownload(t *testing.T) {
	video, repo := Prepare()
	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo

	err := videoService.Download("micro-encoder")
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)
}
