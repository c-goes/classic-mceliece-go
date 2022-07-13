package secretkey

import (
	"cme/constants"
	"cme/fieldelement"
	"cme/fieldordering"
	"cme/polynomial"
	"cme/types"
	"cme/util"
	log "github.com/sirupsen/logrus"
)

type SecretKey struct {
	seed  types.Seed
	g     polynomial.Polynomial
	alpha fieldordering.FieldOrdering
	s     [constants.SBytes]byte
}

func New(seed types.Seed, g polynomial.Polynomial, alpha fieldordering.FieldOrdering, s [constants.SBytes]byte) SecretKey {
	return SecretKey{seed, g, alpha, s}
}

// Decapsulate uses Decode to decode c0
// returns session key
func (sk SecretKey) Decapsulate(ciphertext [constants.C0Bytes + constants.C1Bytes]byte) [constants.LBytes]byte {
	c0 := ciphertext[:constants.C0Bytes]
	var c0Array [constants.C0Bytes]byte
	copy(c0Array[:], c0)
	c1 := ciphertext[constants.C0Bytes:]
	var c1Array [constants.C1Bytes]byte
	copy(c1Array[:], c1)
	var error_ [constants.NBytes]byte

	error_, status := sk.Decode(c0Array, error_)
	if status == 1 {
		log.Warn("Error decapsulating")
	}

	// spec 2.3.3 -> 6.
	b := 1

	c1Strich := util.Hash2v(error_)
	if c1Strich != c1Array {
		log.Info("c1 != c1'  ", c1Array, c1Strich)
		b = 0
		for i := 0; i < constants.SBytes; i++ {

			error_[i] = sk.s[i]

		}
	}

	return util.HashXvC(uint8(b), error_, ciphertext)
}

// Decode uses Synd() and BerlekampMassey()
// Decoding failed when the returned int is == 1
func (sk SecretKey) Decode(c0 [constants.C0Bytes]byte,
	error [constants.NBytes]byte,
) ([constants.NBytes]byte, uint16) {

	var v [constants.NBytes]byte

	copy(v[:constants.C0Bytes], c0[:])

	support := sk.alpha.GenerateSupport()
	syndrome := Synd(sk.g, support, v)

	locator := BerlekampMassey(syndrome)

	images := polynomial.Root(locator, support)

	var weight uint16 = 0
	var t uint16
	for i := 0; i < constants.N/8; i++ {
		error[i] = 0
	}
	// generate error
	for i := 0; i < constants.N; i++ {
		t = images[i].IsZeroMask().Get() & 1
		error[i/8] |= uint8(t << (i % 8))
		weight += t
	}

	compareSynd := Synd(sk.g, support, error)
	var check uint16 = weight

	check ^= constants.T
	for i := 0; i < constants.T; i++ {
		check |= syndrome[i].Get() ^ compareSynd[i].Get()
	}

	check -= 1
	check >>= 15
	return error, check ^ 1
}

func Synd(f polynomial.Polynomial, support [constants.N]fieldelement.FieldElement, r [constants.NBytes]byte) [2 * constants.T]fieldelement.FieldElement {
	var syn [2 * constants.T]fieldelement.FieldElement
	for i := 0; i < constants.N; i++ {
		c := (uint16(r[i/8]) >> (i % 8)) & 1
		e := polynomial.EvalAt(f, support[i])
		invertedE := e.Square().Inverse()

		for j := 0; j < 2*constants.T; j++ {
			syn[j] = syn[j].Add(invertedE.Mul(fieldelement.New(c)))
			invertedE = invertedE.Mul(support[i])
		}

	}
	return syn
}

func Min(a uint16, b uint16) uint16 {
	if a < b {
		return a
	} else {
		return b
	}
}

// BerlekampMassey taskes as input the syndrome (polynomial) S(x), degree 2t-1
// https://eprint.iacr.org/2017/793.pdf
func BerlekampMassey(s [2 * constants.T]fieldelement.FieldElement) polynomial.Polynomial {

	var N uint16 = 0
	var L uint16 = 0
	var mle uint16
	var mne uint16

	var T [constants.T + 1]fieldelement.FieldElement
	var B [constants.T + 1]fieldelement.FieldElement
	var C [constants.T + 1]fieldelement.FieldElement

	// b is \delta
	var b = fieldelement.One
	var d fieldelement.FieldElement
	var f fieldelement.FieldElement

	//Beta is polynomial beta(x) = x
	B[1] = fieldelement.One

	//Sigma is polynomial sigma(x) = 1
	C[0] = fieldelement.One

	for N = 0; N < 2*constants.T; N++ {
		d = fieldelement.Zero

		for i := 0; uint16(i) <= Min(N, constants.T); i++ {
			d = d.Add(C[i].Mul(s[N-uint16(i)]))
		}

		// two masks mne and mle
		// Check if d is zero with logical operations
		// mne is zero when d is zero, else 65535
		mne = d.Get()
		mne -= 1
		mne >>= 15
		mne -= 1
		// Check if N < 2*L with logical operations
		mle = N
		mle -= 2 * L
		mle >>= 15
		mle -= 1
		// combining the two conditions for the cases in the formula
		mle &= mne

		// Copy of the Sigma Polynomial C
		for i := 0; i <= constants.T; i++ {
			T[i] = C[i]
		}

		// factor
		f = d.Div(b)

		for i := 0; i <= constants.T; i++ {
			// Updating \sigma(x) always works the same, we only do a logical operation to calculate a diffent result depending on d == 0
			C[i] = fieldelement.New(C[i].Add(fieldelement.New(f.Mul(B[i]).Get() & mne)).Get())
		}

		// cases: setting L to L when d=0 or N<2*L, else set it to N+1-L
		L = (L & ^mle) | ((N + 1 - L) & mle)

		for i := 0; i <= constants.T; i++ {
			// Updating \beta(x) polynomial to \sigma(x) or \beta(x) depending on the case
			// the multiplication with x happens in the loop that shifts down below
			B[i] = fieldelement.New((B[i].Get() & ^mle) | (T[i].Get() & mle))
		}

		// setting \delta field element to either b (\delta) or to d depending on the case
		b = fieldelement.New((b.Get() & ^mle) | (d.Get() & mle))

		// multiply \beta(x) with x polynomial
		for i := constants.T; i >= 1; i-- {
			B[i] = B[i-1]
		}
		B[0] = fieldelement.Zero

	}

	// field inversion to force a monic output polynomial
	var res polynomial.Polynomial
	for i := 0; i < constants.T; i++ {
		res.Elements[i] = C[constants.T-i]
	}
	return res

}
