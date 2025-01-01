package builder

import (
	"context"
	"errors"
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
		return err
	}
	defer dockerClient.Close()

	reader, err := dockerClient.ImagePull(ctx, "docker.io/library/debian", image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		return err
	}

	resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
		Image: "debian",
		Cmd: []string{
			"cat", ">/package-getter", "<<END",
			script,
			"END",
			"&&", "source", "/package-getter",
		},
		Tty: false,
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	if err := dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := dockerClient.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
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
			return ErrRunBuildInDocker
		}
	}

	out, err := dockerClient.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		return err
	}

	return nil
}
