package main

import (
	"fmt"
	"log"
	"os"
	"sync"
)

/*
这里定义了一个工作池
*/

//定义工作
type Job struct {
	Fn  func(Job) Result //执行函数，返回Result
	Url string
}

//定义结果
type Result struct {
	Url    string
	Length int64
	Status int
	Body   string
	Err    error
}

//结果输出
func (r Result) Print() string {
	return fmt.Sprintf("%3d  -  %7d   -   %s", r.Status, r.Length, r.Url)
}

//结果保存
func (r Result) Save(filename string) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		log.Fatalf("result save err: %v", err)
	}
	defer file.Close()
	file.WriteString("")
}

//定义工作池
type WorkPool struct {
	Count   int
	Jobs    chan Job
	Results chan Result
}

//产生新的工作池
func NewPool(c int, j chan Job, r chan Result) WorkPool {
	return WorkPool{
		Count:   c,
		Jobs:    j,
		Results: r,
	}
}

//运行工作池
func (wp *WorkPool) Run(wg *sync.WaitGroup) {
	for i := 1; i < wp.Count; i++ {
		wg.Add(1)
		go Worker(wg, wp.Jobs, wp.Results)
	}
	wg.Wait()

	close(wp.Results)

}

//定义工人
func Worker(wg *sync.WaitGroup, Jobs chan Job, Results chan Result) {
	for {
		job, ok := <-Jobs
		if !ok {
			wg.Done()
			return
		}
		Results <- job.Fn(job)
	}
}
