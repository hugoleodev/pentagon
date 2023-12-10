package main

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/hugoleodev/pentagon/internal/docker"
	"github.com/hugoleodev/pentagon/task"
	"github.com/hugoleodev/pentagon/worker"
)

func createContainer() (*docker.Docker, *docker.DockerResult) {
	c := docker.Config{
		Name:  "test-container-1",
		Image: "postgres",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
		},
	}

	dc, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	d := docker.Docker{
		Client: dc,
		Config: c,
	}

	result := d.Run(context.Background())

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
		return nil, nil
	}

	fmt.Printf("Container ID: %s is running with config %v\n", result.ContainerId, c)

	return &d, &result
}

func stopContainer(d *docker.Docker, id string) *docker.DockerResult {
	result := d.Stop(context.Background(), id)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
		return nil
	}

	fmt.Printf("Container ID: %s is stopped\n", result.ContainerId)

	return &result
}

func main() {
	db := make(map[uuid.UUID]*task.Task)

	w := worker.Worker{
		Name:  "worker#001",
		Queue: *queue.New(),
		Db:    db,
	}

	t := task.Task{
		ID:    uuid.New(),
		Name:  "task-container-1",
		State: task.Scheduled,
		Image: "lmmendes/http-hello-world",
	}

	fmt.Println("starting task")
	w.AddTask(t)

	result := w.RunTask()

	if result.Error != nil {
		panic(result.Error)
	}

	t.ContainerID = result.ContainerId

	fmt.Printf("Task %s with id %s is running with container ID %s\n", t.Name, t.ID, t.ContainerID)

	fmt.Println("Sleepy time")

	time.Sleep(30 * time.Second)

	fmt.Printf("stopping task %s with id %s in container %s\n", t.Name, t.ID, t.ContainerID)
	t.State = task.Completed
	w.AddTask(t)
	result = w.RunTask()
	if result.Error != nil {
		panic(result.Error)
	}
}
