package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"testing"
)

var addressTest string = ":8080"

// TestListenAndServe 一个简单的 Echo 服务器，它会接受客户端连接并将客户端发送的内容原样传回客户端
func TestListenAndServe(t *testing.T) {
	// 绑定监听地址
	listener, err := net.Listen("tcp", addressTest)
	if err != nil {
		log.Fatalf(fmt.Sprintf("listen err: %v", err))
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)
	log.Println(fmt.Sprintf("bind: %s, start listening...", addressTest))

	for {
		// Accept 会一直阻塞直到有新的连接建立或者 listen 中断才会返回
		conn, err := listener.Accept()
		if err != nil {
			// 通常是 listen 被关闭导致的错误
			log.Fatalf(fmt.Sprintf("accept err: %v", err))
		}
		// 开启新的 goroutine 处理该请求
		go Handle(conn)
	}
}

// Handle 处理请求的逻辑
func Handle(conn net.Conn) {
	// 使用 bu-fio 标准库提供的缓冲区功能
	reader := bufio.NewReader(conn)
	for {
		// ReadString 会一直阻塞到遇到分隔符 '\n'
		// 遇到分隔符后会返回上次遇到分隔符或连接建立后收到的所有数据，包括分隔符本身
		// 如果在遇到分隔符之前遇到异常，ReadString 会返回已收到的数据和错误信息
		msg, err := reader.ReadString('\n')
		if err != nil {
			// 通常遇到的错误是连接中断或被关闭，用 io.EOF 表示
			if err == io.EOF {
				log.Println("connection close")
			} else {
				log.Println(err)
			}
			return
		}
		b := []byte(msg)
		// 将收到的信息发送给客户端
		_, err = conn.Write(b)
		if err != nil {
			return
		}
	}
}
