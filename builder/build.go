package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var ErrRunBuildInDocker = errors.New("running the build script inside docker has failed")

func Build(script string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}
	defer dockerClient.Close()

	reader, err := dockerClient.ImagePull(ctx, "docker.io/library/debian", image.PullOptions{})
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}
	defer reader.Close()

	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}

	createResponse, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:      "debian",
			WorkingDir: "/app",
			Cmd:        []string{"bash", "-c", "'" + script + "'"},
			Tty:        false,
		},
		nil,
		nil,
		nil,
		"",
	)
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}

	if err := dockerClient.ContainerStart(ctx, createResponse.ID, container.StartOptions{}); err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}

	statusCh, errCh := dockerClient.ContainerWait(
		ctx,
		createResponse.ID,
		container.WaitConditionNotRunning,
	)
	select {
	case err := <-errCh:
		if err != nil {
			return errors.Join(
				ErrRunBuildInDocker,
				err,
			)
		}
	case status := <-statusCh:
		if status.Error != nil {
			return errors.Join(
				ErrRunBuildInDocker,
				errors.New(status.Error.Message),
			)
		}
		if status.StatusCode != 0 {
			return errors.Join(
				ErrRunBuildInDocker,
				fmt.Errorf("bad status code: %d", status.StatusCode),
			)
		}
	}

	out, err := dockerClient.ContainerLogs(
		ctx,
		createResponse.ID,
		container.LogsOptions{ShowStdout: true},
	)
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}
	defer out.Close()

	if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, out); err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}

	return nil
}
