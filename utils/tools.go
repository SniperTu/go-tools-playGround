package utils

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"

	"golang.org/x/time/rate"
)

/*
	环境变量及配置处理
*/
// GetEnv 获取环境变量的值，如果环境变量不存在，则返回默认值
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// SetEnv 设置环境变量的值
func SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// LoadConfig 从配置文件加载配置
func LoadConfig(path string) error {
	viper.SetConfigFile(path)
	return viper.ReadInConfig()
}

// GetConfigValue 获取配置项的值
func GetConfigValue(key string) interface{} {
	return viper.Get(key)
}

/*
时间处理
*/
// GetCurrentTime 获取当前时间
func GetCurrentTime() time.Time {
	return time.Now()
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// 时间戳转换（精确到毫秒）
func TimeInt642String(dateTime int64) string {
	unixTime := time.Unix(dateTime, 0)
	timeFormat := unixTime.In(time.FixedZone("CST", 8*60*60)).Format("2006-01-02 15:04:05:01")
	return timeFormat
}

/*
	字符串处理及数值处理
*/
// float保留指定位数小数
func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.1/pow10_n)*pow10_n) / pow10_n
}

// 去除字符串中的标点符号
func RemovePunctuation(s string) string {
	var punctuation = ",.!?;:_'\"，。！、～·（）"
	for _, char := range punctuation {
		s = strings.ReplaceAll(s, string(char), "")
	}
	return s
}

// TrimWhitespace 去除字符串两端的空白字符
func TrimWhitespace(s string) string {
	return strings.TrimSpace(s)
}

// ContainsSubstring 检查字符串是否包含子字符串
func ContainsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// SplitString 按分隔符分割字符串
func SplitString(s, sep string) []string {
	return strings.Split(s, sep)
}

/*
	JSON处理
*/
// ToJSON 将结构体转换为JSON字符串
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 将JSON字符串转换为结构体
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

/*
	struct和map之间的转换
*/
// mongo中结构体转map
func Struct2Map(v interface{}) map[string]interface{} {
	keys := reflect.TypeOf(v)
	vals := reflect.ValueOf(v)
	var rs = make(map[string]interface{})
	var length = keys.NumField()
	for i := 0; i < length; i++ {
		key := keys.Field(i).Tag.Get("bson")
		val := vals.Field(i).Interface()
		switch t := val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint32, uint64, float32, float64:
			if t == 0 {
				continue
			}
		case string:
			if t == "" {
				continue
			}
		}
		rs[key] = val
	}
	return rs
}

/*
	并发处理
*/

// SafeMap 并发安全的Map
type SafeMap struct {
	sync.RWMutex
	data map[string]interface{}
}

func NewSafeMap() *SafeMap {
	return &SafeMap{data: make(map[string]interface{})}
}

// Get 获取值
func (m *SafeMap) Get(key string) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()
	val, exists := m.data[key]
	return val, exists
}

// Delete 删除值
func (m *SafeMap) Delete(key string) {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)
}

// Set 设置值
func (m *SafeMap) Set(key string, value interface{}) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value
}

/*
	文件操作
*/

// JoinPath 拼接文件路径
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// AbsPath 获取绝对路径
func AbsPath(path string) (string, error) {
	return filepath.Abs(path)
}

// ReadFile 读取文件内容
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile 写入文件内容
func WriteFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

// 文件是否存在
func FileIsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// 如果文件不存在，创建目录
func IsFolder(name string) (err error) {
	_, err = os.Stat(name)
	if os.IsNotExist(err) {
		err = os.MkdirAll(name, os.ModePerm)
	}
	return
}

// 文件目录压缩
func AddDirToZip(zipFileName string, dir string, pathInZip string) error {
	// 创建zip文件
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()
	// 获取目录信息
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	// 判断是否为目录
	if dirInfo.IsDir() {
		// 读取目录下的文件列表
		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		// 遍历文件列表
		for _, file := range files {
			// 构造在zip文件中的路径
			newPathInZip := filepath.Join(pathInZip, file.Name())
			if file.IsDir() {
				// 如果是目录，递归调用addDirToZip函数
				if err = AddDirToZip(zipFileName, filepath.Join(dir, file.Name()), newPathInZip); err != nil {
					return err
				}
			} else {
				// 如果是文件，打开文件并写入zip文件
				fileToZip, err := os.Open(filepath.Join(dir, file.Name()))
				if err != nil {
					return err
				}
				defer fileToZip.Close()
				zipFile, err := writer.Create(newPathInZip)
				if err != nil {
					return err
				}
				_, err = io.Copy(zipFile, fileToZip)
				if err != nil {
					return err
				}
			}
		}
	} else {
		fileToZip, err := os.Open(dir)
		if err != nil {
			return err
		}
		defer fileToZip.Close()
		zipFile, err := writer.Create(pathInZip)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fileToZip)
		if err != nil {
			return err
		}
	}
	return nil
}

// zip文件解压
func Unzip2(src string, dest string) (string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	var filenames = r.File[0].Name
	defer r.Close()
	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

/*
	加密解密
*/
// MD5加密
func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

/*
	随机数生成
*/

// RandomInt 生成指定范围内的随机整数
func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min
}

// RandomString 生成指定长度的随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

/*
	网络请求
*/

// HTTPGet 发送HTTP GET请求
func HTTPGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// HTTPPost 发送HTTP POST请求
func HTTPPost(url string, body []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

/*
	日志处理
*/

var lgr = logrus.New()

func Debugf(f string, args ...interface{}) {
	lgr.Debugf(f, args...)
}
func Infof(f string, args ...interface{}) {
	lgr.Infof(f, args...)
}
func Warnf(f string, args ...interface{}) {
	lgr.Warnf(f, args...)
}
func Errorf(f string, args ...interface{}) {
	lgr.Errorf(f, args...)
}
func Fatalf(f string, args ...interface{}) {
	lgr.Fatalf(f, args...)
}

/*
	分布式任务处理
*/

// Retry 重试函数
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
		}
		err = fn()
		if err == nil {
			return nil
		}
	}
	return err
}

// Limiter 限流器
type Limiter struct {
	limiter *rate.Limiter
}

func NewLimiter(r rate.Limit, b int) *Limiter {
	return &Limiter{limiter: rate.NewLimiter(r, b)}
}

// Allow 检查是否允许执行
func (l *Limiter) Allow() bool {
	return l.limiter.Allow()
}

// Wait 等待直到允许执行
func (l *Limiter) Wait(ctx context.Context) error {
	return l.limiter.Wait(ctx)
}

// RedisLock 基于Redis的分布式锁
type RedisLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

func NewRedisLock(client *redis.Client, key, value string, ttl time.Duration) *RedisLock {
	return &RedisLock{client: client, key: key, value: value, ttl: ttl}
}

// Lock 获取锁
func (l *RedisLock) Lock(ctx context.Context) (bool, error) {
	return l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
}

// Unlock 释放锁
func (l *RedisLock) Unlock(ctx context.Context) error {
	script := redis.NewScript(`
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `)
	return script.Run(ctx, l.client, []string{l.key}, l.value).Err()
}

// EtcdLock 基于etcd的分布式锁
type EtcdLock struct {
	client *clientv3.Client
	key    string
	ttl    time.Duration
}

func NewEtcdLock(client *clientv3.Client, key string, ttl time.Duration) *EtcdLock {
	return &EtcdLock{client: client, key: key, ttl: ttl}
}

// Lock 获取锁
func (l *EtcdLock) Lock(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, l.ttl)
	defer cancel()
	_, err := l.client.Put(ctx, l.key, "locked", clientv3.WithLease(clientv3.LeaseID(l.ttl)))
	return err
}

// Unlock 释放锁
func (l *EtcdLock) Unlock(ctx context.Context) error {
	_, err := l.client.Delete(ctx, l.key)
	return err
}

/*
	缓存
*/
// Cache 内存缓存
type Cache struct {
	sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

func NewCache() *Cache {
	return &Cache{items: make(map[string]cacheItem)}
}

// Set 设置缓存项
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.Lock()
	defer c.Unlock()
	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = cacheItem{value: value, expiration: expiration}
}

// Get 获取缓存项
func (c *Cache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}
	return item.value, true
}

// Delete 删除缓存项
func (c *Cache) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
}

// RedisCache Redis缓存
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建一个新的RedisCache
func NewRedisCache(addr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCache{client: client}
}

// Set 设置缓存项
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取缓存项
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Delete 删除缓存项
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
