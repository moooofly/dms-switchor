// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v2/services/parental_status_view_service.proto

package services

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	resources "google.golang.org/genproto/googleapis/ads/googleads/v2/resources"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// Request message for [ParentalStatusViewService.GetParentalStatusView][google.ads.googleads.v2.services.ParentalStatusViewService.GetParentalStatusView].
type GetParentalStatusViewRequest struct {
	// Required. The resource name of the parental status view to fetch.
	ResourceName         string   `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetParentalStatusViewRequest) Reset()         { *m = GetParentalStatusViewRequest{} }
func (m *GetParentalStatusViewRequest) String() string { return proto.CompactTextString(m) }
func (*GetParentalStatusViewRequest) ProtoMessage()    {}
func (*GetParentalStatusViewRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3677998f9a8d2f2f, []int{0}
}

func (m *GetParentalStatusViewRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetParentalStatusViewRequest.Unmarshal(m, b)
}
func (m *GetParentalStatusViewRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetParentalStatusViewRequest.Marshal(b, m, deterministic)
}
func (m *GetParentalStatusViewRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetParentalStatusViewRequest.Merge(m, src)
}
func (m *GetParentalStatusViewRequest) XXX_Size() int {
	return xxx_messageInfo_GetParentalStatusViewRequest.Size(m)
}
func (m *GetParentalStatusViewRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetParentalStatusViewRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetParentalStatusViewRequest proto.InternalMessageInfo

func (m *GetParentalStatusViewRequest) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

func init() {
	proto.RegisterType((*GetParentalStatusViewRequest)(nil), "google.ads.googleads.v2.services.GetParentalStatusViewRequest")
}

func init() {
	proto.RegisterFile("google/ads/googleads/v2/services/parental_status_view_service.proto", fileDescriptor_3677998f9a8d2f2f)
}

var fileDescriptor_3677998f9a8d2f2f = []byte{
	// 442 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x53, 0xcf, 0x6b, 0x13, 0x41,
	0x18, 0x25, 0x2b, 0x08, 0x2e, 0x7a, 0x59, 0x10, 0xdb, 0x58, 0x34, 0x94, 0x1e, 0xa4, 0xe2, 0x0c,
	0x6c, 0x29, 0xc2, 0xf8, 0x03, 0x26, 0x1e, 0xe2, 0x45, 0x09, 0x16, 0x82, 0x48, 0x60, 0x99, 0xee,
	0x7e, 0xae, 0x03, 0xbb, 0x33, 0x71, 0xbe, 0xc9, 0x56, 0x10, 0x2f, 0x82, 0x7f, 0x81, 0x17, 0xcf,
	0x1e, 0xfd, 0x53, 0x7a, 0xf5, 0x26, 0x08, 0x1e, 0x3c, 0xf9, 0x27, 0x78, 0x92, 0x64, 0x66, 0x36,
	0x29, 0xed, 0x9a, 0xdb, 0x63, 0xdf, 0xfb, 0xde, 0xf7, 0xcd, 0x7b, 0x6c, 0xfc, 0xa4, 0xd4, 0xba,
	0xac, 0x80, 0x8a, 0x02, 0xa9, 0x83, 0x0b, 0xd4, 0xa4, 0x14, 0xc1, 0x34, 0x32, 0x07, 0xa4, 0x33,
	0x61, 0x40, 0x59, 0x51, 0x65, 0x68, 0x85, 0x9d, 0x63, 0xd6, 0x48, 0x38, 0xc9, 0x3c, 0x4b, 0x66,
	0x46, 0x5b, 0x9d, 0x0c, 0xdc, 0x24, 0x11, 0x05, 0x92, 0xd6, 0x84, 0x34, 0x29, 0x09, 0x26, 0xfd,
	0x87, 0x5d, 0x6b, 0x0c, 0xa0, 0x9e, 0x9b, 0xae, 0x3d, 0xce, 0xbf, 0xbf, 0x13, 0xa6, 0x67, 0x92,
	0x0a, 0xa5, 0xb4, 0x15, 0x56, 0x6a, 0x85, 0x9e, 0xbd, 0xb1, 0xc6, 0xe6, 0x95, 0x04, 0x65, 0x3d,
	0x71, 0x7b, 0x8d, 0x78, 0x2d, 0xa1, 0x2a, 0xb2, 0x63, 0x78, 0x23, 0x1a, 0xa9, 0x8d, 0x17, 0x6c,
	0xaf, 0x09, 0xc2, 0x21, 0x8e, 0xda, 0x7d, 0x17, 0xef, 0x8c, 0xc0, 0x8e, 0xfd, 0x4d, 0x47, 0xcb,
	0x93, 0x26, 0x12, 0x4e, 0x5e, 0xc0, 0xdb, 0x39, 0xa0, 0x4d, 0x5e, 0xc6, 0xd7, 0xc2, 0x44, 0xa6,
	0x44, 0x0d, 0x5b, 0xbd, 0x41, 0xef, 0xce, 0x95, 0xe1, 0xc1, 0x2f, 0x1e, 0xfd, 0xe5, 0xf7, 0xe2,
	0xbb, 0xab, 0x18, 0x3c, 0x9a, 0x49, 0x24, 0xb9, 0xae, 0xe9, 0x05, 0x96, 0x57, 0x83, 0xd3, 0x73,
	0x51, 0x43, 0xfa, 0x25, 0x8a, 0xb7, 0xcf, 0x8b, 0x8e, 0x5c, 0x92, 0xc9, 0xcf, 0x5e, 0x7c, 0xfd,
	0xc2, 0xc3, 0x92, 0xc7, 0x64, 0x53, 0x0b, 0xe4, 0x7f, 0x2f, 0xea, 0x1f, 0x76, 0xce, 0xb7, 0x1d,
	0x91, 0xf3, 0xd3, 0xbb, 0xcf, 0x7e, 0xf0, 0xb3, 0x49, 0x7c, 0xfc, 0xfe, 0xfb, 0x73, 0x74, 0x3f,
	0x39, 0x5c, 0xb4, 0xfb, 0xfe, 0x0c, 0xf3, 0x28, 0x9f, 0xa3, 0xd5, 0x35, 0x18, 0xa4, 0xfb, 0x6d,
	0xdd, 0x2b, 0x2b, 0xa4, 0xfb, 0x1f, 0xfa, 0x37, 0x4f, 0xf9, 0x56, 0x57, 0x76, 0xc3, 0x4f, 0x51,
	0xbc, 0x97, 0xeb, 0x7a, 0xe3, 0x43, 0x87, 0xb7, 0x3a, 0x03, 0x1c, 0x2f, 0xda, 0x1d, 0xf7, 0x5e,
	0x3d, 0xf5, 0x1e, 0xa5, 0xae, 0x84, 0x2a, 0x89, 0x36, 0x25, 0x2d, 0x41, 0x2d, 0xbb, 0xa7, 0xab,
	0xad, 0xdd, 0xbf, 0xc5, 0x83, 0x00, 0xbe, 0x46, 0x97, 0x46, 0x9c, 0x7f, 0x8b, 0x06, 0x23, 0x67,
	0xc8, 0x0b, 0x24, 0x0e, 0x2e, 0xd0, 0x24, 0x25, 0x7e, 0x31, 0x9e, 0x06, 0xc9, 0x94, 0x17, 0x38,
	0x6d, 0x25, 0xd3, 0x49, 0x3a, 0x0d, 0x92, 0x3f, 0xd1, 0x9e, 0xfb, 0xce, 0x18, 0x2f, 0x90, 0xb1,
	0x56, 0xc4, 0xd8, 0x24, 0x65, 0x2c, 0xc8, 0x8e, 0x2f, 0x2f, 0xef, 0x3c, 0xf8, 0x17, 0x00, 0x00,
	0xff, 0xff, 0x52, 0xe6, 0x78, 0x50, 0xbd, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ParentalStatusViewServiceClient is the client API for ParentalStatusViewService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ParentalStatusViewServiceClient interface {
	// Returns the requested parental status view in full detail.
	GetParentalStatusView(ctx context.Context, in *GetParentalStatusViewRequest, opts ...grpc.CallOption) (*resources.ParentalStatusView, error)
}

type parentalStatusViewServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewParentalStatusViewServiceClient(cc grpc.ClientConnInterface) ParentalStatusViewServiceClient {
	return &parentalStatusViewServiceClient{cc}
}

func (c *parentalStatusViewServiceClient) GetParentalStatusView(ctx context.Context, in *GetParentalStatusViewRequest, opts ...grpc.CallOption) (*resources.ParentalStatusView, error) {
	out := new(resources.ParentalStatusView)
	err := c.cc.Invoke(ctx, "/google.ads.googleads.v2.services.ParentalStatusViewService/GetParentalStatusView", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ParentalStatusViewServiceServer is the server API for ParentalStatusViewService service.
type ParentalStatusViewServiceServer interface {
	// Returns the requested parental status view in full detail.
	GetParentalStatusView(context.Context, *GetParentalStatusViewRequest) (*resources.ParentalStatusView, error)
}

// UnimplementedParentalStatusViewServiceServer can be embedded to have forward compatible implementations.
type UnimplementedParentalStatusViewServiceServer struct {
}

func (*UnimplementedParentalStatusViewServiceServer) GetParentalStatusView(ctx context.Context, req *GetParentalStatusViewRequest) (*resources.ParentalStatusView, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetParentalStatusView not implemented")
}

func RegisterParentalStatusViewServiceServer(s *grpc.Server, srv ParentalStatusViewServiceServer) {
	s.RegisterService(&_ParentalStatusViewService_serviceDesc, srv)
}

func _ParentalStatusViewService_GetParentalStatusView_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetParentalStatusViewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ParentalStatusViewServiceServer).GetParentalStatusView(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.ads.googleads.v2.services.ParentalStatusViewService/GetParentalStatusView",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ParentalStatusViewServiceServer).GetParentalStatusView(ctx, req.(*GetParentalStatusViewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ParentalStatusViewService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.ads.googleads.v2.services.ParentalStatusViewService",
	HandlerType: (*ParentalStatusViewServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetParentalStatusView",
			Handler:    _ParentalStatusViewService_GetParentalStatusView_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/ads/googleads/v2/services/parental_status_view_service.proto",
}
