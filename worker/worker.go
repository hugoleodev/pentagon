package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/hugoleodev/pentagon/internal/docker"
	"github.com/hugoleodev/pentagon/task"
)

type Worker struct {
	Name      string
	Queue     queue.Queue
	Db        map[uuid.UUID]*task.Task
	TaskCount int
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) CollectStats() {
	fmt.Println("I will collect the stats")
}

// RunTask runs a task from the worker's queue and
// returns the result of the task execution.
func (w *Worker) RunTask() docker.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Println("No task to run")
		return docker.DockerResult{Error: nil}
	}

	taskQueued := t.(task.Task)

	taskPersisted := w.Db[taskQueued.ID]

	if taskPersisted == nil {
		taskPersisted = &taskQueued
		w.Db[taskQueued.ID] = &taskQueued
	}

	var result docker.DockerResult
	if task.ValidStateTransition(taskPersisted.State, taskQueued.State) {
		switch taskQueued.State {
		case task.Scheduled:
			result = w.StartTask(*taskPersisted)
		case task.Completed:
			result = w.StopTask(*taskPersisted)
		default:
			result.Error = fmt.Errorf("We should not be here")
		}
	} else {
		err := fmt.Errorf("Invalid state transition from %v to %v", taskPersisted.State, taskQueued.State)
		result.Error = err
	}

	return result
}

func (w *Worker) StartTask(t task.Task) docker.DockerResult {
	ctx := context.Background()
	t.StartTime = time.Now().UTC()
	config := docker.NewConfig(&t)
	d := docker.New(config)
	result := d.Run(ctx)

	if result.Error != nil {
		log.Printf("Error running task %s: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = &t
		return result
	}

	t.ContainerID = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = &t

	log.Printf("Started container %s with ID %v for task %v", config.Name, t.ContainerID, t.ID)

	return result

}

func (w *Worker) StopTask(t task.Task) docker.DockerResult {
	ctx := context.Background()

	config := docker.NewConfig(&t)
	d := docker.New(config)

	result := d.Stop(ctx, t.ContainerID)

	if result.Error != nil {
		log.Printf("Error stopping container %s with ID %s: %v\n", config.Name, t.ContainerID, result.Error)
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = &t
	log.Printf("Stopped and removed container %s with ID %v for task %v", config.Name, t.ContainerID, t.ID)

	return result

}
