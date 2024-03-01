package dict

// TestDick 光标停在 TestDick 结构体的名称上，Ctrl + i，搜索想要实现的 interface，就能一键生成下列代码了
type TestDick struct {
}

// MakeTestDick 实现一个 Make 创建函数，就能直接对外提供了
func MakeTestDick() *TestDick {
	return &TestDick{}
}

func (t TestDick) Get(key string) (val interface{}, exists bool) {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) Len() int {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) Put(key string, val interface{}) (result int) {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) PutIfAbsent(key string, val interface{}) (result int) {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) PutIfExists(key string, val interface{}) (result int) {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) Remove(key string) (result int) {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) ForEach(consumer Consumer) {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) Keys() []string {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) RandomKeys(limit int) []string {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) RandomDistinctKeys(limit int) []string {
	//TODO implement me
	panic("implement me")
}

func (t TestDick) Clear() {
	//TODO implement me
	panic("implement me")
}
