package main

import (
	"bufio"
	"log"
	"math/rand"
	"os"
)

//产生长度为n的字符串
func RandomWords(n int) string {

	const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

//返回随机agent
func RandomAgent() string {
	file, err := os.Open("db/user-agents.txt")
	if err != nil {
		log.Fatalf("get random agent err: %v", err)
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var line []string
	for scanner.Scan() {
		line = append(line, scanner.Text())
	}
	return line[rand.Intn(len(line))]
}
