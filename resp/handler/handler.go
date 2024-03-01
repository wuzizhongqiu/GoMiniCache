package handler

/*
 * 根据 RESP 协议处理客户端的 tcp 连接
 */

import (
	"GoMiniCache/database/database"
	databaseface "GoMiniCache/interface/database"
	"GoMiniCache/lib/logger"
	"GoMiniCache/lib/sync/atomic"
	"GoMiniCache/resp/connection"
	"GoMiniCache/resp/parser"
	"GoMiniCache/resp/reply"
	"context"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

// RespHandler 实现 tcp.Handler
type RespHandler struct {
	activeConn sync.Map              // 存储客户端连接
	db         databaseface.Database // 数据库内核
	closing    atomic.Boolean        // 表示处于关闭状态，拒绝新请求
}

// MakeRespHandler 创建 RespHandler
func MakeRespHandler() *RespHandler {
	var db databaseface.Database
	db = database.NewDatabase()
	return &RespHandler{
		db: db,
	}
}

// closeClient 关闭这个客户端连接
func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Handle 接收并执行 Redis 命令
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	// 处于关闭状态
	if h.closing.Get() == true {
		_ = conn.Close()
	}

	// 创建并存储新连接
	client := connection.NewConn(conn)
	h.activeConn.Store(client, 1)

	// 将连接交给 parser.ParseStream 解析，他会将解析好的内容返回到这个管道
	ch := parser.ParseStream(conn)
	// 我们只需要监听这个管道即可
	for payload := range ch {
		// 如果出现错误
		if payload.Err != nil {
			// 如果出现 io 错误; 客户端断开连接; 使用一个已经关闭的连接; 就关闭客户端连接
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// 如果是出现协议错误，就返回错误回复，继续监听管道等待用户下一次数据
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		// 如果用户发送的参数为空，continue
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		// 转换成二维字符组
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		// 把结果传给内核数据库执行指令
		result := h.db.Exec(client, r.Args)
		// 将结果写回客户端
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else { // 如果结果为空，只能返回未知错误了（前面排了无数错误了）
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close 关闭客户端连接
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool { // 将所有客户端的连接都关闭
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}
