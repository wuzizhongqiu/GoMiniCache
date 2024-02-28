package resp

// Reply 序列号协议接口
type Reply interface {
	ToBytes() []byte // 序列化接口
}
