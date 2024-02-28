package tcp

/*
 * 一个 echo 服务器，用于测试服务器是否正常运行
 */

import (
	"GoMiniCache/lib/logger"
	"GoMiniCache/lib/sync/atomic"
	"GoMiniCache/lib/sync/wait"
	"bufio"
	"context"
	"io"
	"net"
	"sync"
	"time"
)

// EchoHandler 用于测试的 echo 服务
type EchoHandler struct {
	activeConn sync.Map       // 记录连接数
	closing    atomic.Boolean // 记录状态（如果处于关闭状态，停止接收请求）（原子操作）
}

// MakeEchoHandler 用于创建 EchoHandler
func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

// EchoClient 是 EchoHandler 的客户端结构体
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// Close 关闭连接
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second) // 等待一些可能还未完成的请求处理完成
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// Handle 处理客户端的请求
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// 如果 EchoHandler 处于关闭状态，就停止接收请求，关闭请求的连接
		_ = conn.Close()
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{}) // 存储这个客户端连接到 activeConn

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
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		// 将收到的信息发送给客户端
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

// Close 停止 echo handler 的服务
func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true) // 进入关闭状态
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
