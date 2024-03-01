package resp

/*
 * 与客户端连接的抽象
 */

// Connection 抽象的与客户端的连接
type Connection interface {
	Write([]byte) error // 向客户端写消息
	GetDBIndex() int    // 获取数据库编号
	SelectDB(int)       // 选择数据库编号
}
