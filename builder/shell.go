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
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var ErrRunBuildInDocker = errors.New("running the build script inside docker has failed")

type BuildType string

const (
	BuildTypeBuild BuildType = "build"
	BuildTypeFetch BuildType = "fetch"
)

func BuildShell(packageName, packagDir string, buildType BuildType) error {
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

	reader, err := dockerClient.ImagePull(ctx, "docker.io/library/debian:11", image.PullOptions{})
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}
	defer reader.Close()

	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}

	finalDataDir, err := createPackageBinDir(packageName)
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}

	createResponse, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:      "debian:11",
			WorkingDir: "/app",
			Cmd:        []string{"bash", "./scripts/" + string(buildType) + ".bash"},
			Tty:        false,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   packagDir,
					Target:   "/app/scripts/",
					ReadOnly: true,
				},
				{
					Type:   mount.TypeBind,
					Source: finalDataDir,
					Target: "/app/final/",
				},
			},
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
	}

	err = dockerClient.ContainerStart(ctx, createResponse.ID, container.StartOptions{})
	if err != nil {
		return errors.Join(ErrRunBuildInDocker, err)
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

	return nil
}
