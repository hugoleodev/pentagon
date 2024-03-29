package api

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hugoleodev/pentagon/task"
	"github.com/hugoleodev/pentagon/worker"
)

type API struct {
	Address string
	Port    int
	Worker  *worker.Worker
	Router  fiber.Router
}

func (a *API) initRouter(app *fiber.App) {
	a.Router = app.Group("/api")
	a.Router.Get("/tasks", a.GetTasksHandler)
	a.Router.Post("/tasks", a.StartTaskHandler)
	a.Router.Delete("/tasks/:taskId", a.StopTaskHandler)

	a.Router.Get("/stats", a.GetStatsHandler)
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	result := a.Worker.StartTask(&te.Task)
	log.Info().Msgf("Adding task %s: %v\n", te.Task.ID, result.Error)

	return ctx.Status(fiber.StatusCreated).JSON(te.Task)

}

func (a *API) GetTasksHandler(ctx *fiber.Ctx) error {

	return ctx.Status(fiber.StatusOK).JSON(a.Worker.GetTasks())
}

func (a *API) StopTaskHandler(ctx *fiber.Ctx) error {
	taskID := ctx.Params("taskId")

	if taskID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "task id is required",
		})
	}

	tID, err := uuid.Parse(taskID)
	taskToStop, ok := a.Worker.Db[tID]
	if !ok || err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "task not found",
		})
	}

	tTSCopy := *taskToStop
	tTSCopy.State = task.Completed
	a.Worker.AddTask(tTSCopy)

	log.Info().Msgf("Added task %v to stop container %v\n", taskToStop.ID, taskToStop.ContainerID)

	responseMessage := fiber.Map{
		"message": fmt.Sprintf("added task %v to stop container %v", taskToStop.ID, taskToStop.ContainerID),
	}

	return ctx.Status(fiber.StatusNoContent).JSON(responseMessage)
}

func (a *API) GetStatsHandler(ctx *fiber.Ctx) error {
	log.Info().Msg("Getting stats")
	return ctx.Status(fiber.StatusOK).JSON(a.Worker.Stats)
}
