package main

import (
	"encoding/json"
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

// 一个辅助函数，用于将响应转为json格式
// 这里使用了空接口，类似泛型？用于接收任何类型的值，在这里方便转为json
func sendJSONResponse(writer http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(writer, errorInternal, http.StatusInternalServerError)
		return
	}
	// 设置响应头
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	// 写入json回包
	_, err = writer.Write(response)
	if err != nil {
		fmt.Printf("Error writing JSON response: %v\n", err)
	}
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
		sendJSONResponse(writer, http.StatusBadRequest, map[string]string{"error": errorBadRequest})
		return
	}

	file, handler, err := request.FormFile("myFile") // file为文件本身，handler表示这个文件的一些元数据,如文件名
	if err != nil {
		sendJSONResponse(writer, http.StatusBadRequest, map[string]string{"error": errorBadRequest})
		return
	}

	defer file.Close() // 关闭文件

	// 在服务器上保存文件，其实也就是在本地的一个路径
	filePath := filepath.Join(uploadPath, handler.Filename) // 自动处理字段为路径
	destination, err := os.Create(filePath)                 // 创建这个路径的文件
	if err != nil {
		sendJSONResponse(writer, http.StatusInternalServerError, map[string]string{"error": errorInternal}) // 这里如果出现上传错误，是服务器，也就是500
		return
	}
	defer destination.Close() // 关闭路径

	_, err = destination.ReadFrom(file) // 表示将file文件流复制给destination
	if err != nil {
		sendJSONResponse(writer, http.StatusInternalServerError, map[string]string{"error": errorInternal})
		return
	}

	sendJSONResponse(writer, http.StatusOK, map[string]string{"message": "File uploaded successfully", "filePath": filePath})

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
		sendJSONResponse(writer, http.StatusNotFound, map[string]string{"error": errorStateNotFound})
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
		sendJSONResponse(writer, http.StatusInternalServerError, map[string]string{"error": errorInternal})
		return
	}

	var fileNames []string
	// 将所有的文件名放在一个array里
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	// json文件中放所有的文件名
	sendJSONResponse(writer, http.StatusOK, map[string]interface{}{"files": fileNames})
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
		sendJSONResponse(writer, http.StatusNotFound, map[string]string{"error": errorStateNotFound}) //明确表示文件不存在
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(fileName)
	if err != nil {
		sendJSONResponse(writer, http.StatusInternalServerError, map[string]string{"error": errorInternal}) //可能有多种原因
		return
	}

	// 做一个map，key为类型，value为该类型内容，因为类型多样，所以用空接口
	fileDetails := map[string]interface{}{
		"name":    fileStat.Name(),
		"size":    fmt.Sprintf("%d bytes", fileStat.Size()), // fmt.Sprintf，类似java中的format
		"modTime": fileStat.ModTime(),
		"content": string(content),
	}

	// 将json写入回包
	sendJSONResponse(writer, http.StatusOK, fileDetails)

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
