package openapi

import (
	"RealtimeDB/rpc"
	"RealtimeDB/rtdb"
	"context"
	"encoding/json"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)
func (s *Server) QueryNewPoint(ctx context.Context,in *rpc.QueryNewPointRequest)(*rpc.QueryNewPointResponse,error){
	p, _ := peer.FromContext(ctx)
	var slice map[string]interface{}
	err := json.Unmarshal([]byte(in.Tag), &slice)
	if err != nil {
		return nil, err
	}
	metric:=slice["metric"].(string)
	delete(slice,"metric")
	var tags rtdb.TagMatcherSet
	for k,v:=range slice{
		tags=append(tags,rtdb.TagMatcher{Name: k,Value: v.(string),IsRegx: true})
	}
	ret, err := Store[Aps[p.Addr.String()]].QueryNewPoint(metric,tags)
	if err != nil {
		return nil, err
	}
	retString, _ := json.Marshal(ret)
	return &rpc.QueryNewPointResponse{Reply: string(retString)},nil
}
func (s *Server) QuerySeries(ctx context.Context, in *rpc.QuerySeriesRequest)(*rpc.QuerySeriesResponse, error){
	p, _ := peer.FromContext(ctx)
	var slice map[string]interface{}
	err := json.Unmarshal([]byte(in.Tags), &slice)
	if err != nil {
		return nil, err
	}
	start:=int64(slice["start"].(float64))
	end:=int64(slice["end"].(float64))
	delete(slice,"start")
	delete(slice,"end")
	var tags rtdb.TagMatcherSet
	for k,v:=range slice{
		tags=append(tags,rtdb.TagMatcher{Name: k,Value: v.(string),IsRegx: true})
		log.Println(k,v)
	}
	ret, err := Store[Aps[p.Addr.String()]].QuerySeries(tags, start,end)
	if err != nil {
		return nil, err
	}
	retString, _ := json.Marshal(ret)
	return &rpc.QuerySeriesResponse{Reply: string(retString)},nil
}
func (s *Server) QueryRange(ctx context.Context, in *rpc.QueryRangeRequest)(*rpc.QueryRangeResponse, error){
	p, _ := peer.FromContext(ctx)
	var slice map[string]interface{}
	err := json.Unmarshal([]byte(in.MetricTags), &slice)
	if err != nil {
		return nil, err
	}
	metric:=slice["metric"].(string)
	start:=int64(slice["start"].(float64))
	end:=int64(slice["end"].(float64))
	delete(slice,"metric")
	delete(slice,"start")
	delete(slice,"end")
	var tags rtdb.TagMatcherSet
	for k,v:=range slice{
		tags=append(tags,rtdb.TagMatcher{Name: k,Value: v.(string),IsRegx: true})
	}

	ret, err := Store[Aps[p.Addr.String()]].QueryRange(metric,tags,start,end)
	if err != nil {
		return nil, err
	}
	retString, _ := json.Marshal(ret)
	return &rpc.QueryRangeResponse{Reply: string(retString)},nil
}
func (s *Server) QueryTagValues(ctx context.Context, in *rpc.QueryTagValuesRequest)(*rpc.QueryTagValuesResponse, error){
	p, _ := peer.FromContext(ctx)
	var slice map[string]interface{}
	err := json.Unmarshal([]byte(in.Tag), &slice)
	if err != nil {
		return nil, err
	}
	tag:=slice["tag"].(string)
	start:=int64(slice["start"].(float64))
	end:=int64(slice["end"].(float64))
	delete(slice,"start")
	delete(slice,"end")
	delete(slice,"tag")
	ret := Store[Aps[p.Addr.String()]].QueryTagValues(tag, start,end)
	retString, _ := json.Marshal(ret)
	return &rpc.QueryTagValuesResponse{Reply: string(retString)},nil
}
func (s *Server) QuerySeriesAllData(ctx context.Context, in *rpc.QuerySeriesAllDataRequest)(*rpc.QuerySeriesAllDataResponse, error){
	p, _ := peer.FromContext(ctx)
	var slice map[string]interface{}
	err := json.Unmarshal([]byte(in.MetricTags), &slice)
	if err != nil {
		return nil, err
	}
	metric:=slice["metric"].(string)
	delete(slice,"metric")
	delete(slice,"start")
	delete(slice,"end")
	var tags rtdb.TagMatcherSet
	for k,v:=range slice{
		tags=append(tags,rtdb.TagMatcher{Name: k,Value: v.(string),IsRegx: true})
	}
	ret, err := Store[Aps[p.Addr.String()]].QueryRange(metric,tags,0,3200000000000)
	if err != nil {
		return nil, err
	}
	retString, _ := json.Marshal(ret)
	return &rpc.QuerySeriesAllDataResponse{Reply: string(retString)},nil
}
func (s *Server) QueryAllData(ctx context.Context,empty *emptypb.Empty)(*rpc.QueryAllDataResponse, error){
	p, _ := peer.FromContext(ctx)
	rets, err := Store[Aps[p.Addr.String()]].QuerySeries(rtdb.TagMatcherSet{
		{Name: "metric", Value: "(.*?)", IsRegx: true},
	},0,3200000000000)
	if err != nil {
		return nil, err
	}
	var datasString string
	for _,ret:=range rets{
		metric:=ret["metric"]
		var tags rtdb.TagMatcherSet
		for k,v:=range ret{
			tags=append(tags,rtdb.TagMatcher{Name: k,Value: v,IsRegx: true})
		}
		data, err := Store[Aps[p.Addr.String()]].QueryRange(metric,tags,0,3200000000000)
		if err != nil {
			return nil, err
		}
		dataBytes, _ := json.Marshal(data)
		dataString:=string(dataBytes)

		datasString+=dataString
	}
	return &rpc.QueryAllDataResponse{Reply: datasString},nil
}