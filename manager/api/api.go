package api

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hugoleodev/pentagon/manager"
	"github.com/hugoleodev/pentagon/task"
)

type API struct {
	Address string
	Port    int
	Manager *manager.Manager
	Router  fiber.Router
}

func (a *API) initRouter(app *fiber.App) {
	a.Router = app.Group("/api/tasks")
	a.Router.Post("/", a.StartTaskHandler)
	a.Router.Get("/", a.GetTasksHandler)
	a.Router.Delete("/:taskId", a.StopTaskHandler)
}

func (a *API) Start() {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	a.initRouter(app)

	log.Fatal().Err(app.Listen(fmt.Sprintf("%s:%d", a.Address, a.Port)))
}

func (a *API) StartTaskHandler(ctx *fiber.Ctx) error {
	te := task.TaskEvent{}
	err := ctx.BodyParser(&te)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	a.Manager.AddTask(te)
	log.Info().Msgf("Added task %s\n", te.Task.ID)

	return ctx.Status(fiber.StatusCreated).JSON(te.Task)
}

func (a *API) GetTasksHandler(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(a.Manager.GetTasks())
}

func (a *API) StopTaskHandler(ctx *fiber.Ctx) error {
	taskID := ctx.Params("taskId")

	if taskID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "task id is required",
		})
	}

	tID, err := uuid.Parse(taskID)
	taskToStop, ok := a.Manager.TaskDb[tID.String()]
	if !ok || err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "task not found",
		})
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Completed,
		Timestamp: time.Now(),
	}

	tTSCopy := *taskToStop
	tTSCopy.State = task.Completed
	te.Task = tTSCopy
	a.Manager.AddTask(te)

	log.Info().Msgf("Added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerID)

	responseMessage := fiber.Map{
		"message": fmt.Sprintf("added task %v to stop container %v", taskToStop.ID, taskToStop.ContainerID),
	}

	return ctx.Status(fiber.StatusNoContent).JSON(responseMessage)
}
