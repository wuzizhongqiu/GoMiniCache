package tcp

/**
 * tcp 服务器
 */

import (
	"GoMiniCache/interface/tcp"
	"GoMiniCache/lib/logger"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Config tcp 服务器属性
type Config struct {
	Address string
}

// ListenAndServeWithSignal 监听和处理请求，阻塞，直到接收到停止信号
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() { // 阻塞接收信号
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	// 监听 tcp 协议上的连接
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	// 处理客户端的连接 listener
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe 接收和处理请求，阻塞，直到关闭
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// 阻塞等待关闭的信号
	go func() {
		<-closeChan
		logger.Info("shutting down...")
		_ = listener.Close() // 结束接听，停止 listener.Accept()，并返回 err
		_ = handler.Close()  // 关闭链接
	}()
	defer func() {
		// 发生意外错误时关闭
		_ = listener.Close()
		_ = handler.Close()
	}()

	// 接收并处理请求
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		// 接收客户端的连接
		conn, err := listener.Accept()
		if err != nil { // 当 listener.Accept() 返回 err，就跳出循环，停止接收和处理请求
			break
		}

		// 启动一个协程处理请求
		logger.Info("accept link...")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			// 调用处理请求的函数
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}
