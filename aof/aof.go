package aof

import (
	"GoMiniCache/config"
	databaseface "GoMiniCache/interface/database"
	"GoMiniCache/lib/logger"
	"GoMiniCache/lib/utils"
	"GoMiniCache/resp/connection"
	"GoMiniCache/resp/parser"
	"GoMiniCache/resp/reply"
	"io"
	"os"
	"strconv"
)

const (
	aofQueueSize = 1 << 16 // 65535
)

type payload struct {
	cmdLine [][]byte
	dbIndex int
}

// HandlerAof 从通道接收消息并写入AOF文件
type HandlerAof struct {
	db          databaseface.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	currentDB   int
}

// NewAOFHandler 创建 aof.HandlerAof
func NewAOFHandler(db databaseface.Database) (*HandlerAof, error) {
	handler := &HandlerAof{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.db = db
	// 恢复曾经的AOF文件
	handler.LoadAof()
	// 打开AOF文件: os.O_APPEND: 表示在文件末尾追加数据; os.O_CREATE: 如果文件不存在，则创建一个新文件; os.O_RDWR: 表示以读写模式打开文件。
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize)
	go func() { // 起一个协程执行AOF
		handler.handleAof()
	}()
	logger.Info("start aof persistence...")
	return handler, nil
}

// AddAof 通过管道向处理AOF的协程发送命令
func (handler *HandlerAof) AddAof(dbIndex int, cmdLine [][]byte) {
	// 判断AOF是否启用
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// handleAof 监听管道传输的数据并写入文件
func (handler *HandlerAof) handleAof() {
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {
			// 使用其他数据库编号，编好格式，写入文件
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue
			}
			handler.currentDB = p.dbIndex
		}
		// 用的同一个数据库编号，直接写入即可
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
	}
}

// LoadAof 读取AOF文件
func (handler *HandlerAof) LoadAof() {
	file, err := os.Open(handler.aofFilename)
	if err != nil { // 如果文件不存在，就返回
		logger.Warn(err)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	ch := parser.ParseStream(file)       // 解析命令
	fakeConn := &connection.Connection{} // 用于记录 dbIndex
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF { // 读到文件结束符，读完这个AOF文件了
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil { // 遇到空指令，跳过
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply) // 需要转成二维
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		ret := handler.db.Exec(fakeConn, r.Args) // 执行命令
		if reply.IsErrorReply(ret) {
			logger.Error("exec err", err)
		}
	}
}
