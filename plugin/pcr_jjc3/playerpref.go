package pcrjjc3

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"io"
	"net/url"
	"os"
	"strconv"
)

var key = []byte("e806f6")

func deckey(s string) string {
	unquoted, err := url.QueryUnescape(s)
	if err != nil {
		return ""
	}

	// 对解码后的字符串进行Base64解码
	decoded, err := base64.StdEncoding.DecodeString(unquoted)
	if err != nil {
		return ""
	}

	// 进行解密操作
	result := make([]byte, len(decoded))
	for i := 0; i < len(decoded); i++ {
		result[i] = key[i%len(key)] ^ decoded[i]
	}

	return string(result)
}

func decval(k, s string) []byte {
	unquoted, err := url.QueryUnescape(s)
	if err != nil {
		return nil
	}

	// 对解码后的字符串进行Base64解码
	decoded, err := base64.StdEncoding.DecodeString(unquoted)
	if err != nil {
		return nil
	}

	key2 := append([]byte(k), key...)

	// Adjust length of decoded based on condition
	if decoded[len(decoded)-5] != 0 {
		decoded = decoded[:len(decoded)-11]
	} else {
		decoded = decoded[:len(decoded)-7]
	}

	// Decrypt bytes
	result := make([]byte, len(decoded))
	for i := 0; i < len(decoded); i++ {
		result[i] = key2[i%len(key2)] ^ decoded[i]
	}

	return result
}

func decryptxml(path string) map[string]string {
	//xmlData := `<?xml version='1.0' encoding='utf-8' standalone='yes' ?><map><string name="BzN2">RU9</string><string name="K3d">T0</string><int name="Scre" value="720" /></map>`
	file, _ := os.Open(path)
	type String struct {
		Name  string `xml:"name,attr"`
		Value string `xml:",innerxml"`
	}

	type Int struct {
		Name  string `xml:"name,attr"`
		Value int    `xml:"value,attr"`
	}
	// 定义结构体来映射 XML 中的数据
	type XMLData struct {
		XMLName xml.Name `xml:"map"` // 这里的xml:"map"对应XML中的根节点名称
		Strings []String `xml:"string"`
		Ints    []Int    `xml:"int"`
	}

	// 实例化一个 XMLData 结构体变量
	var data XMLData

	raw, _ := io.ReadAll(file)
	// 使用 xml 包解析 XML 数据
	xml.Unmarshal(raw, &data)
	// if err := decoder.Decode(&data); err != nil {
	// 	fmt.Println("Error decoding XML:", err)
	// 	return
	// }
	infoMap := map[string]string{}
	for _, v := range data.Strings {
		k := deckey(v.Name)
		if k == "" {
			continue
		}

		val := decval(k, v.Value)
		var value string
		if k == "UDID" {
			// 根据 Python 中的代码逻辑，遍历范围为 0 到 35
			tmp := []byte{}
			for i := 0; i < 36; i++ {
				// 计算索引，类似于 Python 中的 val[4 * i + 6]
				index := 4*i + 6

				// 取出对应索引处的字节
				b := val[index]

				// 将字节的整数值减去 10
				tmp = append(tmp, b-10) // 使用 rune 将整数值转换为 Unicode 字符
			}

			value = string(tmp)

		} else if len(val) == 4 {
			// 声明一个变量用来存储解析后的整数值
			var result uint32

			// 使用 binary 包中的 Read 函数解析字节数组
			buf := bytes.NewReader(val)
			binary.Read(buf, binary.LittleEndian, &result)
			value = strconv.Itoa(int(result))
		} else {
			value = string(val)
		}

		infoMap[k] = value
	}

	// 将解析后的数据存入 map

	// 打印解析后的 map 数据
	return infoMap
}
