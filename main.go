package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"playGround/config"
	"playGround/utils"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/time/rate"
)

func main() {
	// 设置/获取环境变量
	utils.SetEnv("MY_ENV_VAR", "my_value")
	envValue := utils.GetEnv("MY_ENV_VAR", "default_value")
	fmt.Println("Environment variable:", envValue) // Environment variable: my_value

	// 设置/获取配置
	err := utils.LoadConfig("config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
	} else {
		configValue := utils.GetConfigValue("some_key")
		fmt.Println("Config value:", configValue) //Config value: some_value
	}
	// 格式化当前时间
	fmt.Println("Current Time =", utils.FormatTime(utils.GetCurrentTime())) //Current Time = 2024-08-30T14:38:08+08:00
	// 时间戳转换
	fmt.Println(utils.TimeInt642String(time.Now().Unix())) // 2024-08-30 14:42:21:08

	// float保留指定位数小数
	fmt.Println(utils.Round(3.1415926, 2)) // 3.14
	// 去除字符串中的标点符号
	fmt.Println(utils.RemovePunctuation("fsfwfefafadf_808?!;.")) //fsfwfefafadf808
	// 去除字符串两端的空白字符
	fmt.Println(utils.TrimWhitespace("    hello world  ")) //hello world
	// 字符串是否包含子字符串
	fmt.Println(utils.ContainsSubstring("abcdefghijklmn", "abc")) //true
	// 按分隔符分割字符串
	fmt.Println(utils.SplitString("abc,def,ghi,kl", ",")) //[abc def ghi kl]

	// JSON处理
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	p := &Person{
		Name: "John",
		Age:  18,
	}
	jsonStr, err := utils.ToJSON(p)
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
	} else {
		fmt.Println("JSON string:", jsonStr) //JSON string: {"name":"John","age":18}
	}
	pe := &Person{}
	err = utils.FromJSON(jsonStr, pe)
	if err == nil {
		fmt.Println("persion is", pe.Name) // persion is John
	}

	// 并发安全的Map
	safeMap := utils.NewSafeMap()
	safeMap.Set("key1", "value1")
	value, exists := safeMap.Get("key1")
	if exists {
		fmt.Println("SafeMap value:", value) //SafeMap value: value1
	}

	// 文件路径处理
	// 路径拼接
	path := utils.JoinPath("dir", "subdir", "file.txt")
	fmt.Println("Joined path:", path)
	absPath, err := utils.AbsPath(".")
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
	} else {
		// 绝对路径
		fmt.Println("Absolute path:", absPath)
	}

	// 文件操作
	if utils.FileIsExist("example.txt") {
		content, err := utils.ReadFile("example.txt")
		if err != nil {
			fmt.Println("Error reading file:", err)
		} else {
			fmt.Println("File content:", string(content)) //File content: 1234567890abc
		}
	}

	// 加密解密
	hash := utils.Md5("secret")
	fmt.Println("MD5 hash:", hash) //MD5 hash: 5ebe2294ecd0e0f08eab7690d2a6ee69

	// 随机数生成
	randomInt := utils.RandomInt(1, 100)
	fmt.Println("Random int:", randomInt) //Random int: 5
	randomStr := utils.RandomString(10)
	fmt.Println("Random string:", randomStr) //Random string: HnCObXyiF6

	// 网络请求
	response, err := utils.HTTPGet("https://www.baidu.com")
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
	} else {
		fmt.Println("HTTP response:", string(response))
	}

	// 重试机制
	err = utils.Retry(3, 1*time.Second, func() error {
		fmt.Println("Retrying...")
		return fmt.Errorf("some error")
	})
	if err != nil {
		fmt.Println("Retry failed:", err)
	}

	// 限流器
	limiter := utils.NewLimiter(rate.Limit(1), 1)
	if limiter.Allow() {
		fmt.Println("Limiter allowed")
	}

	// 缓存
	cache := utils.NewCache()
	cache.Set("cachekey1", "cachevalue1", 5*time.Second)
	cachedValue, found := cache.Get("cachekey1")
	if found {
		fmt.Println("Cache value:", cachedValue) //Cache value: cachevalue1
	}

	// 分布式锁（基于etcd）
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	if err != nil {
		fmt.Println("Etcd client error:", err)
	}
	etcdLock := utils.NewEtcdLock(etcdClient, "mylock", 10*time.Second)
	err = etcdLock.Lock(context.Background())
	if err != nil {
		fmt.Println("EtcdLock error:", err)
	} else {
		fmt.Println("EtcdLock acquired")
		etcdLock.Unlock(context.Background())
	}

	// 缓存（基于Redis）
	redisCache := utils.NewRedisCache("localhost:6379")
	err = redisCache.Set(context.Background(), "cachekey1", "cachevalue1", 5*time.Second)
	if err != nil {
		fmt.Println("RedisCache set error:", err)
	}
	cachedValue, err_ := redisCache.Get(context.Background(), "cachekey1")
	if err_ != nil {
		fmt.Println("RedisCache get error:", err)
	} else {
		fmt.Println("RedisCache value:", cachedValue)
	}
}

// 上传压缩文件
func UploadProgram(examId string, program map[string]string, scoreStandardId string) (ncPath string, err error) {
	// 创建一个临时目录来存放文本文件
	nowDate := time.Now().Format("20060102")
	filePath := config.Conf.UploadPath + nowDate + "/" + examId + "/"
	fileUrl := config.Conf.UploadUrl + nowDate + "/" + examId + "/"
	if err != nil {
		return "", err
	}
	// 单一考核标准情况
	if program != nil {
		// 检查zip文件是否存在，存在则删除
		zipFileName := filepath.Join(filePath, scoreStandardId+".zip")
		zipfilePath := filepath.Join(fileUrl, scoreStandardId+".zip")
		if _, err := os.Stat(zipFileName); !os.IsNotExist(err) {
			os.RemoveAll(zipFileName)
		}
		// 在函数结束时清理临时目录
		defer os.RemoveAll(filePath + scoreStandardId)
		// 遍历ncProgram，创建文件并写入内容
		for key, value := range program {
			fileName := filepath.Join(filePath+scoreStandardId, key+".txt")
			if _, err := os.Stat(zipFileName); !os.IsNotExist(err) {
				os.RemoveAll(fileName)
			}
			err = os.MkdirAll(filePath+scoreStandardId, 0755)
			if err != nil {
				return "", err
			}
			err := os.WriteFile(fileName, []byte(value), 0666)
			if err != nil {
				return "", err
			}
		}
		// 创建zip文件目录并添加内容到zip文件中
		err = utils.AddDirToZip(zipFileName, filePath+scoreStandardId, scoreStandardId)
		if err != nil {
			log.Println("Error adding directory to zip:", err)
			return "", err
		}
		return zipfilePath, nil
	}
	return
}

func Upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	file, fileHeader, err := r.FormFile("file")
	name := r.FormValue("name")
	if err != nil {
		log.Println("read file err", err)
		return
	}
	defer file.Close()
	objectId := primitive.NewObjectID().Hex()
	fileName := objectId + path.Ext(fileHeader.Filename)
	nowDate := time.Now().Format("20060102")
	filePath := config.Conf.UploadPath + config.Conf.UploadOfficeUrl + nowDate + "/"
	err = utils.IsFolder(filePath)
	if err != nil {
		log.Println("mkdir err", err)
		return
	}
	newFile, err := os.Create(filePath + fileName)
	if err != nil {
		log.Println("create file err", err)
		return
	}
	defer newFile.Close()
	_, err = io.Copy(newFile, file)
	if err != nil {
		log.Println("write file err", err)
		return
	}
	var resp interface{}
	if name == "zip" && path.Ext(fileHeader.Filename) == ".zip" {
		zipName, err := utils.Unzip2(filePath+fileName, filePath+objectId+"/")
		if err != nil {
			log.Println("unzip fail", err)
			return
		}
		resp = &struct {
			ZipPath   string `json:"zip_path"`
			ModelPath string `json:"model_path"`
		}{
			ZipPath:   "/" + config.Conf.UploadOfficeUrl + nowDate + "/" + fileName,
			ModelPath: "/" + config.Conf.UploadOfficeUrl + nowDate + "/" + objectId + "/" + zipName + "index.html",
		}
	} else {
		resp = &UploadResp{Path: "/" + config.Conf.UploadOfficeUrl + nowDate + "/" + fileName}
	}
	err = JSON(w, resp)
	if err != nil {
		log.Println("response err", err)
	}
}

type UploadResp struct {
	Path string
}

func JSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}
