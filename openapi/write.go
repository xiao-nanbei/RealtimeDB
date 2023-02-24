package openapi

import (
	"RealtimeDB/rtdb"
	"time"
)

func Write(datas []float64,strings []string)error{
	Config("./testdata")
	now:=time.Now().UnixMilli()
	points := make([]*rtdb.Row, 0)
	tags:=make([]rtdb.Tag,0)
	for i:=2;i<len(strings)-1;i+=2{
		tags=append(tags,rtdb.Tag{Name: strings[i], Value: strings[i+1]})
	}
	for _,data:=range datas{
		points = append(points, &rtdb.Row{
			Metric: strings[1],
			Tags: tags,
			Point: rtdb.Point{TimeStamp: now, Value: data},
		})
	}
	err := Store.InsertRows(points)
	if err != nil {
		return err
	}
	return nil
}
