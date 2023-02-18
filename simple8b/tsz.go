package simple8b

import (
	"bytes"
	"encoding/binary"
	"github.com/jwilder/encoding/simple8b"
	"io"
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
	tDelta uint32
	src	[]uint64
}

// New series
func New(t0 uint64) *Series {
	//set t0
	s := Series{
		T0:      t0,
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
	w.writeBits(0x0f, 4)
	w.writeBits(0xffffffff, 32)
	w.writeBit(zero)
}

// Finish the series by writing an end-of-stream record
func (s *Series) Finish() {
	s.Lock()
	if !s.finished {
		for len(s.src)>0{
			simple8bTs,n,_:=simple8b.Encode(s.src)
			s.src=s.src[n:]
			s.bts.writeBits(simple8bTs,64)
		}
		finish(&s.bts)
		finish(&s.bv)
		s.finished = true
	}
	s.Unlock()
}

// Push a timestamp and value to the series
func (s *Series) Push(t uint64, v float64) {
	s.Lock()
	defer s.Unlock()
	tDelta := uint32(t - s.t)
	//dod: Next tDelta minus previous tDelta
	dod := uint64(tDelta - s.tDelta)
	s.src=append(s.src,dod)
	if len(s.src)>=30{
		simple8bTs,n,_:=simple8b.Encode(s.src)
		s.src=s.src[n:]
		s.bts.writeBits(simple8bTs,64)
	}
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
	s.tDelta = tDelta
	s.t = t
	s.val = v

}

// Iter lets you iterate over a series.  It is not concurrency-safe.
func (s *Series) Iter() (*Iter) {
	s.Lock()
	v := s.bv.clone()
	ts := s.bts.clone()
	s.Unlock()
	finish(v)
	finish(ts)
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
}

func bstreamIterator(bts *bstream,bv *bstream) (*Iter, error) {

	bts.count = 8

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

// NewIterator for the series
func NewIterator(bts ,bv[]byte) (*Iter, error) {
	return bstreamIterator(newBReader(bts),newBReader(bv))
}

// Next iteration of the series iterator
func (it *Iter) Next() bool {
	if it.err != nil || it.finished {
		return false
	}
	var dst [240]uint64

	bitByte,_:=it.bts.readBits(64)
	n,_:=simple8b.Decode(&dst,bitByte)
	for i:=0;i<n;i++ {
		bit, err := it.bv.readBit()
		if err != nil {
			it.err = err
			return false
		}
		if bit == zero {
			// it.val = it.val
		} else {
			bit, itErr := it.bv.readBit()
			if itErr != nil {
				it.err = err
				return false
			}
			if bit == zero {
				// reuse leading/trailing zero bits
				// it.leading, it.trailing = it.leading, it.trailing
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
			vbits ^= bits << it.trailing
			it.val = math.Float64frombits(vbits)
		}
		tDelta := it.tDelta + uint32(dst[i])
		it.tDelta = tDelta
		it.t = it.t + uint64(it.tDelta)
	}
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

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (s *Series) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	em := &errMarshal{w: buf}
	em.write(s.T0)
	em.write(s.leading)
	em.write(s.t)
	em.write(s.tDelta)
	em.write(s.trailing)
	em.write(s.val)
	bvStream, err := s.bv.MarshalBinary()
	btsStream, err := s.bts.MarshalBinary()
	if err != nil {
		return nil, err
	}
	em.write(btsStream)
	em.write(bvStream)
	if em.err != nil {
		return nil, em.err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (s *Series) UnmarshalBinary(b []byte) error {
	buf := bytes.NewReader(b)
	em := &errMarshal{r: buf}
	em.read(&s.T0)
	em.read(&s.leading)
	em.read(&s.t)
	em.read(&s.tDelta)
	em.read(&s.trailing)
	em.read(&s.val)
	outBuf := make([]byte, buf.Len())
	em.read(outBuf)
	err := s.bts.UnmarshalBinary(outBuf)
	err2 := s.bv.UnmarshalBinary(outBuf)
	if err != nil|| err2 != nil {
		return err
	}
	if em.err != nil {
		return em.err
	}
	return nil
}
