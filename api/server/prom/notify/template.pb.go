// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: server/prom/notify/template.proto

package notify

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	api "prometheus-manager/api"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 创建模板请求参数
type CreateTemplateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 模板内容
	Content string `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	// 所属策略
	StrategyId uint32 `protobuf:"varint,3,opt,name=strategyId,proto3" json:"strategyId,omitempty"`
	// 模板类型
	NotifyType api.NotifyTemplateType `protobuf:"varint,4,opt,name=notifyType,proto3,enum=api.NotifyTemplateType" json:"notifyType,omitempty"`
}

func (x *CreateTemplateRequest) Reset() {
	*x = CreateTemplateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateTemplateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateTemplateRequest) ProtoMessage() {}

func (x *CreateTemplateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateTemplateRequest.ProtoReflect.Descriptor instead.
func (*CreateTemplateRequest) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{0}
}

func (x *CreateTemplateRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *CreateTemplateRequest) GetStrategyId() uint32 {
	if x != nil {
		return x.StrategyId
	}
	return 0
}

func (x *CreateTemplateRequest) GetNotifyType() api.NotifyTemplateType {
	if x != nil {
		return x.NotifyType
	}
	return api.NotifyTemplateType(0)
}

type CreateTemplateReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CreateTemplateReply) Reset() {
	*x = CreateTemplateReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateTemplateReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateTemplateReply) ProtoMessage() {}

func (x *CreateTemplateReply) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateTemplateReply.ProtoReflect.Descriptor instead.
func (*CreateTemplateReply) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{1}
}

// 更新模板请求参数
type UpdateTemplateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// 模板内容
	Content string `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	// 所属策略
	StrategyId uint32 `protobuf:"varint,3,opt,name=strategyId,proto3" json:"strategyId,omitempty"`
	// 模板类型
	NotifyType api.NotifyTemplateType `protobuf:"varint,4,opt,name=notifyType,proto3,enum=api.NotifyTemplateType" json:"notifyType,omitempty"`
}

func (x *UpdateTemplateRequest) Reset() {
	*x = UpdateTemplateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateTemplateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTemplateRequest) ProtoMessage() {}

func (x *UpdateTemplateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTemplateRequest.ProtoReflect.Descriptor instead.
func (*UpdateTemplateRequest) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateTemplateRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateTemplateRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *UpdateTemplateRequest) GetStrategyId() uint32 {
	if x != nil {
		return x.StrategyId
	}
	return 0
}

func (x *UpdateTemplateRequest) GetNotifyType() api.NotifyTemplateType {
	if x != nil {
		return x.NotifyType
	}
	return api.NotifyTemplateType(0)
}

type UpdateTemplateReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UpdateTemplateReply) Reset() {
	*x = UpdateTemplateReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateTemplateReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTemplateReply) ProtoMessage() {}

func (x *UpdateTemplateReply) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTemplateReply.ProtoReflect.Descriptor instead.
func (*UpdateTemplateReply) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{3}
}

// 删除模板请求参数
type DeleteTemplateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteTemplateRequest) Reset() {
	*x = DeleteTemplateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteTemplateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteTemplateRequest) ProtoMessage() {}

func (x *DeleteTemplateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteTemplateRequest.ProtoReflect.Descriptor instead.
func (*DeleteTemplateRequest) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteTemplateRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteTemplateReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteTemplateReply) Reset() {
	*x = DeleteTemplateReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteTemplateReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteTemplateReply) ProtoMessage() {}

func (x *DeleteTemplateReply) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteTemplateReply.ProtoReflect.Descriptor instead.
func (*DeleteTemplateReply) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{5}
}

// 获取模板请求参数
type GetTemplateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetTemplateRequest) Reset() {
	*x = GetTemplateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTemplateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTemplateRequest) ProtoMessage() {}

func (x *GetTemplateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTemplateRequest.ProtoReflect.Descriptor instead.
func (*GetTemplateRequest) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{6}
}

func (x *GetTemplateRequest) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type GetTemplateReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Detail *api.NotifyTemplateItem `protobuf:"bytes,1,opt,name=detail,proto3" json:"detail,omitempty"`
}

func (x *GetTemplateReply) Reset() {
	*x = GetTemplateReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTemplateReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTemplateReply) ProtoMessage() {}

func (x *GetTemplateReply) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTemplateReply.ProtoReflect.Descriptor instead.
func (*GetTemplateReply) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{7}
}

func (x *GetTemplateReply) GetDetail() *api.NotifyTemplateItem {
	if x != nil {
		return x.Detail
	}
	return nil
}

// 获取模板列表请求参数
type ListTemplateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page       *api.PageRequest `protobuf:"bytes,1,opt,name=page,proto3" json:"page,omitempty"`
	StrategyId uint32           `protobuf:"varint,2,opt,name=strategyId,proto3" json:"strategyId,omitempty"`
}

func (x *ListTemplateRequest) Reset() {
	*x = ListTemplateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListTemplateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListTemplateRequest) ProtoMessage() {}

func (x *ListTemplateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListTemplateRequest.ProtoReflect.Descriptor instead.
func (*ListTemplateRequest) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{8}
}

func (x *ListTemplateRequest) GetPage() *api.PageRequest {
	if x != nil {
		return x.Page
	}
	return nil
}

func (x *ListTemplateRequest) GetStrategyId() uint32 {
	if x != nil {
		return x.StrategyId
	}
	return 0
}

type ListTemplateReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Page *api.PageReply            `protobuf:"bytes,1,opt,name=page,proto3" json:"page,omitempty"`
	List []*api.NotifyTemplateItem `protobuf:"bytes,2,rep,name=list,proto3" json:"list,omitempty"`
}

func (x *ListTemplateReply) Reset() {
	*x = ListTemplateReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_prom_notify_template_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListTemplateReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListTemplateReply) ProtoMessage() {}

func (x *ListTemplateReply) ProtoReflect() protoreflect.Message {
	mi := &file_server_prom_notify_template_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListTemplateReply.ProtoReflect.Descriptor instead.
func (*ListTemplateReply) Descriptor() ([]byte, []int) {
	return file_server_prom_notify_template_proto_rawDescGZIP(), []int{9}
}

func (x *ListTemplateReply) GetPage() *api.PageReply {
	if x != nil {
		return x.Page
	}
	return nil
}

func (x *ListTemplateReply) GetList() []*api.NotifyTemplateItem {
	if x != nil {
		return x.List
	}
	return nil
}

var File_server_prom_notify_template_proto protoreflect.FileDescriptor

var file_server_prom_notify_template_proto_rawDesc = []byte{
	0x0a, 0x21, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x6d, 0x2f, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x16, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x1a, 0x1c, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0a, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa6, 0x01, 0x0a, 0x15,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x27, 0x0a, 0x0a, 0x73, 0x74, 0x72, 0x61,
	0x74, 0x65, 0x67, 0x79, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42,
	0x04, 0x2a, 0x02, 0x20, 0x00, 0x52, 0x0a, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x49,
	0x64, 0x12, 0x41, 0x0a, 0x0a, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x79, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x42, 0x08,
	0xfa, 0x42, 0x05, 0x82, 0x01, 0x02, 0x10, 0x01, 0x52, 0x0a, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79,
	0x54, 0x79, 0x70, 0x65, 0x22, 0x15, 0x0a, 0x13, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x65,
	0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0xbf, 0x01, 0x0a, 0x15,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02, 0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x12, 0x21,
	0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x12, 0x27, 0x0a, 0x0a, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x49, 0x64, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02, 0x20, 0x00, 0x52, 0x0a,
	0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x49, 0x64, 0x12, 0x41, 0x0a, 0x0a, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x82, 0x01, 0x02, 0x10,
	0x01, 0x52, 0x0a, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x22, 0x15, 0x0a,
	0x13, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x30, 0x0a, 0x15, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x65,
	0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02,
	0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x22, 0x15, 0x0a, 0x13, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x2d, 0x0a,
	0x12, 0x47, 0x65, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42,
	0x07, 0xfa, 0x42, 0x04, 0x2a, 0x02, 0x20, 0x00, 0x52, 0x02, 0x69, 0x64, 0x22, 0x43, 0x0a, 0x10,
	0x47, 0x65, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x12, 0x2f, 0x0a, 0x06, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x65, 0x6d,
	0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x06, 0x64, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x22, 0x6e, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x61, 0x67,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01, 0x02,
	0x10, 0x01, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x27, 0x0a, 0x0a, 0x73, 0x74, 0x72, 0x61,
	0x74, 0x65, 0x67, 0x79, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x07, 0xfa, 0x42,
	0x04, 0x2a, 0x02, 0x20, 0x00, 0x52, 0x0a, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67, 0x79, 0x49,
	0x64, 0x22, 0x64, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x22, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x2b, 0x0a, 0x04, 0x6c, 0x69,
	0x73, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x49, 0x74, 0x65,
	0x6d, 0x52, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x32, 0xf8, 0x05, 0x0a, 0x08, 0x54, 0x65, 0x6d, 0x70,
	0x6c, 0x61, 0x74, 0x65, 0x12, 0x97, 0x01, 0x0a, 0x0e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54,
	0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12, 0x2d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x22, 0x29, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x23, 0x22, 0x1e, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2f, 0x74, 0x65, 0x6d, 0x70,
	0x6c, 0x61, 0x74, 0x65, 0x2f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x3a, 0x01, 0x2a, 0x12, 0x97,
	0x01, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74,
	0x65, 0x12, 0x2d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x2b, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x29, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x23, 0x22, 0x1e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x75,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x3a, 0x01, 0x2a, 0x12, 0x97, 0x01, 0x0a, 0x0e, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12, 0x2d, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2b, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x79, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x29, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x23, 0x22,
	0x1e, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2f,
	0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x3a,
	0x01, 0x2a, 0x12, 0x8b, 0x01, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61,
	0x74, 0x65, 0x12, 0x2a, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x47, 0x65, 0x74, 0x54,
	0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x6d,
	0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x6c,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x26, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x20,
	0x22, 0x1b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79,
	0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x67, 0x65, 0x74, 0x3a, 0x01, 0x2a,
	0x12, 0x8f, 0x01, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74,
	0x65, 0x12, 0x2b, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x54,
	0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x6d,
	0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x54, 0x65, 0x6d, 0x70,
	0x6c, 0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x27, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x21, 0x22, 0x1c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x2f, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x6c, 0x69, 0x73, 0x74, 0x3a,
	0x01, 0x2a, 0x42, 0x4c, 0x0a, 0x16, 0x61, 0x70, 0x69, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x50, 0x01, 0x5a, 0x30,
	0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2d, 0x6d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x70, 0x72,
	0x6f, 0x6d, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x3b, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_server_prom_notify_template_proto_rawDescOnce sync.Once
	file_server_prom_notify_template_proto_rawDescData = file_server_prom_notify_template_proto_rawDesc
)

func file_server_prom_notify_template_proto_rawDescGZIP() []byte {
	file_server_prom_notify_template_proto_rawDescOnce.Do(func() {
		file_server_prom_notify_template_proto_rawDescData = protoimpl.X.CompressGZIP(file_server_prom_notify_template_proto_rawDescData)
	})
	return file_server_prom_notify_template_proto_rawDescData
}

var file_server_prom_notify_template_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_server_prom_notify_template_proto_goTypes = []interface{}{
	(*CreateTemplateRequest)(nil),  // 0: api.server.prom.notify.CreateTemplateRequest
	(*CreateTemplateReply)(nil),    // 1: api.server.prom.notify.CreateTemplateReply
	(*UpdateTemplateRequest)(nil),  // 2: api.server.prom.notify.UpdateTemplateRequest
	(*UpdateTemplateReply)(nil),    // 3: api.server.prom.notify.UpdateTemplateReply
	(*DeleteTemplateRequest)(nil),  // 4: api.server.prom.notify.DeleteTemplateRequest
	(*DeleteTemplateReply)(nil),    // 5: api.server.prom.notify.DeleteTemplateReply
	(*GetTemplateRequest)(nil),     // 6: api.server.prom.notify.GetTemplateRequest
	(*GetTemplateReply)(nil),       // 7: api.server.prom.notify.GetTemplateReply
	(*ListTemplateRequest)(nil),    // 8: api.server.prom.notify.ListTemplateRequest
	(*ListTemplateReply)(nil),      // 9: api.server.prom.notify.ListTemplateReply
	(api.NotifyTemplateType)(0),    // 10: api.NotifyTemplateType
	(*api.NotifyTemplateItem)(nil), // 11: api.NotifyTemplateItem
	(*api.PageRequest)(nil),        // 12: api.PageRequest
	(*api.PageReply)(nil),          // 13: api.PageReply
}
var file_server_prom_notify_template_proto_depIdxs = []int32{
	10, // 0: api.server.prom.notify.CreateTemplateRequest.notifyType:type_name -> api.NotifyTemplateType
	10, // 1: api.server.prom.notify.UpdateTemplateRequest.notifyType:type_name -> api.NotifyTemplateType
	11, // 2: api.server.prom.notify.GetTemplateReply.detail:type_name -> api.NotifyTemplateItem
	12, // 3: api.server.prom.notify.ListTemplateRequest.page:type_name -> api.PageRequest
	13, // 4: api.server.prom.notify.ListTemplateReply.page:type_name -> api.PageReply
	11, // 5: api.server.prom.notify.ListTemplateReply.list:type_name -> api.NotifyTemplateItem
	0,  // 6: api.server.prom.notify.Template.CreateTemplate:input_type -> api.server.prom.notify.CreateTemplateRequest
	2,  // 7: api.server.prom.notify.Template.UpdateTemplate:input_type -> api.server.prom.notify.UpdateTemplateRequest
	4,  // 8: api.server.prom.notify.Template.DeleteTemplate:input_type -> api.server.prom.notify.DeleteTemplateRequest
	6,  // 9: api.server.prom.notify.Template.GetTemplate:input_type -> api.server.prom.notify.GetTemplateRequest
	8,  // 10: api.server.prom.notify.Template.ListTemplate:input_type -> api.server.prom.notify.ListTemplateRequest
	1,  // 11: api.server.prom.notify.Template.CreateTemplate:output_type -> api.server.prom.notify.CreateTemplateReply
	3,  // 12: api.server.prom.notify.Template.UpdateTemplate:output_type -> api.server.prom.notify.UpdateTemplateReply
	5,  // 13: api.server.prom.notify.Template.DeleteTemplate:output_type -> api.server.prom.notify.DeleteTemplateReply
	7,  // 14: api.server.prom.notify.Template.GetTemplate:output_type -> api.server.prom.notify.GetTemplateReply
	9,  // 15: api.server.prom.notify.Template.ListTemplate:output_type -> api.server.prom.notify.ListTemplateReply
	11, // [11:16] is the sub-list for method output_type
	6,  // [6:11] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_server_prom_notify_template_proto_init() }
func file_server_prom_notify_template_proto_init() {
	if File_server_prom_notify_template_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_server_prom_notify_template_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateTemplateRequest); i {
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
		file_server_prom_notify_template_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateTemplateReply); i {
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
		file_server_prom_notify_template_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateTemplateRequest); i {
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
		file_server_prom_notify_template_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateTemplateReply); i {
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
		file_server_prom_notify_template_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteTemplateRequest); i {
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
		file_server_prom_notify_template_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteTemplateReply); i {
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
		file_server_prom_notify_template_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTemplateRequest); i {
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
		file_server_prom_notify_template_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTemplateReply); i {
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
		file_server_prom_notify_template_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListTemplateRequest); i {
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
		file_server_prom_notify_template_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListTemplateReply); i {
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
			RawDescriptor: file_server_prom_notify_template_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_server_prom_notify_template_proto_goTypes,
		DependencyIndexes: file_server_prom_notify_template_proto_depIdxs,
		MessageInfos:      file_server_prom_notify_template_proto_msgTypes,
	}.Build()
	File_server_prom_notify_template_proto = out.File
	file_server_prom_notify_template_proto_rawDesc = nil
	file_server_prom_notify_template_proto_goTypes = nil
	file_server_prom_notify_template_proto_depIdxs = nil
}
