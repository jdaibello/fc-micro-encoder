package services

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
	"github.com/jdaibello/fc-micro-encoder/application/repositories"
	"github.com/jdaibello/fc-micro-encoder/domain"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (vs *VideoService) Download(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	if err != nil {
		return err
	}

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(vs.Video.FilePath)
	r, err := obj.NewReader(ctx)

	if err != nil {
		return err
	}

	defer r.Close()

	body, err := ioutil.ReadAll(r)

	if err != nil {
		return err
	}

	f, err := os.Create(os.Getenv("LOCAL_STORAGE_PATH") + "/" + vs.Video.ID + ".mp4")

	if err != nil {
		return err
	}

	_, err = f.Write(body)

	if err != nil {
		return err
	}

	defer f.Close()

	log.Printf("Video %v has been downloaded", vs.Video.ID)

	return nil
}

func (vs *VideoService) Fragment() error {
	err := os.Mkdir(os.Getenv("LOCAL_STORAGE_PATH")+"/"+vs.Video.ID, os.ModePerm)

	if err != nil {
		return err
	}

	source := os.Getenv("LOCAL_STORAGE_PATH") + "/" + vs.Video.ID + ".mp4"
	target := os.Getenv("LOCAL_STORAGE_PATH") + "/" + vs.Video.ID + ".frag"

	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Encode() error {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, os.Getenv("LOCAL_STORAGE_PATH")+"/"+v.Video.ID+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, os.Getenv("LOCAL_STORAGE_PATH")+"/"+v.Video.ID)
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, "/opt/bento4/bin/")
	cmd := exec.Command("mp4dash", cmdArgs...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Finish() error {
	err := os.Remove(os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID + ".mp4")

	if err != nil {
		log.Println("error removing mp4 ", v.Video.ID, ".mp4")
		return err
	}

	err = os.Remove(os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID + ".frag")

	if err != nil {
		log.Println("error removing frag ", v.Video.ID, ".frag")
		return err
	}

	err = os.RemoveAll(os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID)

	if err != nil {
		log.Println("error removing folder ", v.Video.ID)
		return err
	}

	log.Println("Files have been removed: ", v.Video.ID)

	return nil
}

func (v *VideoService) InsertVideo() error {
	_, err := v.VideoRepository.Insert(v.Video)

	if err != nil {
		return err
	}

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("====> Output: %s\n", string(out))
	}
}
