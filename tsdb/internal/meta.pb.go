// Code generated by protoc-gen-gogo.
// source: internal/meta.proto
// DO NOT EDIT!

/*
Package meta is a generated protocol buffer package.

It is generated from these files:
	internal/meta.proto

It has these top-level messages:
	Series
	Tag
	MeasurementFields
	Field
*/
package meta

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type Series struct {
	Key              *string `protobuf:"bytes,1,req,name=Key" json:"Key,omitempty"`
	Tags             []*Tag  `protobuf:"bytes,2,rep,name=Tags" json:"Tags,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Series) Reset()                    { *m = Series{} }
func (m *Series) String() string            { return proto.CompactTextString(m) }
func (*Series) ProtoMessage()               {}
func (*Series) Descriptor() ([]byte, []int) { return fileDescriptorMeta, []int{0} }

func (m *Series) GetKey() string {
	if m != nil && m.Key != nil {
		return *m.Key
	}
	return ""
}

func (m *Series) GetTags() []*Tag {
	if m != nil {
		return m.Tags
	}
	return nil
}

type Tag struct {
	Key              *string `protobuf:"bytes,1,req,name=Key" json:"Key,omitempty"`
	Value            *string `protobuf:"bytes,2,req,name=Value" json:"Value,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Tag) Reset()                    { *m = Tag{} }
func (m *Tag) String() string            { return proto.CompactTextString(m) }
func (*Tag) ProtoMessage()               {}
func (*Tag) Descriptor() ([]byte, []int) { return fileDescriptorMeta, []int{1} }

func (m *Tag) GetKey() string {
	if m != nil && m.Key != nil {
		return *m.Key
	}
	return ""
}

func (m *Tag) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

type MeasurementFields struct {
	Fields           []*Field `protobuf:"bytes,1,rep,name=Fields" json:"Fields,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *MeasurementFields) Reset()                    { *m = MeasurementFields{} }
func (m *MeasurementFields) String() string            { return proto.CompactTextString(m) }
func (*MeasurementFields) ProtoMessage()               {}
func (*MeasurementFields) Descriptor() ([]byte, []int) { return fileDescriptorMeta, []int{2} }

func (m *MeasurementFields) GetFields() []*Field {
	if m != nil {
		return m.Fields
	}
	return nil
}

type Field struct {
	ID               *int32  `protobuf:"varint,1,req,name=ID" json:"ID,omitempty"`
	Name             *string `protobuf:"bytes,2,req,name=Name" json:"Name,omitempty"`
	Type             *int32  `protobuf:"varint,3,req,name=Type" json:"Type,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Field) Reset()                    { *m = Field{} }
func (m *Field) String() string            { return proto.CompactTextString(m) }
func (*Field) ProtoMessage()               {}
func (*Field) Descriptor() ([]byte, []int) { return fileDescriptorMeta, []int{3} }

func (m *Field) GetID() int32 {
	if m != nil && m.ID != nil {
		return *m.ID
	}
	return 0
}

func (m *Field) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

func (m *Field) GetType() int32 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func init() {
	proto.RegisterType((*Series)(nil), "meta.Series")
	proto.RegisterType((*Tag)(nil), "meta.Tag")
	proto.RegisterType((*MeasurementFields)(nil), "meta.MeasurementFields")
	proto.RegisterType((*Field)(nil), "meta.Field")
}

func init() { proto.RegisterFile("internal/meta.proto", fileDescriptorMeta) }

var fileDescriptorMeta = []byte{
	// 197 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x8d, 0x3f, 0x4b, 0xc0, 0x30,
	0x10, 0x47, 0x69, 0xd2, 0x16, 0x7a, 0x05, 0xd1, 0xe8, 0x90, 0x45, 0x28, 0x71, 0xe9, 0x62, 0x05,
	0x27, 0x9d, 0x5c, 0x8a, 0x50, 0x44, 0x87, 0x18, 0xdc, 0x0f, 0x3c, 0x4a, 0xa1, 0xff, 0x48, 0xd2,
	0xa1, 0xdf, 0x5e, 0x7a, 0xd5, 0xcd, 0xed, 0xe5, 0xe5, 0x7e, 0x3c, 0xb8, 0x1e, 0xe6, 0x48, 0x7e,
	0xc6, 0xf1, 0x61, 0xa2, 0x88, 0xcd, 0xea, 0x97, 0xb8, 0xa8, 0xf4, 0x60, 0xf3, 0x0c, 0xf9, 0x27,
	0xf9, 0x81, 0x82, 0xba, 0x04, 0xf9, 0x46, 0xbb, 0x4e, 0x2a, 0x51, 0x17, 0xf6, 0x40, 0x75, 0x0b,
	0xa9, 0xc3, 0x3e, 0x68, 0x51, 0xc9, 0xba, 0x7c, 0x2c, 0x1a, 0x1e, 0x3b, 0xec, 0x2d, 0x6b, 0x73,
	0x0f, 0xd2, 0x61, 0xff, 0xcf, 0xee, 0x06, 0xb2, 0x2f, 0x1c, 0x37, 0xd2, 0x82, 0xdd, 0xf9, 0x30,
	0x4f, 0x70, 0xf5, 0x4e, 0x18, 0x36, 0x4f, 0x13, 0xcd, 0xf1, 0x75, 0xa0, 0xf1, 0x3b, 0xa8, 0x3b,
	0xc8, 0x4f, 0xd2, 0x09, 0x47, 0xca, 0x33, 0xc2, 0xce, 0xfe, 0x7e, 0x99, 0x17, 0xc8, 0x98, 0xd4,
	0x05, 0x88, 0xae, 0xe5, 0x52, 0x66, 0x45, 0xd7, 0x2a, 0x05, 0xe9, 0x07, 0x4e, 0x7f, 0x1d, 0xe6,
	0xc3, 0xb9, 0x7d, 0x25, 0x2d, 0xf9, 0x8a, 0xf9, 0x27, 0x00, 0x00, 0xff, 0xff, 0x64, 0xe7, 0xff,
	0xac, 0x00, 0x01, 0x00, 0x00,
}
