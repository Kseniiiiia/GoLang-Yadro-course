// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v6.30.0--rc1
// source: yolo/yolo.proto

package yolo

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DetectRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ImageData     []byte                 `protobuf:"bytes,1,opt,name=image_data,json=imageData,proto3" json:"image_data,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DetectRequest) Reset() {
	*x = DetectRequest{}
	mi := &file_yolo_yolo_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DetectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DetectRequest) ProtoMessage() {}

func (x *DetectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_yolo_yolo_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DetectRequest.ProtoReflect.Descriptor instead.
func (*DetectRequest) Descriptor() ([]byte, []int) {
	return file_yolo_yolo_proto_rawDescGZIP(), []int{0}
}

func (x *DetectRequest) GetImageData() []byte {
	if x != nil {
		return x.ImageData
	}
	return nil
}

type DetectResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Results       []*Detection           `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DetectResponse) Reset() {
	*x = DetectResponse{}
	mi := &file_yolo_yolo_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DetectResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DetectResponse) ProtoMessage() {}

func (x *DetectResponse) ProtoReflect() protoreflect.Message {
	mi := &file_yolo_yolo_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DetectResponse.ProtoReflect.Descriptor instead.
func (*DetectResponse) Descriptor() ([]byte, []int) {
	return file_yolo_yolo_proto_rawDescGZIP(), []int{1}
}

func (x *DetectResponse) GetResults() []*Detection {
	if x != nil {
		return x.Results
	}
	return nil
}

type Detection struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Bboxes        []float32              `protobuf:"fixed32,1,rep,packed,name=bboxes,proto3" json:"bboxes,omitempty"`
	Confidence    float32                `protobuf:"fixed32,2,opt,name=confidence,proto3" json:"confidence,omitempty"`
	Label         string                 `protobuf:"bytes,3,opt,name=label,proto3" json:"label,omitempty"`
	LabelNum      int32                  `protobuf:"varint,4,opt,name=label_num,json=labelNum,proto3" json:"label_num,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Detection) Reset() {
	*x = Detection{}
	mi := &file_yolo_yolo_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Detection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Detection) ProtoMessage() {}

func (x *Detection) ProtoReflect() protoreflect.Message {
	mi := &file_yolo_yolo_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Detection.ProtoReflect.Descriptor instead.
func (*Detection) Descriptor() ([]byte, []int) {
	return file_yolo_yolo_proto_rawDescGZIP(), []int{2}
}

func (x *Detection) GetBboxes() []float32 {
	if x != nil {
		return x.Bboxes
	}
	return nil
}

func (x *Detection) GetConfidence() float32 {
	if x != nil {
		return x.Confidence
	}
	return 0
}

func (x *Detection) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *Detection) GetLabelNum() int32 {
	if x != nil {
		return x.LabelNum
	}
	return 0
}

var File_yolo_yolo_proto protoreflect.FileDescriptor

var file_yolo_yolo_proto_rawDesc = string([]byte{
	0x0a, 0x0f, 0x79, 0x6f, 0x6c, 0x6f, 0x2f, 0x79, 0x6f, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x04, 0x79, 0x6f, 0x6c, 0x6f, 0x22, 0x2e, 0x0a, 0x0d, 0x44, 0x65, 0x74, 0x65, 0x63,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x69, 0x6d, 0x61, 0x67,
	0x65, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x69, 0x6d,
	0x61, 0x67, 0x65, 0x44, 0x61, 0x74, 0x61, 0x22, 0x3b, 0x0a, 0x0e, 0x44, 0x65, 0x74, 0x65, 0x63,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x29, 0x0a, 0x07, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x79, 0x6f, 0x6c,
	0x6f, 0x2e, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x07, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x73, 0x22, 0x76, 0x0a, 0x09, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x62, 0x6f, 0x78, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x02, 0x52, 0x06, 0x62, 0x62, 0x6f, 0x78, 0x65, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0a, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x64, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x61, 0x62,
	0x65, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12,
	0x1b, 0x0a, 0x09, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x5f, 0x6e, 0x75, 0x6d, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x08, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x4e, 0x75, 0x6d, 0x32, 0x42, 0x0a, 0x0b,
	0x59, 0x6f, 0x6c, 0x6f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x33, 0x0a, 0x06, 0x44,
	0x65, 0x74, 0x65, 0x63, 0x74, 0x12, 0x13, 0x2e, 0x79, 0x6f, 0x6c, 0x6f, 0x2e, 0x44, 0x65, 0x74,
	0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x79, 0x6f, 0x6c,
	0x6f, 0x2e, 0x44, 0x65, 0x74, 0x65, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x1d, 0x5a, 0x1b, 0x79, 0x61, 0x64, 0x72, 0x6f, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x79, 0x6f, 0x6c, 0x6f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_yolo_yolo_proto_rawDescOnce sync.Once
	file_yolo_yolo_proto_rawDescData []byte
)

func file_yolo_yolo_proto_rawDescGZIP() []byte {
	file_yolo_yolo_proto_rawDescOnce.Do(func() {
		file_yolo_yolo_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_yolo_yolo_proto_rawDesc), len(file_yolo_yolo_proto_rawDesc)))
	})
	return file_yolo_yolo_proto_rawDescData
}

var file_yolo_yolo_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_yolo_yolo_proto_goTypes = []any{
	(*DetectRequest)(nil),  // 0: yolo.DetectRequest
	(*DetectResponse)(nil), // 1: yolo.DetectResponse
	(*Detection)(nil),      // 2: yolo.Detection
}
var file_yolo_yolo_proto_depIdxs = []int32{
	2, // 0: yolo.DetectResponse.results:type_name -> yolo.Detection
	0, // 1: yolo.YoloService.Detect:input_type -> yolo.DetectRequest
	1, // 2: yolo.YoloService.Detect:output_type -> yolo.DetectResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_yolo_yolo_proto_init() }
func file_yolo_yolo_proto_init() {
	if File_yolo_yolo_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_yolo_yolo_proto_rawDesc), len(file_yolo_yolo_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_yolo_yolo_proto_goTypes,
		DependencyIndexes: file_yolo_yolo_proto_depIdxs,
		MessageInfos:      file_yolo_yolo_proto_msgTypes,
	}.Build()
	File_yolo_yolo_proto = out.File
	file_yolo_yolo_proto_goTypes = nil
	file_yolo_yolo_proto_depIdxs = nil
}
