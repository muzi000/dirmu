go 学习产物——地址扫描，实现了一小部分功能



# 使用方法

直接使用exe文件

```bash
.\dirmu.exe -h

Usage ofdirmu.exe:
  -delay int
        最长等待时间,单位秒 (default 10)
  -extension string
        文件后缀 js,php.... (default "php,js,do")
  -file string
        爆破字典 (default "db/dicc.txt")
  -method string
        请求方法 GET POST (default "GET")
  -num int
        线程数量 (default 10)
  -redi
        是否跟踪重定向 (default false)
  -retry int
        错误重复次数 (default 3)
  -ssl
        是否采用https (default false)
  -url string
        url or ip (default "127.0.0.1")
```

