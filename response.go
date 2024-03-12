package main

import (
	"net/http"
)

type (
	Responder interface {
		Encode(w http.ResponseWriter) error
	}

	UploadResponse struct {
		Message  string `json:"msg"`
		FilePath string `json:"filePath"`
	}
	ListFilesResponse struct {
		Files []string `json:"files"`
	}
	ViewFileResponse struct {
		Name    string `json:"name"`
		Size    int64  `json:"sizeInBytes"`
		ModTime string `json:"modTime"`
	}
	DownloadResponse struct {
		Request *DownloadFileReq
	}
)

// UploadResponse 实现了 Responder 接口
func (resp *UploadResponse) Encode(w http.ResponseWriter) error {
	return sendJSONResponse(w, http.StatusOK, resp)
}

// FilesListResponse 实现了 Responder 接口
func (resp *ListFilesResponse) Encode(w http.ResponseWriter) error {

	return sendJSONResponse(w, http.StatusOK, resp)
}

// ViewFileResponse 实现了 Responder 接口
func (resp *ViewFileResponse) Encode(w http.ResponseWriter) error {
	return sendJSONResponse(w, http.StatusOK, resp)
}

// ViewFileResponse 实现了 Responder 接口
func (resp *DownloadResponse) Encode(w http.ResponseWriter) error {
	http.ServeFile(w, resp.Request.Request, resp.Request.FilePath)
	return nil
}
