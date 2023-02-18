package tests
import (
	"RealtimeDB/rtdb"
	"fmt"
	"github.com/satori/go.uuid"
	"math/rand"
	"testing"
	"time"
)

// 模拟一些监控指标
var m = []string{
	"cpu.busy",
}

// 增加 Tag 数量
var tags1, tags2, tags3 []string

func init() {
	for i := 0; i < len(m); i++ {
		tags1 = append(tags1, uuid.NewV4().String())
		tags2 = append(tags2, uuid.NewV4().String())
		tags3 = append(tags3, uuid.NewV4().String())
	}
}

func Points(ts int64) []*rtdb.Row {
	points := make([]*rtdb.Row, 0)//init point
	for idx, metric := range m {
		points = append(points, &rtdb.Row{
			Metric: metric,
			Tags: []rtdb.Tag{
				{Name: "foo", Value: tags1[idx]},
				{Name: "bar", Value: tags2[idx]},
				{Name: "zoo", Value: tags3[idx]},
			},
			Point: rtdb.Point{TimeStamp: ts, Value: float64(rand.Int31n(60))},
		})
	}

	return points
}

func Test_WRITE_READ(t *testing.T) {
	store := rtdb.OpenRTDB()
	defer store.Close()
	for i := 0; i < 36000; i++ {
		_ = store.InsertRows(Points(time.Now().UnixMilli()))
	}
	fmt.Println("finished")

	select {}
}

