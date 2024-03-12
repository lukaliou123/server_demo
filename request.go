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
)

// ViewFileReq 实现 Requester 接口
func (req *ViewFileReq) Decode(r *http.Request) error {
	// 从 r 中解析出需要的信息，比如从 URL 中提取路径参数
	// 只要文件名字
	filePath := r.URL.Path[len(RouteFile):]

	// 把文件名字和路径结合,并将其赋值给 req.FilePath
	req.FilePath = filepath.Join(uploadPath, filePath)
	//fmt.Printf("in viewFileReq Decode, the req is: %s\n", req.FilePath)
	return nil
}

// 创建一个新的ViewFileReq，使其不为nil
func (req *ViewFileReq) CreateInstance() Requester {
	//return &ViewFileReq{}
	return new(ViewFileReq)
}

// ListFilesReq 实现 Requester 接口
func (req *ListFilesReq) Decode(r *http.Request) error {
	return nil
}

// 创建一个新的ListFilesReq，使其不为nil
func (req *ListFilesReq) CreateInstance() Requester {
	return new(ListFilesReq)
}
