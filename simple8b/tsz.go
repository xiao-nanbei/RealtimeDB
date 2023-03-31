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
	br		  bstream
	leading  uint8
	trailing uint8
	finished bool
	tDelta uint64
	srcts	[]uint64
	srcv    []float64
	flagv bool
}

// New series
func New(t0 uint64) *Series {
	//set t0
	s := Series{
		T0:      t0,
		t: t0,
		leading: ^uint8(0), // 0xff
		srcts: make([]uint64,0),
		srcv: make([]float64,0),
	}
	s.br.writeBits(t0, 64)//enter the t0
	return &s

}

// Bytes value of the series stream
func (s *Series) Bytes() []byte {
	s.Lock()
	defer s.Unlock()
	return s.br.bytes()
}

func finish(w *bstream) {
	// write an end-of-stream record
	w.writeBits(0xffffffffffffffff, 64)
}

// Finish the series by writing an end-of-stream record
func (s *Series) Finish() {
	s.Lock()
	if !s.finished {
		finish(&s.br)
		s.finished = true
	}
	s.Unlock()
}
func (s *Series) End(){
	s.Lock()
	defer s.Unlock()
	if s.flagv==false{
		if len(s.srcv)==0{
			return
		}
		s.br.writeBits(math.Float64bits(s.srcv[0]), 64)
		s.val=s.srcv[0]
		s.flagv=true
	}
	for len(s.srcts)>0{
		simple8bTs,n,_:=simple8b.Encode(s.srcts)
		s.srcts=s.srcts[n:]
		s.br.writeBits(simple8bTs,64)
		for i:=0;i<n;i++{
			vDelta := math.Float64bits(s.srcv[0]) ^ math.Float64bits(s.val)
			if vDelta == 0 {
				//log.Println(false)
				s.br.writeBit(zero)
			} else {
				//log.Println(true)
				s.br.writeBit(one)
				leading := uint8(bits.LeadingZeros64(vDelta))
				trailing := uint8(bits.TrailingZeros64(vDelta))
				// clamp number of leading zeros to avoid overflow when encoding
				if leading >= 32 {
					leading = 31
				}
				// TODO(dgryski): check if it's 'cheaper' to reset the leading/trailing bits instead
				if s.leading != ^uint8(0) && leading >= s.leading && trailing >= s.trailing {
					s.br.writeBit(zero)
					s.br.writeBits(vDelta>>s.trailing, 64-int(s.leading)-int(s.trailing))
				} else {
					s.leading, s.trailing = leading, trailing
					s.br.writeBit(one)
					s.br.writeBits(uint64(leading), 5)

					// Note that if leading == trailing == 0, then sigbits == 64.  But that value doesn't actually fit into the 6 bits we have.
					// Luckily, we never need to encode 0 significant bits, since that would put us in the other case (vdelta == 0).
					// So instead we write out a 0 and adjust it back to 64 on unpacking.
					sigbits := 64 - leading - trailing
					s.br.writeBits(uint64(sigbits), 6)
					s.br.writeBits(vDelta>>trailing, int(sigbits))
				}
			}
			s.val = s.srcv[0]
			s.srcv=s.srcv[1:]
		}
	}
}
// Push a timestamp and value to the series
func (s *Series) Push(t uint64, v float64) {
	s.Lock()
	defer s.Unlock()
	if s.flagv==false{
		s.br.writeBits(math.Float64bits(v), 64)
		s.val=v
		s.flagv=true

	}
	tDelta := t - s.t
	//dod: Next tDelta minus previous tDelta
	s.srcts=append(s.srcts,tDelta)
	s.srcv=append(s.srcv,v)
	if len(s.srcts)>=60 {
		simple8bTs, n, _ := simple8b.Encode(s.srcts)
		s.srcts = s.srcts[n:]
		s.br.writeBits(simple8bTs, 64)
		for i := 0; i < n; i++ {
			vDelta := math.Float64bits(s.srcv[0]) ^ math.Float64bits(s.val)

			if vDelta == 0 {
				//log.Println(false)
				s.br.writeBit(zero)
			} else {
				//log.Println(true)
				s.br.writeBit(one)
				leading := uint8(bits.LeadingZeros64(vDelta))
				trailing := uint8(bits.TrailingZeros64(vDelta))
				// clamp number of leading zeros to avoid overflow when encoding
				if leading >= 32 {
					leading = 31
				}
				// TODO(dgryski): check if it's 'cheaper' to reset the leading/trailing bits instead
				if s.leading != ^uint8(0) && leading >= s.leading && trailing >= s.trailing {
					s.br.writeBit(zero)
					s.br.writeBits(vDelta>>s.trailing, 64-int(s.leading)-int(s.trailing))
				} else {
					s.leading, s.trailing = leading, trailing
					s.br.writeBit(one)
					s.br.writeBits(uint64(leading), 5)

					// Note that if leading == trailing == 0, then sigbits == 64.  But that value doesn't actually fit into the 6 bits we have.
					// Luckily, we never need to encode 0 significant bits, since that would put us in the other case (vdelta == 0).
					// So instead we write out a 0 and adjust it back to 64 on unpacking.
					sigbits := 64 - leading - trailing
					s.br.writeBits(uint64(sigbits), 6)
					s.br.writeBits(vDelta>>trailing, int(sigbits))
				}
			}
			s.val = s.srcv[0]
			s.srcv = s.srcv[1:]
		}
	}
	s.tDelta = tDelta
	s.t = t
}

// Iter lets you iterate over a series.  It is not concurrency-safe.
func (s *Series) Iter() (*Iter) {
	s.Lock()
	s.End()
	br := s.br.clone()
	s.Unlock()
	finish(br)
	iter, _ := bstreamIterator(br)
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

	tDelta uint32
	err    error
	nextT []uint64
	nextV []float64
	flagv bool
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
func (s *Series) GetNewT()(uint64,float64){
	return s.t,s.val
}
// NewIterator for the series
func NewIterator(br []byte) (*Iter, error) {
	return bstreamIterator(newBReader(br))
}

// Next iteration of the series iterator
func (it *Iter) Next() bool {
	if (it.err != nil || it.finished)&&len(it.nextT)==0{
		return false
	}
	if it.flagv==false{
		v,_:=it.br.readBits(64)
		it.val=math.Float64frombits(v)
		it.flagv=true
	}
	if len(it.nextT)==0{
		var dst [240]uint64
		bitByte,_:=it.br.readBits(64)
		if bitByte==0xffffffffffffffff{
			it.finished=true
			return false
		}
		n,_:=simple8b.Decode(&dst,bitByte)
		tnow:=it.T0
		for i:=0;i<n;i++ {
			tnow+=dst[i]
			it.nextT=append(it.nextT,tnow)
			it.T0=tnow
			bit, err := it.br.readBit()
			//log.Println(bit)
			if err != nil {
				it.err = err
				return false
			}
			if bit == zero {
			} else {
				bit, itErr := it.br.readBit()
				if itErr != nil {
					it.err = err
					return false
				}
				if bit == zero {
				} else {
					bits, err := it.br.readBits(5)
					if err != nil {
						it.err = err
						return false
					}
					it.leading = uint8(bits)
					bits, err = it.br.readBits(6)
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
				bits, err := it.br.readBits(mbits)
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
	//log.Println(it.nextV)
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
	bStream, err := s.br.MarshalBinary()
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
	em.read(&s.tDelta)
	em.read(&s.trailing)
	em.read(&s.val)
	outBuf := make([]byte, buf.Len())
	em.read(outBuf)
	err := s.br.UnmarshalBinary(outBuf)
	if err != nil {
		return err
	}
	if em.err != nil {
		return em.err
	}
	return nil
}
