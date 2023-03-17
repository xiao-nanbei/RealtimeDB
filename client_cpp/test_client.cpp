//
// Created by databrains on 23-3-12.
//
#include <gtest/gtest.h>
#include <string>
#include <grpcpp/create_channel.h>
#include "client.cpp"
#include <fstream>
#include "iostream"
#include <mutex>
#include <thread>
#include <time.h>
#include <chrono>
std::mutex mtx;
void read_from_file(std::ifstream in_file,TestClientWritePoints* writePoints)
{
    std::lock_guard<std::mutex> lock(mtx);
    std::string line;
    while (std::getline(in_file, line))
    {
        writePoints->WritePoints(line);
    }
}
TEST(WritePointsTest,WritePoints){
    /*
    std::lock_guard<std::mutex> lock(mtx);
    TestClientConfig client(grpc::CreateChannel("localhost:8086",grpc::InsecureChannelCredentials()));
    EXPECT_EQ(true,client.Config("test"));
    TestClientWritePoints writePoints;
    writePoints.stub_=client.stub_;

    std::vector<std::string> metrics={"cpu.busy", "cpu.load1", "cpu.load5", "cpu.load15", "cpu.iowait","disk.write.ops", "disk.read.ops", "disk.used","net.in.bytes", "net.out.bytes", "net.in.packages", "net.out.packages","mem.used", "mem.idle", "mem.used.bytes", "mem.total.bytes"};
    std::vector<std::string> tags={"host","core"};
    long long start=1600000000000;
    std::fstream f;
    f.open("/home/databrains/data2.txt",std::ios::in);
    /*
    //输入你想写入的内容
    for(auto & metric : metrics){
        for (int j=0;j<1000;j++){
            for (int k=0;k<10;k++){
                for (int m=0;m<10;m++){
                    //std::string row="{\"row\":""\"{"+std::string(R"(\"metric\")")+":"+"\\\""+metric+"\\\""+","+"\\\""+tags[0]+"\\\""+":"+"\\\""+tags[0]+std::to_string(k)+"\\\""+","+"\\\""+tags[1]+"\\\""+":"+"\\\""+tags[1]+std::to_string(m)+"\\\""+","+"\\\""+"value"+"\\\""+":"+std::to_string(j)+","+R"(\"timestamp\")"+":"+std::to_string(start)+"}\"}";
                    std::string row="{"+std::string("\"metric\"")+":"+"\""+metric+"\""+","+"\""+tags[0]+"\""+":"+"\""+tags[0]+std::to_string(k)+"\""+","+"\""+tags[1]+"\""+":"+"\""+tags[1]+std::to_string(m)+"\""+","+"\""+"value"+"\""+":"+std::to_string(j)+","+"\"timestamp\""+":"+std::to_string(start)+"}";
                    f<<row<<std::endl;

                }
                //EXPECT_EQ(true,writePoints.WritePoints(row));
                start++;
            }

        }
    }

    std::string row;
    while(getline(f,row))
    {
        EXPECT_EQ(true,writePoints.WritePoints(row));
    }

    f.close();
     */
    TestClientConfig client(grpc::CreateChannel("localhost:8086",grpc::InsecureChannelCredentials()));
    //EXPECT_EQ(true,client.Config("./testdata","test"));
    TestClientWritePoints writePoints;
    writePoints.stub_=client.stub_;
    std::ifstream in_file("/home/databrains/data2.txt");
    for(int i=0;i<10;i++){
        std::thread write_thread(read_from_file,std::move(in_file),&writePoints);
        write_thread.join();
    }
    in_file.close();
}


int main(int argc, char **argv){
    printf("Running main() from %s\n", __FILE__);
    testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}