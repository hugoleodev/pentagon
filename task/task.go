package task

import (
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type Task struct {
	ID            uuid.UUID         `json:"id"`
	ContainerID   string            `json:"container_id"`
	Name          string            `json:"name"`
	State         State             `json:"state"`
	Image         string            `json:"image"`
	Cpu           float64           `json:"cpu"`
	Memory        int64             `json:"memory"`
	Disk          int64             `json:"disk"`
	ExposedPorts  nat.PortSet       `json:"exposed_ports"`
	PortBindings  map[string]string `json:"port_bindings"`
	RestartPolicy string            `json:"restart_policy"`
	StartTime     time.Time         `json:"start_time"`
	FinishTime    time.Time         `json:"finish_time"`
}

type TaskEvent struct {
	ID        uuid.UUID `json:"id"`
	State     State     `json:"state"`
	Timestamp time.Time `json:"timestamp"`
	Task      Task      `json:"task"`
}
