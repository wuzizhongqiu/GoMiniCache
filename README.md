# GoMiniCache

用 Golang 实现一个简易的内存型数据库，通过核心代码，迅速上手如何用 Go 实现内存型数据库。

## Golang 实现 tcp 服务器

Golang 实现一个简易的 echo 服务器

```go
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
```

## Golang 自己封装拥有超时功能的 wait

go 原生的 sync.WaitGroup 不支持超时，而我们需要超时兜底，所以自己封装实现了拥有超时功能的 wait

```go
// Wait 跟库里的 sync.WaitGroup 基本一致，额外实现了超时功能
type Wait struct {
	wg sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

func (w *Wait) Done() {
	w.wg.Done()
}

func (w *Wait) Wait() {
	w.wg.Wait() // 阻塞，直到计数器为 0
}

// WaitWithTimeout 会阻塞，直到计数器为 0，或者超时
// 如果出现超时，返回 true
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan bool, 1)
	go func() {
		defer close(c)
		w.wg.Wait()
		c <- true
	}()
	select {
	case <-c:
		return false // 正常完成
	case <-time.After(timeout):
		return true // 出现超时
	}
}
```

## Golang 实现 RESP 协议解析器

### 给客户端的回复

以拼装字符串回复为例

```go
// BulkReply 存储二进制安全的字符串（也就是用 []byte 来存）
type BulkReply struct {
	Arg []byte
}

// MakeBulkReply 创建 BulkReply
// 比如说: 从数据库中查询一个信息，就是这样拼装返回的
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

// ToBytes 序列化 resp.Reply
func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	// 序列化成 RESP 协议的形式
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}
```

对外暴露 MakeBulkReply 方法调用，需要回复时调用即可

### 解析客户端的请求

以解析一个正常的 RESP请求为例

```go
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
```

### 实现类 Redis 的处理 Handle

```go
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
```

### 实现 Redis 风格的存储引擎

这是 database 的抽象

```go
// Database Redis 风格的存储引擎
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply // 执行指令
	AfterClientClose(c resp.Connection)                    // 关闭后的操作（善后工作）
	Close()                                                // 关闭连接
}

// DataEntity 指代 Redis 的数据结构，包括 string, list, hash, set 等等
type DataEntity struct {
	Data interface{}
}
```

## Golang 实现内存型数据库



## Golang 实现 Redis 持久化



## Golang 实现 Redis 集群