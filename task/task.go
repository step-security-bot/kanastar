package task

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type Task struct {
	ID            uuid.UUID
	ContainerID   string
	Name          string
	State         State
	Image         string
	Memory        int64
	Disk          int64
	Cpu           float64
	ExposedPorts  nat.PortSet
	HostPorts     nat.PortMap
	PortBindings  map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
	HealthCheck   string
	RestartCount  int
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
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

type Docker struct {
	Client      *client.Client
	Config      Config
	ContainerId string
}

type DockerInspectResponse struct {
	Error     error
	Container *types.ContainerJSON
}

func NewConfig(t *Task) *Config {
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

type DockerResult struct {
	Error       error
	Action      string
	ContainerID string
	Result      string
}

func NewDocker(c *Config) *Docker {
	dc, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		log.Fatalf("[task] error creating docker client, is the daemon running?")
		return nil
	}

	return &Docker{
		Client: dc,
		Config: *c,
	}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("[task] attempting to stop container %v", id)

	ctx := context.Background()

	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})

	if err != nil {
		log.Printf("[task] error attempting to stop container %v", id)
		return DockerResult{Error: err}
	}

	remove_err := d.Client.ContainerRemove(ctx, id, container.RemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		/* The RemoveLinks option is typically used when you want to remove links between containers,
		but in most modern Docker applications, links are deprecated in favor of networks.
		If you're not specifically using container links, you probably don't need this option.
		WARNING: setting this to true will cause remove_err to result in
		"Conflict, cannot remove the default link name of the container" */
		Force: false,
	})

	if remove_err != nil {
		log.Printf("[task] error attempting to remove container %v", id)
		return DockerResult{Error: err}
	}

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()

	reader, err := d.Client.ImagePull(
		ctx, d.Config.Image, image.PullOptions{})

	if err != nil {
		log.Printf("[task] error pulling image %s: %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}

	io.Copy(os.Stdout, reader)

	rp := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	r := container.Resources{
		Memory: d.Config.Memory,
	}

	cc := container.Config{
		Image: d.Config.Image,
		Env:   d.Config.Env,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	resp, create_err := d.Client.ContainerCreate(
		ctx, &cc, &hc, nil, nil, d.Config.Name)

	if create_err != nil {
		log.Printf("[task] error creating container using image %s: %v\n", d.Config.Image, create_err)
		return DockerResult{Error: err}
	}

	start_err := d.Client.ContainerStart(
		ctx, resp.ID, container.StartOptions{})

	if start_err != nil {
		log.Printf("[task] error starting container %s: %v\n", resp.ID, start_err)
		return DockerResult{Error: err}
	}

	out, err := d.Client.ContainerLogs(
		ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})

	if err != nil {
		log.Printf("[task] error getting logs for container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{ContainerID: resp.ID, Action: "start", Result: "success"}
}

func (d *Docker) Inspect(containerID string) DockerInspectResponse {
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	ctx := context.Background()
	resp, err := dc.ContainerInspect(ctx, containerID)

	if err != nil {
		log.Printf("[task] error inspecting container: %s\n", err.Error())
		return DockerInspectResponse{Error: err}
	}

	return DockerInspectResponse{Container: &resp}
}
