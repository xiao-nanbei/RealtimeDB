cmake_minimum_required(VERSION 3.25)
project(client_cc)

find_package(Threads REQUIRED)
find_package(GTest REQUIRED)
include_directories(${GTEST_INCLUDE_DIRS})
find_package(Protobuf CONFIG REQUIRED)
set(_PROTOBUF_LIBPROTOBUF protobuf::libprotobuf)
set(_REFLECTION gRPC::grpc++_reflection)
find_package(gRPC CONFIG REQUIRED)
set(_GRPC_GRPCPP gRPC::grpc++)
set(CMAKE_CXX_STANDARD 17)

include_directories("${CMAKE_CURRENT_BINARY_DIR}/../client_cc/rpc")
set(rpc_proto_srcs "${CMAKE_CURRENT_BINARY_DIR}/../client_cc/rpc/client.pb.cc")
set(rpc_proto_hdrs "${CMAKE_CURRENT_BINARY_DIR}/../client_cc/rpc/client.pb.h")
set(rpc_grpc_srcs "${CMAKE_CURRENT_BINARY_DIR}/../client_cc/rpc/client.grpc.pb.cc")
set(rpc_grpc_hdrs "${CMAKE_CURRENT_BINARY_DIR}/../client_cc/rpc/client.grpc.pb.h" client.cpp)

add_library(rpc_grpc_proto
        ${rpc_grpc_srcs}
        ${rpc_grpc_hdrs}
        ${rpc_proto_srcs}
        ${rpc_proto_hdrs})
target_link_libraries(rpc_grpc_proto
        ${_REFLECTION}
        ${_GRPC_GRPCPP}
        ${_PROTOBUF_LIBPROTOBUF}
        )
foreach(_target
        test_client
        )
    add_executable(${_target} "${_target}.cpp")
    target_link_libraries(${_target}
            rpc_grpc_proto
            ${_REFLECTION}
            ${_GRPC_GRPCPP}
            ${_PROTOBUF_LIBPROTOBUF}
            ${GTEST_BOTH_LIBRARIES}
            )
    gtest_discover_tests(${_target})
endforeach()
