package tests

import (
	"RealtimeDB/rtdb"
	"github.com/chenjiandongx/logger"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)


// 模拟一些监控指标
var metrics = []string{
	"cpu.busy", "cpu.load1", "cpu.load5", "cpu.load15", "cpu.iowait",
	"disk.write.ops", "disk.read.ops", "disk.used",
	"net.in.bytes", "net.out.bytes", "net.in.packages", "net.out.packages",
	"mem.used", "mem.idle", "mem.used.bytes", "mem.total.bytes",
}

func genPoints(ts int64, node, dc int) []*rtdb.Row {
	points := make([]*rtdb.Row, 0)
	for _, metric := range metrics {
		points = append(points, &rtdb.Row{
			Metric: metric,
			Tags: []rtdb.Tag{
				{Name: "node", Value: "vm" + strconv.Itoa(node)},
				{Name: "dc", Value: strconv.Itoa(dc)},
			},
			Point: rtdb.Point{TimeStamp: ts, Value: float64(ts)},
		})
	}

	return points
}

func TestRTDB_QueryRange(t *testing.T) {
	tmpdir := "/tmp/rtdb8888888"

	store := rtdb.OpenRTDB(rtdb.WithDataPath(tmpdir), rtdb.WithLoggerConfig(&logger.Options{
		Stdout:      true,
		ConsoleMode: true,
		Level:       logger.ErrorLevel,
	}))
	defer store.Close()
	defer os.RemoveAll(tmpdir)

	var start int64 = 1600000000000

	var now = start
	for i := 0; i < 720; i++ {
		for n := 0; n < 3; n++ {
			for j := 0; j < 24; j++ {
				_ = store.InsertRows(genPoints(now, n, j))
			}
		}

		now += 60 //1min
	}

	time.Sleep(time.Millisecond * 20)

	ret, err := store.QueryRange("cpu.busy", rtdb.TagMatcherSet{
		{Name: "node", Value: "vm1"},
		{Name: "dc", Value: "0"},
	}, start, start+120)
	assert.NoError(t, err)

	ret[0].Tags.Sorted()
	labels := rtdb.TagSet{
		{"__name__", "cpu.busy"},
		{"dc", "0"},
		{"node", "vm1"},
	}
	assert.Equal(t, ret[0].Tags, labels)

	values := []int64{start, start + 60, start + 120}

	for idx, d := range ret[0].Points {
		assert.Equal(t, d.TimeStamp, values[idx])
		assert.Equal(t, d.Value, float64(values[idx]))
	}

	ret, err = store.QueryRange("cpu.busy", rtdb.TagMatcherSet{
		{Name: "node", Value: "vm1"},
		{Name: "dc", Value: "0"},
	}, now-120, now)
	assert.NoError(t, err)
	assert.Equal(t, len(ret[0].Points), 2)
}

func TestRTDB_QuerySeries(t *testing.T) {
	tmpdir := "/tmp/rtdb6"

	store := rtdb.OpenRTDB(rtdb.WithDataPath(tmpdir))
	defer store.Close()
	defer os.RemoveAll(tmpdir)

	var start int64 = 1600000000000

	var now = start
	for i := 0; i < 720; i++ {
		for n := 0; n < 3; n++ {
			for j := 0; j < 24; j++ {
				_ = store.InsertRows(genPoints(now, n, j))
			}
		}

		now += 60 //1min
	}

	time.Sleep(time.Millisecond * 20)

	ret, err := store.QuerySeries(rtdb.TagMatcherSet{
		{Name: "__name__", Value: "disk.*", IsRegx: true},
		{Name: "node", Value: "vm1"},
		{Name: "dc", Value: "0"},
	}, start, start+120)
	assert.NoError(t, err)
	assert.Equal(t, len(ret), 3)
}

func TestRTDB_QueryTagValues(t *testing.T) {
	tmpdir := "/tmp/rtdb8"

	store := rtdb.OpenRTDB(rtdb.WithDataPath(tmpdir))
	defer store.Close()
	defer os.RemoveAll(tmpdir)

	var start int64 = 1600000000000

	var now = start
	for i := 0; i < 720; i++ {
		for n := 0; n < 3; n++ {
			for j := 0; j < 24; j++ {
				_ = store.InsertRows(genPoints(now, n, j))
			}
		}

		now += 60 //1min
	}

	time.Sleep(time.Millisecond * 20)

	ret := store.QueryTagValues("node", start, start+120)
	assert.Equal(t, ret, []string{"vm0", "vm1", "vm2"})
}
