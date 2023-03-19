package simple8b

import (
	"encoding/binary"
	"github.com/jwilder/encoding/simple8b"
	"io"
	"log"
	"math"
	"math/bits"
	"sync"
)

// Series is the basic series primitive
// you can concurrently put values, finish the stream, and create iterators



type Series struct {
	sync.Mutex

	// TODO(dgryski): timestamps in the paper are uint64
	T0  uint64
	t   uint64
	val float64
	bts       bstream//btimestamp
	bv		  bstream//bvalue
	leading  uint8
	trailing uint8
	finished bool
	tDelta uint64
	src	[]uint64
	flagv bool
}

// New series
func New(t0 uint64) *Series {
	//set t0
	s := Series{
		T0:      t0,
		t: t0,
		leading: ^uint8(0), // 0xff
		src: make([]uint64,0),
	}
	s.bts.writeBits(t0, 64)//enter the t0
	return &s

}

// Bytes value of the series stream
func (s *Series) Bytes() []byte {
	s.Lock()
	defer s.Unlock()
	return s.bts.bytes()
}

func finish(w *bstream) {
	// write an end-of-stream record
	w.writeBits(0xffffffffffffffff, 64)
}

// Finish the series by writing an end-of-stream record
func (s *Series) Finish() {
	s.Lock()
	if !s.finished {
		finish(&s.bts)
		finish(&s.bv)
		s.finished = true
	}
	s.Unlock()
}
func (s *Series) End(){
	for len(s.src)>0{
		simple8bTs,n,_:=simple8b.Encode(s.src)
		s.src=s.src[n:]
		s.bts.writeBits(simple8bTs,64)
	}
}
// Push a timestamp and value to the series
func (s *Series) Push(t uint64, v float64) {
	s.Lock()
	defer s.Unlock()
	tDelta := t - s.t
	//dod: Next tDelta minus previous tDelta
	s.src=append(s.src,tDelta)
	if len(s.src)>=30{
		simple8bTs,n,_:=simple8b.Encode(s.src)
		s.src=s.src[n:]
		s.bts.writeBits(simple8bTs,64)
	}
	if s.flagv==false{
		s.bv.writeBits(math.Float64bits(v), 64)
		s.flagv=true
	}else{
		vDelta := math.Float64bits(v) ^ math.Float64bits(s.val)
		if vDelta == 0 {
			s.bv.writeBit(zero)
		} else {
			s.bv.writeBit(one)
			leading := uint8(bits.LeadingZeros64(vDelta))
			trailing := uint8(bits.TrailingZeros64(vDelta))
			// clamp number of leading zeros to avoid overflow when encoding
			if leading >= 32 {
				leading = 31
			}
			// TODO(dgryski): check if it's 'cheaper' to reset the leading/trailing bits instead
			if s.leading != ^uint8(0) && leading >= s.leading && trailing >= s.trailing {
				s.bv.writeBit(zero)
				s.bv.writeBits(vDelta>>s.trailing, 64-int(s.leading)-int(s.trailing))
			} else {
				s.leading, s.trailing = leading, trailing
				s.bv.writeBit(one)
				s.bv.writeBits(uint64(leading), 5)

				// Note that if leading == trailing == 0, then sigbits == 64.  But that value doesn't actually fit into the 6 bits we have.
				// Luckily, we never need to encode 0 significant bits, since that would put us in the other case (vdelta == 0).
				// So instead we write out a 0 and adjust it back to 64 on unpacking.
				sigbits := 64 - leading - trailing
				s.bv.writeBits(uint64(sigbits), 6)
				s.bv.writeBits(vDelta>>trailing, int(sigbits))
			}
		}
	}
	s.tDelta = tDelta
	s.t = t
	s.val = v
}

// Iter lets you iterate over a series.  It is not concurrency-safe.
func (s *Series) Iter() (*Iter) {
	s.Lock()
	s.End()
	v := s.bv.clone()
	ts := s.bts.clone()
	s.Unlock()
	finish(ts)
	finish(v)
	iter, _ := bstreamIterator(ts,v)
	return iter
}

// Iter lets you iterate over a series.  It is not concurrency-safe.
type Iter struct {
	T0 uint64

	t   uint64
	val float64

	bts       bstream
	bv		  bstream
	leading  uint8
	trailing uint8

	finished bool

	tDelta uint32
	err    error
	nextT []uint64
	nextV []float64
	flagv bool
}

func bstreamIterator(bts *bstream,bv *bstream) (*Iter, error) {

	bts.count = 8
	bv.count = 8
	t0, err := bts.readBits(64)
	if err != nil {
		return nil, err
	}

	return &Iter{
		T0: t0,
		bts: *bts,
		bv: *bv,
	}, nil
}
func (s *Series) GetNewT()(uint64,float64){
	return s.t,s.val
}
// NewIterator for the series
func NewIterator(bts ,bv []byte) (*Iter, error) {
	return bstreamIterator(newBReader(bts),newBReader(bv))
}

// Next iteration of the series iterator
func (it *Iter) Next() bool {
	if (it.err != nil || it.finished)&&len(it.nextT)==0{
		return false
	}
	if len(it.nextT)==0{
		var dst [240]uint64
		bitByte,_:=it.bts.readBits(64)
		if bitByte==0xffffffffffffffff{
			it.finished=true
			return false
		}
		n,_:=simple8b.Decode(&dst,bitByte)
		log.Println(n)
		tnow:=it.T0
		for i:=0;i<n;i++ {
			tnow+=dst[i]
			it.nextT=append(it.nextT,tnow)
			it.T0=tnow
			if it.flagv==false{
				v,_:=it.bv.readBits(64)
				it.val=math.Float64frombits(v)
				it.nextV=append(it.nextV,it.val)
				it.flagv=true
				continue
			}
			bit, err := it.bv.readBit()
			if err != nil {
				it.err = err
				return false
			}
			if bit == zero {

			} else {
				bit, itErr := it.bv.readBit()
				if itErr != nil {
					it.err = err
					return false
				}
				if bit == zero {
				} else {
					bits, err := it.bv.readBits(5)
					if err != nil {
						it.err = err
						return false
					}
					it.leading = uint8(bits)
					bits, err = it.bv.readBits(6)
					if err != nil {
						it.err = err
						return false
					}
					mbits := uint8(bits)
					// 0 significant bits here means we overflowed and we actually need 64; see comment in encoder
					if mbits == 0 {
						mbits = 64
					}
					it.trailing = 64 - it.leading - mbits
				}
				mbits := int(64 - it.leading - it.trailing)
				bits, err := it.bv.readBits(mbits)
				if err != nil {
					it.err = err
					return false
				}
				vbits := math.Float64bits(it.val)
				vbits ^= (bits << it.trailing)
				it.val = math.Float64frombits(vbits)
			}
			it.nextV=append(it.nextV,it.val)
		}
	}
	it.t=it.nextT[0]
	it.val=it.nextV[0]
	it.nextT=it.nextT[1:]
	it.nextV=it.nextV[1:]
	return true
}

// Values at the current iterator position
func (it *Iter) Values() (uint64, float64) {
	return it.t, it.val
}

// Err error at the current iterator position
func (it *Iter) Err() error {
	return it.err
}

type errMarshal struct {
	w   io.Writer
	r   io.Reader
	err error
}

func (em *errMarshal) write(t interface{}) {
	if em.err != nil {
		return
	}
	em.err = binary.Write(em.w, binary.BigEndian, t)
}

func (em *errMarshal) read(t interface{}) {
	if em.err != nil {
		return
	}
	em.err = binary.Read(em.r, binary.BigEndian, t)
}



