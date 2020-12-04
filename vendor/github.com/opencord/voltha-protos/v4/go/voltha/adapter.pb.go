// Code generated by protoc-gen-go. DO NOT EDIT.
// source: voltha_protos/adapter.proto

package voltha

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/opencord/voltha-protos/v4/go/common"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type AdapterConfig struct {
	// Custom (vendor-specific) configuration attributes
	AdditionalConfig     *any.Any `protobuf:"bytes,64,opt,name=additional_config,json=additionalConfig,proto3" json:"additional_config,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AdapterConfig) Reset()         { *m = AdapterConfig{} }
func (m *AdapterConfig) String() string { return proto.CompactTextString(m) }
func (*AdapterConfig) ProtoMessage()    {}
func (*AdapterConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e998ce153307274, []int{0}
}

func (m *AdapterConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdapterConfig.Unmarshal(m, b)
}
func (m *AdapterConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdapterConfig.Marshal(b, m, deterministic)
}
func (m *AdapterConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdapterConfig.Merge(m, src)
}
func (m *AdapterConfig) XXX_Size() int {
	return xxx_messageInfo_AdapterConfig.Size(m)
}
func (m *AdapterConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_AdapterConfig.DiscardUnknown(m)
}

var xxx_messageInfo_AdapterConfig proto.InternalMessageInfo

func (m *AdapterConfig) GetAdditionalConfig() *any.Any {
	if m != nil {
		return m.AdditionalConfig
	}
	return nil
}

// Adapter (software plugin)
type Adapter struct {
	// the adapter ID has to be unique,
	// it will be generated as Type + CurrentReplica
	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Vendor  string `protobuf:"bytes,2,opt,name=vendor,proto3" json:"vendor,omitempty"`
	Version string `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	// Adapter configuration
	Config *AdapterConfig `protobuf:"bytes,16,opt,name=config,proto3" json:"config,omitempty"`
	// Custom descriptors and custom configuration
	AdditionalDescription *any.Any `protobuf:"bytes,64,opt,name=additional_description,json=additionalDescription,proto3" json:"additional_description,omitempty"`
	LogicalDeviceIds      []string `protobuf:"bytes,4,rep,name=logical_device_ids,json=logicalDeviceIds,proto3" json:"logical_device_ids,omitempty"`
	// timestamp when the adapter last sent a message to the core
	LastCommunication *timestamp.Timestamp `protobuf:"bytes,5,opt,name=last_communication,json=lastCommunication,proto3" json:"last_communication,omitempty"`
	CurrentReplica    int32                `protobuf:"varint,6,opt,name=currentReplica,proto3" json:"currentReplica,omitempty"`
	TotalReplicas     int32                `protobuf:"varint,7,opt,name=totalReplicas,proto3" json:"totalReplicas,omitempty"`
	Endpoint          string               `protobuf:"bytes,8,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	// all replicas of the same adapter will have the same type
	// it is used to associate a device to an adapter
	Type                 string   `protobuf:"bytes,9,opt,name=type,proto3" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Adapter) Reset()         { *m = Adapter{} }
func (m *Adapter) String() string { return proto.CompactTextString(m) }
func (*Adapter) ProtoMessage()    {}
func (*Adapter) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e998ce153307274, []int{1}
}

func (m *Adapter) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Adapter.Unmarshal(m, b)
}
func (m *Adapter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Adapter.Marshal(b, m, deterministic)
}
func (m *Adapter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Adapter.Merge(m, src)
}
func (m *Adapter) XXX_Size() int {
	return xxx_messageInfo_Adapter.Size(m)
}
func (m *Adapter) XXX_DiscardUnknown() {
	xxx_messageInfo_Adapter.DiscardUnknown(m)
}

var xxx_messageInfo_Adapter proto.InternalMessageInfo

func (m *Adapter) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Adapter) GetVendor() string {
	if m != nil {
		return m.Vendor
	}
	return ""
}

func (m *Adapter) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *Adapter) GetConfig() *AdapterConfig {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *Adapter) GetAdditionalDescription() *any.Any {
	if m != nil {
		return m.AdditionalDescription
	}
	return nil
}

func (m *Adapter) GetLogicalDeviceIds() []string {
	if m != nil {
		return m.LogicalDeviceIds
	}
	return nil
}

func (m *Adapter) GetLastCommunication() *timestamp.Timestamp {
	if m != nil {
		return m.LastCommunication
	}
	return nil
}

func (m *Adapter) GetCurrentReplica() int32 {
	if m != nil {
		return m.CurrentReplica
	}
	return 0
}

func (m *Adapter) GetTotalReplicas() int32 {
	if m != nil {
		return m.TotalReplicas
	}
	return 0
}

func (m *Adapter) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

func (m *Adapter) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

type Adapters struct {
	Items                []*Adapter `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Adapters) Reset()         { *m = Adapters{} }
func (m *Adapters) String() string { return proto.CompactTextString(m) }
func (*Adapters) ProtoMessage()    {}
func (*Adapters) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e998ce153307274, []int{2}
}

func (m *Adapters) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Adapters.Unmarshal(m, b)
}
func (m *Adapters) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Adapters.Marshal(b, m, deterministic)
}
func (m *Adapters) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Adapters.Merge(m, src)
}
func (m *Adapters) XXX_Size() int {
	return xxx_messageInfo_Adapters.Size(m)
}
func (m *Adapters) XXX_DiscardUnknown() {
	xxx_messageInfo_Adapters.DiscardUnknown(m)
}

var xxx_messageInfo_Adapters proto.InternalMessageInfo

func (m *Adapters) GetItems() []*Adapter {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
	proto.RegisterType((*AdapterConfig)(nil), "voltha.AdapterConfig")
	proto.RegisterType((*Adapter)(nil), "voltha.Adapter")
	proto.RegisterType((*Adapters)(nil), "voltha.Adapters")
}

func init() { proto.RegisterFile("voltha_protos/adapter.proto", fileDescriptor_7e998ce153307274) }

var fileDescriptor_7e998ce153307274 = []byte{
	// 455 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0x56, 0xda, 0xf5, 0xcf, 0x53, 0xa1, 0x33, 0x14, 0x99, 0xa2, 0x69, 0x51, 0x05, 0x28, 0x17,
	0x2c, 0x11, 0x83, 0x07, 0xa0, 0xdd, 0x6e, 0x76, 0x6b, 0x4d, 0x5c, 0x70, 0x53, 0xb9, 0xb6, 0x97,
	0x59, 0x4a, 0x7c, 0xa2, 0xd8, 0x8d, 0xd4, 0x87, 0xe4, 0x05, 0x78, 0x02, 0x9e, 0x80, 0x6b, 0x54,
	0xdb, 0xa1, 0x3f, 0x48, 0xbb, 0x4a, 0xce, 0xf7, 0x73, 0x3e, 0x9f, 0x63, 0xa3, 0x77, 0x0d, 0x14,
	0xf6, 0x89, 0xad, 0xaa, 0x1a, 0x2c, 0x98, 0x8c, 0x09, 0x56, 0x59, 0x59, 0xa7, 0xae, 0xc4, 0x7d,
	0x4f, 0xce, 0xde, 0xe6, 0x00, 0x79, 0x21, 0x33, 0x87, 0xae, 0x37, 0x8f, 0x19, 0xd3, 0x5b, 0x2f,
	0x99, 0x91, 0x63, 0x7f, 0x29, 0x2d, 0x0b, 0xcc, 0xd5, 0xa9, 0xc9, 0xaa, 0x52, 0x1a, 0xcb, 0xca,
	0xca, 0x0b, 0xe6, 0x14, 0x8d, 0x17, 0x3e, 0xee, 0x16, 0xf4, 0xa3, 0xca, 0xf1, 0x02, 0x5d, 0x30,
	0x21, 0x94, 0x55, 0xa0, 0x59, 0xb1, 0xe2, 0x0e, 0x24, 0xdf, 0xe2, 0x28, 0x39, 0xbf, 0x79, 0x9d,
	0xfa, 0x6e, 0x69, 0xdb, 0x2d, 0x5d, 0xe8, 0x2d, 0x9d, 0xec, 0xe5, 0xbe, 0xc5, 0xfc, 0x57, 0x17,
	0x0d, 0x42, 0x53, 0x3c, 0x45, 0x1d, 0x25, 0x48, 0x14, 0x47, 0xc9, 0x68, 0xd9, 0xfb, 0xfd, 0xe7,
	0xe7, 0x65, 0x44, 0x3b, 0x4a, 0xe0, 0x4b, 0xd4, 0x6f, 0xa4, 0x16, 0x50, 0x93, 0xce, 0x21, 0x15,
	0x40, 0x7c, 0x85, 0x06, 0x8d, 0xac, 0x8d, 0x02, 0x4d, 0xba, 0x87, 0x7c, 0x8b, 0xe2, 0x6b, 0xd4,
	0x0f, 0x47, 0x9b, 0xb8, 0xa3, 0x4d, 0x53, 0xbf, 0x82, 0xf4, 0x68, 0x18, 0x1a, 0x44, 0x98, 0xa2,
	0x37, 0x07, 0x43, 0x09, 0x69, 0x78, 0xad, 0xaa, 0x5d, 0xf5, 0xdc, 0x64, 0x6d, 0xe8, 0x74, 0x6f,
	0xbd, 0xdb, 0x3b, 0xf1, 0x27, 0x84, 0x0b, 0xc8, 0x15, 0x77, 0x0d, 0x1b, 0xc5, 0xe5, 0x4a, 0x09,
	0x43, 0xce, 0xe2, 0x6e, 0x32, 0xa2, 0x93, 0xc0, 0xdc, 0x39, 0xe2, 0x5e, 0x18, 0x7c, 0x8f, 0x70,
	0xc1, 0x8c, 0x5d, 0x71, 0x28, 0xcb, 0x8d, 0x56, 0x9c, 0xb9, 0xf4, 0x9e, 0x4b, 0x9f, 0xfd, 0x97,
	0xfe, 0xd0, 0xde, 0x12, 0xbd, 0xd8, 0xb9, 0x6e, 0x0f, 0x4d, 0xf8, 0x23, 0x7a, 0xc1, 0x37, 0x75,
	0x2d, 0xb5, 0xa5, 0xb2, 0x2a, 0x14, 0x67, 0xa4, 0x1f, 0x47, 0x49, 0x8f, 0x9e, 0xa0, 0xf8, 0x3d,
	0x1a, 0x5b, 0xb0, 0xac, 0x08, 0xb5, 0x21, 0x03, 0x27, 0x3b, 0x06, 0xf1, 0x0c, 0x0d, 0xa5, 0x16,
	0x15, 0x28, 0x6d, 0xc9, 0x70, 0xb7, 0x6b, 0xfa, 0xaf, 0xc6, 0x18, 0x9d, 0xd9, 0x6d, 0x25, 0xc9,
	0xc8, 0xe1, 0xee, 0x7f, 0xfe, 0x19, 0x0d, 0xc3, 0x8e, 0x0d, 0xfe, 0x80, 0x7a, 0xca, 0xca, 0xd2,
	0x90, 0x28, 0xee, 0x26, 0xe7, 0x37, 0x2f, 0x4f, 0x2e, 0x81, 0x7a, 0x76, 0xf9, 0x80, 0x5e, 0x41,
	0x9d, 0xa7, 0x50, 0x49, 0xcd, 0xa1, 0x16, 0x41, 0xb5, 0x1c, 0x7f, 0x77, 0xdf, 0x20, 0xfe, 0x91,
	0xe6, 0xca, 0x3e, 0x6d, 0xd6, 0x29, 0x87, 0x32, 0x6b, 0xa5, 0x99, 0x97, 0x5e, 0x87, 0x87, 0xdd,
	0x7c, 0xcd, 0x72, 0x08, 0xd8, 0xba, 0xef, 0xc0, 0x2f, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x07,
	0xa4, 0x0d, 0x2b, 0x3d, 0x03, 0x00, 0x00,
}