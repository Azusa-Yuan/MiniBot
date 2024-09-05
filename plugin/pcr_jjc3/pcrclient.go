package pcrjjc3

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"

	// MessagePack 是一个轻量级的、速度快的二进制序列化格式

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	alphabet = "0123456789"
	hexChars = "0123456789abcdef"
)

type pcrclient struct {
	viewer_id  string
	short_udid string
	udid       string
	header     map[string]string
	apiroot    string
	login      bool
	client     *http.Client
	sync.RWMutex
}

func CreatePcrclient(udid, short_udid, viewer_id, platform, proxy string, header map[string]string) (p *pcrclient, err error) {
	if platform != "1" {
		platform = "5"
	}
	client := &http.Client{}
	proxyURL := &url.URL{}
	if proxy != "" {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			logrus.Errorln("proxy error", err)
			return
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client = &http.Client{
			Transport: transport,
		}
	} else {
		proxyURL = nil
	}

	header["SID"] = makemd5(viewer_id + udid)
	header["platform"] = "1"
	apiroot := "https://api" + platform + "-pc.so-net.tw"
	p = &pcrclient{viewer_id: viewer_id, short_udid: short_udid,
		udid: udid, header: header, apiroot: apiroot,
		login: false, client: client}
	return
}

func (p *pcrclient) createkey(str string) []byte {
	randomBytes := make([]byte, 32)

	for i := 0; i < 32; i++ {
		randomBytes[i] = str[rand.Intn(len(str))]
	}
	return randomBytes
}

func (p *pcrclient) _ivstring() string {
	// return "83206083111718785322348355702973"
	return string(p.createkey(alphabet))
}

func (p *pcrclient) _encode(dat string) string {
	// Step 1: Calculate length of dat in hexadecimal format
	lengthHex := fmt.Sprintf("%04x", len(dat))
	// Step 2: Encode each character based on the rules specified
	var encodedChars []string
	for i := 0; i < len(dat)*4; i++ {
		if i%4 == 2 {
			// Modify character at position i/4 (integer division)
			index := i / 4
			char := dat[index]

			// Perform a simple transformation (e.g., adding 10 to ASCII value)
			transformed := byte(int(char) + 10)
			encodedChars = append(encodedChars, string(transformed))
		} else {
			// Choose a random character from a predefined alphabet
			// randomIndex, _ := random.Int(random.Reader, big.NewInt(int64(len(alphabet))))
			// encodedChars = append(encodedChars, string(alphabet[randomIndex.Int64()]))
			encodedChars = append(encodedChars, "2")

		}
	}

	// Step 3: Generate the final encoded string by concatenating all parts
	encodedString := lengthHex + strings.Join(encodedChars, "") + p._ivstring()
	return encodedString
}

// iv是恒定的，根据用户信息来
func (p *pcrclient) getiv() []byte {
	cleanedUdid := strings.ReplaceAll(p.udid, "-", "")
	shortenedUdid := cleanedUdid[:16]
	return []byte(shortenedUdid)
}

func (p *pcrclient) pack(data map[string]interface{}, key []byte) ([]byte, []byte, error) {

	// 将数据编码为 msgpack 格式,这里和python的行为不太一样,但好像不影响
	packed, err := msgpack.Marshal(data)
	if err != nil {
		return nil, nil, err
	}

	// packed := packedData.Bytes()

	encryptData, err := p.encrypt(packed, key)
	if err != nil {
		return nil, nil, err
	}

	return packed, encryptData, nil
}

// 这里是携带的数据就带key了
func (p *pcrclient) encrypt(data, key []byte) ([]byte, error) {
	// 这里采用的key为32字节，
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesIV := p.getiv()
	aesBlockMode := cipher.NewCBCEncrypter(block, aesIV)

	// Pad the data as per PKCS#7 standard
	paddedData := pkcs7Padding(data, aes.BlockSize)

	encrypted := make([]byte, len(paddedData))
	aesBlockMode.CryptBlocks(encrypted, paddedData)

	return append(encrypted, key...), nil
}

func (p *pcrclient) decrypt(data []byte) ([]byte, []byte, error) {
	key := data[len(data)-32:]
	encryptedData := data[:len(data)-32]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesIV := p.getiv()
	aesBlockMode := cipher.NewCBCDecrypter(block, aesIV)

	decrypted := make([]byte, len(encryptedData))
	aesBlockMode.CryptBlocks(decrypted, encryptedData)

	return decrypted, key, nil
}

func (p *pcrclient) unpack(rawdata []byte) (gjson.Result, []byte, error) {
	data := make([]byte, len(rawdata))
	n, _ := base64.StdEncoding.Decode(data, rawdata)
	data = data[:n]
	decrypted, key, _ := p.decrypt(data)
	dec, _ := pkcs7Unpadding(decrypted)

	// 定义结构体来存储解析后的数据
	response := map[string]interface{}{}

	// msgpack可能有问题
	err := msgpack.Unmarshal(dec, &response)
	if err != nil {
		logrus.Errorln("解析 MessagePack 失败:", err)
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		logrus.Errorln("转化json失败:", err)
	}

	res := gjson.ParseBytes(jsonData)
	return res, key, err
}

func (p *pcrclient) createHeader(packed []byte, apiUrl string) map[string]string {

	// 将 packed 数据进行 Base64 编码
	b64packed := base64.StdEncoding.EncodeToString(packed)
	// 计算 SHA1 散列值
	hasher := sha1.New()
	hasher.Write([]byte(p.udid + apiUrl + b64packed + p.viewer_id))
	sha1Hash := hex.EncodeToString(hasher.Sum(nil))

	// 获取header
	p.RLock()
	defer p.RUnlock()
	header := map[string]string{}
	for k, v := range p.header {
		header[k] = v
	}
	header["PARAM"] = sha1Hash
	header["SHORT-UDID"] = p._encode(p.short_udid)
	return header
}

func (p *pcrclient) CallApi(apiUrl string, request map[string]interface{}) (res gjson.Result, err error) {

	key := p.createkey(hexChars)

	if p.viewer_id != "" {
		encryptViewerId, err := p.encrypt([]byte(p.viewer_id), key)
		if err != nil {
			logrus.Error("encrypt fail", err)
		}
		base64ViewerId := make([]byte, base64.StdEncoding.EncodedLen(len(encryptViewerId)))
		base64.StdEncoding.Encode(base64ViewerId, encryptViewerId)
		request["viewer_id"] = base64ViewerId
	}

	packed, crypted, err := p.pack(request, key)
	if err != nil {
		logrus.Error("pack fail", err)
	}

	header := p.createHeader(packed, apiUrl)

	// 构建POST请求
	url := p.apiroot + apiUrl
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(crypted))
	if err != nil {
		logrus.Error("pcrclient request erro", err)
		return
	}

	// 设置自定义的Header
	for key, value := range header {
		req.Header.Set(key, value)
	}
	// 发送请求
	respRaw, err := p.client.Do(req)
	if err != nil {
		logrus.Error("请求失败:", err)
		fmt.Println(err)
		return
	}
	defer respRaw.Body.Close()
	if respRaw.StatusCode != http.StatusOK {
		err = fmt.Errorf(respRaw.Status)
		return
	}

	// 读取响应的内容
	body, err := io.ReadAll(respRaw.Body)
	if err != nil {
		logrus.Error("读取响应失败:", err)
		return
	}

	resp, _, err := p.unpack(body)
	if err != nil {
		logrus.Error("Fail to unpack body")
	}

	respHeader := resp.Get("data_headers")

	if respHeader.Get("viewer_id").Exists() {
		p.viewer_id = respHeader.Get("viewer_id").String()
	}

	if respHeader.Get("required_res_ver").Exists() {
		p.header["RES-VER"] = respHeader.Get("required_res_ver").String()
	}

	data := resp.Get("data")

	if data.Get("server_error").Exists() {
		err = fmt.Errorf("pcrclient: %s failed: %v", url, resp.Get("server_error").String())
		logrus.Error("[pcr]", err)
		p.login = false
		return
	}

	res = data
	return
}

func (p *pcrclient) updateVersion(version string) {
	p.Lock()
	defer p.Unlock()
	p.header["APP-VER"] = version

}

func (p *pcrclient) Login() {
	p.CallApi("/check/check_agreement", map[string]interface{}{})
	p.CallApi("/check/game_start", map[string]interface{}{})
	// p.CallApi("/load/index", map[string]interface{}{"carrier": "Android"})
	p.login = true
}

// PKCS#7 padding and unpadding
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("invalid padding size")
	}
	padding := data[length-1]
	return data[:length-int(padding)], nil
}

func makemd5(str string) string {
	salt := "r!I@nt8e5i="

	// Concatenate the string and the salt
	concatenated := str + salt

	// Compute MD5 hash
	hasher := md5.New()
	hasher.Write([]byte(concatenated))
	hashBytes := hasher.Sum(nil)

	// Convert the hash to a hexadecimal string
	hashHex := hex.EncodeToString(hashBytes)

	return hashHex
}
