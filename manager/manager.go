package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/hugoleodev/pentagon/task"
	"github.com/hugoleodev/pentagon/worker"
)

type Manager struct {
	Pending       queue.Queue
	TaskDb        map[string]*task.Task
	EventDb       map[string]*task.TaskEvent
	Workers       []string
	WorkerTaskMap map[string][]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
	LastWorker    int
}

func New(workers []string) *Manager {
	taskDb := make(map[string]*task.Task)
	eventDb := make(map[string]*task.TaskEvent)
	workerTaskMap := make(map[string][]uuid.UUID)
	taskWorkerMap := make(map[uuid.UUID]string)

	for worker := range workers {
		workerTaskMap[workers[worker]] = []uuid.UUID{}
	}

	return &Manager{
		Pending:       *queue.New(),
		TaskDb:        taskDb,
		EventDb:       eventDb,
		Workers:       workers,
		WorkerTaskMap: workerTaskMap,
		TaskWorkerMap: taskWorkerMap,
	}
}

func (m *Manager) SelectWorker() string {
	var newWorker int

	if m.LastWorker+1 < len(m.Workers) {
		newWorker = m.LastWorker + 1
		m.LastWorker++
	} else {
		newWorker = 0
		m.LastWorker = 0
	}

	return m.Workers[newWorker]
}

func (m *Manager) updateTasks() {
	for _, w := range m.Workers {
		log.Info().Msgf("Worker: %v", w)
		log.Info().Msgf("Checking worker %v for task updates\n", w)
		url := fmt.Sprintf("http://%s/api/tasks", w)
		resp, err := http.Get(url)

		if err != nil {
			log.Info().Msgf("Unable to get tasks from worker %v: %v\n", w, err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Info().Msgf("Error sending request: %v\n", err)
		}

		d := json.NewDecoder(resp.Body)
		var tasks []*task.Task
		err = d.Decode(&tasks)
		if err != nil {
			log.Info().Msgf("Unable unmarshaling tasks: %s\n", err.Error())
		}

		for _, t := range tasks {
			log.Info().Msgf("Attempting to update task %s\n", t.ID)

			_, ok := m.TaskDb[t.ID.String()]
			if !ok {
				log.Info().Msgf("Task %s not found\n", t.ID)
			}

			if m.TaskDb[t.ID.String()].State != t.State {
				m.TaskDb[t.ID.String()].State = t.State
				log.Info().Msgf("Task %s state updated\n", t.ID)
			}

			m.TaskDb[t.ID.String()].StartTime = t.StartTime
			m.TaskDb[t.ID.String()].FinishTime = t.FinishTime
			m.TaskDb[t.ID.String()].ContainerID = t.ContainerID
		}
	}
}

func (m *Manager) SendWork() {
	if m.Pending.Len() > 0 {
		w := m.SelectWorker()

		e := m.Pending.Dequeue()
		te := e.(task.TaskEvent)
		t := te.Task
		log.Info().Msgf("Pulled %v of pending queue\n", t)

		m.EventDb[te.ID.String()] = &te
		m.WorkerTaskMap[w] = append(m.WorkerTaskMap[w], te.Task.ID)

		m.TaskWorkerMap[t.ID] = w

		t.State = task.Scheduled
		m.TaskDb[t.ID.String()] = &t

		data, err := json.Marshal(te)
		if err != nil {
			log.Info().Msgf("Unable to marshal task event: %v\n", err)
		}

		url := fmt.Sprintf("http://%s/api/tasks", w)
		log.Info().Msgf("sending request to %v", url)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))

		if err != nil {
			log.Info().Msgf("Error connecting to %v: %v\n", w, err)
			m.Pending.Enqueue(te)
			return
		}

		d := json.NewDecoder(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			e := worker.ErrResponse{}
			err := d.Decode(&e)
			if err != nil {
				log.Info().Msgf("Error decoding response: %s\n", err.Error())
				return
			}
			log.Info().Msgf("Response error (%d): %s", e.HTTPStatusCode, e.Message)
			return
		}

		t = task.Task{}
		err = d.Decode(&t)
		if err != nil {
			log.Info().Msgf("Error decoding response: %s\n", err.Error())
			return
		}
		log.Info().Msgf("%#v\n", t)
	} else {
		log.Info().Msgf("Pending queue is empty\n")
	}
}

func (m *Manager) AddTask(te task.TaskEvent) {
	m.Pending.Enqueue(te)
}

func (m *Manager) GetTasks() []*task.Task {
	tasks := []*task.Task{}
	for _, t := range m.TaskDb {
		tasks = append(tasks, t)
	}
	return tasks
}

func (m *Manager) ProcessTasks() {
	for {
		log.Info().Msg("Processing any tasks in the queue")
		m.SendWork()
		log.Info().Msg("Sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
	}
}

func (m *Manager) UpdateTasks() {
	for {
		log.Info().Msg("Checking for task updates from workers")
		m.updateTasks()
		log.Info().Msg("Task updates completed")
		log.Info().Msg("Sleeping for 15 seconds")
		time.Sleep(15 * time.Second)
	}
}
