package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/DigDeeply/bloom"
)

/**
* @brief safeHandler
*
* @param http.HandlerFunc
*
* @return
 */
func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				log.Println("WARN: paninc in %v - %v", fn, e)
				log.Println(string(debug.Stack()))
			}
		}()
		fn(w, r)
	}
}

/**
* @brief statusHandler 检查bloom的运行状态
*
* @param http.ResponseWriter
* @param http.Request
*
* @return
 */
func statusHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok\n")
}

/**
* @brief addbloomHandler 向bloom中添加数据
*
* @param http.ResponseWriter
* @param http.Request
*
* @return
 */
func addbloomHandler(w http.ResponseWriter, r *http.Request) {

	// 指定bloom容器
	id := r.FormValue("id")

	// 检查该容器是否已经存在
	bh, exist := bhMap[id]
	if exist == false {
		io.WriteString(w, "-1")
		return
	}

	// 关键字
	kw := r.FormValue("keyword")

	// 校验位
	tk := r.FormValue("token")

	// token算法.  加盐,md5,取第7 到 第12位
	salt := "tt.bloom"
	needle_tk := fmt.Sprintf("%x", md5.Sum([]byte(id+kw+salt)))[6:12]

	if needle_tk != tk {
		io.WriteString(w, "-2")
	} else {
		kwb := []byte(kw)
		check := bh.Add(kwb)
		if check != nil {
			io.WriteString(w, "1")
		} else {
			io.WriteString(w, "0")
		}
	}

	return
}

/**
* @brief bloomHandler 检查bloom filter的handler
*
* @param http.ResponseWriter
* @param http.Request
*
* @return
 */
func bloomHandler(w http.ResponseWriter, r *http.Request) {

	// 待检查的关键字
	kw := r.FormValue("keyword")

	// 指定具体的bloom过滤器
	id := r.FormValue("id")

	// 检查该容器是否已经存在
	bh, exist := bhMap[id]
	if exist == false {
		io.WriteString(w, "-1")
		return
	}

	// 开始检查
	check := bh.Test([]byte(kw))
	if check {
		io.WriteString(w, "1")
	} else {
		io.WriteString(w, "0")
	}

	return
}

// 命令参数
var fileNames *string = flag.String("f", "mid", "The files of id1 and id2 to be bloom filtered. The fileter id1 will be isolated from id2.")
var port *int = flag.Int("p", 8080, "listen port")

// 保存bloom实例
var bhMap = make(map[string]*bloom.BloomFilter)

// 协程通信控制相关
var msg chan *Msg

// 协程通信数据结构体
type Msg struct {
	Name        string
	BloomFilter *bloom.BloomFilter
	Err         error
}

/**
* @brief main 主函数
*
* @return
 */
func main() {

	// 初始化命令行参数
	flag.Parse()
	log.Printf("start load bloom filter dict [%s]...\n", *fileNames)

	// 处理文件参数，用来初始化bloom
	fileNameArr := strings.Split(*fileNames, ",")

	// 创建一个缓存长度为数组长度的channel
	msg = make(chan *Msg, len(fileNameArr))

	// 遍历数组文件，逐个初始化容器
	for i := 0; i < len(fileNameArr); i++ {
		go initBloom(fileNameArr[i])
	}

	// channel中发送的值
	log.Printf("Wait child processes start...")
	for i := 1; i <= len(fileNameArr); i++ {
		ret := <-msg
		log.Print("Bloom filter of " + ret.Name + " infomation:")
		if ret.Err != nil {
			log.Print("Init failed!", ret.Err.Error())
			return
		} else {
			log.Print("Success!")
			bhMap[ret.Name] = ret.BloomFilter
		}
	}
	log.Printf("Child processes completed!")

	http.HandleFunc("/status", safeHandler(statusHandler))
	http.HandleFunc("/bloom", safeHandler(bloomHandler))
	http.HandleFunc("/addbloom", safeHandler(addbloomHandler))

	log.Printf("server start at port  %d...\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Print("bloon listen error", err.Error())
	}

	// 关闭channel
	defer close(msg)

	return
}

/**
* @brief initBloom 初始化容器
*
* @param string
* @param int
*
* @return
 */
func initBloom(fileName string) {

	// 返回值
	ret := new(Msg)
	defer func(ret *Msg) {
		msg <- ret
	}(ret)

	// 根据给出的filename读取里边的id值，即去掉后缀名的文件名字
	basename := filepath.Base(fileName)
	ret.Name = basename

	// 创建一个新实例
	bh := bloom.New(10000000000, 5)

	// 添加词典
	file, err := os.Open(fileName)
	if err != nil {
		ret.Err = err
		return
	}

	defer func(file *os.File) {
		file.Close()
	}(file)

	// 读取文件line-by-line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		bh.Add([]byte(scanner.Text()))
	}

	// 检查是否出现错误
	err = scanner.Err()
	if err != nil {
		ret.Err = err
		return
	}

	// 过滤器
	ret.BloomFilter = bh
	return
}
