package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http" // 核心件，用于各种http请求
	"os"
	"path/filepath"
	"time"
)

const (
	// controller层路径
	RouteDownload = "/download/"
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
	http.HandleFunc("/upload", uploadFileHandler)      // 上传文件
	http.HandleFunc("/download/", downloadFileHandler) // 处理文件下载
	http.HandleFunc("/files", listFilesHandler)        // 列出文件
	http.HandleFunc("/file/", viewFileHandler)         // 查看特定文件的详情

	fmt.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("HTTP server failed to start: %v\n", err)
		os.Exit(1) // 无法连接服务器时，直接退出
	}

}

// 错误输出信息
type ErrResponse struct {
	Error string `json:"error"`
}

// 一个辅助函数，用来判断是否符合需要的请求类型
func checkRequestMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {

	if r.Method == expectedMethod {
		return true
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	_, err := fmt.Fprintf(w, errorMethodNotAllowed)
	if err != nil {
		log.Fatalf("Check Method failed, error: %s", err)
		//http.Error(w, errorInternal, http.StatusInternalServerError)
		return false
	}
	return false

}

// 一个辅助函数，用于将响应转为json格式
// 这里使用了空接口，类似泛型？用于接收任何类型的值，在这里方便转为json
func sendJSONResponse(w http.ResponseWriter, status int, payload any) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, errorInternal, http.StatusInternalServerError)
		return
	}
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// 写入json回包
	_, err = w.Write(response)
	if err != nil {
		fmt.Printf("Error writing JSON response: %v\n", err)
	}
}

type UploadResponse struct {
	Message  string `json:"msg"`
	FilePath string `json:"filePath"`
}

type FilesListResponse struct {
	Files []string `json:"files"`
}

type FileDetailsResponse struct {
	Name    string `json:"name"`
	Size    int64  `json:"sizeInBytes"`
	ModTime string `json:"modTime"`
}

// 上传文件功能
// 一般来说，接口可以直接调用，然而结构体要用地址值，以免直接调用会导致每次使用调用结构体
// 通过使用指针（即*http.Request），可以确保所有函数都使用相同请求实例。
// 这对于修改请求的状态或内容是必要的，如设置请求头、改变请求的URL等
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if !checkRequestMethod(w, r, "POST") {
		return
	}

	// 解析上传的文件
	err := r.ParseMultipartForm(10 << 20) // 表示10*2^20，也就是10MB，这里表示限制上传大小10MB
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, ErrResponse{Error: errorBadRequest})
		return
	}

	file, handler, err := r.FormFile("myFile") // file为文件本身，handler表示这个文件的一些元数据,如文件名
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, ErrResponse{Error: errorBadRequest})
		return
	}

	defer file.Close() // 关闭文件

	// 在服务器上保存文件，其实也就是在本地的一个路径
	filePath := filepath.Join(uploadPath, handler.Filename) // 自动处理字段为路径
	destination, err := os.Create(filePath)                 // 创建这个路径的文件
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, ErrResponse{Error: errorInternal}) // 这里如果出现上传错误，是服务器，也就是500
		return
	}
	defer destination.Close() // 关闭路径

	_, err = destination.ReadFrom(file) // 表示将file文件流复制给destination
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, ErrResponse{Error: errorInternal})
		return
	}

	response := UploadResponse{
		Message:  "File uploaded successfully",
		FilePath: filePath,
	}

	sendJSONResponse(w, http.StatusOK, response)

}

// 下载文件
func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if !checkRequestMethod(w, r, "GET") {
		return
	}
	// 寻找代码名,这里是一个切片操作，得到路径后的名字
	fileName := r.URL.Path[len(RouteDownload):]
	// 构成完整的路径，在服务器中或者上传路径的完整路径名
	filePath := filepath.Join(uploadPath, fileName)

	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if err != nil {
		sendJSONResponse(w, http.StatusNotFound, ErrResponse{Error: errorStateNotFound})
		return
	}
	// 设置相应的头信息
	// attachment指下载，并规定下载下来的名字
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	// octet-stream指规定文件类型为二进制数据，方便io传输
	w.Header().Set("Content-Type", "application/octet-stream")

	// 发送文件
	http.ServeFile(w, r, filePath)

}

// 列出所有文件
func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if !checkRequestMethod(w, r, "GET") {
		return
	}
	// 使用IO工具，读取路径
	files, err := os.ReadDir(uploadPath)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, ErrResponse{Error: errorInternal})
		return
	}

	var fileNames []string
	// 将所有的文件名放在一个array里
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	response := FilesListResponse{
		Files: fileNames,
	}
	// json文件中放所有的文件名
	sendJSONResponse(w, http.StatusOK, response)
}

// 查看特定文件元数据
func viewFileHandler(w http.ResponseWriter, r *http.Request) {
	// 只接受GET请求
	if !checkRequestMethod(w, r, "GET") {
		return
	}
	// 只要文件名字
	filePath := r.URL.Path[len(RouteFile):]

	// 把文件名字和路径结合
	fileName := filepath.Join(uploadPath, filePath)

	// 获取文件元数据
	fileStat, err := os.Stat(fileName)
	if err != nil {
		sendJSONResponse(w, http.StatusNotFound, ErrResponse{Error: errorStateNotFound}) //明确表示文件不存在
		return
	}

	response := FileDetailsResponse{
		Name:    fileStat.Name(),
		Size:    fileStat.Size(),
		ModTime: fileStat.ModTime().Format(time.RFC3339),
	}

	// 将json写入回包
	sendJSONResponse(w, http.StatusOK, response)

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
