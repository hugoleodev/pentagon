package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/hugoleodev/pentagon/internal/docker"
	"github.com/hugoleodev/pentagon/task"
	"github.com/hugoleodev/pentagon/worker"
	"github.com/hugoleodev/pentagon/worker/api"
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
	host := os.Getenv("PENTAGON_HOST")
	port, _ := strconv.Atoi(os.Getenv("PENTAGON_PORT"))

	fmt.Println("Starting Pentagon Worker")

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	api := api.API{
		Address: host,
		Port:    port,
		Worker:  &w,
	}

	go runTasks(&w)
	go w.CollectStats()
	api.Start(host, port, &w)
}

func runTasks(w *worker.Worker) {
	for {
		if w.Queue.Len() != 0 {
			result := w.RunTask()

			if result.Error != nil {
				fmt.Printf("Error running task: %v\n", result.Error)
			}
		} else {
			fmt.Println("No tasks to run yet")
		}
		fmt.Println("Sleeping for 10 seconds")

		time.Sleep(10 * time.Second)
	}
}
