package tcp

import (
	"context"
	"net"
)

// Handler 基于 tcp 的应用服务器
type Handler interface {
	Handle(ctx context.Context, conn net.Conn) // 处理函数
	Close() error                              // 关闭连接
}
