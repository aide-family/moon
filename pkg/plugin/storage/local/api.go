package local

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/pkg/merr"
)

func (l *Local) RegisterHandler(srv *kratoshttp.Server) {
	c := l.c
	route := srv.Route("/")
	route.Handle(c.GetUploadMethod(), c.GetUploadURL(), func(c kratoshttp.Context) error {
		return l.UploadHandler(c.Response(), c.Request())
	})

	route.GET(c.GetPreviewURL(), func(c kratoshttp.Context) error {
		return l.PreviewHandler(c.Response(), c.Request())
	})
}

func (l *Local) UploadHandler(w http.ResponseWriter, r *http.Request) error {
	if strings.ToUpper(r.Method) != l.uploadMethod {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return merr.ErrorMethodNotAllowed("method not allowed %s", r.Method)
	}

	uploadID := r.URL.Query().Get("uploadID")
	partNumberStr := r.URL.Query().Get("partNumber")
	if uploadID == "" || partNumberStr == "" {
		http.Error(w, "missing uploadID or partNumber", http.StatusBadRequest)
		return merr.ErrorParamsError("missing uploadID or partNumber")
	}

	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber <= 0 {
		http.Error(w, "invalid partNumber", http.StatusBadRequest)
		return merr.ErrorParamsError("invalid partNumber").WithCause(err)
	}

	defer r.Body.Close()

	session, exists := l.uploads.Get(uploadID)
	if !exists {
		http.Error(w, "upload session not found", http.StatusNotFound)
		return merr.ErrorParamsError("upload session not found")
	}

	tempDir := filepath.Join(l.root, "tmp", uploadID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("failed to create temp directory: %v", err), http.StatusInternalServerError)
		return merr.ErrorInternalServerError("system err").WithCause(err)
	}

	tempFile, err := os.CreateTemp(tempDir, fmt.Sprintf("part_%d", partNumber))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create temp file: %v", err), http.StatusInternalServerError)
		return merr.ErrorInternalServerError("system err").WithCause(err)
	}
	defer tempFile.Close()

	hasher := md5.New()
	multiWriter := io.MultiWriter(tempFile, hasher)

	if _, err := io.Copy(multiWriter, r.Body); err != nil {
		http.Error(w, fmt.Sprintf("failed to write part data: %v", err), http.StatusInternalServerError)
		return merr.ErrorInternalServerError("system err").WithCause(err)
	}

	eTag := hex.EncodeToString(hasher.Sum(nil))

	session.parts.Set(partNumber, tempFile.Name())

	w.Header().Set("ETag", eTag)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"uploadID":   uploadID,
		"partNumber": partNumber,
		"eTag":       eTag,
		"size":       hasher.Size(),
	}); err != nil {
		return merr.ErrorInternalServerError("system err").WithCause(err)
	}
	return nil
}

func (l *Local) PreviewHandler(w http.ResponseWriter, r *http.Request) error {
	objectKey := r.URL.Query().Get("objectKey")
	if objectKey == "" {
		http.Error(w, "missing objectKey", http.StatusBadRequest)
		return merr.ErrorParamsError("missing objectKey")
	}

	filePath := objectKey
	if !strings.HasPrefix(filePath, l.root) {
		filePath = filepath.Join(l.root, filePath)
	}
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open file: %v", err), http.StatusInternalServerError)
		return merr.ErrorInternalServerError("system err").WithCause(err)
	}
	defer file.Close()
	defer r.Body.Close()

	ext := filepath.Ext(file.Name())
	contentType := getContentType(ext)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(file.Name())))
	if _, err := io.Copy(w, file); err != nil {
		return merr.ErrorInternalServerError("system err").WithCause(err)
	}
	return nil
}

func getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}
