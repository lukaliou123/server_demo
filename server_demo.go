package main

import (
	"fmt"
	"net/http" // 核心件，用于各种http请求
	"os"
	"path/filepath"
)

const (
	// controller层路径
	RouteUpload   = "/upload"
	RouteDownload = "/download/"
	RouteFiles    = "/files"
	RouteFile     = "/file/"

	// 本地储存文件路径
	uploadPath = "./uploads" //文件上传的目录

	// 错误提示
	errorInternal         = "Internal server error"
	errorBadRequest       = "Bad request"
	errorStateNotFound    = "State not found"
	errorMethodNotAllowed = "Method Not Allowed"
)

func main() {
	// 类似springboot中的controller层

	http.HandleFunc(RouteUpload, uploadFileHandler)     // 上传文件
	http.HandleFunc(RouteDownload, downloadFileHandler) // 处理文件下载
	http.HandleFunc(RouteFiles, listFilesHandler)       // 列出文件
	http.HandleFunc(RouteFile, viewFileHandler)         // 查看特定文件的详情

	fmt.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("HTTP server failed to start: %v\n", err)
		os.Exit(1) // 无法连接服务器时，直接退出
	}

}

// 一个辅助函数，用来判断是否符合需要的请求类型
func checkRequestMethod(writer http.ResponseWriter, request *http.Request, expectedMethod string) bool {
	if request.Method != expectedMethod {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		_, err := fmt.Fprintf(writer, errorMethodNotAllowed)
		if err != nil {
			http.Error(writer, errorInternal, http.StatusInternalServerError)
			return false
		}
		return false
	}
	return true
}

// 上传文件功能
// 一般来说，接口可以直接调用，然而结构体要用地址值，以免直接调用会导致每次使用调用结构体
// 通过使用指针（即*http.Request），可以确保所有函数都使用相同的请求实例。
// 这对于修改请求的状态或内容是必要的，如设置请求头、改变请求的URL等
func uploadFileHandler(writer http.ResponseWriter, request *http.Request) {
	// 只接受POST请求
	if !checkRequestMethod(writer, request, "POST") {
		return
	}

	// 解析上传的文件
	err := request.ParseMultipartForm(10 << 20) // 表示10*2^20，也就是10MB，这里表示限制上传大小10MB
	if err != nil {
		http.Error(writer, errorBadRequest, http.StatusBadRequest)
		return
	}

	file, handler, err := request.FormFile("myFile") // file为文件本身，handler表示这个文件的一些元数据,如文件名
	if err != nil {
		http.Error(writer, errorBadRequest, http.StatusBadRequest)
		return
	}

	defer file.Close() // 关闭文件

	// 在服务器上保存文件，其实也就是在本地的一个路径
	filePath := filepath.Join(uploadPath, handler.Filename) // 自动处理字段为路径
	destination, err := os.Create(filePath)                 // 创建这个路径的文件
	if err != nil {
		http.Error(writer, errorInternal, http.StatusInternalServerError) // 这里如果出现上传错误，是服务器，也就是500
		return
	}
	defer destination.Close() // 关闭路径

	_, err = destination.ReadFrom(file) // 表示将file文件流复制给destination
	if err != nil {
		http.Error(writer, errorInternal, http.StatusInternalServerError)
		return
	}

	if _, err := fmt.Fprintf(writer, "File uploaded successfully: %s", filePath); err != nil {
		http.Error(writer, errorInternal, http.StatusInternalServerError)
		return
	}

}

// 下载文件
func downloadFileHandler(writer http.ResponseWriter, request *http.Request) {
	// 只接受GET请求
	if !checkRequestMethod(writer, request, "GET") {
		return
	}
	// 寻找代码名,这里是一个切片操作，得到路径后的名字
	fileName := request.URL.Path[len(RouteDownload):]
	// 构成完整的路径，在服务器中或者上传路径的完整路径名
	filePath := filepath.Join(uploadPath, fileName)

	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	// 设置相应的头信息
	// attachment指下载，并规定下载下来的名字
	writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	// octet-stream指规定文件类型为二进制数据，方便io传输
	writer.Header().Set("Content-Type", "application/octet-stream")

	// 发送文件
	http.ServeFile(writer, request, filePath)

}

// 列出所有文件
func listFilesHandler(writer http.ResponseWriter, request *http.Request) {
	// 只接受GET请求
	if !checkRequestMethod(writer, request, "GET") {
		return
	}
	// 使用IO工具，读取路径
	files, err := os.ReadDir(uploadPath)
	if err != nil {
		http.Error(writer, errorInternal, http.StatusInternalServerError)
		return
	}

	// 一个循环遍历
	for _, file := range files {
		if _, err := fmt.Fprintf(writer, "%s\n", file.Name()); err != nil {
			fmt.Printf("Error writing file name to response: %v\n", err)      // 记录日志
			http.Error(writer, errorInternal, http.StatusInternalServerError) // 向用户发送通用错误消息
			return
		}
	}
}

// 查看特定文件内容和元数据
func viewFileHandler(writer http.ResponseWriter, request *http.Request) {
	// 只接受GET请求
	if !checkRequestMethod(writer, request, "GET") {
		return
	}
	// 查看文件名字
	fileName := request.URL.Path[len(RouteFile):]

	// 获取文件元数据
	fileStat, err := os.Stat(fileName)
	if err != nil {
		http.Error(writer, errorStateNotFound, http.StatusNotFound) //明确表示文件不存在
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(fileName)
	if err != nil {
		http.Error(writer, errorInternal, http.StatusInternalServerError) //可能有多种原因
		return
	}

	if _, err := fmt.Fprintf(writer, "Name: %s\nSize: %d\nModTime: %s\nContent:\n%s",
		fileStat.Name(), fileStat.Size(), fileStat.ModTime(), string(content)); err != nil {
		fmt.Println("Error writing file details to response:", err) // 记录日志
		return
	}

}

func init() {
	// 创建上传目录
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		err := os.Mkdir(uploadPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}
