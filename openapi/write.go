package openapi

import (
	"RealtimeDB/rtdb"
	"strconv"
	"time"
)

func Write(datas []float64,host string,core string,process string)error{
	points := make([]*rtdb.Row, 0)
	for number,data:=range datas{
		points = append(points, &rtdb.Row{
			Metric: "ADDATA",
			Tags: []rtdb.Tag{
				{Name: "host", Value: host},
				{Name: "core", Value: core},
				{Name: "process", Value: process},
				{Name: "ad", Value: "ad" + strconv.Itoa(number)},
			},
			Point: rtdb.Point{TimeStamp: time.Now().UnixMilli(), Value: data},
		})
	}
	err := Store.InsertRows(points)
	if err != nil {
		return err
	}
	return nil
}
