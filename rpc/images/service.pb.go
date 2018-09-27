// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rpc/images/service.proto

package images

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ListBuildsRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListBuildsRequest) Reset()         { *m = ListBuildsRequest{} }
func (m *ListBuildsRequest) String() string { return proto.CompactTextString(m) }
func (*ListBuildsRequest) ProtoMessage()    {}
func (*ListBuildsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_03e4ae55c7c0e319, []int{0}
}

func (m *ListBuildsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListBuildsRequest.Unmarshal(m, b)
}
func (m *ListBuildsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListBuildsRequest.Marshal(b, m, deterministic)
}
func (m *ListBuildsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListBuildsRequest.Merge(m, src)
}
func (m *ListBuildsRequest) XXX_Size() int {
	return xxx_messageInfo_ListBuildsRequest.Size(m)
}
func (m *ListBuildsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListBuildsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListBuildsRequest proto.InternalMessageInfo

type ListBuildsResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListBuildsResponse) Reset()         { *m = ListBuildsResponse{} }
func (m *ListBuildsResponse) String() string { return proto.CompactTextString(m) }
func (*ListBuildsResponse) ProtoMessage()    {}
func (*ListBuildsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_03e4ae55c7c0e319, []int{1}
}

func (m *ListBuildsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListBuildsResponse.Unmarshal(m, b)
}
func (m *ListBuildsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListBuildsResponse.Marshal(b, m, deterministic)
}
func (m *ListBuildsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListBuildsResponse.Merge(m, src)
}
func (m *ListBuildsResponse) XXX_Size() int {
	return xxx_messageInfo_ListBuildsResponse.Size(m)
}
func (m *ListBuildsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListBuildsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListBuildsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ListBuildsRequest)(nil), "travisci.images.ListBuildsRequest")
	proto.RegisterType((*ListBuildsResponse)(nil), "travisci.images.ListBuildsResponse")
}

func init() { proto.RegisterFile("rpc/images/service.proto", fileDescriptor_03e4ae55c7c0e319) }

var fileDescriptor_03e4ae55c7c0e319 = []byte{
	// 135 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x28, 0x2a, 0x48, 0xd6,
	0xcf, 0xcc, 0x4d, 0x4c, 0x4f, 0x2d, 0xd6, 0x2f, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x2f, 0x29, 0x4a, 0x2c, 0xcb, 0x2c, 0x4e, 0xce, 0xd4, 0x83,
	0x48, 0x2b, 0x09, 0x73, 0x09, 0xfa, 0x64, 0x16, 0x97, 0x38, 0x95, 0x66, 0xe6, 0xa4, 0x14, 0x07,
	0xa5, 0x16, 0x96, 0xa6, 0x16, 0x97, 0x28, 0x89, 0x70, 0x09, 0x21, 0x0b, 0x16, 0x17, 0xe4, 0xe7,
	0x15, 0xa7, 0x1a, 0xc5, 0x73, 0xb1, 0x79, 0x82, 0x35, 0x09, 0x85, 0x72, 0x71, 0x21, 0xe4, 0x85,
	0x94, 0xf4, 0xd0, 0x0c, 0xd5, 0xc3, 0x30, 0x51, 0x4a, 0x19, 0xaf, 0x1a, 0x88, 0x05, 0x4e, 0x1c,
	0x51, 0x6c, 0x10, 0xc9, 0x24, 0x36, 0xb0, 0x6b, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0xdf,
	0x30, 0x75, 0x27, 0xc9, 0x00, 0x00, 0x00,
}
