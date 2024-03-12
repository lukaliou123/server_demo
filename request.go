package main

import (
	"net/http"
	"path/filepath"
)

type (
	Requester interface {
		Decode(r *http.Request) error
		CreateInstance() Requester
	}
	ViewFileReq struct {
		FilePath string `json:"filePath"`
	}
	ListFilesReq struct {
		FilePath string `json:"filePath"`
	}
	DownloadFileReq struct {
		FileName string `json:"fileName"`
		FilePath string `json:"filePath"`
		Request  *http.Request
	}
)

// Decode ViewFileReq 实现 Requester 接口
func (req *ViewFileReq) Decode(r *http.Request) error {
	// 从 r 中解析出需要的信息，比如从 URL 中提取路径参数
	// 只要文件名字
	filePath := r.URL.Path[len(RouteFile):]

	// 把文件名字和路径结合,并将其赋值给 req.FilePath
	req.FilePath = filepath.Join(uploadPath, filePath)
	//fmt.Printf("in viewFileReq Decode, the req is: %s\n", req.FilePath)
	return nil
}

// CreateInstance 创建一个新的ViewFileReq，使其不为nil
func (req *ViewFileReq) CreateInstance() Requester {
	//return &ViewFileReq{}
	return new(ViewFileReq)
}

// Decode ListFilesReq 实现 Requester 接口
func (req *ListFilesReq) Decode(r *http.Request) error {
	req.FilePath = uploadPath
	return nil
}

// CreateInstance 创建一个新的ListFilesReq，使其不为nil
func (req *ListFilesReq) CreateInstance() Requester {
	return new(ListFilesReq)
}

// Decode DownloadFileReq 实现 Requester 接口
func (req *DownloadFileReq) Decode(r *http.Request) error {
	// 寻找代码名,这里是一个切片操作，得到路径后的名字
	fileName := r.URL.Path[len(RouteDownload):]
	// 构成完整的路径，在服务器中或者上传路径的完整路径名
	req.FileName = fileName
	req.FilePath = filepath.Join(uploadPath, fileName)
	req.Request = r
	return nil
}

// CreateInstance 创建一个新的DownloadFileReq，使其不为nil
func (req *DownloadFileReq) CreateInstance() Requester {
	return new(DownloadFileReq)
}
