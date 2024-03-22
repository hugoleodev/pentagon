package main

import (
	"fmt"

	"github.com/hugoleodev/pentagon/manager"
	mapi "github.com/hugoleodev/pentagon/manager/api"
	"github.com/hugoleodev/pentagon/worker"
	wapi "github.com/hugoleodev/pentagon/worker/api"
	"github.com/rs/zerolog/log"
)

func main() {
	mhost := "localhost"
	mport := 8888
	whost := "localhost"
	wport := 7777

	log.Info().Msg("Starting Pentagon worker")

	w1 := worker.New("worker-1")
	wapi1 := wapi.API{Address: whost, Port: wport, Worker: w1}

	w2 := worker.New("worker-2")
	wapi2 := wapi.API{Address: whost, Port: wport + 1, Worker: w2}

	w3 := worker.New("worker-3")
	wapi3 := wapi.API{Address: whost, Port: wport + 2, Worker: w3}

	go w1.RunTasks()
	go wapi1.Start()

	go w2.RunTasks()
	go wapi2.Start()

	go w3.RunTasks()
	go wapi3.Start()

	log.Info().Msg("Starting Pentagon manager...")

	workers := []string{
		fmt.Sprintf("%s:%d", whost, wport),
		fmt.Sprintf("%s:%d", whost, wport+1),
		fmt.Sprintf("%s:%d", whost, wport+2),
	}

	m := manager.New(workers)
	managerApi := mapi.API{Address: mhost, Port: mport, Manager: m}

	go m.ProcessTasks()
	go m.UpdateTasks()

	managerApi.Start()
}
