package emojimix

import "fmt"

type KeyPair struct {
	Key1 string
	Key2 string
}

// 实现接口，以便可以用作 map 的键
func (k KeyPair) String() string {
	// 确保 Key1 和 Key2 的顺序不影响组合键的唯一性
	if k.Key1 < k.Key2 {
		return fmt.Sprintf("%s|%s", k.Key1, k.Key2)
	}
	return fmt.Sprintf("%s|%s", k.Key2, k.Key1)
}

// 添加值的函数
func addValue(data map[string]string, key1 string, key2 string, value string) {
	pair := KeyPair{key1, key2}
	data[pair.String()] = value
}

// 获取值的函数
func getValue(data map[string]string, key1 string, key2 string) string {
	pair := KeyPair{key1, key2}
	return data[pair.String()]
}
