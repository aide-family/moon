// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: merr/err.proto

package merr

import (
	_ "github.com/go-kratos/kratos/v2/errors"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 错误类型
type ErrorReason int32

const (
	// 用于表单验证错误
	ErrorReason_ALERT ErrorReason = 0
	// 用于弹窗验证错误, 需要提供确认按钮和确认请求的幂等键
	ErrorReason_MODAL ErrorReason = 1
	// 用于toast验证错误， 资源不存在或者已存在时候提示
	ErrorReason_TOAST ErrorReason = 2
	// 用于通知验证错误， 系统级别错误
	ErrorReason_NOTIFICATION ErrorReason = 3
	// 用于重定向验证错误, 跳转到指定页面， 认证级别提示
	ErrorReason_UNAUTHORIZED ErrorReason = 4
	// 权限不足时候提示, toast提示 权限级别提示
	ErrorReason_FORBIDDEN ErrorReason = 5
	// 触发频率限制
	ErrorReason_TOO_MANY_REQUESTS ErrorReason = 6
	// 文件相关
	ErrorReason_FILE_RELATED ErrorReason = 7
)

// Enum value maps for ErrorReason.
var (
	ErrorReason_name = map[int32]string{
		0: "ALERT",
		1: "MODAL",
		2: "TOAST",
		3: "NOTIFICATION",
		4: "UNAUTHORIZED",
		5: "FORBIDDEN",
		6: "TOO_MANY_REQUESTS",
		7: "FILE_RELATED",
	}
	ErrorReason_value = map[string]int32{
		"ALERT":             0,
		"MODAL":             1,
		"TOAST":             2,
		"NOTIFICATION":      3,
		"UNAUTHORIZED":      4,
		"FORBIDDEN":         5,
		"TOO_MANY_REQUESTS": 6,
		"FILE_RELATED":      7,
	}
)

func (x ErrorReason) Enum() *ErrorReason {
	p := new(ErrorReason)
	*p = x
	return p
}

func (x ErrorReason) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ErrorReason) Descriptor() protoreflect.EnumDescriptor {
	return file_merr_err_proto_enumTypes[0].Descriptor()
}

func (ErrorReason) Type() protoreflect.EnumType {
	return &file_merr_err_proto_enumTypes[0]
}

func (x ErrorReason) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ErrorReason.Descriptor instead.
func (ErrorReason) EnumDescriptor() ([]byte, []int) {
	return file_merr_err_proto_rawDescGZIP(), []int{0}
}

var File_merr_err_proto protoreflect.FileDescriptor

var file_merr_err_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x6d, 0x65, 0x72, 0x72, 0x2f, 0x65, 0x72, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x70, 0x6b, 0x67, 0x2e, 0x6d, 0x65, 0x72, 0x72, 0x1a, 0x13, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x73, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2a,
	0xea, 0x29, 0x0a, 0x0b, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12,
	0xdc, 0x10, 0x0a, 0x05, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x10, 0x00, 0x1a, 0xd0, 0x10, 0xa8, 0x45,
	0x90, 0x03, 0xb2, 0x45, 0x0c, 0xe5, 0x8f, 0x82, 0xe6, 0x95, 0xb0, 0xe9, 0x94, 0x99, 0xe8, 0xaf,
	0xaf, 0xba, 0x45, 0x05, 0x41, 0x4c, 0x45, 0x52, 0x54, 0xca, 0x45, 0x91, 0x01, 0x0a, 0x20, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x6c, 0x61, 0x72, 0x6d, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x4e, 0x61, 0x6d, 0x65, 0x5f, 0x4c, 0x65, 0x6e, 0x12,
	0x0f, 0xe7, 0x94, 0xa8, 0xe6, 0x88, 0xb7, 0xe5, 0x90, 0x8d, 0xe9, 0x94, 0x99, 0xe8, 0xaf, 0xaf,
	0x1a, 0x5c, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x27, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f,
	0x5f, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x6c, 0x61, 0x72, 0x6d, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x4e, 0x61, 0x6d, 0x65, 0x5f, 0x4c, 0x65,
	0x6e, 0x1a, 0x2b, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7, 0xbb, 0x84, 0xe5, 0x90, 0x8d, 0xe7,
	0xa7, 0xb0, 0xe9, 0x95, 0xbf, 0xe5, 0xba, 0xa6, 0xe9, 0x99, 0x90, 0xe5, 0x88, 0xb6, 0xe5, 0x9c,
	0xa8, 0x31, 0x2d, 0x32, 0x30, 0xe4, 0xb8, 0xaa, 0xe5, 0xad, 0x97, 0xe7, 0xac, 0xa6, 0xca, 0x45,
	0xb1, 0x01, 0x0a, 0x22, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x6c, 0x61, 0x72, 0x6d, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x52, 0x65, 0x6d, 0x61,
	0x72, 0x6b, 0x5f, 0x4c, 0x65, 0x6e, 0x12, 0x2c, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7, 0xbb,
	0x84, 0xe8, 0xaf, 0xb4, 0xe6, 0x98, 0x8e, 0xe9, 0x95, 0xbf, 0xe5, 0xba, 0xa6, 0xe9, 0x99, 0x90,
	0xe5, 0x88, 0xb6, 0xe5, 0x9c, 0xa8, 0x30, 0x2d, 0x32, 0x30, 0x30, 0xe4, 0xb8, 0xaa, 0xe5, 0xad,
	0x97, 0xe7, 0xac, 0xa6, 0x1a, 0x5d, 0x0a, 0x02, 0x69, 0x64, 0x12, 0x29, 0x41, 0x4c, 0x45, 0x52,
	0x54, 0x5f, 0x5f, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x6c, 0x61, 0x72, 0x6d, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x52, 0x65, 0x6d, 0x61, 0x72,
	0x6b, 0x5f, 0x4c, 0x65, 0x6e, 0x1a, 0x2c, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7, 0xbb, 0x84,
	0xe8, 0xaf, 0xb4, 0xe6, 0x98, 0x8e, 0xe9, 0x95, 0xbf, 0xe5, 0xba, 0xa6, 0xe9, 0x99, 0x90, 0xe5,
	0x88, 0xb6, 0xe5, 0x9c, 0xa8, 0x30, 0x2d, 0x32, 0x30, 0x30, 0xe4, 0xb8, 0xaa, 0xe5, 0xad, 0x97,
	0xe7, 0xac, 0xa6, 0xca, 0x45, 0x44, 0x0a, 0x0c, 0x50, 0x41, 0x53, 0x53, 0x57, 0x4f, 0x52, 0x44,
	0x5f, 0x45, 0x52, 0x52, 0x12, 0x0c, 0xe5, 0xaf, 0x86, 0xe7, 0xa0, 0x81, 0xe9, 0x94, 0x99, 0xe8,
	0xaf, 0xaf, 0x1a, 0x26, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x0c,
	0x50, 0x41, 0x53, 0x53, 0x57, 0x4f, 0x52, 0x44, 0x5f, 0x45, 0x52, 0x52, 0x1a, 0x0c, 0xe5, 0xaf,
	0x86, 0xe7, 0xa0, 0x81, 0xe9, 0x94, 0x99, 0xe8, 0xaf, 0xaf, 0xca, 0x45, 0x69, 0x0a, 0x11, 0x50,
	0x41, 0x53, 0x53, 0x57, 0x4f, 0x52, 0x44, 0x5f, 0x53, 0x41, 0x4d, 0x45, 0x5f, 0x45, 0x52, 0x52,
	0x12, 0x18, 0xe6, 0x96, 0xb0, 0xe6, 0x97, 0xa7, 0xe5, 0xaf, 0x86, 0xe7, 0xa0, 0x81, 0xe4, 0xb8,
	0x8d, 0xe8, 0x83, 0xbd, 0xe7, 0x9b, 0xb8, 0xe5, 0x90, 0x8c, 0x1a, 0x3a, 0x0a, 0x0b, 0x6e, 0x65,
	0x77, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x11, 0x50, 0x41, 0x53, 0x53, 0x57,
	0x4f, 0x52, 0x44, 0x5f, 0x53, 0x41, 0x4d, 0x45, 0x5f, 0x45, 0x52, 0x52, 0x1a, 0x18, 0xe6, 0x96,
	0xb0, 0xe6, 0x97, 0xa7, 0xe5, 0xaf, 0x86, 0xe7, 0xa0, 0x81, 0xe4, 0xb8, 0x8d, 0xe8, 0x83, 0xbd,
	0xe7, 0x9b, 0xb8, 0xe5, 0x90, 0x8c, 0xca, 0x45, 0x49, 0x0a, 0x13, 0x54, 0x45, 0x41, 0x4d, 0x5f,
	0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x12, 0x15,
	0xe5, 0x9b, 0xa2, 0xe9, 0x98, 0x9f, 0xe5, 0x90, 0x8d, 0xe7, 0xa7, 0xb0, 0xe5, 0xb7, 0xb2, 0xe5,
	0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x1b, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x13, 0x54,
	0x45, 0x41, 0x4d, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x5f, 0x45,
	0x52, 0x52, 0xca, 0x45, 0x44, 0x0a, 0x0b, 0x43, 0x41, 0x50, 0x54, 0x43, 0x48, 0x41, 0x5f, 0x45,
	0x52, 0x52, 0x12, 0x0f, 0xe9, 0xaa, 0x8c, 0xe8, 0xaf, 0x81, 0xe7, 0xa0, 0x81, 0xe9, 0x94, 0x99,
	0xe8, 0xaf, 0xaf, 0x1a, 0x24, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x0b, 0x43, 0x41, 0x50,
	0x54, 0x43, 0x48, 0x41, 0x5f, 0x45, 0x52, 0x52, 0x1a, 0x0f, 0xe9, 0xaa, 0x8c, 0xe8, 0xaf, 0x81,
	0xe7, 0xa0, 0x81, 0xe9, 0x94, 0x99, 0xe8, 0xaf, 0xaf, 0xca, 0x45, 0x50, 0x0a, 0x0e, 0x43, 0x41,
	0x50, 0x54, 0x43, 0x48, 0x41, 0x5f, 0x45, 0x58, 0x50, 0x49, 0x52, 0x45, 0x12, 0x12, 0xe9, 0xaa,
	0x8c, 0xe8, 0xaf, 0x81, 0xe7, 0xa0, 0x81, 0xe5, 0xb7, 0xb2, 0xe8, 0xbf, 0x87, 0xe6, 0x9c, 0x9f,
	0x1a, 0x2a, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x0e, 0x43, 0x41, 0x50, 0x54, 0x43, 0x48,
	0x41, 0x5f, 0x45, 0x58, 0x50, 0x49, 0x52, 0x45, 0x1a, 0x12, 0xe9, 0xaa, 0x8c, 0xe8, 0xaf, 0x81,
	0xe7, 0xa0, 0x81, 0xe5, 0xb7, 0xb2, 0xe8, 0xbf, 0x87, 0xe6, 0x9c, 0x9f, 0xca, 0x45, 0xa4, 0x01,
	0x0a, 0x19, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50,
	0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x12, 0x31, 0xe7, 0xad, 0x96,
	0xe7, 0x95, 0xa5, 0xe7, 0xbb, 0x84, 0x5b, 0x25, 0x73, 0x5d, 0xe6, 0x9c, 0xaa, 0xe5, 0x90, 0xaf,
	0xe7, 0x94, 0xa8, 0x2c, 0x20, 0xe4, 0xb8, 0x8d, 0xe5, 0x85, 0x81, 0xe8, 0xae, 0xb8, 0xe5, 0xbc,
	0x80, 0xe5, 0x90, 0xaf, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0x5b, 0x25, 0x73, 0x5d, 0x1a, 0x54,
	0x0a, 0x0d, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12,
	0x19, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x45, 0x4e, 0x41, 0x42, 0x4c, 0x45, 0x1a, 0x28, 0xe7, 0xad, 0x96, 0xe7,
	0x95, 0xa5, 0xe7, 0xbb, 0x84, 0xe6, 0x9c, 0xaa, 0xe5, 0x90, 0xaf, 0xe7, 0x94, 0xa8, 0x2c, 0xe4,
	0xb8, 0x8d, 0xe5, 0x85, 0x81, 0xe8, 0xae, 0xb8, 0xe5, 0xbc, 0x80, 0xe5, 0x90, 0xaf, 0xe7, 0xad,
	0x96, 0xe7, 0x95, 0xa5, 0xca, 0x45, 0x67, 0x0a, 0x16, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x4f,
	0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x44, 0x55, 0x50, 0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x12,
	0x12, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe5, 0xaf, 0xb9, 0xe8, 0xb1, 0xa1, 0xe9, 0x87, 0x8d,
	0xe5, 0xa4, 0x8d, 0x1a, 0x39, 0x0a, 0x0b, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x4f, 0x62, 0x6a, 0x65,
	0x63, 0x74, 0x12, 0x16, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54,
	0x5f, 0x44, 0x55, 0x50, 0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x1a, 0x12, 0xe5, 0x91, 0x8a, 0xe8,
	0xad, 0xa6, 0xe5, 0xaf, 0xb9, 0xe8, 0xb1, 0xa1, 0xe9, 0x87, 0x8d, 0xe5, 0xa4, 0x8d, 0xca, 0x45,
	0x70, 0x0a, 0x15, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x44,
	0x55, 0x50, 0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x12, 0x18, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5,
	0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7, 0xad, 0x89, 0xe7, 0xba, 0xa7, 0xe9, 0x87, 0x8d, 0xe5,
	0xa4, 0x8d, 0x1a, 0x3d, 0x0a, 0x0a, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x4c, 0x65, 0x76, 0x65, 0x6c,
	0x12, 0x15, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x44, 0x55,
	0x50, 0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x1a, 0x18, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe5,
	0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7, 0xad, 0x89, 0xe7, 0xba, 0xa7, 0xe9, 0x87, 0x8d, 0xe5, 0xa4,
	0x8d, 0xca, 0x45, 0x45, 0x0a, 0x11, 0x45, 0x4d, 0x41, 0x49, 0x4c, 0x5f, 0x43, 0x41, 0x50, 0x54,
	0x43, 0x48, 0x41, 0x5f, 0x45, 0x52, 0x52, 0x12, 0x15, 0xe9, 0x82, 0xae, 0xe7, 0xae, 0xb1, 0xe9,
	0xaa, 0x8c, 0xe8, 0xaf, 0x81, 0xe7, 0xa0, 0x81, 0xe9, 0x94, 0x99, 0xe8, 0xaf, 0xaf, 0x1a, 0x19,
	0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x11, 0x45, 0x4d, 0x41, 0x49, 0x4c, 0x5f, 0x43, 0x41,
	0x50, 0x54, 0x43, 0x48, 0x41, 0x5f, 0x45, 0x52, 0x52, 0xca, 0x45, 0x43, 0x0a, 0x15, 0x53, 0x45,
	0x4c, 0x45, 0x43, 0x54, 0x5f, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x50, 0x41, 0x47, 0x45, 0x5f,
	0x45, 0x52, 0x52, 0x12, 0x2a, 0xe9, 0x80, 0x89, 0xe6, 0x8b, 0xa9, 0xe5, 0x91, 0x8a, 0xe8, 0xad,
	0xa6, 0xe9, 0xa1, 0xb5, 0xe9, 0x9d, 0xa2, 0xe9, 0x94, 0x99, 0xe8, 0xaf, 0xaf, 0xef, 0xbc, 0x8c,
	0xe8, 0xaf, 0xb7, 0xe9, 0x87, 0x8d, 0xe6, 0x96, 0xb0, 0xe9, 0x80, 0x89, 0xe6, 0x8b, 0xa9, 0xca,
	0x45, 0x56, 0x0a, 0x13, 0x48, 0x4f, 0x4f, 0x4b, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55,
	0x50, 0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x12, 0x10, 0x68, 0x6f, 0x6f, 0x6b, 0xe5, 0x90, 0x8d,
	0xe7, 0xa7, 0xb0, 0xe9, 0x87, 0x8d, 0xe5, 0xa4, 0x8d, 0x1a, 0x2d, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x13, 0x48, 0x4f, 0x4f, 0x4b, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55, 0x50,
	0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x1a, 0x10, 0x68, 0x6f, 0x6f, 0x6b, 0xe5, 0x90, 0x8d, 0xe7,
	0xa7, 0xb0, 0xe9, 0x87, 0x8d, 0xe5, 0xa4, 0x8d, 0xca, 0x45, 0x57, 0x0a, 0x1a, 0x41, 0x4c, 0x45,
	0x52, 0x54, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55,
	0x50, 0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x12, 0x15, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7,
	0xbb, 0x84, 0xe5, 0x90, 0x8d, 0xe7, 0xa7, 0xb0, 0xe9, 0x87, 0x8d, 0xe5, 0xa4, 0x8d, 0x1a, 0x22,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x47, 0x52,
	0x4f, 0x55, 0x50, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55, 0x50, 0x4c, 0x49, 0x43, 0x41,
	0x54, 0x45, 0xca, 0x45, 0x5d, 0x0a, 0x1d, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f,
	0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55, 0x50, 0x4c, 0x49,
	0x43, 0x41, 0x54, 0x45, 0x12, 0x15, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe7, 0xbb, 0x84, 0xe5,
	0x90, 0x8d, 0xe7, 0xa7, 0xb0, 0xe9, 0x87, 0x8d, 0xe5, 0xa4, 0x8d, 0x1a, 0x25, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x1d, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x47, 0x52,
	0x4f, 0x55, 0x50, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55, 0x50, 0x4c, 0x49, 0x43, 0x41,
	0x54, 0x45, 0xca, 0x45, 0x4e, 0x0a, 0x17, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f,
	0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55, 0x50, 0x4c, 0x49, 0x43, 0x41, 0x54, 0x45, 0x12, 0x12,
	0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe5, 0x90, 0x8d, 0xe7, 0xa7, 0xb0, 0xe9, 0x87, 0x8d, 0xe5,
	0xa4, 0x8d, 0x1a, 0x1f, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x17, 0x53, 0x54, 0x52, 0x41,
	0x54, 0x45, 0x47, 0x59, 0x5f, 0x4e, 0x41, 0x4d, 0x45, 0x5f, 0x44, 0x55, 0x50, 0x4c, 0x49, 0x43,
	0x41, 0x54, 0x45, 0xca, 0x45, 0x60, 0x0a, 0x1d, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59,
	0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f,
	0x45, 0x58, 0x49, 0x53, 0x54, 0x12, 0x18, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe7, 0xbb, 0x84,
	0xe7, 0xb1, 0xbb, 0xe5, 0x9e, 0x8b, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a,
	0x25, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1d, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47,
	0x59, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4e, 0x4f, 0x54,
	0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0xca, 0x45, 0x51, 0x0a, 0x17, 0x53, 0x54, 0x52, 0x41, 0x54,
	0x45, 0x47, 0x59, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x45, 0x58, 0x49,
	0x53, 0x54, 0x12, 0x15, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe7, 0xb1, 0xbb, 0xe5, 0x9e, 0x8b,
	0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x1f, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x17, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0xca, 0x45, 0x2b, 0x0a, 0x15, 0x41,
	0x4c, 0x45, 0x52, 0x54, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46,
	0x4f, 0x55, 0x4e, 0x44, 0x12, 0x12, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7, 0xbb, 0x84, 0xe4,
	0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x2e, 0x0a, 0x18, 0x53, 0x54, 0x52,
	0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4e, 0x4f, 0x54, 0x5f,
	0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x12, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe7, 0xbb, 0x84,
	0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x2a, 0x0a, 0x14, 0x44, 0x41,
	0x54, 0x41, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55,
	0x4e, 0x44, 0x12, 0x12, 0xe6, 0x95, 0xb0, 0xe6, 0x8d, 0xae, 0xe6, 0xba, 0x90, 0xe4, 0xb8, 0x8d,
	0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x2d, 0x0a, 0x14, 0x41, 0x4c, 0x45, 0x52, 0x54,
	0x5f, 0x50, 0x41, 0x47, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12,
	0x15, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe9, 0xa1, 0xb5, 0xe9, 0x9d, 0xa2, 0xe4, 0xb8, 0x8d,
	0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x2e, 0x0a, 0x15, 0x41, 0x4c, 0x45, 0x52, 0x54,
	0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44,
	0x12, 0x15, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7, 0xad, 0x89, 0xe7, 0xba, 0xa7, 0xe4, 0xb8,
	0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x34, 0x0a, 0x1b, 0x53, 0x54, 0x52, 0x41,
	0x54, 0x45, 0x47, 0x59, 0x5f, 0x54, 0x45, 0x4d, 0x50, 0x4c, 0x41, 0x54, 0x45, 0x5f, 0x4e, 0x4f,
	0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x15, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe6,
	0xa8, 0xa1, 0xe6, 0x9d, 0xbf, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x12, 0xe8,
	0x01, 0x0a, 0x05, 0x4d, 0x4f, 0x44, 0x41, 0x4c, 0x10, 0x01, 0x1a, 0xdc, 0x01, 0xa8, 0x45, 0x95,
	0x03, 0xb2, 0x45, 0x09, 0xe8, 0xaf, 0xb7, 0xe7, 0xa1, 0xae, 0xe8, 0xae, 0xa4, 0xba, 0x45, 0x05,
	0x4d, 0x4f, 0x44, 0x41, 0x4c, 0xca, 0x45, 0x5f, 0x0a, 0x0e, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x52,
	0x4d, 0x5f, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x12, 0x0c, 0xe7, 0xa1, 0xae, 0xe8, 0xae, 0xa4,
	0xe5, 0x88, 0xa0, 0xe9, 0x99, 0xa4, 0x1a, 0x19, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x72,
	0x6d, 0x12, 0x0e, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x52, 0x4d, 0x5f, 0x44, 0x45, 0x4c, 0x45, 0x54,
	0x45, 0x1a, 0x17, 0x0a, 0x06, 0x63, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x12, 0x0d, 0x43, 0x41, 0x4e,
	0x43, 0x45, 0x4c, 0x5f, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x1a, 0x0b, 0x0a, 0x09, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44, 0xca, 0x45, 0x5f, 0x0a, 0x0e, 0x43, 0x4f, 0x4e, 0x46,
	0x49, 0x52, 0x4d, 0x5f, 0x55, 0x50, 0x44, 0x41, 0x54, 0x45, 0x12, 0x0c, 0xe7, 0xa1, 0xae, 0xe8,
	0xae, 0xa4, 0xe4, 0xbf, 0xae, 0xe6, 0x94, 0xb9, 0x1a, 0x19, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x72, 0x6d, 0x12, 0x0e, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x52, 0x4d, 0x5f, 0x55, 0x50, 0x44,
	0x41, 0x54, 0x45, 0x1a, 0x17, 0x0a, 0x06, 0x63, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x12, 0x0d, 0x43,
	0x41, 0x4e, 0x43, 0x45, 0x4c, 0x5f, 0x55, 0x50, 0x44, 0x41, 0x54, 0x45, 0x1a, 0x0b, 0x0a, 0x09,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44, 0x12, 0xed, 0x0d, 0x0a, 0x05, 0x54, 0x4f,
	0x41, 0x53, 0x54, 0x10, 0x02, 0x1a, 0xe1, 0x0d, 0xa8, 0x45, 0x94, 0x03, 0xb2, 0x45, 0x0f, 0xe8,
	0xb5, 0x84, 0xe6, 0xba, 0x90, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xba, 0x45,
	0x05, 0x54, 0x4f, 0x41, 0x53, 0x54, 0xca, 0x45, 0x25, 0x0a, 0x12, 0x52, 0x45, 0x53, 0x4f, 0x55,
	0x52, 0x43, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x0f, 0xe8,
	0xb5, 0x84, 0xe6, 0xba, 0x90, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45,
	0x21, 0x0a, 0x0e, 0x52, 0x45, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45, 0x5f, 0x45, 0x58, 0x49, 0x53,
	0x54, 0x12, 0x0f, 0xe8, 0xb5, 0x84, 0xe6, 0xba, 0x90, 0xe5, 0xb7, 0xb2, 0xe5, 0xad, 0x98, 0xe5,
	0x9c, 0xa8, 0xca, 0x45, 0x21, 0x0a, 0x0e, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f,
	0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x0f, 0xe7, 0x94, 0xa8, 0xe6, 0x88, 0xb7, 0xe4, 0xb8, 0x8d,
	0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x24, 0x0a, 0x0e, 0x55, 0x53, 0x45, 0x52, 0x4e,
	0x41, 0x4d, 0x45, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x12, 0x12, 0xe7, 0x94, 0xa8, 0xe6, 0x88,
	0xb7, 0xe5, 0x90, 0x8d, 0xe5, 0xb7, 0xb2, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x2b,
	0x0a, 0x15, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f, 0x4e, 0x4f,
	0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x12, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe7,
	0xbb, 0x84, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x28, 0x0a, 0x12,
	0x44, 0x41, 0x54, 0x41, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45, 0x5f, 0x53, 0x59, 0x4e, 0x43, 0x49,
	0x4e, 0x47, 0x12, 0x12, 0xe6, 0x95, 0xb0, 0xe6, 0x8d, 0xae, 0xe6, 0xba, 0x90, 0xe5, 0x90, 0x8c,
	0xe6, 0xad, 0xa5, 0xe4, 0xb8, 0xad, 0xca, 0x45, 0x2e, 0x0a, 0x12, 0x55, 0x53, 0x45, 0x52, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x53, 0x55, 0x42, 0x53, 0x43, 0x52, 0x49, 0x42, 0x45, 0x12, 0x18, 0xe7,
	0x94, 0xa8, 0xe6, 0x88, 0xb7, 0xe6, 0x9c, 0xaa, 0xe8, 0xae, 0xa2, 0xe9, 0x98, 0x85, 0xe6, 0xad,
	0xa4, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xca, 0x45, 0x21, 0x0a, 0x0e, 0x54, 0x45, 0x41, 0x4d,
	0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x0f, 0xe5, 0x9b, 0xa2, 0xe9,
	0x98, 0x9f, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x33, 0x0a, 0x1a,
	0x55, 0x53, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x41, 0x4c, 0x4c, 0x4f, 0x57, 0x5f, 0x52,
	0x45, 0x4d, 0x4f, 0x56, 0x45, 0x5f, 0x53, 0x45, 0x4c, 0x46, 0x12, 0x15, 0xe4, 0xb8, 0x8d, 0xe5,
	0x85, 0x81, 0xe8, 0xae, 0xb8, 0xe7, 0xa7, 0xbb, 0xe9, 0x99, 0xa4, 0xe8, 0x87, 0xaa, 0xe5, 0xb7,
	0xb1, 0xca, 0x45, 0x3d, 0x0a, 0x1b, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x41,
	0x4c, 0x4c, 0x4f, 0x57, 0x5f, 0x52, 0x45, 0x4d, 0x4f, 0x56, 0x45, 0x5f, 0x41, 0x44, 0x4d, 0x49,
	0x4e, 0x12, 0x1e, 0xe4, 0xb8, 0x8d, 0xe5, 0x85, 0x81, 0xe8, 0xae, 0xb8, 0xe7, 0xa7, 0xbb, 0xe9,
	0x99, 0xa4, 0xe5, 0x9b, 0xa2, 0xe9, 0x98, 0x9f, 0xe7, 0xae, 0xa1, 0xe7, 0x90, 0x86, 0xe5, 0x91,
	0x98, 0xca, 0x45, 0x47, 0x0a, 0x1c, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x41,
	0x4c, 0x4c, 0x4f, 0x57, 0x5f, 0x4f, 0x50, 0x45, 0x52, 0x41, 0x54, 0x45, 0x5f, 0x41, 0x44, 0x4d,
	0x49, 0x4e, 0x12, 0x27, 0xe4, 0xb8, 0x8d, 0xe5, 0x85, 0x81, 0xe8, 0xae, 0xb8, 0xe6, 0x93, 0x8d,
	0xe4, 0xbd, 0x9c, 0xe8, 0x87, 0xaa, 0xe5, 0xb7, 0xb1, 0xe7, 0x9a, 0x84, 0xe7, 0xae, 0xa1, 0xe7,
	0x90, 0x86, 0xe5, 0x91, 0x98, 0xe8, 0xba, 0xab, 0xe4, 0xbb, 0xbd, 0xca, 0x45, 0x21, 0x0a, 0x0e,
	0x52, 0x4f, 0x4c, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x0f,
	0xe8, 0xa7, 0x92, 0xe8, 0x89, 0xb2, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca,
	0x45, 0x34, 0x0a, 0x1b, 0x54, 0x45, 0x4d, 0x50, 0x4c, 0x41, 0x54, 0x45, 0x5f, 0x53, 0x54, 0x52,
	0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12,
	0x15, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe6, 0xa8, 0xa1, 0xe6, 0x9d, 0xbf, 0xe4, 0xb8, 0x8d,
	0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x21, 0x0a, 0x0e, 0x55, 0x53, 0x45, 0x52, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x12, 0x0f, 0xe7, 0x94, 0xa8, 0xe6, 0x88,
	0xb7, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x2c, 0x0a, 0x13, 0x44,
	0x41, 0x53, 0x48, 0x42, 0x4f, 0x41, 0x52, 0x44, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55,
	0x4e, 0x44, 0x12, 0x15, 0xe5, 0x9b, 0xbe, 0xe8, 0xa1, 0xa8, 0xe5, 0xa4, 0xa7, 0xe7, 0x9b, 0x98,
	0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x31, 0x0a, 0x18, 0x52, 0x45,
	0x41, 0x4c, 0x54, 0x49, 0x4d, 0x45, 0x5f, 0x41, 0x4c, 0x41, 0x52, 0x4d, 0x5f, 0x4e, 0x4f, 0x54,
	0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x15, 0xe5, 0xae, 0x9e, 0xe6, 0x97, 0xb6, 0xe5, 0x91,
	0x8a, 0xe8, 0xad, 0xa6, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x30,
	0x0a, 0x17, 0x48, 0x49, 0x53, 0x54, 0x4f, 0x52, 0x59, 0x5f, 0x41, 0x4c, 0x41, 0x52, 0x4d, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x15, 0xe5, 0x8e, 0x86, 0xe5, 0x8f,
	0xb2, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8,
	0xca, 0x45, 0x50, 0x0a, 0x15, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45,
	0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x12, 0xe6, 0x95, 0xb0, 0xe6,
	0x8d, 0xae, 0xe6, 0xba, 0x90, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x23,
	0x0a, 0x0a, 0x64, 0x61, 0x74, 0x61, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x15, 0x44, 0x41,
	0x54, 0x41, 0x5f, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f,
	0x55, 0x4e, 0x44, 0xca, 0x45, 0x39, 0x0a, 0x0e, 0x44, 0x49, 0x43, 0x54, 0x5f, 0x4e, 0x4f, 0x54,
	0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x0f, 0xe5, 0xad, 0x97, 0xe5, 0x85, 0xb8, 0xe4, 0xb8,
	0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x16, 0x0a, 0x04, 0x64, 0x69, 0x63, 0x74, 0x12,
	0x0e, 0x44, 0x49, 0x43, 0x54, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0xca,
	0x45, 0x4e, 0x0a, 0x14, 0x41, 0x4c, 0x41, 0x52, 0x4d, 0x5f, 0x48, 0x4f, 0x4f, 0x4b, 0x5f, 0x4e,
	0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x13, 0xe5, 0x91, 0x8a, 0xe8, 0xad, 0xa6,
	0x68, 0x6f, 0x6f, 0x6b, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x21, 0x0a,
	0x09, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x48, 0x6f, 0x6f, 0x6b, 0x12, 0x14, 0x41, 0x4c, 0x45, 0x52,
	0x54, 0x5f, 0x48, 0x4f, 0x4f, 0x4b, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44,
	0xca, 0x45, 0x39, 0x0a, 0x0e, 0x4d, 0x45, 0x4e, 0x55, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f,
	0x55, 0x4e, 0x44, 0x12, 0x0f, 0xe8, 0x8f, 0x9c, 0xe5, 0x8d, 0x95, 0xe4, 0xb8, 0x8d, 0xe5, 0xad,
	0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x16, 0x0a, 0x04, 0x6d, 0x65, 0x6e, 0x75, 0x12, 0x0e, 0x4d, 0x45,
	0x4e, 0x55, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0xca, 0x45, 0x3f, 0x0a,
	0x10, 0x4d, 0x45, 0x54, 0x52, 0x49, 0x43, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e,
	0x44, 0x12, 0x0f, 0xe6, 0x8c, 0x87, 0xe6, 0xa0, 0x87, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5,
	0x9c, 0xa8, 0x1a, 0x1a, 0x0a, 0x06, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x10, 0x4d, 0x45,
	0x54, 0x52, 0x49, 0x43, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0xca, 0x45,
	0x33, 0x0a, 0x0d, 0x41, 0x50, 0x49, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44,
	0x12, 0x0c, 0x41, 0x50, 0x49, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x14,
	0x0a, 0x03, 0x61, 0x70, 0x69, 0x12, 0x0d, 0x41, 0x50, 0x49, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46,
	0x4f, 0x55, 0x4e, 0x44, 0xca, 0x45, 0x56, 0x0a, 0x12, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47,
	0x59, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x15, 0xe5, 0x91, 0x8a,
	0xe8, 0xad, 0xa6, 0xe7, 0xad, 0x96, 0xe7, 0x95, 0xa5, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5,
	0x9c, 0xa8, 0x1a, 0x29, 0x0a, 0x0d, 0x61, 0x6c, 0x65, 0x72, 0x74, 0x53, 0x74, 0x72, 0x61, 0x74,
	0x65, 0x67, 0x79, 0x12, 0x18, 0x41, 0x4c, 0x45, 0x52, 0x54, 0x5f, 0x53, 0x54, 0x52, 0x41, 0x54,
	0x45, 0x47, 0x59, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0xca, 0x45, 0x59,
	0x0a, 0x18, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50,
	0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x12, 0xe7, 0xad, 0x96, 0xe7,
	0x95, 0xa5, 0xe7, 0xbb, 0x84, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x29,
	0x0a, 0x0d, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12,
	0x18, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0xca, 0x45, 0x7a, 0x0a, 0x1a, 0x54, 0x45,
	0x41, 0x4d, 0x5f, 0x49, 0x4e, 0x56, 0x49, 0x54, 0x45, 0x5f, 0x41, 0x4c, 0x52, 0x45, 0x41, 0x44,
	0x59, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x53, 0x12, 0x32, 0x25, 0x73, 0x2c, 0xe9, 0x82, 0x80,
	0xe8, 0xaf, 0xb7, 0xe8, 0xae, 0xb0, 0xe5, 0xbd, 0x95, 0xe5, 0xb7, 0xb2, 0xe5, 0xad, 0x98, 0xe5,
	0x9c, 0xa8, 0x2c, 0xe6, 0x88, 0x96, 0xe8, 0x80, 0x85, 0xe5, 0xb7, 0xb2, 0xe7, 0xbb, 0x8f, 0xe5,
	0x8a, 0xa0, 0xe5, 0x85, 0xa5, 0xe5, 0x9b, 0xa2, 0xe9, 0x98, 0x9f, 0x21, 0x1a, 0x28, 0x0a, 0x0a,
	0x74, 0x65, 0x61, 0x6d, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x12, 0x1a, 0x54, 0x45, 0x41, 0x4d,
	0x5f, 0x49, 0x4e, 0x56, 0x49, 0x54, 0x45, 0x5f, 0x41, 0x4c, 0x52, 0x45, 0x41, 0x44, 0x59, 0x5f,
	0x45, 0x58, 0x49, 0x53, 0x54, 0x53, 0xca, 0x45, 0x53, 0x0a, 0x15, 0x54, 0x45, 0x41, 0x4d, 0x5f,
	0x49, 0x4e, 0x56, 0x49, 0x54, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44,
	0x12, 0x15, 0xe9, 0x82, 0x80, 0xe8, 0xaf, 0xb7, 0xe8, 0xae, 0xb0, 0xe5, 0xbd, 0x95, 0xe4, 0xb8,
	0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x1a, 0x23, 0x0a, 0x0a, 0x74, 0x65, 0x61, 0x6d, 0x49,
	0x6e, 0x76, 0x69, 0x74, 0x65, 0x12, 0x15, 0x54, 0x45, 0x41, 0x4d, 0x5f, 0x49, 0x4e, 0x56, 0x49,
	0x54, 0x45, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0xca, 0x45, 0x2d, 0x0a,
	0x11, 0x52, 0x4f, 0x4c, 0x45, 0x5f, 0x48, 0x41, 0x53, 0x5f, 0x52, 0x45, 0x4c, 0x41, 0x54, 0x49,
	0x4f, 0x4e, 0x12, 0x18, 0xe8, 0xa7, 0x92, 0xe8, 0x89, 0xb2, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8,
	0xe5, 0x85, 0xb3, 0xe8, 0x81, 0x94, 0xe5, 0x85, 0xb3, 0xe7, 0xb3, 0xbb, 0xca, 0x45, 0x2e, 0x0a,
	0x15, 0x54, 0x45, 0x41, 0x4d, 0x5f, 0x4d, 0x45, 0x4d, 0x42, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54,
	0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x15, 0xe5, 0x9b, 0xa2, 0xe9, 0x98, 0x9f, 0xe6, 0x88,
	0x90, 0xe5, 0x91, 0x98, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x29,
	0x0a, 0x16, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45, 0x5f, 0x4e,
	0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x0f, 0xe6, 0xb6, 0x88, 0xe6, 0x81, 0xaf,
	0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0x12, 0xf3, 0x01, 0x0a, 0x0c, 0x4e, 0x4f,
	0x54, 0x49, 0x46, 0x49, 0x43, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x03, 0x1a, 0xe0, 0x01, 0xa8,
	0x45, 0xf4, 0x03, 0xb2, 0x45, 0x30, 0xe6, 0x9c, 0x8d, 0xe5, 0x8a, 0xa1, 0xe5, 0x99, 0xa8, 0xe5,
	0x8f, 0xaf, 0xe8, 0x83, 0xbd, 0xe9, 0x81, 0x87, 0xe5, 0x88, 0xb0, 0xe4, 0xba, 0x86, 0xe6, 0x84,
	0x8f, 0xe5, 0xa4, 0x96, 0xef, 0xbc, 0x8c, 0xe9, 0x9d, 0x9e, 0xe5, 0xb8, 0xb8, 0xe6, 0x8a, 0xb1,
	0xe6, 0xad, 0x89, 0xef, 0xbc, 0x81, 0xba, 0x45, 0x0c, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x49, 0x43,
	0x41, 0x54, 0x49, 0x4f, 0x4e, 0xca, 0x45, 0x5e, 0x0a, 0x0c, 0x53, 0x59, 0x53, 0x54, 0x45, 0x4d,
	0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x12, 0x4e, 0xe6, 0x9c, 0x8d, 0xe5, 0x8a, 0xa1, 0xe5, 0x99,
	0xa8, 0xe9, 0x81, 0xad, 0xe9, 0x81, 0x87, 0xe4, 0xba, 0x86, 0xe5, 0xa4, 0x96, 0xe6, 0x98, 0x9f,
	0xe4, 0xba, 0xba, 0xe6, 0x94, 0xbb, 0xe5, 0x87, 0xbb, 0xef, 0xbc, 0x8c, 0xe6, 0x94, 0xbb, 0xe5,
	0x9f, 0x8e, 0xe7, 0x8b, 0xae, 0xe5, 0x92, 0x8c, 0xe7, 0xa8, 0x8b, 0xe5, 0xba, 0x8f, 0xe7, 0x8c,
	0xbf, 0xe4, 0xbb, 0xac, 0xe6, 0xad, 0xa3, 0xe5, 0x9c, 0xa8, 0xe6, 0x8a, 0xa2, 0xe4, 0xbf, 0xae,
	0x2e, 0x2e, 0x2e, 0x2e, 0x2e, 0x2e, 0xca, 0x45, 0x36, 0x0a, 0x17, 0x55, 0x4e, 0x53, 0x55, 0x50,
	0x50, 0x4f, 0x52, 0x54, 0x45, 0x44, 0x5f, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x53, 0x4f, 0x55, 0x52,
	0x43, 0x45, 0x12, 0x1b, 0xe4, 0xb8, 0x8d, 0xe6, 0x94, 0xaf, 0xe6, 0x8c, 0x81, 0xe7, 0x9a, 0x84,
	0xe6, 0x95, 0xb0, 0xe6, 0x8d, 0xae, 0xe6, 0xba, 0x90, 0xe7, 0xb1, 0xbb, 0xe5, 0x9e, 0x8b, 0x12,
	0xbc, 0x02, 0x0a, 0x0c, 0x55, 0x4e, 0x41, 0x55, 0x54, 0x48, 0x4f, 0x52, 0x49, 0x5a, 0x45, 0x44,
	0x10, 0x04, 0x1a, 0xa9, 0x02, 0xa8, 0x45, 0x91, 0x03, 0xb2, 0x45, 0x0c, 0xe8, 0xaf, 0xb7, 0xe5,
	0x85, 0x88, 0xe7, 0x99, 0xbb, 0xe5, 0xbd, 0x95, 0xba, 0x45, 0x0c, 0x55, 0x4e, 0x41, 0x55, 0x54,
	0x48, 0x4f, 0x52, 0x49, 0x5a, 0x45, 0x44, 0xc2, 0x45, 0x12, 0x0a, 0x08, 0x72, 0x65, 0x64, 0x69,
	0x72, 0x65, 0x63, 0x74, 0x12, 0x06, 0x2f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0xca, 0x45, 0x1d, 0x0a,
	0x0a, 0x4a, 0x57, 0x54, 0x5f, 0x45, 0x58, 0x50, 0x49, 0x52, 0x45, 0x12, 0x0f, 0xe7, 0x99, 0xbb,
	0xe5, 0xbd, 0x95, 0xe5, 0xb7, 0xb2, 0xe8, 0xbf, 0x87, 0xe6, 0x9c, 0x9f, 0xca, 0x45, 0x31, 0x0a,
	0x0f, 0x4a, 0x57, 0x54, 0x5f, 0x4f, 0x54, 0x48, 0x45, 0x52, 0x5f, 0x4c, 0x4f, 0x47, 0x49, 0x4e,
	0x12, 0x1e, 0xe8, 0xb4, 0xa6, 0xe5, 0x8f, 0xb7, 0xe5, 0xb7, 0xb2, 0xe5, 0x9c, 0xa8, 0xe5, 0x85,
	0xb6, 0xe4, 0xbb, 0x96, 0xe5, 0x9c, 0xb0, 0xe6, 0x96, 0xb9, 0xe7, 0x99, 0xbb, 0xe5, 0xbd, 0x95,
	0xca, 0x45, 0x32, 0x0a, 0x07, 0x4a, 0x57, 0x54, 0x5f, 0x42, 0x41, 0x4e, 0x12, 0x27, 0xe8, 0xae,
	0xa4, 0xe8, 0xaf, 0x81, 0xe4, 0xbf, 0xa1, 0xe6, 0x81, 0xaf, 0xe5, 0xb7, 0xb2, 0xe7, 0x99, 0xbb,
	0xe5, 0x87, 0xba, 0xef, 0xbc, 0x8c, 0xe8, 0xaf, 0xb7, 0xe9, 0x87, 0x8d, 0xe6, 0x96, 0xb0, 0xe7,
	0x99, 0xbb, 0xe5, 0xbd, 0x95, 0xca, 0x45, 0x21, 0x0a, 0x0e, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4e,
	0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x12, 0x0f, 0xe8, 0xb4, 0xa6, 0xe5, 0x8f, 0xb7,
	0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98, 0xe5, 0x9c, 0xa8, 0xca, 0x45, 0x42, 0x0a, 0x08, 0x55, 0x53,
	0x45, 0x52, 0x5f, 0x42, 0x41, 0x4e, 0x12, 0x36, 0xe6, 0x82, 0xa8, 0xe5, 0xb7, 0xb2, 0xe8, 0xa2,
	0xab, 0xe7, 0xa6, 0x81, 0xe6, 0xad, 0xa2, 0xe4, 0xbd, 0xbf, 0xe7, 0x94, 0xa8, 0xe8, 0xaf, 0xa5,
	0xe7, 0xb3, 0xbb, 0xe7, 0xbb, 0x9f, 0xef, 0xbc, 0x8c, 0xe8, 0xaf, 0xb7, 0xe8, 0x81, 0x94, 0xe7,
	0xb3, 0xbb, 0xe5, 0xae, 0x98, 0xe6, 0x96, 0xb9, 0xe8, 0xa7, 0xa3, 0xe9, 0x99, 0xa4, 0x12, 0xd3,
	0x01, 0x0a, 0x09, 0x46, 0x4f, 0x52, 0x42, 0x49, 0x44, 0x44, 0x45, 0x4e, 0x10, 0x05, 0x1a, 0xc3,
	0x01, 0xa8, 0x45, 0x93, 0x03, 0xb2, 0x45, 0x38, 0xe6, 0x82, 0xa8, 0xe6, 0xb2, 0xa1, 0xe6, 0x9c,
	0x89, 0xe6, 0x93, 0x8d, 0xe4, 0xbd, 0x9c, 0xe6, 0x9d, 0x83, 0xe9, 0x99, 0x90, 0x2c, 0x20, 0xe8,
	0xaf, 0xb7, 0xe8, 0x81, 0x94, 0xe7, 0xb3, 0xbb, 0xe7, 0xae, 0xa1, 0xe7, 0x90, 0x86, 0xe5, 0x91,
	0x98, 0xe5, 0xbc, 0x80, 0xe9, 0x80, 0x9a, 0xe8, 0xaf, 0xa5, 0xe6, 0x9d, 0x83, 0xe9, 0x99, 0x90,
	0xba, 0x45, 0x09, 0x46, 0x4f, 0x52, 0x42, 0x49, 0x44, 0x44, 0x45, 0x4e, 0xca, 0x45, 0x2c, 0x0a,
	0x10, 0x55, 0x53, 0x45, 0x52, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x49, 0x4e, 0x5f, 0x54, 0x45, 0x41,
	0x4d, 0x12, 0x18, 0xe6, 0x82, 0xa8, 0xe5, 0xb7, 0xb2, 0xe4, 0xb8, 0x8d, 0xe5, 0xb1, 0x9e, 0xe4,
	0xba, 0x8e, 0xe8, 0xaf, 0xa5, 0xe5, 0x9b, 0xa2, 0xe9, 0x98, 0x9f, 0xca, 0x45, 0x46, 0x0a, 0x0f,
	0x4d, 0x45, 0x4d, 0x42, 0x45, 0x52, 0x5f, 0x44, 0x49, 0x53, 0x41, 0x42, 0x4c, 0x45, 0x44, 0x12,
	0x33, 0xe6, 0x82, 0xa8, 0xe5, 0xb7, 0xb2, 0xe8, 0xa2, 0xab, 0xe8, 0xaf, 0xa5, 0xe5, 0x9b, 0xa2,
	0xe9, 0x98, 0x9f, 0xe7, 0xa6, 0x81, 0xe7, 0x94, 0xa8, 0xe6, 0x93, 0x8d, 0xe4, 0xbd, 0x9c, 0xef,
	0xbc, 0x8c, 0xe8, 0xaf, 0xb7, 0xe8, 0x81, 0x94, 0xe7, 0xb3, 0xbb, 0xe7, 0xae, 0xa1, 0xe7, 0x90,
	0x86, 0xe5, 0x91, 0x98, 0x12, 0x53, 0x0a, 0x11, 0x54, 0x4f, 0x4f, 0x5f, 0x4d, 0x41, 0x4e, 0x59,
	0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x53, 0x10, 0x06, 0x1a, 0x3c, 0xa8, 0x45, 0xad,
	0x03, 0xb2, 0x45, 0x21, 0xe8, 0xaf, 0xb7, 0xe6, 0xb1, 0x82, 0xe5, 0xa4, 0xaa, 0xe9, 0xa2, 0x91,
	0xe7, 0xb9, 0x81, 0xef, 0xbc, 0x8c, 0xe8, 0xaf, 0xb7, 0xe7, 0xa8, 0x8d, 0xe5, 0x90, 0x8e, 0xe5,
	0x86, 0x8d, 0xe8, 0xaf, 0x95, 0xba, 0x45, 0x11, 0x54, 0x4f, 0x4f, 0x5f, 0x4d, 0x41, 0x4e, 0x59,
	0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x53, 0x12, 0xba, 0x02, 0x0a, 0x0c, 0x46, 0x49,
	0x4c, 0x45, 0x5f, 0x52, 0x45, 0x4c, 0x41, 0x54, 0x45, 0x44, 0x10, 0x07, 0x1a, 0xa7, 0x02, 0xa8,
	0x45, 0xae, 0x03, 0xb2, 0x45, 0x12, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0xe7, 0x9b, 0xb8, 0xe5,
	0x85, 0xb3, 0xe5, 0xbc, 0x82, 0xe5, 0xb8, 0xb8, 0xba, 0x45, 0x0c, 0x46, 0x49, 0x4c, 0x45, 0x5f,
	0x52, 0x45, 0x4c, 0x41, 0x54, 0x45, 0x44, 0xca, 0x45, 0x3c, 0x0a, 0x1b, 0x46, 0x49, 0x4c, 0x45,
	0x5f, 0x43, 0x4f, 0x4e, 0x54, 0x45, 0x4e, 0x54, 0x5f, 0x44, 0x4f, 0x45, 0x53, 0x5f, 0x4e, 0x4f,
	0x54, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x12, 0x1d, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0x3a,
	0x5b, 0x25, 0x73, 0x5d, 0xe5, 0x86, 0x85, 0xe5, 0xae, 0xb9, 0xe4, 0xb8, 0x8d, 0xe5, 0xad, 0x98,
	0xe5, 0x9c, 0xa8, 0xef, 0xbc, 0x81, 0xca, 0x45, 0x40, 0x0a, 0x19, 0x46, 0x49, 0x4c, 0x45, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x53, 0x55, 0x50, 0x50, 0x4f, 0x52, 0x54, 0x45, 0x44, 0x5f, 0x55, 0x50,
	0x4c, 0x4f, 0x41, 0x44, 0x12, 0x23, 0xe4, 0xb8, 0x8d, 0xe6, 0x94, 0xaf, 0xe6, 0x8c, 0x81, 0xe8,
	0xaf, 0xa5, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0xe7, 0xb1, 0xbb, 0xe5, 0x9e, 0x8b, 0xef, 0xbc,
	0x9a, 0x25, 0x73, 0xe4, 0xb8, 0x8a, 0xe4, 0xbc, 0xa0, 0xca, 0x45, 0x42, 0x0a, 0x12, 0x46, 0x49,
	0x4c, 0x45, 0x5f, 0x4d, 0x41, 0x58, 0x49, 0x4d, 0x55, 0x4d, 0x5f, 0x4c, 0x49, 0x4d, 0x49, 0x54,
	0x12, 0x2c, 0xe8, 0xaf, 0xa5, 0xe7, 0xb1, 0xbb, 0xe5, 0x9e, 0x8b, 0x5b, 0x25, 0x73, 0x5d, 0xe6,
	0x96, 0x87, 0xe4, 0xbb, 0xb6, 0xe5, 0xa4, 0xa7, 0xe5, 0xb0, 0x8f, 0xe8, 0xb6, 0x85, 0xe8, 0xbf,
	0x87, 0xe6, 0x9c, 0x80, 0xe5, 0xa4, 0xa7, 0xe9, 0x99, 0x90, 0xe5, 0x88, 0xb6, 0x21, 0xca, 0x45,
	0x35, 0x0a, 0x0e, 0x4f, 0x53, 0x53, 0x5f, 0x4e, 0x4f, 0x54, 0x5f, 0x4f, 0x50, 0x45, 0x4e, 0x45,
	0x44, 0x12, 0x23, 0x6f, 0x73, 0x73, 0xe6, 0x9c, 0xaa, 0xe6, 0x89, 0x93, 0xe5, 0xbc, 0x80, 0x2c,
	0xe4, 0xb8, 0x8d, 0xe5, 0x85, 0x81, 0xe8, 0xae, 0xb8, 0xe4, 0xb8, 0x8a, 0xe4, 0xbc, 0xa0, 0xe6,
	0x96, 0x87, 0xe4, 0xbb, 0xb6, 0x21, 0x1a, 0x04, 0xa0, 0x45, 0x93, 0x03, 0x42, 0x37, 0x0a, 0x08,
	0x70, 0x6b, 0x67, 0x2e, 0x6d, 0x65, 0x72, 0x72, 0x50, 0x01, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x69, 0x64, 0x65, 0x2d, 0x66, 0x61, 0x6d, 0x69,
	0x6c, 0x79, 0x2f, 0x6d, 0x6f, 0x6f, 0x6e, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x65, 0x72, 0x72,
	0x3b, 0x6d, 0x65, 0x72, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_merr_err_proto_rawDescOnce sync.Once
	file_merr_err_proto_rawDescData = file_merr_err_proto_rawDesc
)

func file_merr_err_proto_rawDescGZIP() []byte {
	file_merr_err_proto_rawDescOnce.Do(func() {
		file_merr_err_proto_rawDescData = protoimpl.X.CompressGZIP(file_merr_err_proto_rawDescData)
	})
	return file_merr_err_proto_rawDescData
}

var file_merr_err_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_merr_err_proto_goTypes = []any{
	(ErrorReason)(0), // 0: pkg.merr.ErrorReason
}
var file_merr_err_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_merr_err_proto_init() }
func file_merr_err_proto_init() {
	if File_merr_err_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_merr_err_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_merr_err_proto_goTypes,
		DependencyIndexes: file_merr_err_proto_depIdxs,
		EnumInfos:         file_merr_err_proto_enumTypes,
	}.Build()
	File_merr_err_proto = out.File
	file_merr_err_proto_rawDesc = nil
	file_merr_err_proto_goTypes = nil
	file_merr_err_proto_depIdxs = nil
}
