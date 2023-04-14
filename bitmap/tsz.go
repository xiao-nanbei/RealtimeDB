package bitmap

import (
	"bytes"
	"encoding/binary"
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

	bw       bstream
	leading  uint8
	trailing uint8
	finished bool
}

// New series
func New(t0 uint64) *Series {
	//set t0
	s := Series{
		T0:      t0,
		leading: ^uint8(0), // 0xff
	}
	s.bw.writeBits(t0, 64)
	return &s

}

// Bytes value of the series stream
func (s *Series) Bytes() []byte {
	s.Lock()
	defer s.Unlock()
	return s.bw.bytes()
}

func finish(w *bstream) {
	// write an end-of-stream record
	w.writeBits(0xffffffff, 32)
	w.writeBit(zero)
}

// Finish the series by writing an end-of-stream record
func (s *Series) Finish() {
	s.Lock()
	if !s.finished {
		finish(&s.bw)
		s.finished = true
	}
	s.Unlock()
}

func (s *Series) GetNewPoint()(uint64,float64){
	return s.t,s.val
}

// Push a timestamp and value to the series
func (s *Series) Push(t uint64, v float64) {
	s.Lock()
	defer s.Unlock()
	if s.t == 0 {
		// first point
		s.t = t
		s.val = v
		tDelta := t - s.T0
		s.bw.writeBits(math.Float64bits(v), 64)
		var i uint64=0
		for i<tDelta{
			s.bw.writeBit(zero)
			i++
		}
		s.bw.writeBit(one)
		return
	}



	vDelta := math.Float64bits(v) ^ math.Float64bits(s.val)

	if vDelta == 0 {
		s.bw.writeBit(zero)
	} else {
		s.bw.writeBit(one)

		leading := uint8(bits.LeadingZeros64(vDelta))
		trailing := uint8(bits.TrailingZeros64(vDelta))

		// clamp number of leading zeros to avoid overflow when encoding
		if leading >= 32 {
			leading = 31
		}

		// TODO(dgryski): check if it's 'cheaper' to reset the leading/trailing bits instead
		if s.leading != ^uint8(0) && leading >= s.leading && trailing >= s.trailing {
			s.bw.writeBit(zero)
			s.bw.writeBits(vDelta>>s.trailing, 64-int(s.leading)-int(s.trailing))
		} else {
			s.leading, s.trailing = leading, trailing

			s.bw.writeBit(one)
			s.bw.writeBits(uint64(leading), 5)

			// Note that if leading == trailing == 0, then sigbits == 64.  But that value doesn't actually fit into the 6 bits we have.
			// Luckily, we never need to encode 0 significant bits, since that would put us in the other case (vdelta == 0).
			// So instead we write out a 0 and adjust it back to 64 on unpacking.
			sigbits := 64 - leading - trailing
			s.bw.writeBits(uint64(sigbits), 6)
			s.bw.writeBits(vDelta>>trailing, int(sigbits))
		}
	}
	tDelta := t - s.t
	var i uint64=0
	for i<tDelta{
		s.bw.writeBit(zero)
		i++
	}
	s.bw.writeBit(one)
	s.t = t
	s.val = v

}

// Iter lets you iterate over a series.  It is not concurrency-safe.
func (s *Series) Iter() *Iter {
	s.Lock()
	w := s.bw.clone()
	s.Unlock()

	finish(w)
	iter, _ := bstreamIterator(w)
	return iter
}

// Iter lets you iterate over a series.  It is not concurrency-safe.
type Iter struct {
	T0 uint64

	t   uint64
	val float64

	br       bstream
	leading  uint8
	trailing uint8

	finished bool
	err    error
}

func bstreamIterator(br *bstream) (*Iter, error) {

	br.count = 8

	t0, err := br.readBits(64)
	if err != nil {
		return nil, err
	}

	return &Iter{
		T0: t0,
		br: *br,
	}, nil
}

// NewIterator for the series
func NewIterator(b []byte) (*Iter, error) {
	return bstreamIterator(newBReader(b))
}

// Next iteration of the series iterator
func (it *Iter) Next() bool {

	if it.err != nil || it.finished {
		return false
	}

	if it.t == 0 {
		// read first t and v
		it.t=it.T0
		val, err := it.br.readBits(64)
		if err != nil {
			it.err = err
			return false
		}
		it.val = math.Float64frombits(val)
	}else{
		// read compressed value
		bit, err := it.br.readBit()
		if err != nil {
			it.err = err
			return false
		}

		if bit == zero {
			// it.val = it.val
		} else {
			bit, itErr := it.br.readBit()
			if itErr != nil {
				it.err = err
				return false
			}
			if bit == zero {
				// reuse leading/trailing zero bits
				// it.leading, it.trailing = it.leading, it.trailing
			} else {
				flag_1:=false
				flag_2:=false
				bits, err := it.br.readBits(5)
				if err != nil {
					it.err = err
					return false
				}
				if bits==31{
					flag_1=true
				}
				it.leading = uint8(bits)

				bits, err = it.br.readBits(6)
				if err != nil {
					it.err = err
					return false
				}
				if bits==63{
					flag_2=true
				}
				if flag_1&&flag_2{
					it.finished=true
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
			bits, err := it.br.readBits(mbits)
			if err != nil {
				it.err = err
				return false
			}
			vbits := math.Float64bits(it.val)
			vbits ^= (bits << it.trailing)
			it.val = math.Float64frombits(vbits)
		}
	}
	var count uint64=0
	readBit, _ := it.br.readBit()
	for readBit==zero{
		count++
		readBit, _ = it.br.readBit()
	}
	it.t=it.t+count
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
	em.write(s.trailing)
	em.write(s.val)
	bStream, err := s.bw.MarshalBinary()
	if err != nil {
		return nil, err
	}
	em.write(bStream)
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
	em.read(&s.trailing)
	em.read(&s.val)
	outBuf := make([]byte, buf.Len())
	em.read(outBuf)
	err := s.bw.UnmarshalBinary(outBuf)
	if err != nil {
		return err
	}
	if em.err != nil {
		return em.err
	}
	return nil
}