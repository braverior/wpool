package wpool


import (
	"sync"
	"time"
)


var MaxJobPoolSize    = 100 //任务队列缓冲器


// define job channel
type JobChan chan Job
type WorkerChan chan *Worker


type WorkerPool struct {
	Workers       []*Worker
	WorkerChan    chan *Worker
	quit          chan bool
	jobQueue      JobChan
	workerNum     int
	freeNumber    int
	qps           int
	free          chan bool
	mu            sync.Mutex
	currQps       int
	statMap       map[int64]int
}

func NewWorkerPool(workerNum, qps int) *WorkerPool {
	WorkerPool := &WorkerPool{
		workerNum:  workerNum,
		WorkerChan: make(WorkerChan, workerNum),
		jobQueue:   make(JobChan, MaxJobPoolSize),
		quit:       make(chan bool),
		free:       make(chan bool, 1),
		qps:        qps,
	}

	for idx := 0; idx < workerNum; idx ++ {
		worker := NewWorker(int(idx), WorkerPool)
		WorkerPool.Workers = append(WorkerPool.Workers, worker)
		WorkerPool.WorkerChan <- worker
		go worker.Start()
	}
	return WorkerPool
}

func (this *WorkerPool) WaitJobComplete() {
	for {
		if len(this.WorkerChan) == this.workerNum {
			this.free <- true
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (this *WorkerPool) GetQps() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	ts := time.Now().Unix() - 1
	currQps, ok := this.statMap[ts];
	if ok {
		return currQps
	}
	return 0
}

func (this *WorkerPool) Run() error {
	this.statMap = make(map[int64]int)
	for {
		ts := time.Now().Unix()
		this.mu.Lock()
		if len(this.statMap) > 10000 {
			this.statMap = make(map[int64]int)
		}
		currQps, ok := this.statMap[ts];
		if !ok {
			this.statMap[ts] = 0
		}
		this.statMap[ts] ++
		this.mu.Unlock()
		currQps++
		//如果QPS超限，则需要等一会儿
		if currQps > this.qps {
			time.Sleep(time.Millisecond * 100)
		}
		select {
		case job := <- this.jobQueue:
			//有任务了，看谁闲着，把活派给他
			worker := <- this.WorkerChan
			worker.job <- job
		case <-this.quit:
			// stop WorkerPool
			return nil
		}
	}
}


func (this *WorkerPool) PutJob(job Job) error {
	this.jobQueue <- job
	return nil
}



func (this *WorkerPool) Stop() {
	for _, worker := range this.Workers {
		worker.Stop()
	}
	this.quit <- true
}
