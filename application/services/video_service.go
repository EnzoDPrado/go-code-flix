package services

import (
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
)

type VideoService struct {
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) Download(video *domain.Video, bucketName string) error {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)

	if err != nil {
		return err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(video.FilePath)

	r, err := obj.NewReader(ctx)

	if err != nil {
		return err
	}

	defer r.Close()

	body, err := ioutil.ReadAll(r)

	if err != nil {
		return err
	}

	f, err := os.Create(os.Getenv("localStoragePath") + "/" + video.ID + ".mp4")

	if err != nil {
		return err
	}

	_, err = f.Write(body)

	defer f.Close()

	log.Printf("video %v has been stored", video.ID)

	return nil
}

func (v *VideoService) Fragment(video *domain.Video) error {
	err := os.Mkdir(os.Getenv("localStoragePath")+"/"+video.ID, os.ModePerm)

	if err != nil {
		return err
	}

	source := os.Getenv("localStoragePath") + "/" + video.ID + ".mp4"
	target := os.Getenv("localStoragePath") + "/" + video.ID + ".frag"

	cmd := exec.Command("mp4fragment", source, target)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Encode(video *domain.Video) error {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, os.Getenv("localStoragePath")+"/"+video.ID+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, os.Getenv("localStoragePath")+"/"+video.ID)
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

func (v *VideoService) Finish(video *domain.Video) error {
	err := os.Remove(os.Getenv("localStoragePath") + "/" + video.ID + ".mp4")

	if err != nil {
		log.Println("error removing mp4", video.ID, ".mp4")
		return err
	}

	err = os.Remove(os.Getenv("localStoragePath") + "/" + video.ID + ".frag")

	if err != nil {
		log.Println("error removing frag", video.ID, ".frag")
		return err
	}

	err = os.RemoveAll(os.Getenv("localStoragePath") + "/" + video.ID)

	if err != nil {
		log.Println("error removing video file", video.ID)
		return err
	}

	return nil
}

func (v *VideoService) InsertVideo(video *domain.Video) error {
	_, err := v.VideoRepository.Insert(video)

	if err != nil {
		return err
	}

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("==========> Output: %s\n", string(out))
	}
}
