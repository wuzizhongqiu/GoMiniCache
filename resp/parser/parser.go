package parser

/*
 * 解析用户发送过来的请求
 */

import (
	"GoMiniCache/interface/resp"
	"GoMiniCache/lib/logger"
	"GoMiniCache/resp/reply"
	"bufio"
	"errors"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

// Payload 存储 resp.Reply 和错误 err
type Payload struct {
	Data resp.Reply
	Err  error
}

// readState 读取解析器的状态
type readState struct {
	readingMultiLine  bool     // 解析的数据是单行还是多行（只有最开始是单行模式，读取第一行数据后保持多行模式）
	expectedArgsCount int      // 需要解析多少个参数
	msgType           byte     // 解析的数据类型
	args              [][]byte // 存储已经解析的数据
	bulkLen           int64    // 需要读取的数据长度（例：$？）
}

// finished 解析完成了
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// ParseStream 从 io.Reader 读取数据（io.Reader 是客户端给我们传来的字节流） 并通过管道异步的发送回客户端
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// parse0 解析客户端传来的数据的主逻辑
func parse0(reader io.Reader, ch chan<- *Payload) {
	// 接收错误信息，确保一个用户的 goroutine panic 后不会导致主协程崩溃
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	// 开启死循环读取用户输入的信息，直到他退出
	for {
		var ioErr bool
		// readLine 读一行数据
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil { // 出现错误
			if ioErr == true { // 如果出现 io 错误，就返回错误，关闭通信通道，结束通信
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			// 如果不是 io 错误，那就是协议解析出错，直接给用户返回错误
			ch <- &Payload{
				Err: err,
			}
			state = readState{} // 重置 readState 状态，让用户重新发送数据
			continue
		}

		// 解析读到的这一行数据，判断是不是多行解析模式（其实就是判断这行数据有没有解析）
		if !state.readingMultiLine { // 开始解析
			if msg[0] == '*' { // 第一个字符是'*'的情况
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 { // 需要解析的参数为0
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' { // 第一个字符是'$'的情况
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 { // 数据为空
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{}
					continue
				}
			} else { // 类似 +OK 这类单行数据
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{ // 直接返回内容了
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else { // 已经是多行读取模式（也就是初步解析过数据行了）
			err = readBody(msg, &state) // 最后对 body 进行解析
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{}
				continue
			}
			// 数据解析完成了，根据请求类型发送返回的数据
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}
}

// readLine 精确读取一行指令
// 例1: *3\r\n（向 Redis 服务器发送一个包含 3 个元素的数组）
// 例2: $3 SET\r\n（表示一个包含一个元素的数组，其中元素是一个长度为 3 的字符串"SET"）
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 { // 正常情况，直接根据 \r\n 进行切分
		msg, err = bufReader.ReadBytes('\n') // 根据 \n 切分数据流
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' { // 如果数据是空或者 \n 前面不是 \r 也要返回错误
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { // 根据之前读取的 $ 数字，严格读取字符的个数
		msg = make([]byte, state.bulkLen+2)  // 加上 \r\n 的两个字节
		_, err = io.ReadFull(bufReader, msg) // 读取 bulkLen+2 字节的数据
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' { // 数据不为空且结尾是 \r\n 才正确
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0 // $ 数据读完了，重置
	}
	return msg, false, nil
}

// parseMultiBulkHeader 前面 readLine 读取完一行之后，需要解析这一行数据的含义（正常的解析情况）
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	// 把无意义的部分切走，留下数字（例：*300\r\n, 切走第一个字符和最后两个字符）（注：base 是进制，bitSize 位数）
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 { // 用户没加参数，返回
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 { // 用户有加参数，处理
		state.msgType = msg[0]                       // 例：*3\r\n, msgType = * 表示他是个数组
		state.readingMultiLine = true                // 进入多行模式
		state.expectedArgsCount = int(expectedLine)  // 数据长度
		state.args = make([][]byte, 0, expectedLine) // 初始化 args
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// parseBulkHeader 如果是遇到 $ 开头的数据行，使用这个解析
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // 空
		return nil
	} else if state.bulkLen > 0 { // 修改解析器的状态
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1) // 给该数据初始化了一个切片，元素类型为[]byte，长度为0，容量为1
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// parseSingleLineReply 如果客户端发送类似 +OK 的信息，使用这个方法解析
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n") // 去除后缀"\r\n"
	var result resp.Reply
	switch msg[0] {
	case '+': // 状态回复
		result = reply.MakeStatusReply(str[1:])
	case '-': // 错误回复
		result = reply.MakeErrReply(str[1:])
	case ':': // 整数回复
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// readBody 前面解析完头数字，后续的 body 需要根据数字解析解析
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // 空
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else { // 正常情况就直接塞进 args 里面
		state.args = append(state.args, line)
	}
	return nil
}
