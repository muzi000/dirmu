package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

//随机发包，保存的数据，用于结果判断
type Meaning struct {
	Length int64
	Status int
	Body   string
}

var (
	config Config
	wg     sync.WaitGroup
)

//发送http，这是工作函数
func HttpSent(job Job) Result {
	url := config.BaseUrl + job.Url
	for i := 0; i < config.Retry; i++ {

		req, err := http.NewRequest(config.Method, url, nil)
		if err != nil {
			log.Fatalf("request err: %v", err)
		}
		req.Header.Set("User-Agent", RandomAgent())
		req.Header.Set("Referer", config.BaseUrl)
		req.Header.Set("Connect", "close")
		if config.Method == "POST" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		resp, err := config.Client.Do(req)
		if err != nil && i < config.Retry-1 {
			//fmt.Printf("err: %v\n", err)
			continue
		} else if err != nil && i == config.Retry-1 {

			return Result{
				Url:    job.Url,
				Length: resp.ContentLength,
				Body:   "",
				Err:    err,
				Status: resp.StatusCode,
			}

		} else {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return Result{
					Url:    job.Url,
					Length: resp.ContentLength,
					Body:   "",
					Err:    err,
					Status: resp.StatusCode,
				}

			}
			return Result{
				Url:    job.Url,
				Length: resp.ContentLength,
				Body:   string(body),
				Err:    err,
				Status: resp.StatusCode,
			}

		}

	}
	return Result{
		Url: job.Url,
		Err: errors.New("未知错误"),
	}
}

func init() {

	flag.StringVar(&config.BaseUrl, "url", "127.0.0.1", "url or ip")
	flag.StringVar(&config.FileName, "file", "db/dicc.txt", "爆破字典")
	flag.StringVar(&config.Extensions, "extension", "php,js,do", "文件后缀 js,php....")
	flag.IntVar(&config.Num, "num", 10, "线程数量")
	flag.IntVar(&config.Delay, "delay", 10, "最长等待时间,单位秒")
	flag.IntVar(&config.Retry, "retry", 3, "错误重复次数")
	flag.BoolVar(&config.Redi, "redi", false, "是否跟踪重定向 (default false)")
	flag.BoolVar(&config.Ssl, "ssl", false, "是否采用https (default false)")
	flag.StringVar(&config.Method, "method", "GET", "请求方法 GET POST")
	flag.Parse()

	flag.Parse()

	config.Client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if config.Redi {
				return nil
			}
			return http.ErrUseLastResponse
		},
		Timeout: time.Second * time.Duration(config.Delay),
	}

	rand.Seed(time.Now().Unix())

	if config.BaseUrl == "" {
		log.Fatal("url is empty")
	}

	//判断是否加有http或https，先删除，后加上
	if strings.Index(config.BaseUrl, "://") == 4 || strings.Index(config.BaseUrl, "://") == 5 {
		_, config.BaseUrl, _ = strings.Cut(config.BaseUrl, "://")
	}
	if config.Ssl {
		config.BaseUrl = "https://" + config.BaseUrl
	} else {
		config.BaseUrl = "http://" + config.BaseUrl
	}

	//判断最后是否是/，没有则加上
	if (config.BaseUrl)[len(config.BaseUrl)-1] != '/' {
		config.BaseUrl = config.BaseUrl + "/"
	}

	//判断请求方法
	if config.Method != "GET" && config.Method != "POST" {
		log.Fatalf("can not support the way %s", config.Method)
	}
}

//
func testRead(Jobs chan<- Job) {
	for i := 0; i < 3; i++ {

		Jobs <- Job{
			Url: RandomWords(12),
			Fn:  HttpSent,
		}
	}
	defer close(Jobs)
}

//读取文件
func readFile(filename string, Jobs chan<- Job) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("open err: %v", err)
	}
	defer file.Close()
	defer close(Jobs)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		s := strings.Split(config.Extensions, ",")
		if strings.Contains(scanner.Text(), "%EXT%") {
			for i := range s {
				Jobs <- Job{
					Url: strings.Replace(scanner.Text(), "%EXT%", s[i], -1),
					Fn:  HttpSent,
				}
			}
		} else {
			Jobs <- Job{
				Url: scanner.Text(),
				Fn:  HttpSent,
			}

		}
	}
}

func main() {

	fir := time.Now()

	//一下是获取测试数据

	var mean Meaning
	var TestCode []int
	var TestLeng []int64
	var TestBody []string
	TestJobs := make(chan Job, 1)
	TestResults := make(chan Result, 1)

	Testwp := NewPool(config.Num, TestJobs, TestResults)
	go testRead(TestJobs)
	go Testwp.Run(&wg)

	for {
		result, ok := <-TestResults
		if !ok {
			break
		}
		if result.Err != nil {
			continue
		}
		TestBody = append(TestBody, result.Body)
		TestCode = append(TestCode, result.Status)
		TestLeng = append(TestLeng, result.Length)

	}

	for i := 0; i < len(TestCode)-1; i++ {
		if TestCode[i] != TestCode[i+1] {
			log.Fatalf("fuzzing test failed")
		}

	}
	mean.Length = TestLeng[0]
	mean.Status = TestCode[0]
	mean.Body = TestBody[0]

	//以下是爆破

	var wgtwo sync.WaitGroup
	Jobs := make(chan Job)
	Results := make(chan Result)
	wp := NewPool(config.Num, Jobs, Results)
	go readFile(config.FileName, Jobs)
	go wp.Run(&wgtwo)
	for {
		result, ok := <-Results
		if !ok {
			break
		}
		if result.Err != nil {
			continue
		}
		//如果状态不同则有效
		//如果状态相同，则比较长度，长度不同，则有效
		//最佳的方法是判断body的相似程度，待实现
		if result.Status != mean.Status {
			result.Print()
		} else if float32(result.Length) > 1.2*float32(mean.Length) && float32(result.Length) < 0.8*float32(mean.Length) {
			result.Print()
		}
	}
	fmt.Printf("耗时：%d  s", time.Since(fir)/1000000000)
}
