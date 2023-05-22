package simple8b

import (
	"fmt"
	"testing"
	"time"
)

func TestMarshalBinary(t *testing.T) {
	s1 := New(TwoHoursData[0].T)
	for _, p := range TwoHoursData {
		s1.Push(p.T, p.V)
	}
	it1 := s1.Iter()
	it1.Next()
	b, err := s1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	s2 := New(s1.T0)
	err = s2.UnmarshalBinary(b)
	if err != nil {
		t.Error(err)
	}
	it := s2.Iter()
	for _, w := range TwoHoursData {
		if !it.Next() {
			t.Fatalf("Next()=false, want true")
		}
		tt, vv := it.Values()
		// t.Logf("it.Values()=(%+v, %+v)\n", time.Unix(int64(tt), 0), vv)
		if w.T != tt || w.V != vv {
			t.Errorf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.T, w.V)
		}
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	var err error
	b.StopTimer()
	s1 := New(TwoHoursData[0].T)
	for _, p := range TwoHoursData {
		s1.Push(p.T, p.V)
	}
	s1.Finish()
	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err = s1.MarshalBinary()
	}
	if err != nil {
		b.Errorf("Unexpected error: %v\n", err)
	}
}

func BenchmarkUnmarshalBinary(b *testing.B) {
	var err error
	b.StopTimer()
	s1 := New(TwoHoursData[0].T)
	for _, p := range TwoHoursData {
		s1.Push(p.T, p.V)
	}
	s1.Finish()
	buf, err := s1.MarshalBinary()
	if err != nil {
		b.Error(err)
	}
	b.ReportAllocs()
	s2 := New(s1.T0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err = s2.UnmarshalBinary(buf)
	}
	if err != nil {
		b.Errorf("Unexpected error: %v\n", err)
	}
}
func TestExampleEncoding_Decoding(t *testing.T) {

	// Example from the paper
	t0, _ := time.ParseInLocation("Jan _2 2006 15:04:05", "Mar 24 2015 02:00:00", time.Local)
	tunix := t0.Unix()
	tunixUint64 :=uint64(tunix)
	s := New(tunixUint64)
	tunix += 62
	s.Push(uint64(tunix), 12)
	tunix += 60
	s.Push(uint64(tunix), 12)
	tunix += 60
	s.Push(uint64(tunix), 24)
	// extra tests

	// floating point masking/shifting bug
	tunix += 60
	s.Push(uint64(tunix), 13)
	tunix += 60
	s.Push(uint64(tunix), 24)

	// delta-of-delta sizes
	tunix += 300 // == delta-of-delta of 240
	s.Push(uint64(tunix), 24)
	tunix += 900 // == delta-of-delta of 600
	s.Push(uint64(tunix), 24)
	tunix += 900 + 2050 // == delta-of-delta of 600
	s.Push(uint64(tunix), 24)
	s.Finish()
	it := s.Iter()

	want := []struct {
		t uint64
		v float64
	}{
		{tunixUint64 + 62, 12},
		{tunixUint64 + 122, 12},
		{tunixUint64 + 182, 24},

		{tunixUint64 + 242, 13},
		{tunixUint64 + 302, 24},

		{tunixUint64 + 602, 24},
		{tunixUint64 + 1502, 24},
		{tunixUint64 + 4452, 24},
	}

	for _, w := range want {
		if !it.Next() {
			t.Fatalf("Next()=false, want true")
		}
		tt, vv := it.Values()
		if w.t != tt || w.v != vv {
			t.Errorf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.t, w.v)
		}
		fmt.Printf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.t, w.v)
	}

	if it.Next() {
		t.Fatalf("Next()=true, want false")
	}

	if err := it.Err(); err != nil {
		t.Errorf("it.Err()=%v, want nil", err)
	}
}

func TestRoundtrip(t *testing.T) {

	s := New(uint64(TwoHoursData[0].T))
	for _, p := range TwoHoursData {
		s.Push(uint64(p.T), p.V)
	}

	it := s.Iter()
	for _, w := range TwoHoursData {
		if !it.Next() {
			t.Fatalf("Next()=false, want true")
		}
		tt, vv := it.Values()
		// t.Logf("it.Values()=(%+v, %+v)\n", time.Unix(int64(tt), 0), vv)
		if w.T != tt || w.V != vv {
			t.Errorf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.T, w.V)
		}
	}

	if it.Next() {
		t.Fatalf("Next()=true, want false")
	}

	if err := it.Err(); err != nil {
		t.Errorf("it.Err()=%v, want nil", err)
	}
}

func TestConcurrentRoundtripImmediateWrites(t *testing.T) {
	testConcurrentRoundtrip(t, time.Duration(0))
}
func TestConcurrentRoundtrip1MsBetweenWrites(t *testing.T) {
	testConcurrentRoundtrip(t, time.Millisecond)
}
func TestConcurrentRoundtrip10MsBetweenWrites(t *testing.T) {
	testConcurrentRoundtrip(t, 10*time.Millisecond)
}

// Test reading while writing at the same time.
func testConcurrentRoundtrip(t *testing.T, sleep time.Duration) {
	s := New(uint64(TwoHoursData[0].T))

	//notify the reader about the number of points that have been written.
	writeNotify := make(chan int)

	// notify the reader when we have finished.
	done := make(chan struct{})

	// continuously iterate over the values of the series.
	// when a write is made, the total number of points in the series
	// will be sent over the channel, so we can make sure we are reading
	// the correct amount of values.
	go func(numPoints chan int, finished chan struct{}) {
		written := 0
		for {
			select {
			case written = <-numPoints:
			default:
				read := 0
				it := s.Iter()
				// read all of the points in the series.
				for it.Next() {
					tt, vv := it.Values()
					expectedT := TwoHoursData[read].T
					expectedV := TwoHoursData[read].V
					if uint64(expectedT) != tt || expectedV != vv {
						t.Errorf("metric values dont match what was written. (%d, %f) != (%d, %f)\n", tt, vv, expectedT, expectedV)
					}
					read++
				}
				//check that the number of points read matches the number of points
				// written to the series.
				if read != written && read != written+1 {
					// check if a point was written while we were running
					select {
					case written = <-numPoints:
						// a new point was written.
						if read != written && read != written+1 {
							t.Errorf("expexcted %d values in series, got %d", written, read)
						}
					default:
						t.Errorf("expexcted %d values in series, got %d", written, read)
					}
				}
			}
			//check if we have finished writing points.
			select {
			case <-finished:
				return
			default:
			}
		}
	}(writeNotify, done)

	// write points to the series.
	for i := 0; i < len(TwoHoursData); i++ {
		s.Push(uint64(TwoHoursData[i].T), TwoHoursData[i].V)
		writeNotify <- i + 1
		time.Sleep(sleep)
	}
	done <- struct{}{}
}

func BenchmarkEncode(b *testing.B) {
	b.SetBytes(int64(len(TwoHoursData) * 12))
	for i := 0; i < b.N; i++ {
		s := New(uint64(TwoHoursData[0].T))
		for _, tt := range TwoHoursData {
			s.Push(uint64(tt.T), tt.V)
		}
	}
}

func BenchmarkDecodeSeries(b *testing.B) {
	b.SetBytes(int64(len(TwoHoursData) * 12))
	s := New(uint64(TwoHoursData[0].T))
	for _, tt := range TwoHoursData {
		s.Push(uint64(tt.T), tt.V)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		it := s.Iter()
		var j int
		for it.Next() {
			j++
		}
	}
}

func BenchmarkDecodeByteSlice(b *testing.B) {
	b.SetBytes(int64(len(TwoHoursData) * 12))
	s := New(uint64(TwoHoursData[0].T))
	for _, tt := range TwoHoursData {
		s.Push(uint64(tt.T), tt.V)
	}

	s.Finish()
	bytes := s.Bytes()
	buf := make([]byte, len(bytes))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		copy(buf, bytes)
		it, _ := NewIterator(buf)//
		var j int
		for it.Next() {
			j++
		}
	}
}

func TestEncodeSimilarFloats(t *testing.T) {
	tunix := time.Unix(0, 0).Unix()
	s := New(uint64(tunix))
	want := []struct {
		t uint64
		v float64
	}{
		{uint64(tunix), 6.00065e+06},
		{uint64(tunix + 1), 6.000656e+06},
		{uint64(tunix + 2), 6.000657e+06},
		{uint64(tunix + 3), 6.000659e+06},
		{uint64(tunix + 4), 6.000661e+06},
	}
	for _, v := range want {
		s.Push(v.t, v.v)
	}


	it := s.Iter()

	for _, w := range want {
		if !it.Next() {
			t.Fatalf("Next()=false, want true")
		}
		tt, vv := it.Values()
		if w.t != tt || w.v != vv {
			t.Errorf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.t, w.v)
		}
	}

	if it.Next() {
		t.Fatalf("Next()=true, want false")
	}

	if err := it.Err(); err != nil {
		t.Errorf("it.Err()=%v, want nil", err)
	}
}

func TestBstreamIteratorError(t *testing.T) {
	bts := newBReader([]byte(""))
	_, err := bstreamIterator(bts)
	if err == nil {
		t.Errorf("An error was expected")
	}
}


