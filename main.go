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
	"github.com/hugoleodev/pentagon/manager"
	"github.com/hugoleodev/pentagon/task"
	"github.com/hugoleodev/pentagon/worker"
	"github.com/hugoleodev/pentagon/worker/api"
	"github.com/shirou/gopsutil/v3/mem"
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

	workers := []string{fmt.Sprintf("%s:%d", host, port)}
	m := manager.New(workers)

	for i := 0; i < 5; i++ {
		t := task.Task{
			ID:    uuid.New(),
			Name:  fmt.Sprintf("task-container-%d", i),
			State: task.Scheduled,
			Image: "postgres",
		}
		te := task.TaskEvent{
			ID:    uuid.New(),
			Task:  t,
			State: task.Running,
		}
		m.AddTask(te)
		m.SendWork()
	}

	go func() {
		time.Sleep(20 * time.Second)
		for {
			fmt.Printf("[Manager] Updating tasks from %d workers\n", len(m.Workers))
			m.UpdateTasks()
			time.Sleep(15 * time.Second)
		}
	}()

	go func() {
		time.Sleep(20 * time.Second)
		for {
			for _, t := range m.TaskDb {
				fmt.Printf("[Manager] Task: id: %s, state: %d\n", t.ID, t.State)
				time.Sleep(15 * time.Second)
			}
		}
	}()

	// GetStatsFromGoPsUtil()

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

func GetStatsFromGoPsUtil() {
	v, err := mem.VirtualMemory()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Total: %v, Free:%v, Available:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.Available, v.UsedPercent)

	fmt.Println(v)
}
