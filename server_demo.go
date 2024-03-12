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
	errorFailedRequest    = "Failed to create request type"
)

func main() {

	// 类似springboot中的controller层
	http.HandleFunc("/upload", uploadFileHandler)               // 上传文件
	http.HandleFunc("/download/", downloadFileHandler)          // 处理文件下载
	http.HandleFunc("/files", handlerWrapper("GET", listFiles)) // 列出文件  也可以添加附加功能，如http.HandleFunc("/files", handlerWrapper(listFiles，“GET”))
	http.HandleFunc("/file/", handlerWrapper("GET", viewFile))  // 查看特定文件的详情

	fmt.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("HTTP server failed to start: %v\n", err)
		os.Exit(1) // 无法连接服务器时，直接退出
	}

}

// ErrResponse 错误输出信息
type ErrResponse struct {
	Msg string `json:"error"`
}

func (r ErrResponse) Error() string {
	return r.Msg
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
func sendJSONResponse(w http.ResponseWriter, status int, payload any) error {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, errorInternal, http.StatusInternalServerError)
		return err
	}
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// 写入json回包
	_, err = w.Write(response)
	if err != nil {
		// 也许这里我也该改一下？改为io?
		fmt.Printf("Error writing JSON response: %v\n", err)
	}
	return err
}

// 这里涉及到go的Generics
// 这个method要接受不同的参数
// 可以把一些东西加入到中间层，比如假如请求类属性的判断
func handlerWrapper[T Requester, U Responder](method string, f func(T) (U, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 方法检查
		if !checkRequestMethod(w, r, method) {
			return
		}
		// 初始化请求对象
		var prototype T
		// 这里利用类型断言转换形态
		reqInterface := prototype.CreateInstance()
		req, ok := reqInterface.(T)
		if !ok {
			// 失败了就说转换请求类型失败
			http.Error(w, errorFailedRequest, http.StatusInternalServerError)
			return
		}
		err := req.Decode(r)
		resp, err := f(req)
		// 3. 响应
		if err != nil {
			http.Error(w, errorInternal, http.StatusInternalServerError)
			return
		}
		// 3.1 成功
		err = resp.Encode(w)
		// 3.2 失败
		if err != nil {
			http.Error(w, errorInternal, http.StatusInternalServerError)
			return
		}
	}
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
		sendJSONResponse(w, http.StatusBadRequest, ErrResponse{Msg: errorBadRequest})
		return
	}

	file, handler, err := r.FormFile("myFile") // file为文件本身，handler表示这个文件的一些元数据,如文件名
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, ErrResponse{Msg: errorBadRequest})
		return
	}

	defer file.Close() // 关闭文件

	// 在服务器上保存文件，其实也就是在本地的一个路径
	filePath := filepath.Join(uploadPath, handler.Filename) // 自动处理字段为路径
	destination, err := os.Create(filePath)                 // 创建这个路径的文件
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, ErrResponse{Msg: errorInternal}) // 这里如果出现上传错误，是服务器，也就是500
		return
	}
	defer destination.Close() // 关闭路径

	_, err = destination.ReadFrom(file) // 表示将file文件流复制给destination
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, ErrResponse{Msg: errorInternal})
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
		sendJSONResponse(w, http.StatusNotFound, ErrResponse{Msg: errorStateNotFound})
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

// 列出所有文件的单独版本
func listFiles(req *ListFilesReq) (*ListFilesResponse, error) {
	files, err := os.ReadDir(uploadPath)
	if err != nil {
		fmt.Printf(err.Error())
		return nil, &ErrResponse{Msg: errorInternal}
	}
	var fileNames []string
	// 将所有的文件名放在一个array里
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return &ListFilesResponse{
		Files: fileNames,
	}, nil
}

func viewFile(req *ViewFileReq) (*ViewFileResponse, error) {
	//fmt.Printf("in viewFile, the req is: %s\n", req.FilePath)
	// 获取文件元数据
	fileStat, err := os.Stat(req.FilePath)
	if err != nil {
		fmt.Printf(err.Error())
		return nil, &ErrResponse{Msg: errorInternal}
	}

	return &ViewFileResponse{
		Name:    fileStat.Name(),
		Size:    fileStat.Size(),
		ModTime: fileStat.ModTime().Format(time.RFC3339),
	}, nil

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
