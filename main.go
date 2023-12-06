package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/hugoleodev/pentagon/internal/docker"
	"github.com/hugoleodev/pentagon/manager"
	"github.com/hugoleodev/pentagon/node"
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
	t := task.Task{
		ID:     uuid.New(),
		Name:   "task#001",
		State:  task.Pending,
		Image:  "alpine",
		Memory: 1024,
		Disk:   1,
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		Timestamp: time.Now(),
		Task:      t,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("task event: %v\n", te)

	w := worker.Worker{
		Name:  "worker#001",
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	fmt.Printf("worker: %v\n", w)
	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StartTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string][]*task.Task),
		EventDb: make(map[string][]*task.TaskEvent),
		Workers: []string{w.Name},
	}

	fmt.Printf("manager: %v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SendWork()

	n := node.Node{
		Name:   "node#001",
		Ip:     "127.0.0.1",
		Cores:  3,
		Memory: 1024,
		Disk:   25,
		Role:   "worker",
	}

	fmt.Printf("node: %v\n", n)

	fmt.Printf("Creating a test container...\n")
	dockerTask, createResult := createContainer()

	if createResult.Error != nil {
		fmt.Printf("Error: %v\n", createResult.Error)
		os.Exit(1)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("Stopping the test container %s\n", createResult.ContainerId)

	_ = stopContainer(dockerTask, createResult.ContainerId)
}
