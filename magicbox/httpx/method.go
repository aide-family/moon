package httpx

import "net/http"

type Method string

const (
	MethodGet     Method = http.MethodGet
	MethodPost    Method = http.MethodPost
	MethodPut     Method = http.MethodPut
	MethodDelete  Method = http.MethodDelete
	MethodPatch   Method = http.MethodPatch
	MethodOptions Method = http.MethodOptions
	MethodHead    Method = http.MethodHead
	MethodTrace   Method = http.MethodTrace
	MethodConnect Method = http.MethodConnect
)
