package polynomial

import (
	"cme/constants"
	"cme/fieldelement"
	"errors"
	log "github.com/sirupsen/logrus"
)

type Polynomial struct {
	Elements [constants.T]fieldelement.FieldElement
}

const (
	Len = constants.Sigma1Bytes * constants.T
)

func InterpretAsPolynomial(input [Len]byte) Polynomial {
	p := Polynomial{}
	for i := 0; i < len(p.Elements); i++ {
		var chunk [2]byte
		chunk[0] = input[2*i]
		chunk[1] = input[(2*i)+1]

		p.Elements[i] = fieldelement.NewByte(chunk)
	}
	return p
}

func Irreducible(input [Len]byte) (Polynomial, error) {
	// poly 64 x uint16
	poli := InterpretAsPolynomial(input)

	var mat [constants.T + 1]Polynomial

	mat[0].Elements[0] = fieldelement.One
	mat[1] = poli

	for i := 1; i < constants.T; i++ {
		tmp := Mul(mat[i], poli)
		// from 127 Elements to 64 Elements
		copy(mat[i+1].Elements[:], tmp[:])

	}

	for j := 0; j < constants.T; j++ {

		for k := j + 1; k < constants.T; k++ {

			mask := mat[j].Elements[j].IsZeroMask()
			for c := j; c < constants.T+1; c++ {
				mat[c].Elements[j] = mat[c].Elements[j].Add(mat[c].Elements[k].And(mask.Get()))
			}

		}

		if mat[j].Elements[j].IsZeroMask().Get()&1 == 1 {
			log.Errorln("mat[j].Elements[j] is zero:", mat[j].Elements[j])
			return Polynomial{}, errors.New("error, not systematic")
		}

		inverse := mat[j].Elements[j].Inverse()

		for c := j; c < constants.T+1; c++ {
			mat[c].Elements[j] = mat[c].Elements[j].Mul(inverse)
		}

		for k := 0; k < constants.T; k++ {
			if k != j {
				t := mat[j].Elements[k]
				for c := j; c < constants.T+1; c++ {
					mat[c].Elements[k] = mat[c].Elements[k].Add(mat[c].Elements[j].Mul(t))
				}
			}
		}

	}

	return mat[constants.T], nil
}

// ref: GF_mul

func Mul(ap Polynomial, bp Polynomial) [2*constants.T - 1]fieldelement.FieldElement {
	var tmp [2*constants.T - 1]fieldelement.FieldElement
	for i := 0; i < constants.T; i++ {
		tmp[i] = fieldelement.Zero
	}
	for i, a := range ap.Elements {
		for j, b := range bp.Elements {
			tmp[i+j] = tmp[i+j].Add(a.Mul(b))
		}
	}

	for i := (constants.T - 1) * 2; i >= constants.T; i-- {
		l := tmp[i]
		tmp[i-constants.T+3] = tmp[i-constants.T+3].Add(l)
		tmp[i-constants.T+1] = tmp[i-constants.T+1].Add(l)
		tmp[i-constants.T+0] = tmp[i-constants.T+0].Add(l.Mul(fieldelement.Two))
	}

	return tmp
}

func EvalAt(mp Polynomial, a fieldelement.FieldElement) fieldelement.FieldElement {

	s := mp.Elements[:]
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	tmp := a.Add(s[0])
	for i := 1; i < len(s); i++ {
		tmp = tmp.Mul(a)
		tmp = tmp.Add(s[i])
	}
	return tmp

}

func Root(mp Polynomial, support [constants.N]fieldelement.FieldElement) [constants.N]fieldelement.FieldElement {
	var result [constants.N]fieldelement.FieldElement

	for i := 0; i < constants.N; i++ {
		result[i] = EvalAt(mp, support[i])
	}
	return result
}
