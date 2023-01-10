package collect

// 因为我们希望得到的某些信息是在之前的阶段获得的。在这里我将缓存结构定义为了一个哈希表，并封装了 Get 与 Set 两个函数来获取和设置请求中的缓存。

type Temp struct {
	data map[string]interface{}
}

// Get 返回临时缓存数据
func (t *Temp) Get(key string) interface{} {
	return t.data[key]
}

func (t *Temp) Set(key string, value interface{}) error {
	if t.data == nil { // 这里是初始化 map
		t.data = make(map[string]interface{}, 8)
	}
	t.data[key] = value
	return nil
}
