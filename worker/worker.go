package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

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
	Stats     *Stats
}

func New(name string) *Worker {
	return &Worker{
		Name:  name,
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}
}

func (w *Worker) AddTask(t task.Task) {
	w.Queue.Enqueue(t)
}

func (w *Worker) CollectStats() {
	for {
		log.Info().Msg("Collecting stats")
		w.Stats = GetStats()
		w.Stats.TaskCount = w.TaskCount
		time.Sleep(15 * time.Second)
	}
}

func (w *Worker) GetTasks() []*task.Task {
	tasks := []*task.Task{}

	for _, t := range w.Db {
		tasks = append(tasks, t)
	}

	return tasks
}

// RunTask runs a task from the worker's queue and
// returns the result of the task execution.
func (w *Worker) runTask() docker.DockerResult {
	t := w.Queue.Dequeue()
	if t == nil {
		log.Info().Msg("No task to run")
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
			result = w.StartTask(taskPersisted)
		case task.Completed:
			result = w.StopTask(taskPersisted)
		default:
			result.Error = fmt.Errorf("WE SHOULD NOT BE HERE")
		}
	} else {
		err := fmt.Errorf("INVALID STATE TRANSITION FROM %v TO %v", taskPersisted.State, taskQueued.State)
		result.Error = err
	}

	return result
}

func (w *Worker) StartTask(t *task.Task) docker.DockerResult {
	ctx := context.Background()
	t.StartTime = time.Now().UTC()
	config := docker.NewConfig(t)
	d := docker.New(config)
	result := d.Run(ctx)

	if result.Error != nil {
		log.Info().Msgf("Error running task %s: %v\n", t.ID, result.Error)
		t.State = task.Failed
		w.Db[t.ID] = t
		return result
	}

	t.ContainerID = result.ContainerId
	t.State = task.Running
	w.Db[t.ID] = t

	log.Info().Msgf("Started container %s with ID %v for task %v", config.Name, t.ContainerID, t.ID)

	return result

}

func (w *Worker) RunTasks() {
	for {
		if w.Queue.Len() != 0 {
			result := w.runTask()
			if result.Error != nil {
				log.Info().Msgf("Error running task: %v\n", result.Error)
			}
		} else {
			log.Info().Msg("No tasks to process currently.\n")
		}
		log.Info().Msg("Sleeping for 10 seconds.")
		time.Sleep(10 * time.Second)
	}

}

func (w *Worker) StopTask(t *task.Task) docker.DockerResult {
	ctx := context.Background()

	config := docker.NewConfig(t)
	d := docker.New(config)

	result := d.Stop(ctx, t.ContainerID)

	if result.Error != nil {
		log.Info().Msgf("Error stopping container %s with ID %s: %v\n", config.Name, t.ContainerID, result.Error)
	}

	t.FinishTime = time.Now().UTC()
	t.State = task.Completed
	w.Db[t.ID] = t
	log.Info().Msgf("Stopped and removed container %s with ID %v for task %v", config.Name, t.ContainerID, t.ID)

	return result

}
