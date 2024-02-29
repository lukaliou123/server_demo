package main

import (
	"fmt"
	"io/ioutil"
	"net/http" //核心件，用于各种http请求
	"os"
	"path/filepath"
)

const uploadPath = "./uploads" //文件上传的目录

func main() {
	//类似springboot中的controller层
	http.HandleFunc("/upload", uploadFileHandler)      // 上传文件
	http.HandleFunc("/download/", downloadFileHandler) // 处理文件下载
	http.HandleFunc("/files", listFilesHandler)        // 列出文件
	http.HandleFunc("/file/", viewFileHandler)         // 查看特定文件的详情

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

// 上传文件功能
// 一般来说，接口可以直接调用，然而结构体要用地址值，以免直接调用会导致每次使用调用结构体
// 通过使用指针（即*http.Request），可以确保所有函数都使用相同的请求实例。
// 这对于修改请求的状态或内容是必要的，如设置请求头、改变请求的URL等
func uploadFileHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		// 解析上传的文件
		err := request.ParseMultipartForm(10 << 20) //表示10*2^20，也就是10MB，这里表示限制上传大小10MB
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		file, handler, err := request.FormFile("myFile") //file为文件本身，handler表示这个文件的一些元数据,如文件名
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close() //关闭文件

		//在服务器上保存文件，其实也就是在本地的一个路径
		filePath := filepath.Join(uploadPath, handler.Filename) //自动处理字段为路径
		destination, err := os.Create(filePath)                 //创建这个路径的文件
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError) //这里如果出现上传错误，是服务器，也就是500
		}
		defer destination.Close() //关闭路径

		_, err = destination.ReadFrom(file) //表示将file文件流复制给destination
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(writer, "File uploaded successfully: %s", filePath)

	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(writer, "Only POST method is allowed") //Fprintf将一串字符写进一个io.writer接口中
	}
}

// 下载文件
func downloadFileHandler(writer http.ResponseWriter, request *http.Request) {
	//寻找代码名,这里是一个切片操作，得到路径后的名字
	fileName := request.URL.Path[len("/download/"):]
	//构成完整的路径，在服务器中或者上传路径的完整路径名
	filePath := filepath.Join(uploadPath, fileName)

	//检查文件是否存在
	_, err := os.Stat(filePath)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	//设置相应的头信息
	//attachment指下载，并规定下载下来的名字
	writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	//octet-stream指规定文件类型为二进制数据，方便io传输
	writer.Header().Set("Content-Type", "application/octet-stream")

	//发送文件
	http.ServeFile(writer, request, filePath)

}

// 列出所有文件
func listFilesHandler(writer http.ResponseWriter, request *http.Request) {
	//使用IO工具，读取路径
	files, err := ioutil.ReadDir(uploadPath)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	// 一个循环遍历
	for _, file := range files {
		fmt.Fprintf(writer, "%s\n", file.Name())
	}
}

// 查看特定文件内容和元数据
func viewFileHandler(writer http.ResponseWriter, request *http.Request) {
	//查看文件名字
	fileName := request.URL.Path[len("/file/"):]

	//获取文件元数据
	fileStat, err := os.Stat(fileName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusNotFound) //明确表示文件不存在
		return
	}

	//读取文件内容
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError) //可能有多种原因
		return
	}

	fmt.Fprintf(writer, "Name: %s\nSize: %d\nModTime: %s\nContent:\n%s",
		fileStat.Name(), fileStat.Size(), fileStat.ModTime(), string(content))
}

func init() {
	//创建上传目录
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		err := os.Mkdir(uploadPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}
