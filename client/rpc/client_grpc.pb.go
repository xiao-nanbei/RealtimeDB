// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.6
// source: rpc/client.proto

package rpc

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

// GreeterClient is the client API for Greeter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GreeterClient interface {
	WritePoints(ctx context.Context, in *WritePointsRequest, opts ...grpc.CallOption) (*WritePointsResponse, error)
	QuerySeries(ctx context.Context, in *QuerySeriesRequest, opts ...grpc.CallOption) (*QuerySeriesResponse, error)
	Config(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error)
	QueryRange(ctx context.Context, in *QueryRangeRequest, opts ...grpc.CallOption) (*QueryRangeResponse, error)
	QueryTagValues(ctx context.Context, in *QueryTagValuesRequest, opts ...grpc.CallOption) (*QueryTagValuesResponse, error)
}

type greeterClient struct {
	cc grpc.ClientConnInterface
}

func NewGreeterClient(cc grpc.ClientConnInterface) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) WritePoints(ctx context.Context, in *WritePointsRequest, opts ...grpc.CallOption) (*WritePointsResponse, error) {
	out := new(WritePointsResponse)
	err := c.cc.Invoke(ctx, "/rpc.Greeter/WritePoints", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greeterClient) QuerySeries(ctx context.Context, in *QuerySeriesRequest, opts ...grpc.CallOption) (*QuerySeriesResponse, error) {
	out := new(QuerySeriesResponse)
	err := c.cc.Invoke(ctx, "/rpc.Greeter/QuerySeries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greeterClient) Config(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error) {
	out := new(ConfigResponse)
	err := c.cc.Invoke(ctx, "/rpc.Greeter/Config", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greeterClient) QueryRange(ctx context.Context, in *QueryRangeRequest, opts ...grpc.CallOption) (*QueryRangeResponse, error) {
	out := new(QueryRangeResponse)
	err := c.cc.Invoke(ctx, "/rpc.Greeter/QueryRange", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greeterClient) QueryTagValues(ctx context.Context, in *QueryTagValuesRequest, opts ...grpc.CallOption) (*QueryTagValuesResponse, error) {
	out := new(QueryTagValuesResponse)
	err := c.cc.Invoke(ctx, "/rpc.Greeter/QueryTagValues", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GreeterServer is the server API for Greeter service.
// All implementations must embed UnimplementedGreeterServer
// for forward compatibility
type GreeterServer interface {
	WritePoints(context.Context, *WritePointsRequest) (*WritePointsResponse, error)
	QuerySeries(context.Context, *QuerySeriesRequest) (*QuerySeriesResponse, error)
	Config(context.Context, *ConfigRequest) (*ConfigResponse, error)
	QueryRange(context.Context, *QueryRangeRequest) (*QueryRangeResponse, error)
	QueryTagValues(context.Context, *QueryTagValuesRequest) (*QueryTagValuesResponse, error)
	mustEmbedUnimplementedGreeterServer()
}

// UnimplementedGreeterServer must be embedded to have forward compatible implementations.
type UnimplementedGreeterServer struct {
}

func (UnimplementedGreeterServer) WritePoints(context.Context, *WritePointsRequest) (*WritePointsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WritePoints not implemented")
}
func (UnimplementedGreeterServer) QuerySeries(context.Context, *QuerySeriesRequest) (*QuerySeriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QuerySeries not implemented")
}
func (UnimplementedGreeterServer) Config(context.Context, *ConfigRequest) (*ConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Config not implemented")
}
func (UnimplementedGreeterServer) QueryRange(context.Context, *QueryRangeRequest) (*QueryRangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryRange not implemented")
}
func (UnimplementedGreeterServer) QueryTagValues(context.Context, *QueryTagValuesRequest) (*QueryTagValuesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTagValues not implemented")
}
func (UnimplementedGreeterServer) mustEmbedUnimplementedGreeterServer() {}

// UnsafeGreeterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GreeterServer will
// result in compilation errors.
type UnsafeGreeterServer interface {
	mustEmbedUnimplementedGreeterServer()
}

func RegisterGreeterServer(s grpc.ServiceRegistrar, srv GreeterServer) {
	s.RegisterService(&Greeter_ServiceDesc, srv)
}

func _Greeter_WritePoints_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WritePointsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).WritePoints(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Greeter/WritePoints",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).WritePoints(ctx, req.(*WritePointsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Greeter_QuerySeries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySeriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).QuerySeries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Greeter/QuerySeries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).QuerySeries(ctx, req.(*QuerySeriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Greeter_Config_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).Config(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Greeter/Config",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).Config(ctx, req.(*ConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Greeter_QueryRange_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRangeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).QueryRange(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Greeter/QueryRange",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).QueryRange(ctx, req.(*QueryRangeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Greeter_QueryTagValues_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTagValuesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).QueryTagValues(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.Greeter/QueryTagValues",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).QueryTagValues(ctx, req.(*QueryTagValuesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Greeter_ServiceDesc is the grpc.ServiceDesc for Greeter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Greeter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.Greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WritePoints",
			Handler:    _Greeter_WritePoints_Handler,
		},
		{
			MethodName: "QuerySeries",
			Handler:    _Greeter_QuerySeries_Handler,
		},
		{
			MethodName: "Config",
			Handler:    _Greeter_Config_Handler,
		},
		{
			MethodName: "QueryRange",
			Handler:    _Greeter_QueryRange_Handler,
		},
		{
			MethodName: "QueryTagValues",
			Handler:    _Greeter_QueryTagValues_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc/client.proto",
}
