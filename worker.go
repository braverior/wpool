package wpool

import "log"

type Worker struct {
	no int
	job chan Job
	quit chan bool
	dp *WorkerPool
}

func NewWorker(workerNo int, dp *WorkerPool) *Worker {
	return &Worker{
		no: workerNo,
		job: make(chan Job, 1),
		dp: dp,
	}
}


func (this *Worker) Start() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[WORKER %d] catched error %v", this.no)
			go this.Start()
			this.dp.WorkerChan <- this
		}
	}()
	for {
		select {
		case job := <- this.job:
			if err := job.Do(); err != nil {

			}
			//活干完了，再去排队领任务吧
			this.dp.WorkerChan <- this
		case <- this.quit:
			return
		}
	}
}

func (this *Worker) Stop() {
	this.quit <- true
}


