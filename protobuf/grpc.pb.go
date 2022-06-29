// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: protobuf/grpc.proto

package protobuf

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Entry_Action int32

const (
	Entry_Insert Entry_Action = 0
	Entry_Get    Entry_Action = 1
	Entry_Update Entry_Action = 2
	Entry_Delete Entry_Action = 3
	Entry_Tick   Entry_Action = 4
)

// Enum value maps for Entry_Action.
var (
	Entry_Action_name = map[int32]string{
		0: "Insert",
		1: "Get",
		2: "Update",
		3: "Delete",
		4: "Tick",
	}
	Entry_Action_value = map[string]int32{
		"Insert": 0,
		"Get":    1,
		"Update": 2,
		"Delete": 3,
		"Tick":   4,
	}
)

func (x Entry_Action) Enum() *Entry_Action {
	p := new(Entry_Action)
	*p = x
	return p
}

func (x Entry_Action) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Entry_Action) Descriptor() protoreflect.EnumDescriptor {
	return file_protobuf_grpc_proto_enumTypes[0].Descriptor()
}

func (Entry_Action) Type() protoreflect.EnumType {
	return &file_protobuf_grpc_proto_enumTypes[0]
}

func (x Entry_Action) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Entry_Action.Descriptor instead.
func (Entry_Action) EnumDescriptor() ([]byte, []int) {
	return file_protobuf_grpc_proto_rawDescGZIP(), []int{0, 0}
}

type Entry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Term   int32        `protobuf:"varint,1,opt,name=term,proto3" json:"term,omitempty"`
	Key    string       `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	Value  string       `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	Action Entry_Action `protobuf:"varint,4,opt,name=action,proto3,enum=protobuf.Entry_Action" json:"action,omitempty"`
}

func (x *Entry) Reset() {
	*x = Entry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_grpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entry) ProtoMessage() {}

func (x *Entry) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_grpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entry.ProtoReflect.Descriptor instead.
func (*Entry) Descriptor() ([]byte, []int) {
	return file_protobuf_grpc_proto_rawDescGZIP(), []int{0}
}

func (x *Entry) GetTerm() int32 {
	if x != nil {
		return x.Term
	}
	return 0
}

func (x *Entry) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Entry) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Entry) GetAction() Entry_Action {
	if x != nil {
		return x.Action
	}
	return Entry_Insert
}

// TickRequest is the message containing information inside a tick.
type TickRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Entry *Entry `protobuf:"bytes,2,opt,name=entry,proto3" json:"entry,omitempty"`
}

func (x *TickRequest) Reset() {
	*x = TickRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_grpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TickRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TickRequest) ProtoMessage() {}

func (x *TickRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_grpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TickRequest.ProtoReflect.Descriptor instead.
func (*TickRequest) Descriptor() ([]byte, []int) {
	return file_protobuf_grpc_proto_rawDescGZIP(), []int{1}
}

func (x *TickRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TickRequest) GetEntry() *Entry {
	if x != nil {
		return x.Entry
	}
	return nil
}

// TickResponse is the message containing whether the tick is accepted.
type TickResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Accept bool  `protobuf:"varint,1,opt,name=accept,proto3" json:"accept,omitempty"`
	Term   int32 `protobuf:"varint,2,opt,name=term,proto3" json:"term,omitempty"`
}

func (x *TickResponse) Reset() {
	*x = TickResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_grpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TickResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TickResponse) ProtoMessage() {}

func (x *TickResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_grpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TickResponse.ProtoReflect.Descriptor instead.
func (*TickResponse) Descriptor() ([]byte, []int) {
	return file_protobuf_grpc_proto_rawDescGZIP(), []int{2}
}

func (x *TickResponse) GetAccept() bool {
	if x != nil {
		return x.Accept
	}
	return false
}

func (x *TickResponse) GetTerm() int32 {
	if x != nil {
		return x.Term
	}
	return 0
}

var File_protobuf_grpc_proto protoreflect.FileDescriptor

var file_protobuf_grpc_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x22,
	0xb4, 0x01, 0x0a, 0x05, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x72,
	0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x74, 0x65, 0x72, 0x6d, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x2e, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x06, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x3f, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x0a, 0x0a, 0x06, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x47,
	0x65, 0x74, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x10, 0x02,
	0x12, 0x0a, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x10, 0x03, 0x12, 0x08, 0x0a, 0x04,
	0x54, 0x69, 0x63, 0x6b, 0x10, 0x04, 0x22, 0x44, 0x0a, 0x0b, 0x54, 0x69, 0x63, 0x6b, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x25, 0x0a, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x65, 0x6e, 0x74, 0x72, 0x79, 0x22, 0x3a, 0x0a, 0x0c,
	0x54, 0x69, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x61, 0x63, 0x63, 0x65, 0x70, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x61, 0x63,
	0x63, 0x65, 0x70, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x72, 0x6d, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x04, 0x74, 0x65, 0x72, 0x6d, 0x32, 0x88, 0x01, 0x0a, 0x06, 0x54, 0x69, 0x63,
	0x6b, 0x65, 0x72, 0x12, 0x3e, 0x0a, 0x0b, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x56, 0x6f,
	0x74, 0x65, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x26, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x62, 0x6a, 0x7a, 0x68, 0x61, 0x6e, 0x67, 0x31, 0x31, 0x30, 0x31, 0x2f, 0x72, 0x61,
	0x66, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_protobuf_grpc_proto_rawDescOnce sync.Once
	file_protobuf_grpc_proto_rawDescData = file_protobuf_grpc_proto_rawDesc
)

func file_protobuf_grpc_proto_rawDescGZIP() []byte {
	file_protobuf_grpc_proto_rawDescOnce.Do(func() {
		file_protobuf_grpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_protobuf_grpc_proto_rawDescData)
	})
	return file_protobuf_grpc_proto_rawDescData
}

var file_protobuf_grpc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protobuf_grpc_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protobuf_grpc_proto_goTypes = []interface{}{
	(Entry_Action)(0),    // 0: protobuf.Entry.Action
	(*Entry)(nil),        // 1: protobuf.Entry
	(*TickRequest)(nil),  // 2: protobuf.TickRequest
	(*TickResponse)(nil), // 3: protobuf.TickResponse
}
var file_protobuf_grpc_proto_depIdxs = []int32{
	0, // 0: protobuf.Entry.action:type_name -> protobuf.Entry.Action
	1, // 1: protobuf.TickRequest.entry:type_name -> protobuf.Entry
	2, // 2: protobuf.Ticker.AppendEntry:input_type -> protobuf.TickRequest
	2, // 3: protobuf.Ticker.RequestVote:input_type -> protobuf.TickRequest
	3, // 4: protobuf.Ticker.AppendEntry:output_type -> protobuf.TickResponse
	3, // 5: protobuf.Ticker.RequestVote:output_type -> protobuf.TickResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_protobuf_grpc_proto_init() }
func file_protobuf_grpc_proto_init() {
	if File_protobuf_grpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protobuf_grpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entry); i {
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
		file_protobuf_grpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TickRequest); i {
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
		file_protobuf_grpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TickResponse); i {
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
			RawDescriptor: file_protobuf_grpc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protobuf_grpc_proto_goTypes,
		DependencyIndexes: file_protobuf_grpc_proto_depIdxs,
		EnumInfos:         file_protobuf_grpc_proto_enumTypes,
		MessageInfos:      file_protobuf_grpc_proto_msgTypes,
	}.Build()
	File_protobuf_grpc_proto = out.File
	file_protobuf_grpc_proto_rawDesc = nil
	file_protobuf_grpc_proto_goTypes = nil
	file_protobuf_grpc_proto_depIdxs = nil
}
