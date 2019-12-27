package wpool

import (
	"fmt"
	"testing"
	"time"
)


type job struct {
	id string
}
func NewJob(idx string) *job {
	return &job{
		id: idx,
	}
}
func (j *job) Do() error {
	time.Sleep(time.Millisecond * 50)
	if j.id  == "job100" {
		panic("xxxxxxxx")
	}
	return nil
}

func (j *job) ID() string {
	return j.id
}


func TestNewWorkerPool(t *testing.T) {
	t1 := time.Now()
	dp := NewWorkerPool(200, 1000)

	go dp.Run()

	go func() {
		for {
			qpsNow := dp.GetQps()
			fmt.Println("CurrQps:", qpsNow)
			time.Sleep(time.Millisecond * 400)
		}
	}()
	idx := 1
	for {
		job1 := NewJob("job"+fmt.Sprintf("%d", idx))
		dp.PutJob(job1)
		if idx >= 5000 {
			break
		}
		idx ++
	}

	dp.WaitJobComplete()
	ts := time.Now().Unix() - t1.Unix()

	if ts > 5 {
		t.Errorf("work pool not effective")
	}
}

