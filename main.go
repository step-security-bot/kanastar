package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
	"github.com/surajsharma/kanastar/manager"
	"github.com/surajsharma/kanastar/task"
	"github.com/surajsharma/kanastar/worker"
)

func main() {

	mhost := os.Getenv("MANAGER_HOST")
	mport, _ := strconv.Atoi(os.Getenv("MANAGER_PORT"))

	whost := os.Getenv("WORKER_HOST")
	wport, _ := strconv.Atoi(os.Getenv("WORKER_PORT"))

	fmt.Println("⏳ Starting worker...")

	w1 := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	wapi1 := worker.Api{Address: whost, Port: wport, Worker: &w1}

	w2 := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	wapi2 := worker.Api{Address: whost, Port: wport + 1, Worker: &w2}

	w3 := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]*task.Task),
	}

	wapi3 := worker.Api{Address: whost, Port: wport + 2, Worker: &w3}

	go w1.RunTasks()
	go w1.CollectStats()
	go w1.UpdateTasks()
	go wapi1.Start()

	go w2.RunTasks()
	go w2.CollectStats()
	go w2.UpdateTasks()
	go wapi2.Start()

	go w3.RunTasks()
	go w3.CollectStats()
	go w3.UpdateTasks()
	go wapi3.Start()

	fmt.Println("⏳ Starting manager...")

	workers := []string{
		fmt.Sprintf("%s:%d", whost, wport),
		fmt.Sprintf("%s:%d", whost, wport+1),
		fmt.Sprintf("%s:%d", whost, wport+2),
	}

	m := manager.New(workers, "roundrobin")

	mapi := manager.Api{Address: mhost, Port: mport, Manager: m}

	go m.ProcessTasks()
	go m.UpdateTasks()
	go m.DoHealthChecks()

	mapi.Start() //this cannot be a goroutine for http.ListenAndServe is blocking
}
