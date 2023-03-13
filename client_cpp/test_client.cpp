//
// Created by databrains on 23-3-12.
//
#include <gtest/gtest.h>
#include <grpc/grpc.h>
#include <grpcpp/channel.h>
#include <grpcpp/client_context.h>
#include <grpcpp/create_channel.h>
#include <grpcpp/security/credentials.h>
#include <iostream>
#include <memory>
#include <string>
#include <vector>
#include "client.cpp"

int main(){
    TestClientConfig client(grpc::CreateChannel("localhost:8086",grpc::InsecureChannelCredentials()));
    std::string path = "./testdata";
    std::string name = "test";
    client.Config(path,name);
    TestClientWritePoints writePoints;
    writePoints.stub_=client.stub_;
    std::vector<std::string> metrics={"cpu.busy", "cpu.load1", "cpu.load5", "cpu.load15", "cpu.iowait","disk.write.ops", "disk.read.ops", "disk.used","net.in.bytes", "net.out.bytes", "net.in.packages", "net.out.packages","mem.used", "mem.idle", "mem.used.bytes", "mem.total.bytes"};
    std::vector<std::string> tags={"host","core"};
    long long start=1600000000000;
    for(auto & metric : metrics){
        for (int j=0;j<100000;j++){
            for (int k=0;k<10;k++){
                std::string row="{"+std::string("\"metric\"")+":"+"\""+metric+"\""+","+"\""+tags[0]+"\""+":"+"\""+tags[0]+std::to_string(k)+"\""+","+"\""+"value"+"\""+":"+std::to_string(j)+","+"\"timestamp\""+":"+std::to_string(start)+"}";
                writePoints.WritePoints(row);
            }
            start++;
        }
    }
}