package utils

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Error Error
type Error struct {
	err    error
	params []interface{}
}

func (err *Error) Error() string {
	if err == nil {
		return "<nil>"
	}
	str := fmt.Sprint(err.params...)
	if err.err != nil {
		str += fmt.Sprintf(" err:%s", err.err.Error())
	}
	return str
}

// NewError NewError
func NewError(err error, params ...interface{}) error {
	return &Error{err, params}
}

// JSONEncode JSONEncode
func JSONEncode(data interface{}) []byte {
	d, err := json.Marshal(data)
	if err != nil {
		logrus.Errorln("JSONEncode error:", err)
	}
	return d
}

// JSONDecode JSONDecode
func JSONDecode(data []byte, obj interface{}) error {
	return json.Unmarshal(data, obj)
}

func RandInt(min, max int) int {
	if max < min {
		return 0
	}
	max++
	max -= min
	rand.Seed(time.Now().UnixNano())
	r := rand.Int()
	return r%max + min
}

const (
	letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// RandString https://github.com/kpbird/golang_random_string
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	// fill output
	for pos := range output {
		// get random item
		random := randomness[pos]
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = letterBytes[randomPos]
	}

	return string(output)
}

func timeoutClient() *http.Client {
	connectTimeout := time.Duration(20 * time.Second)
	readWriteTimeout := time.Duration(30 * time.Second)
	return &http.Client{
		Transport: &http.Transport{
			DialContext:         timeoutDialer(connectTimeout, readWriteTimeout),
			MaxIdleConnsPerHost: 200,
			DisableKeepAlives:   true,
		},
	}
}
func timeoutDialer(cTimeout time.Duration,
	rwTimeout time.Duration) func(ctx context.Context, net, addr string) (c net.Conn, err error) {
	return func(ctx context.Context, netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

// PostRequest PostRequest
func PostRequest(url string, bodyType string, body io.Reader) ([]byte, error) {
	client := timeoutClient()
	resp, err := client.Post(url, bodyType, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respbody, nil
}

// PostJSONRequest PostJSONRequest
func PostJSONRequest(url string, data interface{}) ([]byte, error) {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return PostRequest(url, "application/json;charset=UTF-8", bytes.NewReader(bytesData))
}

// GetRequest GetRequest
func GetRequest(url string) ([]byte, error) {
	client := timeoutClient()
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respbody, nil
}

// GetMD5 GetMD5
func GetMD5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// XMLDecode XMLDecode
func XMLDecode(data []byte, v interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(v)
}

// Max Max
func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ResolveSelfIP ResolveSelfIP
func ResolveSelfIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip, nil
		}
	}
	return nil, errors.New("server not connected to any network")
}

// GBK 转 UTF-8
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// UTF-8 转 GBK
func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// ==================== 密码加密工具 ====================

// EncryptPassword 密码加密（SHA256 + MD5 + Salt）
func EncryptPassword(password string) string {
	// 生成随机盐
	salt := RandString(16)
	// SHA256 哈希
	hash := sha256.New()
	hash.Write([]byte(password))
	sha256Hash := hex.EncodeToString(hash.Sum(nil))
	// MD5 哈希
	md5Hash := GetMD5(sha256Hash)
	// 组合：salt$hash
	return fmt.Sprintf("%s$%s", salt, md5Hash)
}

// VerifyPassword 验证密码
func VerifyPassword(password string, encrypted string) bool {
	// 分离盐和哈希
	// 格式：salt$hash
	if len(encrypted) < 33 {
		return false
	}
	parts := splitPassword(encrypted)
	if len(parts) != 2 {
		return false
	}
	// salt := parts[0] // 盐用于未来扩展，当前版本暂不使用
	storedHash := parts[1]

	// 使用相同的盐进行哈希
	hash := sha256.New()
	hash.Write([]byte(password))
	sha256Hash := hex.EncodeToString(hash.Sum(nil))
	md5Hash := GetMD5(sha256Hash)

	return md5Hash == storedHash
}

// splitPassword 分割密码字符串
func splitPassword(encrypted string) []string {
	for i := 0; i < len(encrypted); i++ {
		if encrypted[i] == '$' {
			return []string{encrypted[:i], encrypted[i+1:]}
		}
	}
	return nil
}

// ChangePassword 修改密码（返回新的加密密码）
func ChangePassword(oldPassword string, oldEncrypted string, newPassword string) (string, bool) {
	// 验证旧密码
	if !VerifyPassword(oldPassword, oldEncrypted) {
		return "", false
	}
	// 返回新密码的加密结果
	return EncryptPassword(newPassword), true
}
