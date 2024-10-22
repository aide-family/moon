package bo

import (
	"bytes"
	"mime/multipart"
)

type (

	// UploadFileParams 上传文件参数
	UploadFileParams struct {
		FileName  string        `json:"fileName"`
		BytesBuff *bytes.Buffer `json:"bytesBuff"`
		// 文件后最
		Extension string `json:"extension"`
	}

	// ConvertFileParams 转换文件参数
	ConvertFileParams struct {
		UploadType string         `json:"uploadType"`
		File       multipart.File `json:"file"`
		Filename   string         `json:"filename"`
	}
	// UploadResParams 上传文件返回参数
	UploadResParams struct {
		Url string `json:"url"`
	}

	// DownLoadFileParams 下载文件参数
	DownLoadFileParams struct {
		FilePath string `json:"filePath"`
	}
)
