package main

import (
	"net/http"
	"os"
)

/*
配置文件，设置
*/

type Config struct {
	BaseUrl    string //根url
	FileName   string //爆破字典
	Num        int    //线程数量
	Ssl        bool   //是否采用https
	Extensions string //文件后缀 js,php....
	Method     string //请求方法 GET POST
	Delay      int    //最长等待时间
	Redi       bool   //是否跟踪重定向
	Out        string //输出保存文件名
	Retry      int    //错误重复次数

	Client *http.Client //根据上面确定的Http端
	File   *os.File
}
