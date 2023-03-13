//
// Created by databrains on 23-3-13.
//
#include <gtest/gtest.h>
#include <grpcpp/channel.h>
#include <grpcpp/client_context.h>
#include <grpcpp/security/credentials.h>
#include "./rpc/client.grpc.pb.h"
#include <iostream>
#include <memory>
#include <string>

class TestClientConfig{
public:
    std::shared_ptr<rpc::Greeter::Stub> stub_;
    // create stub
    TestClientConfig(std::shared_ptr<grpc::Channel> channel):stub_(rpc::Greeter::NewStub(channel)){}
    void Config()
    {
        rpc::ConfigRequest configRequest;
        rpc::ConfigResponse configResponse;
        configRequest.set_path("./testdata");
        configRequest.set_name("test");
        GetOneData(configRequest,&configResponse);
    }

private:
    bool GetOneData(const rpc::ConfigRequest& configRequest,rpc::ConfigResponse* configResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->Config(&context,configRequest,configResponse);
        if(!status.ok())
        {
            std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(configResponse->reply().empty())
        {
            std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            std::cout<<"MsgReply:"<<configResponse->reply()<<std::endl;
        }
        return true;
    }
};
class TestClientWritePoints{
public:
    std::shared_ptr<rpc::Greeter::Stub> stub_;
    void WritePoints(std::string row)
    {
        rpc::WritePointsRequest writePointsRequest;
        rpc::WritePointsResponse writePointsResponse;
        writePointsRequest.set_row(row);
        GetOneData(writePointsRequest,&writePointsResponse);
    }

private:
    bool GetOneData(const rpc::WritePointsRequest& writePointsRequest,rpc::WritePointsResponse* writePointsResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->WritePoints(&context,writePointsRequest,writePointsResponse);
        if(!status.ok())
        {
            std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(writePointsResponse->reply().empty())
        {
            std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            std::cout<<"MsgReply:"<<writePointsResponse->reply()<<std::endl;
        }
        return true;
    }

};

class TestClientQuerySeries{
public:
    std::shared_ptr<rpc::Greeter::Stub> stub_;
    void QuerySeries(std::string tags)
    {
        rpc::QuerySeriesRequest querySeriesRequest;
        rpc::QuerySeriesResponse querySeriesResponse;
        querySeriesRequest.set_tags(tags);
        GetOneData(querySeriesRequest,&querySeriesResponse);
    }

private:
    bool GetOneData(const rpc::QuerySeriesRequest& querySeriesRequest,rpc::QuerySeriesResponse* querySeriesResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QuerySeries(&context,querySeriesRequest,querySeriesResponse);
        if(!status.ok())
        {
            std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(querySeriesResponse->reply().empty())
        {
            std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            std::cout<<"MsgReply:"<<querySeriesResponse->reply()<<std::endl;
        }
        return true;
    }

};
class TestClientQueryNewPoint{
public:
    std::shared_ptr<rpc::Greeter::Stub> stub_;
    void QueryNewPoint(std::string tags)
    {
        rpc::QueryNewPointRequest queryNewPointRequest;
        rpc::QueryNewPointResponse queryNewPointResponse;
        queryNewPointRequest.set_tag(tags);
        GetOneData(queryNewPointRequest,&queryNewPointResponse);
    }

private:
    bool GetOneData(const rpc::QueryNewPointRequest& queryNewPointRequest,rpc::QueryNewPointResponse* queryNewPointResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QueryNewPoint(&context,queryNewPointRequest,queryNewPointResponse);
        if(!status.ok())
        {
            std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(queryNewPointResponse->reply().empty())
        {
            std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            std::cout<<"MsgReply:"<<queryNewPointResponse->reply()<<std::endl;
        }
        return true;
    }

};

class TestClientQueryRange{
public:
    std::shared_ptr<rpc::Greeter::Stub> stub_;
    void QueryRange(std::string tags)
    {
        rpc::QueryRangeRequest queryRangeRequest;
        rpc::QueryRangeResponse queryRangeResponse;
        queryRangeRequest.set_metric_tags(tags);
        GetOneData(queryRangeRequest,&queryRangeResponse);
    }

private:
    bool GetOneData(const rpc::QueryRangeRequest& queryRangeRequest,rpc::QueryRangeResponse* queryRangeResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QueryRange(&context,queryRangeRequest,queryRangeResponse);
        if(!status.ok())
        {
            std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(queryRangeResponse->reply().empty())
        {
            std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            std::cout<<"MsgReply:"<<queryRangeResponse->reply()<<std::endl;
        }
        return true;
    }

};

class TestClientQueryTagValues{
public:
    std::shared_ptr<rpc::Greeter::Stub> stub_;
    void QueryTagValues(std::string tags)
    {
        rpc::QueryTagValuesRequest queryTagValuesRequest;
        rpc::QueryTagValuesResponse queryTagValuesResponse;
        queryTagValuesRequest.set_tag(tags);
        GetOneData(queryTagValuesRequest,&queryTagValuesResponse);
    }

private:
    bool GetOneData(const rpc::QueryTagValuesRequest& queryTagValuesRequest,rpc::QueryTagValuesResponse* queryTagValuesResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QueryTagValues(&context,queryTagValuesRequest,queryTagValuesResponse);
        if(!status.ok())
        {
            std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(queryTagValuesResponse->reply().empty())
        {
            std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            std::cout<<"MsgReply:"<<queryTagValuesResponse->reply()<<std::endl;
        }
        return true;
    }

};
