package docker

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/hugoleodev/pentagon/task"
	"github.com/moby/moby/pkg/namesgenerator"
)

type Docker struct {
	Client *client.Client
	Config Config
}

func New(c *Config) *Docker {
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	return &Docker{
		Client: dc,
		Config: *c,
	}
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStderr  bool
	ExposedPorts  nat.PortSet
	Cmd           []string
	Image         string
	Cpu           float64
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

func NewConfig(t *task.Task) *Config {

	if t.Name == "" {
		t.Name = namesgenerator.GetRandomName(1)
	}

	return &Config{
		Name:          t.Name,
		ExposedPorts:  t.ExposedPorts,
		Image:         t.Image,
		Cpu:           t.Cpu,
		Memory:        t.Memory,
		Disk:          t.Disk,
		RestartPolicy: t.RestartPolicy,
	}
}

const (
	DockerResultSuccess = "success"
	DockerResultFailure = "failure"
)

type DockerResult struct {
	ContainerId string
	Action      string
	Error       error
	Result      string
}

func (d *Docker) Run(ctx context.Context) DockerResult {

	reader, err := d.Client.ImagePull(ctx, d.Config.Image, types.ImagePullOptions{})

	if err != nil {
		log.Info().Msgf("Error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	io.Copy(os.Stdout, reader)

	rp := container.RestartPolicy{
		Name: d.Config.RestartPolicy,
	}

	r := container.Resources{
		Memory:   d.Config.Memory,
		NanoCPUs: int64(d.Config.Cpu * 1e9),
	}

	cc := container.Config{
		Env:          d.Config.Env,
		ExposedPorts: d.Config.ExposedPorts,
		Image:        d.Config.Image,
		Tty:          false,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)

	if err != nil {
		log.Info().Msgf("Error creating container %s using image %s: %v\n", d.Config.Name, d.Config.Image, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})

	if err != nil {
		log.Info().Msgf("Error starting container %s with ID %s: %v\n", d.Config.Name, resp.ID, err)
		return DockerResult{Error: err}
	}

	out, err := d.Client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	if err != nil {
		log.Info().Msgf("Error getting logs for container %s with ID %s: %v\n", d.Config.Name, resp.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{
		ContainerId: resp.ID,
		Action:      "start",
		Error:       nil,
		Result:      DockerResultSuccess,
	}
}

func (d *Docker) Stop(ctx context.Context, id string) DockerResult {
	log.Info().Msgf("attempting to stop container %v", id)

	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})

	if err != nil {
		log.Info().Msgf("Error stopping container %s with ID %s: %v\n", d.Config.Name, id, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         false,
		RemoveLinks:   false,
	})

	if err != nil {
		log.Info().Msgf("Error removing container %s with ID %s: %v\n", d.Config.Name, id, err)
		return DockerResult{Error: err}
	}

	return DockerResult{
		ContainerId: id,
		Action:      "stop",
		Error:       nil,
		Result:      DockerResultSuccess,
	}
}
