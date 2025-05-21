package bo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type OperateLogListRequest struct {
	*PaginationRequest
	Operation string
	Keyword   string
	UserID    uint32
	TimeRange []time.Time
}

func (r *OperateLogListRequest) ToListReply(logs []do.OperateLog) *OperateLogListReply {
	return &OperateLogListReply{
		PaginationReply: r.ToReply(),
		Items:           logs,
	}
}

type OperateLogListReply = ListReply[do.OperateLog]

type TeamOperateLogListReply = ListReply[do.OperateLog]

type OperateLogParams struct {
	Operation     string
	MenuName      string
	MenuID        uint32
	Request       string
	Error         string
	OriginRequest string
	Duration      time.Duration
	RequestTime   time.Time
	ReplyTime     time.Time
	ClientIP      string
	UserAgent     string
	UserID        uint32
	UserBaseInfo  string
	TeamID        uint32
}

type HttpRequest struct {
	Method           string          `json:"method"`
	URL              *url.URL        `json:"url"`
	Proto            string          `json:"proto"`
	ProtoMajor       int             `json:"proto_major"`
	ProtoMinor       int             `json:"proto_minor"`
	Header           http.Header     `json:"header"`
	Body             string          `json:"body"`
	ContentLength    int64           `json:"content_length"`
	TransferEncoding []string        `json:"transfer_encoding"`
	Host             string          `json:"host"`
	Form             url.Values      `json:"form"`
	PostForm         url.Values      `json:"post_form"`
	MultipartForm    *multipart.Form `json:"multipart_form"`
	Trailer          http.Header     `json:"trailer"`
	RemoteAddr       string          `json:"remote_addr"`
	RequestURI       string          `json:"request_uri"`
	Pattern          string          `json:"pattern"`
}

func (r *HttpRequest) String() string {
	body, _ := json.Marshal(r)
	return string(body)
}

// NewHttpRequest creates a new HttpRequest from an http.Request.
// It reads and preserves the request body for logging purposes.
func NewHttpRequest(request *http.Request) (*HttpRequest, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	// Restore the body for subsequent reads
	request.Body = io.NopCloser(bytes.NewBuffer(body))

	return &HttpRequest{
		Method:           request.Method,
		URL:              request.URL,
		Proto:            request.Proto,
		ProtoMajor:       request.ProtoMajor,
		ProtoMinor:       request.ProtoMinor,
		Header:           request.Header,
		Body:             string(body),
		ContentLength:    request.ContentLength,
		TransferEncoding: request.TransferEncoding,
		Host:             request.Host,
		Form:             request.Form,
		PostForm:         request.PostForm,
		MultipartForm:    request.MultipartForm,
		Trailer:          request.Trailer,
		RemoteAddr:       request.RemoteAddr,
		RequestURI:       request.RequestURI,
		Pattern:          request.Pattern,
	}, nil
}
