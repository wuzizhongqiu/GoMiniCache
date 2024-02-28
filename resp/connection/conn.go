package connection

/*
 * 与客户端的连接情况
 */

import (
	"GoMiniCache/lib/sync/wait"
	"net"
	"sync"
	"time"
)

// Connection 表示的是与客户端的连接
type Connection struct {
	conn         net.Conn
	waitingReply wait.Wait // 在关闭服务的时候需要 wait 解决未完成的任务
	mu           sync.Mutex
	selectedDB   int // Redis 有 16 个独立的数据库，这里指示的是正在操作的那个
}

// NewConn 创建新连接
func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// RemoteAddr 返回远程的地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close 关闭客户端连接
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// Write 向客户端写消息
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock() // 一个协程一个连接，理论上是不存在并发问题的（防止特殊情况出问题）
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()

	_, err := c.conn.Write(b) // 写给客户端（核心操作）
	return err
}

// GetDBIndex 获取选择的数据库编号
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB 选择数据库
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
