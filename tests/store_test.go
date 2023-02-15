package tests
import (
	"RealtimeDB/rtdb"
	"fmt"
	"github.com/satori/go.uuid"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// 模拟一些监控指标
var Metrics = []string{
	"cpu.busy", "cpu.load1", "cpu.load5", "cpu.load15", "cpu.iowait",
	"disk.write.ops", "disk.read.ops", "disk.used",
	"net.in.bytes", "net.out.bytes", "net.in.packages", "net.out.packages",
	"mem.used", "mem.idle", "mem.used.bytes", "mem.total.bytes",
}

// 增加 Tag 数量
var uid1, uid2, uid3 []string

func init() {
	for i := 0; i < len(Metrics); i++ {
		uid1 = append(uid1, uuid.NewV4().String())
		uid2 = append(uid2, uuid.NewV4().String())
		uid3 = append(uid3, uuid.NewV4().String())
	}
}

func GenPoints(ts int64, node, dc int) []*rtdb.Row {
	points := make([]*rtdb.Row, 0)
	for idx, metric := range Metrics {
		points = append(points, &rtdb.Row{
			Metric: metric,
			Tags: []rtdb.Tag{
				{Name: "node", Value: "vm" + strconv.Itoa(node)},
				{Name: "dc", Value: strconv.Itoa(dc)},
				{Name: "foo", Value: uid1[idx]},
				{Name: "bar", Value: uid2[idx]},
				{Name: "zoo", Value: uid3[idx]},
			},
			Point: rtdb.Point{TimeStamp: ts, Value: float64(rand.Int31n(60))},
		})
	}

	return points
}

func Test_STORE(t *testing.T) {
	store := rtdb.OpenRTDB()
	defer store.Close()

	now := time.Now().Unix() - 36000 // 10h ago

	for i := 0; i < 720; i++ {
		for n := 0; n < 5; n++ {
			for j := 0; j < 1024; j++ {
				_ = store.InsertRows(GenPoints(now, n, j))
			}
		}

		now += 60 //1min
	}

	fmt.Println("finished")

	select {}
}
