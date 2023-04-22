// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: washd_interface.proto

package washd

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// WashServiceClient is the client API for WashService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WashServiceClient interface {
	ListMachines(ctx context.Context, in *MachineListRequest, opts ...grpc.CallOption) (*MachineList, error)
}

type washServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWashServiceClient(cc grpc.ClientConnInterface) WashServiceClient {
	return &washServiceClient{cc}
}

func (c *washServiceClient) ListMachines(ctx context.Context, in *MachineListRequest, opts ...grpc.CallOption) (*MachineList, error) {
	out := new(MachineList)
	err := c.cc.Invoke(ctx, "/washd.WashService/ListMachines", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WashServiceServer is the server API for WashService service.
// All implementations must embed UnimplementedWashServiceServer
// for forward compatibility
type WashServiceServer interface {
	ListMachines(context.Context, *MachineListRequest) (*MachineList, error)
	mustEmbedUnimplementedWashServiceServer()
}

// UnimplementedWashServiceServer must be embedded to have forward compatible implementations.
type UnimplementedWashServiceServer struct {
}

func (UnimplementedWashServiceServer) ListMachines(context.Context, *MachineListRequest) (*MachineList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListMachines not implemented")
}
func (UnimplementedWashServiceServer) mustEmbedUnimplementedWashServiceServer() {}

// UnsafeWashServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WashServiceServer will
// result in compilation errors.
type UnsafeWashServiceServer interface {
	mustEmbedUnimplementedWashServiceServer()
}

func RegisterWashServiceServer(s grpc.ServiceRegistrar, srv WashServiceServer) {
	s.RegisterService(&WashService_ServiceDesc, srv)
}

func _WashService_ListMachines_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MachineListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WashServiceServer).ListMachines(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/washd.WashService/ListMachines",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WashServiceServer).ListMachines(ctx, req.(*MachineListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WashService_ServiceDesc is the grpc.ServiceDesc for WashService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WashService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "washd.WashService",
	HandlerType: (*WashServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListMachines",
			Handler:    _WashService_ListMachines_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "washd_interface.proto",
}
