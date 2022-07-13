package fieldelement

import (
	"cme/constants"
	"encoding/binary"
)

type FieldElement struct {
	value uint16 // eigentlich 12
}

var Zero = FieldElement{0}
var One = FieldElement{1}
var Two = FieldElement{2}

const (
	Mask = (1 << constants.M) - 1
)

func New(val uint16) FieldElement {
	return FieldElement{val & Mask}
}

func NewByte(vals [2]byte) FieldElement {
	return New(binary.LittleEndian.Uint16(vals[:]))
}

func (fe FieldElement) IsValid() bool {
	return fe.value == fe.value&Mask
}

func (fe FieldElement) Get() uint16 {
	return fe.value
}

func (fe FieldElement) GetBit(k uint8) uint8 {
	return uint8((fe.value >> k) & 1)
}

// IsZeroMask is used as follows fe.IsZeroMask() & 1 == 1
func (fe FieldElement) IsZeroMask() FieldElement {
	var t = uint32(fe.value)
	t -= 1
	t >>= 19

	return FieldElement{uint16(t)}
}

func (fe FieldElement) And(val uint16) FieldElement {
	var t = uint16(fe.value)
	t = t & val

	return FieldElement{uint16(t)}
}

func (fe FieldElement) Mul(other FieldElement) FieldElement {
	var t0 = uint32(fe.value)
	var t1 = uint32(other.value)
	var tmp uint32
	var t uint32

	tmp = t0 * (t1 & 1)

	for i := 1; i < constants.M; i++ {
		tmp ^= t0 * (t1 & (1 << i))
	}

	t = tmp & 0x7FC000
	tmp ^= t >> 9
	tmp ^= t >> 12

	t = tmp & 0x3000
	tmp ^= t >> 9
	tmp ^= t >> 12

	return FieldElement{uint16(tmp & ((1 << constants.M) - 1))}

}

func (fe FieldElement) Square() FieldElement {
	x := uint32(fe.value)

	x = (x | (x << 8)) & 0x00FF00FF
	x = (x | (x << 4)) & 0x0F0F0F0F
	x = (x | (x << 2)) & 0x33333333
	x = (x | (x << 1)) & 0x55555555

	t1 := x & 0x7FC000
	x ^= t1 >> 9
	x ^= t1 >> 12

	t2 := x & 0x3000
	x ^= t2 >> 9
	x ^= t2 >> 12

	return FieldElement{uint16(x & ((1 << constants.M) - 1))}
}

func (fe FieldElement) Add(other FieldElement) FieldElement {
	tmp := fe.value ^ other.value
	return FieldElement{tmp}
}

func (fe FieldElement) ReverseBits() FieldElement {
	tmp := fe.value
	tmp = ((tmp & 0x00ff) << 8) | ((tmp & 0xff00) >> 8)
	tmp = ((tmp & 0x0f0f) << 4) | ((tmp & 0xf0f0) >> 4)
	tmp = ((tmp & 0x3333) << 2) | ((tmp & 0xcccc) >> 2)
	tmp = ((tmp & 0x5555) << 1) | ((tmp & 0xaaaa) >> 1)
	tmp >>= 4
	return FieldElement{tmp}
}

func (fe FieldElement) Inverse() FieldElement {
	// 0x001
	tmp1 := fe

	// 0x003
	tmp11 := tmp1.Mul(tmp1.Square())

	// 0x00f
	tmp1111 := tmp11.Mul(tmp11.Square().Square())

	// 0x0ff
	tmp11111111 := tmp1111.Mul(tmp1111.Square().Square().Square().Square())

	// 0x3ff
	tmp1111111111 := tmp11.Mul(tmp11111111.Square().Square())

	// 0x7ff
	tmp11111111111 := tmp1.Mul(tmp1111111111.Square())

	return tmp11111111111.Square()

}

func (fe FieldElement) Div(o FieldElement) FieldElement {
	return fe.Mul(o.Inverse())
}
