// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.19.4
// source: prom/v1/file.proto

package v1

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	api "prometheus-manager/api"
	prom "prometheus-manager/api/prom"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	File *prom.FileItem `protobuf:"bytes,1,opt,name=file,proto3" json:"file,omitempty"`
}

func (x *CreateFileRequest) Reset() {
	*x = CreateFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateFileRequest) ProtoMessage() {}

func (x *CreateFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateFileRequest.ProtoReflect.Descriptor instead.
func (*CreateFileRequest) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{0}
}

func (x *CreateFileRequest) GetFile() *prom.FileItem {
	if x != nil {
		return x.File
	}
	return nil
}

type CreateFileReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response *api.Response `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *CreateFileReply) Reset() {
	*x = CreateFileReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateFileReply) ProtoMessage() {}

func (x *CreateFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateFileReply.ProtoReflect.Descriptor instead.
func (*CreateFileReply) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{1}
}

func (x *CreateFileReply) GetResponse() *api.Response {
	if x != nil {
		return x.Response
	}
	return nil
}

type UpdateFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   uint32         `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	File *prom.FileItem `protobuf:"bytes,2,opt,name=file,proto3" json:"file,omitempty"`
}

func (x *UpdateFileRequest) Reset() {
	*x = UpdateFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateFileRequest) ProtoMessage() {}

func (x *UpdateFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateFileRequest.ProtoReflect.Descriptor instead.
func (*UpdateFileRequest) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateFileRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateFileRequest) GetFile() *prom.FileItem {
	if x != nil {
		return x.File
	}
	return nil
}

type UpdateFileReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response *api.Response `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *UpdateFileReply) Reset() {
	*x = UpdateFileReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateFileReply) ProtoMessage() {}

func (x *UpdateFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateFileReply.ProtoReflect.Descriptor instead.
func (*UpdateFileReply) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{3}
}

func (x *UpdateFileReply) GetResponse() *api.Response {
	if x != nil {
		return x.Response
	}
	return nil
}

type DeleteFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteFileRequest) Reset() {
	*x = DeleteFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteFileRequest) ProtoMessage() {}

func (x *DeleteFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteFileRequest.ProtoReflect.Descriptor instead.
func (*DeleteFileRequest) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteFileRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteFileReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response *api.Response `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
}

func (x *DeleteFileReply) Reset() {
	*x = DeleteFileReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteFileReply) ProtoMessage() {}

func (x *DeleteFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteFileReply.ProtoReflect.Descriptor instead.
func (*DeleteFileReply) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{5}
}

func (x *DeleteFileReply) GetResponse() *api.Response {
	if x != nil {
		return x.Response
	}
	return nil
}

type GetFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetFileRequest) Reset() {
	*x = GetFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFileRequest) ProtoMessage() {}

func (x *GetFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFileRequest.ProtoReflect.Descriptor instead.
func (*GetFileRequest) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{6}
}

func (x *GetFileRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type GetFileReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response *api.Response  `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	File     *prom.FileItem `protobuf:"bytes,2,opt,name=file,proto3" json:"file,omitempty"`
}

func (x *GetFileReply) Reset() {
	*x = GetFileReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFileReply) ProtoMessage() {}

func (x *GetFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFileReply.ProtoReflect.Descriptor instead.
func (*GetFileReply) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{7}
}

func (x *GetFileReply) GetResponse() *api.Response {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *GetFileReply) GetFile() *prom.FileItem {
	if x != nil {
		return x.File
	}
	return nil
}

type ListFileRequestParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keyword string `protobuf:"bytes,1,opt,name=keyword,proto3" json:"keyword,omitempty"`
}

func (x *ListFileRequestParams) Reset() {
	*x = ListFileRequestParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFileRequestParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFileRequestParams) ProtoMessage() {}

func (x *ListFileRequestParams) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFileRequestParams.ProtoReflect.Descriptor instead.
func (*ListFileRequestParams) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{8}
}

func (x *ListFileRequestParams) GetKeyword() string {
	if x != nil {
		return x.Keyword
	}
	return ""
}

type ListFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page   *api.PageRequest       `protobuf:"bytes,1,opt,name=page,proto3" json:"page,omitempty"`
	Params *ListFileRequestParams `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
}

func (x *ListFileRequest) Reset() {
	*x = ListFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFileRequest) ProtoMessage() {}

func (x *ListFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFileRequest.ProtoReflect.Descriptor instead.
func (*ListFileRequest) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{9}
}

func (x *ListFileRequest) GetPage() *api.PageRequest {
	if x != nil {
		return x.Page
	}
	return nil
}

func (x *ListFileRequest) GetParams() *ListFileRequestParams {
	if x != nil {
		return x.Params
	}
	return nil
}

type ListFileReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response *api.Response    `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	Page     *api.PageReply   `protobuf:"bytes,2,opt,name=page,proto3" json:"page,omitempty"`
	List     []*prom.FileItem `protobuf:"bytes,3,rep,name=list,proto3" json:"list,omitempty"`
}

func (x *ListFileReply) Reset() {
	*x = ListFileReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_prom_v1_file_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListFileReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListFileReply) ProtoMessage() {}

func (x *ListFileReply) ProtoReflect() protoreflect.Message {
	mi := &file_prom_v1_file_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListFileReply.ProtoReflect.Descriptor instead.
func (*ListFileReply) Descriptor() ([]byte, []int) {
	return file_prom_v1_file_proto_rawDescGZIP(), []int{10}
}

func (x *ListFileReply) GetResponse() *api.Response {
	if x != nil {
		return x.Response
	}
	return nil
}

func (x *ListFileReply) GetPage() *api.PageReply {
	if x != nil {
		return x.Page
	}
	return nil
}

func (x *ListFileReply) GetList() []*prom.FileItem {
	if x != nil {
		return x.List
	}
	return nil
}

var File_prom_v1_file_proto protoreflect.FileDescriptor

var file_prom_v1_file_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x1a, 0x1c,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0a, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x0f, 0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x6e, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x45, 0x0a, 0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d,
	0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01,
	0x02, 0x10, 0x01, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x22, 0x3c, 0x0a, 0x0f, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x29, 0x0a, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x08, 0x72,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x5e, 0x0a, 0x11, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02, 0x20,
	0x00, 0x52, 0x02, 0x69, 0x64, 0x12, 0x30, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x46,
	0x69, 0x6c, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01, 0x02, 0x10,
	0x01, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x22, 0x3c, 0x0a, 0x0f, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x29, 0x0a, 0x08, 0x72, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2c, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46,
	0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02, 0x20, 0x00, 0x52,
	0x02, 0x69, 0x64, 0x22, 0x3c, 0x0a, 0x0f, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46, 0x69, 0x6c,
	0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x29, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x29, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42,
	0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02, 0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x22, 0x61, 0x0a, 0x0c,
	0x47, 0x65, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x29, 0x0a, 0x08,
	0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x08, 0x72,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x26, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d,
	0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x22,
	0x3c, 0x0a, 0x15, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x23, 0x0a, 0x07, 0x6b, 0x65, 0x79, 0x77,
	0x6f, 0x72, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x09, 0xfa, 0x42, 0x06, 0x72, 0x04,
	0x10, 0x00, 0x18, 0x40, 0x52, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x22, 0x70, 0x0a,
	0x0f, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x24, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x37, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f,
	0x6d, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x22,
	0x86, 0x01, 0x0a, 0x0d, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x29, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x22, 0x0a, 0x04,
	0x70, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x50, 0x61, 0x67, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65,
	0x12, 0x26, 0x0a, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x74,
	0x65, 0x6d, 0x52, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x32, 0xf5, 0x03, 0x0a, 0x04, 0x46, 0x69, 0x6c,
	0x65, 0x12, 0x62, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x12,
	0x1b, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x46, 0x69,
	0x6c, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x16, 0x22,
	0x11, 0x2f, 0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x2f, 0x61,
	0x64, 0x64, 0x3a, 0x01, 0x2a, 0x12, 0x68, 0x0a, 0x0a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x46,
	0x69, 0x6c, 0x65, 0x12, 0x1b, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x19, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x22, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x1c, 0x1a, 0x17, 0x2f, 0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x69,
	0x6c, 0x65, 0x2f, 0x65, 0x64, 0x69, 0x74, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x3a, 0x01, 0x2a, 0x12,
	0x67, 0x0a, 0x0a, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x1b, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46,
	0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x46, 0x69, 0x6c, 0x65,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x21, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1b, 0x2a, 0x19, 0x2f,
	0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x2f, 0x64, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x57, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x46,
	0x69, 0x6c, 0x65, 0x12, 0x18, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x47,
	0x65, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x47, 0x65, 0x74, 0x46, 0x69, 0x6c, 0x65,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x1a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x12, 0x12, 0x2f,
	0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x2f, 0x7b, 0x69, 0x64,
	0x7d, 0x12, 0x5d, 0x0a, 0x08, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x19, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70,
	0x72, 0x6f, 0x6d, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x22, 0x1d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x22, 0x12, 0x2f, 0x70, 0x72, 0x6f, 0x6d,
	0x2f, 0x76, 0x31, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x2f, 0x6c, 0x69, 0x73, 0x74, 0x3a, 0x01, 0x2a,
	0x42, 0x32, 0x0a, 0x0b, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x76, 0x31, 0x50,
	0x01, 0x5a, 0x21, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2d, 0x6d, 0x61,
	0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x76,
	0x31, 0x3b, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_prom_v1_file_proto_rawDescOnce sync.Once
	file_prom_v1_file_proto_rawDescData = file_prom_v1_file_proto_rawDesc
)

func file_prom_v1_file_proto_rawDescGZIP() []byte {
	file_prom_v1_file_proto_rawDescOnce.Do(func() {
		file_prom_v1_file_proto_rawDescData = protoimpl.X.CompressGZIP(file_prom_v1_file_proto_rawDescData)
	})
	return file_prom_v1_file_proto_rawDescData
}

var file_prom_v1_file_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_prom_v1_file_proto_goTypes = []interface{}{
	(*CreateFileRequest)(nil),     // 0: api.prom.CreateFileRequest
	(*CreateFileReply)(nil),       // 1: api.prom.CreateFileReply
	(*UpdateFileRequest)(nil),     // 2: api.prom.UpdateFileRequest
	(*UpdateFileReply)(nil),       // 3: api.prom.UpdateFileReply
	(*DeleteFileRequest)(nil),     // 4: api.prom.DeleteFileRequest
	(*DeleteFileReply)(nil),       // 5: api.prom.DeleteFileReply
	(*GetFileRequest)(nil),        // 6: api.prom.GetFileRequest
	(*GetFileReply)(nil),          // 7: api.prom.GetFileReply
	(*ListFileRequestParams)(nil), // 8: api.prom.ListFileRequestParams
	(*ListFileRequest)(nil),       // 9: api.prom.ListFileRequest
	(*ListFileReply)(nil),         // 10: api.prom.ListFileReply
	(*prom.FileItem)(nil),         // 11: api.prom.FileItem
	(*api.Response)(nil),          // 12: api.Response
	(*api.PageRequest)(nil),       // 13: api.PageRequest
	(*api.PageReply)(nil),         // 14: api.PageReply
}
var file_prom_v1_file_proto_depIdxs = []int32{
	11, // 0: api.prom.CreateFileRequest.file:type_name -> api.prom.FileItem
	12, // 1: api.prom.CreateFileReply.response:type_name -> api.Response
	11, // 2: api.prom.UpdateFileRequest.file:type_name -> api.prom.FileItem
	12, // 3: api.prom.UpdateFileReply.response:type_name -> api.Response
	12, // 4: api.prom.DeleteFileReply.response:type_name -> api.Response
	12, // 5: api.prom.GetFileReply.response:type_name -> api.Response
	11, // 6: api.prom.GetFileReply.file:type_name -> api.prom.FileItem
	13, // 7: api.prom.ListFileRequest.page:type_name -> api.PageRequest
	8,  // 8: api.prom.ListFileRequest.params:type_name -> api.prom.ListFileRequestParams
	12, // 9: api.prom.ListFileReply.response:type_name -> api.Response
	14, // 10: api.prom.ListFileReply.page:type_name -> api.PageReply
	11, // 11: api.prom.ListFileReply.list:type_name -> api.prom.FileItem
	0,  // 12: api.prom.File.CreateFile:input_type -> api.prom.CreateFileRequest
	2,  // 13: api.prom.File.UpdateFile:input_type -> api.prom.UpdateFileRequest
	4,  // 14: api.prom.File.DeleteFile:input_type -> api.prom.DeleteFileRequest
	6,  // 15: api.prom.File.GetFile:input_type -> api.prom.GetFileRequest
	9,  // 16: api.prom.File.ListFile:input_type -> api.prom.ListFileRequest
	1,  // 17: api.prom.File.CreateFile:output_type -> api.prom.CreateFileReply
	3,  // 18: api.prom.File.UpdateFile:output_type -> api.prom.UpdateFileReply
	5,  // 19: api.prom.File.DeleteFile:output_type -> api.prom.DeleteFileReply
	7,  // 20: api.prom.File.GetFile:output_type -> api.prom.GetFileReply
	10, // 21: api.prom.File.ListFile:output_type -> api.prom.ListFileReply
	17, // [17:22] is the sub-list for method output_type
	12, // [12:17] is the sub-list for method input_type
	12, // [12:12] is the sub-list for extension type_name
	12, // [12:12] is the sub-list for extension extendee
	0,  // [0:12] is the sub-list for field type_name
}

func init() { file_prom_v1_file_proto_init() }
func file_prom_v1_file_proto_init() {
	if File_prom_v1_file_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_prom_v1_file_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateFileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateFileReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateFileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateFileReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteFileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteFileReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetFileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetFileReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFileRequestParams); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_prom_v1_file_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListFileReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_prom_v1_file_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_prom_v1_file_proto_goTypes,
		DependencyIndexes: file_prom_v1_file_proto_depIdxs,
		MessageInfos:      file_prom_v1_file_proto_msgTypes,
	}.Build()
	File_prom_v1_file_proto = out.File
	file_prom_v1_file_proto_rawDesc = nil
	file_prom_v1_file_proto_goTypes = nil
	file_prom_v1_file_proto_depIdxs = nil
}
