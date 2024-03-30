// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: base.proto

package api

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
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

// 系统状态, 用于表达数据是否可用
type Status int32

const (
	// UNKNOWN 未知, 用于默认值
	Status_STATUS_UNKNOWN Status = 0
	// ENABLED 启用
	Status_STATUS_ENABLED Status = 1
	// DISABLED 禁用
	Status_STATUS_DISABLED Status = 2
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "STATUS_UNKNOWN",
		1: "STATUS_ENABLED",
		2: "STATUS_DISABLED",
	}
	Status_value = map[string]int32{
		"STATUS_UNKNOWN":  0,
		"STATUS_ENABLED":  1,
		"STATUS_DISABLED": 2,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_base_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_base_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{0}
}

// 告警状态
type AlarmStatus int32

const (
	// UNKNOWN 未知, 用于默认值
	AlarmStatus_ALARM_STATUS_UNKNOWN AlarmStatus = 0
	// ALARM 告警
	AlarmStatus_ALARM_STATUS_ALARM AlarmStatus = 1
	// RESOLVE 已解决
	AlarmStatus_ALARM_STATUS_RESOLVE AlarmStatus = 2
)

// Enum value maps for AlarmStatus.
var (
	AlarmStatus_name = map[int32]string{
		0: "ALARM_STATUS_UNKNOWN",
		1: "ALARM_STATUS_ALARM",
		2: "ALARM_STATUS_RESOLVE",
	}
	AlarmStatus_value = map[string]int32{
		"ALARM_STATUS_UNKNOWN": 0,
		"ALARM_STATUS_ALARM":   1,
		"ALARM_STATUS_RESOLVE": 2,
	}
)

func (x AlarmStatus) Enum() *AlarmStatus {
	p := new(AlarmStatus)
	*p = x
	return p
}

func (x AlarmStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AlarmStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_base_proto_enumTypes[1].Descriptor()
}

func (AlarmStatus) Type() protoreflect.EnumType {
	return &file_base_proto_enumTypes[1]
}

func (x AlarmStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AlarmStatus.Descriptor instead.
func (AlarmStatus) EnumDescriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{1}
}

// 分类, 区分字典中的各个模块数据
type Category int32

const (
	// UNKNOWN 未知, 用于默认值
	Category_CATEGORY_UNKNOWN Category = 0
	// PromLabel 标签
	Category_CATEGORY_PROM_LABEL Category = 1
	// PromAnnotation 注解
	Category_CATEGORY_PROM_ANNOTATION Category = 2
	// PromStrategy 策略
	Category_CATEGORY_PROM_STRATEGY Category = 3
	// PromStrategyGroup 策略组
	Category_CATEGORY_PROM_STRATEGY_GROUP Category = 4
	// AlarmLevel 告警级别
	Category_CATEGORY_ALARM_LEVEL Category = 5
	// AlarmStatus 告警状态
	Category_CATEGORY_ALARM_STATUS Category = 6
	// NotifyType 通知类型
	Category_CATEGORY_NOTIFY_TYPE Category = 7
)

// Enum value maps for Category.
var (
	Category_name = map[int32]string{
		0: "CATEGORY_UNKNOWN",
		1: "CATEGORY_PROM_LABEL",
		2: "CATEGORY_PROM_ANNOTATION",
		3: "CATEGORY_PROM_STRATEGY",
		4: "CATEGORY_PROM_STRATEGY_GROUP",
		5: "CATEGORY_ALARM_LEVEL",
		6: "CATEGORY_ALARM_STATUS",
		7: "CATEGORY_NOTIFY_TYPE",
	}
	Category_value = map[string]int32{
		"CATEGORY_UNKNOWN":             0,
		"CATEGORY_PROM_LABEL":          1,
		"CATEGORY_PROM_ANNOTATION":     2,
		"CATEGORY_PROM_STRATEGY":       3,
		"CATEGORY_PROM_STRATEGY_GROUP": 4,
		"CATEGORY_ALARM_LEVEL":         5,
		"CATEGORY_ALARM_STATUS":        6,
		"CATEGORY_NOTIFY_TYPE":         7,
	}
)

func (x Category) Enum() *Category {
	p := new(Category)
	*p = x
	return p
}

func (x Category) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Category) Descriptor() protoreflect.EnumDescriptor {
	return file_base_proto_enumTypes[2].Descriptor()
}

func (Category) Type() protoreflect.EnumType {
	return &file_base_proto_enumTypes[2]
}

func (x Category) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Category.Descriptor instead.
func (Category) EnumDescriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{2}
}

// 通知类型, 用于区分通知方式
type NotifyType int32

const (
	// UNKNOWN 未知, 用于默认值
	NotifyType_NOTIFY_TYPE_UNKNOWN NotifyType = 0
	// EMAIL 邮件
	NotifyType_NOTIFY_TYPE_EMAIL NotifyType = 2
	// SMS 短信
	NotifyType_NOTIFY_TYPE_SMS NotifyType = 4
	// phone 电话
	NotifyType_NOTIFY_TYPE_PHONE NotifyType = 8
)

// Enum value maps for NotifyType.
var (
	NotifyType_name = map[int32]string{
		0: "NOTIFY_TYPE_UNKNOWN",
		2: "NOTIFY_TYPE_EMAIL",
		4: "NOTIFY_TYPE_SMS",
		8: "NOTIFY_TYPE_PHONE",
	}
	NotifyType_value = map[string]int32{
		"NOTIFY_TYPE_UNKNOWN": 0,
		"NOTIFY_TYPE_EMAIL":   2,
		"NOTIFY_TYPE_SMS":     4,
		"NOTIFY_TYPE_PHONE":   8,
	}
)

func (x NotifyType) Enum() *NotifyType {
	p := new(NotifyType)
	*p = x
	return p
}

func (x NotifyType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NotifyType) Descriptor() protoreflect.EnumDescriptor {
	return file_base_proto_enumTypes[3].Descriptor()
}

func (NotifyType) Type() protoreflect.EnumType {
	return &file_base_proto_enumTypes[3]
}

func (x NotifyType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NotifyType.Descriptor instead.
func (NotifyType) EnumDescriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{3}
}

// 通知应用, 用于区分通知方式
type NotifyApp int32

const (
	// UNKNOWN 未知, 用于默认值
	NotifyApp_NOTIFY_APP_UNKNOWN NotifyApp = 0
	// DINGTALK 钉钉
	NotifyApp_NOTIFY_APP_DINGTALK NotifyApp = 1
	// WECHATWORK 企业微信
	NotifyApp_NOTIFY_APP_WECHATWORK NotifyApp = 2
	// FEISHU 飞书
	NotifyApp_NOTIFY_APP_FEISHU NotifyApp = 3
	// 自定义
	NotifyApp_NOTIFY_APP_CUSTOM NotifyApp = 4
)

// Enum value maps for NotifyApp.
var (
	NotifyApp_name = map[int32]string{
		0: "NOTIFY_APP_UNKNOWN",
		1: "NOTIFY_APP_DINGTALK",
		2: "NOTIFY_APP_WECHATWORK",
		3: "NOTIFY_APP_FEISHU",
		4: "NOTIFY_APP_CUSTOM",
	}
	NotifyApp_value = map[string]int32{
		"NOTIFY_APP_UNKNOWN":    0,
		"NOTIFY_APP_DINGTALK":   1,
		"NOTIFY_APP_WECHATWORK": 2,
		"NOTIFY_APP_FEISHU":     3,
		"NOTIFY_APP_CUSTOM":     4,
	}
)

func (x NotifyApp) Enum() *NotifyApp {
	p := new(NotifyApp)
	*p = x
	return p
}

func (x NotifyApp) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NotifyApp) Descriptor() protoreflect.EnumDescriptor {
	return file_base_proto_enumTypes[4].Descriptor()
}

func (NotifyApp) Type() protoreflect.EnumType {
	return &file_base_proto_enumTypes[4]
}

func (x NotifyApp) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NotifyApp.Descriptor instead.
func (NotifyApp) EnumDescriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{4}
}

// 性别, 用于区分用户性别
type Gender int32

const (
	// UNKNOWN 未知, 用于默认值
	Gender_Gender_UNKNOWN Gender = 0
	// MALE 男
	Gender_Gender_MALE Gender = 1
	// FEMALE 女
	Gender_Gender_FEMALE Gender = 2
)

// Enum value maps for Gender.
var (
	Gender_name = map[int32]string{
		0: "Gender_UNKNOWN",
		1: "Gender_MALE",
		2: "Gender_FEMALE",
	}
	Gender_value = map[string]int32{
		"Gender_UNKNOWN": 0,
		"Gender_MALE":    1,
		"Gender_FEMALE":  2,
	}
)

func (x Gender) Enum() *Gender {
	p := new(Gender)
	*p = x
	return p
}

func (x Gender) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Gender) Descriptor() protoreflect.EnumDescriptor {
	return file_base_proto_enumTypes[5].Descriptor()
}

func (Gender) Type() protoreflect.EnumType {
	return &file_base_proto_enumTypes[5]
}

func (x Gender) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Gender.Descriptor instead.
func (Gender) EnumDescriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{5}
}

// 操作动作
type Action int32

const (
	// UNKNOWN 未知, 用于默认值
	Action_ACTION_UNKNOWN Action = 0
	// CREATE 创建
	Action_ACTION_CREATE Action = 1
	// UPDATE 更新
	Action_ACTION_UPDATE Action = 2
	// DELETE 删除
	Action_ACTION_DELETE Action = 3
	// GET 获取
	Action_ACTION_GET Action = 4
	// ALL 所有
	Action_ACTION_ALL Action = 5
)

// Enum value maps for Action.
var (
	Action_name = map[int32]string{
		0: "ACTION_UNKNOWN",
		1: "ACTION_CREATE",
		2: "ACTION_UPDATE",
		3: "ACTION_DELETE",
		4: "ACTION_GET",
		5: "ACTION_ALL",
	}
	Action_value = map[string]int32{
		"ACTION_UNKNOWN": 0,
		"ACTION_CREATE":  1,
		"ACTION_UPDATE":  2,
		"ACTION_DELETE":  3,
		"ACTION_GET":     4,
		"ACTION_ALL":     5,
	}
)

func (x Action) Enum() *Action {
	p := new(Action)
	*p = x
	return p
}

func (x Action) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Action) Descriptor() protoreflect.EnumDescriptor {
	return file_base_proto_enumTypes[6].Descriptor()
}

func (Action) Type() protoreflect.EnumType {
	return &file_base_proto_enumTypes[6]
}

func (x Action) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Action.Descriptor instead.
func (Action) EnumDescriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{6}
}

// 分页请求参数
type PageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 当前页, 从1开始
	Curr int32 `protobuf:"varint,1,opt,name=curr,proto3" json:"curr,omitempty"`
	// 每页大小, 1 < size <= 200
	Size int32 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *PageRequest) Reset() {
	*x = PageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_base_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PageRequest) ProtoMessage() {}

func (x *PageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_base_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PageRequest.ProtoReflect.Descriptor instead.
func (*PageRequest) Descriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{0}
}

func (x *PageRequest) GetCurr() int32 {
	if x != nil {
		return x.Curr
	}
	return 0
}

func (x *PageRequest) GetSize() int32 {
	if x != nil {
		return x.Size
	}
	return 0
}

// 分页返回参数
type PageReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 当前页码
	Curr int32 `protobuf:"varint,1,opt,name=curr,proto3" json:"curr,omitempty"`
	// 每页大小
	Size int32 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
	// 总数
	Total int64 `protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}

func (x *PageReply) Reset() {
	*x = PageReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_base_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PageReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PageReply) ProtoMessage() {}

func (x *PageReply) ProtoReflect() protoreflect.Message {
	mi := &file_base_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PageReply.ProtoReflect.Descriptor instead.
func (*PageReply) Descriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{1}
}

func (x *PageReply) GetCurr() int32 {
	if x != nil {
		return x.Curr
	}
	return 0
}

func (x *PageReply) GetSize() int32 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *PageReply) GetTotal() int64 {
	if x != nil {
		return x.Total
	}
	return 0
}

// 返回参数
type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 状态码
	Code int32 `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	// 返回信息
	Msg string `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_base_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_base_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_base_proto_rawDescGZIP(), []int{2}
}

func (x *Response) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *Response) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

var File_base_proto protoreflect.FileDescriptor

var file_base_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x61, 0x70,
	0x69, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4a, 0x0a, 0x0b, 0x50, 0x61,
	0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x04, 0x63, 0x75, 0x72,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x1a, 0x02, 0x20, 0x00,
	0x52, 0x04, 0x63, 0x75, 0x72, 0x72, 0x12, 0x1e, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x42, 0x0a, 0xfa, 0x42, 0x07, 0x1a, 0x05, 0x18, 0xc8, 0x01, 0x20, 0x00,
	0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x49, 0x0a, 0x09, 0x50, 0x61, 0x67, 0x65, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x75, 0x72, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x04, 0x63, 0x75, 0x72, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74,
	0x6f, 0x74, 0x61, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61,
	0x6c, 0x22, 0x30, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x63, 0x6f, 0x64,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6d, 0x73, 0x67, 0x2a, 0x45, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a,
	0x0e, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10,
	0x00, 0x12, 0x12, 0x0a, 0x0e, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x45, 0x4e, 0x41, 0x42,
	0x4c, 0x45, 0x44, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f,
	0x44, 0x49, 0x53, 0x41, 0x42, 0x4c, 0x45, 0x44, 0x10, 0x02, 0x2a, 0x59, 0x0a, 0x0b, 0x41, 0x6c,
	0x61, 0x72, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x14, 0x41, 0x4c, 0x41,
	0x52, 0x4d, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x16, 0x0a, 0x12, 0x41, 0x4c, 0x41, 0x52, 0x4d, 0x5f, 0x53, 0x54, 0x41,
	0x54, 0x55, 0x53, 0x5f, 0x41, 0x4c, 0x41, 0x52, 0x4d, 0x10, 0x01, 0x12, 0x18, 0x0a, 0x14, 0x41,
	0x4c, 0x41, 0x52, 0x4d, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x53, 0x4f,
	0x4c, 0x56, 0x45, 0x10, 0x02, 0x2a, 0xe4, 0x01, 0x0a, 0x08, 0x43, 0x61, 0x74, 0x65, 0x67, 0x6f,
	0x72, 0x79, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x55,
	0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x17, 0x0a, 0x13, 0x43, 0x41, 0x54, 0x45,
	0x47, 0x4f, 0x52, 0x59, 0x5f, 0x50, 0x52, 0x4f, 0x4d, 0x5f, 0x4c, 0x41, 0x42, 0x45, 0x4c, 0x10,
	0x01, 0x12, 0x1c, 0x0a, 0x18, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x50, 0x52,
	0x4f, 0x4d, 0x5f, 0x41, 0x4e, 0x4e, 0x4f, 0x54, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x02, 0x12,
	0x1a, 0x0a, 0x16, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x50, 0x52, 0x4f, 0x4d,
	0x5f, 0x53, 0x54, 0x52, 0x41, 0x54, 0x45, 0x47, 0x59, 0x10, 0x03, 0x12, 0x20, 0x0a, 0x1c, 0x43,
	0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x50, 0x52, 0x4f, 0x4d, 0x5f, 0x53, 0x54, 0x52,
	0x41, 0x54, 0x45, 0x47, 0x59, 0x5f, 0x47, 0x52, 0x4f, 0x55, 0x50, 0x10, 0x04, 0x12, 0x18, 0x0a,
	0x14, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x41, 0x4c, 0x41, 0x52, 0x4d, 0x5f,
	0x4c, 0x45, 0x56, 0x45, 0x4c, 0x10, 0x05, 0x12, 0x19, 0x0a, 0x15, 0x43, 0x41, 0x54, 0x45, 0x47,
	0x4f, 0x52, 0x59, 0x5f, 0x41, 0x4c, 0x41, 0x52, 0x4d, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53,
	0x10, 0x06, 0x12, 0x18, 0x0a, 0x14, 0x43, 0x41, 0x54, 0x45, 0x47, 0x4f, 0x52, 0x59, 0x5f, 0x4e,
	0x4f, 0x54, 0x49, 0x46, 0x59, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x10, 0x07, 0x2a, 0x68, 0x0a, 0x0a,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x17, 0x0a, 0x13, 0x4e, 0x4f,
	0x54, 0x49, 0x46, 0x59, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x45, 0x4d, 0x41, 0x49, 0x4c, 0x10, 0x02, 0x12, 0x13, 0x0a, 0x0f, 0x4e, 0x4f,
	0x54, 0x49, 0x46, 0x59, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x4d, 0x53, 0x10, 0x04, 0x12,
	0x15, 0x0a, 0x11, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50,
	0x48, 0x4f, 0x4e, 0x45, 0x10, 0x08, 0x2a, 0x85, 0x01, 0x0a, 0x09, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x41, 0x70, 0x70, 0x12, 0x16, 0x0a, 0x12, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x5f, 0x41,
	0x50, 0x50, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x17, 0x0a, 0x13,
	0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x5f, 0x41, 0x50, 0x50, 0x5f, 0x44, 0x49, 0x4e, 0x47, 0x54,
	0x41, 0x4c, 0x4b, 0x10, 0x01, 0x12, 0x19, 0x0a, 0x15, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x5f,
	0x41, 0x50, 0x50, 0x5f, 0x57, 0x45, 0x43, 0x48, 0x41, 0x54, 0x57, 0x4f, 0x52, 0x4b, 0x10, 0x02,
	0x12, 0x15, 0x0a, 0x11, 0x4e, 0x4f, 0x54, 0x49, 0x46, 0x59, 0x5f, 0x41, 0x50, 0x50, 0x5f, 0x46,
	0x45, 0x49, 0x53, 0x48, 0x55, 0x10, 0x03, 0x12, 0x15, 0x0a, 0x11, 0x4e, 0x4f, 0x54, 0x49, 0x46,
	0x59, 0x5f, 0x41, 0x50, 0x50, 0x5f, 0x43, 0x55, 0x53, 0x54, 0x4f, 0x4d, 0x10, 0x04, 0x2a, 0x40,
	0x0a, 0x06, 0x47, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x0e, 0x47, 0x65, 0x6e, 0x64,
	0x65, 0x72, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b,
	0x47, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x4d, 0x41, 0x4c, 0x45, 0x10, 0x01, 0x12, 0x11, 0x0a,
	0x0d, 0x47, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x46, 0x45, 0x4d, 0x41, 0x4c, 0x45, 0x10, 0x02,
	0x2a, 0x75, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x0e, 0x41, 0x43,
	0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x11,
	0x0a, 0x0d, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x43, 0x52, 0x45, 0x41, 0x54, 0x45, 0x10,
	0x01, 0x12, 0x11, 0x0a, 0x0d, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x55, 0x50, 0x44, 0x41,
	0x54, 0x45, 0x10, 0x02, 0x12, 0x11, 0x0a, 0x0d, 0x41, 0x43, 0x54, 0x49, 0x4f, 0x4e, 0x5f, 0x44,
	0x45, 0x4c, 0x45, 0x54, 0x45, 0x10, 0x03, 0x12, 0x0e, 0x0a, 0x0a, 0x41, 0x43, 0x54, 0x49, 0x4f,
	0x4e, 0x5f, 0x47, 0x45, 0x54, 0x10, 0x04, 0x12, 0x0e, 0x0a, 0x0a, 0x41, 0x43, 0x54, 0x49, 0x4f,
	0x4e, 0x5f, 0x41, 0x4c, 0x4c, 0x10, 0x05, 0x42, 0x23, 0x0a, 0x03, 0x61, 0x70, 0x69, 0x50, 0x01,
	0x5a, 0x1a, 0x70, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x2d, 0x6d, 0x61, 0x6e,
	0x61, 0x67, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x3b, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_base_proto_rawDescOnce sync.Once
	file_base_proto_rawDescData = file_base_proto_rawDesc
)

func file_base_proto_rawDescGZIP() []byte {
	file_base_proto_rawDescOnce.Do(func() {
		file_base_proto_rawDescData = protoimpl.X.CompressGZIP(file_base_proto_rawDescData)
	})
	return file_base_proto_rawDescData
}

var file_base_proto_enumTypes = make([]protoimpl.EnumInfo, 7)
var file_base_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_base_proto_goTypes = []interface{}{
	(Status)(0),         // 0: api.Status
	(AlarmStatus)(0),    // 1: api.AlarmStatus
	(Category)(0),       // 2: api.Category
	(NotifyType)(0),     // 3: api.NotifyType
	(NotifyApp)(0),      // 4: api.NotifyApp
	(Gender)(0),         // 5: api.Gender
	(Action)(0),         // 6: api.Action
	(*PageRequest)(nil), // 7: api.PageRequest
	(*PageReply)(nil),   // 8: api.PageReply
	(*Response)(nil),    // 9: api.Response
}
var file_base_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_base_proto_init() }
func file_base_proto_init() {
	if File_base_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_base_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PageRequest); i {
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
		file_base_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PageReply); i {
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
		file_base_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
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
			RawDescriptor: file_base_proto_rawDesc,
			NumEnums:      7,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_base_proto_goTypes,
		DependencyIndexes: file_base_proto_depIdxs,
		EnumInfos:         file_base_proto_enumTypes,
		MessageInfos:      file_base_proto_msgTypes,
	}.Build()
	File_base_proto = out.File
	file_base_proto_rawDesc = nil
	file_base_proto_goTypes = nil
	file_base_proto_depIdxs = nil
}
