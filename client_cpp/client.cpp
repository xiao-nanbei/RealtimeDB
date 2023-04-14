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

class ClientConfig{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    // create stub
    ClientConfig(std::shared_ptr<grpc::Channel> channel):stub_(rpc::Greeter::NewStub(channel)){}
    bool Config(std::string name)
    {
        rpc::ConfigRequest configRequest;
        rpc::ConfigResponse configResponse;
        configRequest.set_name(name);
        if (GetOneData(configRequest,&configResponse)){
            return true;
        }else{
            return false;
        }

    }

private:
    bool GetOneData(const rpc::ConfigRequest& configRequest,rpc::ConfigResponse* configResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->Config(&context,configRequest,configResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(configResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<configResponse->reply()<<std::endl;
        }
        return true;
    }
};
/*
class ClientWritePoints{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool WritePoints(std::string row){
        rpc::WritePointsRequest writePointsRequest;
        writePointsRequest.set_row(row);
        AsyncClientCall* call = new AsyncClientCall;
        call->response_reader = stub_->AsyncWritePoints(&call->context, writePointsRequest, &cq_);
        call->response_reader->Finish(&call->reply, &call->status, (void*)call);
        return true;
    }
    void AsyncCompleteRpc() {
        void* got_tag;
        bool ok = false;

        // Block until the next result is available in the completion queue "cq".
        while (cq_.Next(&got_tag, &ok)) {
            // The tag in this example is the memory location of the call object
            AsyncClientCall* call = static_cast<AsyncClientCall*>(got_tag);

            // Verify that the request was completed successfully. Note that "ok"
            // corresponds solely to the request for updates introduced by Finish().
            GPR_ASSERT(ok);
            if (call->status.ok())
                std::cout << "Greeter received: " << call->reply.reply() << std::endl;
            else
                std::cout << "RPC failed" << std::endl;
            // Once we're complete, deallocate the call object.
            delete call;
        }
    }
private:
    struct AsyncClientCall {
        rpc::WritePointsResponse reply;

        // Context for the client. It could be used to convey extra information to
        // the server and/or tweak certain RPC behaviors.
        grpc::ClientContext context;

        // Storage for the status of the RPC upon completion.
        grpc::Status status;


        std::unique_ptr<grpc::ClientAsyncResponseReader<rpc::WritePointsResponse>> response_reader;
    };

    grpc::CompletionQueue cq_;

};
*/
class ClientWritePoints{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool WritePoints(std::string tags)
    {
        rpc::WritePointsRequest writePointsRequest;
        rpc::WritePointsResponse writePointsResponse;
        writePointsRequest.set_row(tags);
        if(GetOneData(writePointsRequest,&writePointsResponse)){
            return true;
        }else{
            return false;
        }
    }

private:
    bool GetOneData(const rpc::WritePointsRequest& writePointsRequest,rpc::WritePointsResponse* writePointsResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->WritePoints(&context,writePointsRequest,writePointsResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(writePointsResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<querySeriesResponse->reply()<<std::endl;
        }
        return true;
    }

};
class ClientQuerySeries{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool QuerySeries(std::string tags)
    {
        rpc::QuerySeriesRequest querySeriesRequest;
        rpc::QuerySeriesResponse querySeriesResponse;
        querySeriesRequest.set_tags(tags);
        if(GetOneData(querySeriesRequest,&querySeriesResponse)){
            return true;
        }else{
            return false;
        }
    }

private:
    bool GetOneData(const rpc::QuerySeriesRequest& querySeriesRequest,rpc::QuerySeriesResponse* querySeriesResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QuerySeries(&context,querySeriesRequest,querySeriesResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(querySeriesResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<querySeriesResponse->reply()<<std::endl;
        }
        return true;
    }

};
class ClientQueryNewPoint{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool QueryNewPoint(std::string tags)
    {
        rpc::QueryNewPointRequest queryNewPointRequest;
        rpc::QueryNewPointResponse queryNewPointResponse;
        queryNewPointRequest.set_tag(tags);
        if(GetOneData(queryNewPointRequest,&queryNewPointResponse)){
            return true;
        }else{
            return false;
        }
    }

private:
    bool GetOneData(const rpc::QueryNewPointRequest& queryNewPointRequest,rpc::QueryNewPointResponse* queryNewPointResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QueryNewPoint(&context,queryNewPointRequest,queryNewPointResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(queryNewPointResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<queryNewPointResponse->reply()<<std::endl;
        }
        return true;
    }

};

class ClientQueryRange{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool QueryRange(std::string tags)
    {
        rpc::QueryRangeRequest queryRangeRequest;
        rpc::QueryRangeResponse queryRangeResponse;
        queryRangeRequest.set_metric_tags(tags);
        if (GetOneData(queryRangeRequest,&queryRangeResponse)){
            return true;
        }else{
            return false;
        }
    }

private:
    bool GetOneData(const rpc::QueryRangeRequest& queryRangeRequest,rpc::QueryRangeResponse* queryRangeResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QueryRange(&context,queryRangeRequest,queryRangeResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(queryRangeResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<queryRangeResponse->reply()<<std::endl;
        }
        return true;
    }

};

class ClientQueryTagValues{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool QueryTagValues(std::string tags)
    {
        rpc::QueryTagValuesRequest queryTagValuesRequest;
        rpc::QueryTagValuesResponse queryTagValuesResponse;
        queryTagValuesRequest.set_tag(tags);
        if (GetOneData(queryTagValuesRequest,&queryTagValuesResponse)){
            return true;
        }else{
            return false;
        }
    }

private:
    bool GetOneData(const rpc::QueryTagValuesRequest& queryTagValuesRequest,rpc::QueryTagValuesResponse* queryTagValuesResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QueryTagValues(&context,queryTagValuesRequest,queryTagValuesResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(queryTagValuesResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<queryTagValuesResponse->reply()<<std::endl;
        }
        return true;
    }

};
class ClientQuerySeriesAllData{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool QuerySeriesAllData(std::string metric_tags)
    {
        rpc::QuerySeriesAllDataRequest querySeriesAllDataRequest;
        rpc::QuerySeriesAllDataResponse querySeriesAllDataResponse;
        querySeriesAllDataRequest.set_metric_tags(metric_tags);
        if (GetOneData(querySeriesAllDataRequest,&querySeriesAllDataResponse)){
            return true;
        }else{
            return false;
        }
    }

private:
    bool GetOneData(const rpc::QuerySeriesAllDataRequest& querySeriesAllDataRequest,rpc::QuerySeriesAllDataResponse* querySeriesAllDataResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QuerySeriesAllData(&context,querySeriesAllDataRequest,querySeriesAllDataResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(querySeriesAllDataResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<queryTagValuesResponse->reply()<<std::endl;
        }
        return true;
    }

};
class ClientQueryAllData{
public:
    std::unique_ptr<rpc::Greeter::Stub> stub_;
    bool QueryQueryAllData(std::string metric_tags)
    {
        google::protobuf::Empty request;
        rpc::QueryAllDataResponse queryAllDataResponse;
        if (GetOneData(request,&queryAllDataResponse)){
            return true;
        }else{
            return false;
        }
    }

private:
    bool GetOneData(google::protobuf::Empty &request,rpc::QueryAllDataResponse* queryAllDataResponse)
    {
        grpc::ClientContext context;
        grpc::Status status=stub_->QueryAllData(&context,request,queryAllDataResponse);
        if(!status.ok())
        {
            //std::cout<<"GetData rpc failed."<<std::endl;
            return false;
        }
        if(queryAllDataResponse->reply().empty())
        {
            //std::cout<<"message empty."<<std::endl;
            return false;
        }
        else
        {
            //std::cout<<"MsgReply:"<<queryTagValuesResponse->reply()<<std::endl;
        }
        return true;
    }

};