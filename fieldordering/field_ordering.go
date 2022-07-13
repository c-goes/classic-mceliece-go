package fieldordering

import (
	"cme/constants"
	"cme/fieldelement"
	"encoding/binary"
	"errors"
)

const (
	FOBytes = constants.Sigma2Bytes * constants.Q // 16384
)

type FieldOrdering struct {
	elements [constants.Q]fieldelement.FieldElement
}

func uint64MinxMax(a *uint64, b *uint64) {
	c := *b - *a
	c >>= 63
	c = -c
	c &= *a ^ *b
	*a ^= c
	*b ^= c

}

func (fo FieldOrdering) GenerateSupport() [constants.N]fieldelement.FieldElement {
	var support [constants.N]fieldelement.FieldElement
	for i := 0; i < constants.N; i++ {
		support[i] = fo.elements[i].ReverseBits()
	}
	return support
}

// djbsortUint64 based on https://sorting.cr.yp.to/
func djbsortUint64(x []uint64) []uint64 {

	n := int64(len(x))
	var top int64
	var p int64
	var q int64
	var r int64
	var i int64

	if n < 2 {
		return x
	}
	top = 1
	for top < n-top {
		top += top
	}

	for p = top; p > 0; p >>= 1 {
		for i = 0; i < n-p; i++ {
			if (i & p) == 0 {
				uint64MinxMax(&x[i], &x[i+p])
			}
		}
		i = 0
		for q = top; q > p; q >>= 1 {
			for ; i < n-q; i++ {
				if (i & p) == 0 {
					a := x[i+p]
					for r = q; r > p; r >>= 1 {
						uint64MinxMax(&a, &x[i+r])
					}
					x[i+p] = a

				}
			}
		}
	}

	return x
}

func New(input [FOBytes]byte) (FieldOrdering, error) {
	// Take 4 byte = 32 input bits
	// as 32 bit integers = 4 x uint32
	// a_0, a_1, a_2, a_3
	// little endian

	var pairs [constants.Q]uint64

	for i, _ := range pairs {
		bytes := uint64(binary.LittleEndian.Uint32(input[i*4 : (i*4)+4]))
		pairs[i] = uint64(i)
		pairs[i] |= bytes << 31
	}

	// sort pairs (a_i, i) in lexicographic order
	tmp := djbsortUint64(pairs[:])

	copy(pairs[:], tmp)

	// check if distinct integers
	// and return error when needed
	for i := 1; i < constants.Q; i++ {
		if (pairs[i-1] >> 31) == (pairs[i] >> 31) {
			return FieldOrdering{}, errors.New("non distinct elements")
		}
	}

	alpha := FieldOrdering{}
	for i := 0; i < constants.Q; i++ {
		alpha.elements[i] = fieldelement.New(uint16(pairs[i]))
	}

	return alpha, nil
}
