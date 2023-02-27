package openapi

import (
	"RealtimeDB/rpc"
	"RealtimeDB/rtdb"
	"context"
	"encoding/json"
)

func (s *Server) QuerySeries(ctx context.Context, in *rpc.QuerySeriesRequest)(*rpc.QuerySeriesResponse, error){
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
		tags=append(tags,rtdb.TagMatcher{Name: k,Value: v.(string)})
	}
	ret, err := Store.QuerySeries(tags, start,end)
	if err != nil {
		return nil, err
	}
	retstring, _ := json.Marshal(ret)
	return &rpc.QuerySeriesResponse{Reply: string(retstring)},nil
}
func (s *Server) QueryRange(ctx context.Context, in *rpc.QueryRangeRequest)(*rpc.QueryRangeResponse, error){
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
		tags=append(tags,rtdb.TagMatcher{Name: k,Value: v.(string)})
	}

	ret, err := Store.QueryRange(metric,tags,start,end)
	if err != nil {
		return nil, err
	}
	retstring, _ := json.Marshal(ret)
	return &rpc.QueryRangeResponse{Reply: string(retstring)},nil
}
func (s *Server) QueryTagValues(ctx context.Context, in *rpc.QueryTagValuesRequest)(*rpc.QueryTagValuesResponse, error){
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
	ret := Store.QueryTagValues(tag, start,end)
	retstring, _ := json.Marshal(ret)
	return &rpc.QueryTagValuesResponse{Reply: string(retstring)},nil
}