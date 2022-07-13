package publickey

import (
	"cme/constants"
	"cme/fieldordering"
	"cme/polynomial"
	"cme/util"
	"errors"
)

const (
	Rows     = constants.M * constants.T
	RowBytes = constants.KBytes
	Len      = Rows * RowBytes
)

type PublicKey struct {
	pk [Len]byte
}

func Generate(g polynomial.Polynomial, alpha fieldordering.FieldOrdering) (PublicKey, error) {
	var mat [Rows][constants.NBytes]byte
	support := alpha.GenerateSupport()
	var b byte

	// fill matrix
	inv := polynomial.Root(g, support)

	for i := 0; i < constants.N; i++ {
		inv[i] = inv[i].Inverse()
	}

	for i := 0; i < constants.T; i++ {
		for j := 0; j < constants.N; j += 8 {

			for k := 0; k < constants.M; k++ {

				b = uint8((inv[j+7].Get() >> k) & 1)
				b <<= 1
				b |= uint8((inv[j+6].Get() >> k) & 1)
				b <<= 1
				b |= uint8((inv[j+5].Get() >> k) & 1)
				b <<= 1
				b |= uint8((inv[j+4].Get() >> k) & 1)
				b <<= 1
				b |= uint8((inv[j+3].Get() >> k) & 1)
				b <<= 1
				b |= uint8((inv[j+2].Get() >> k) & 1)
				b <<= 1
				b |= uint8((inv[j+1].Get() >> k) & 1)
				b <<= 1
				b |= uint8((inv[j+0].Get() >> k) & 1)

				mat[i*constants.M+k][j/8] = b
			}

		}
		// multiply the support
		for j := 0; j < constants.N; j++ {
			inv[j] = inv[j].Mul(support[j])
		}

	}

	// Gaussian
	for i := 0; i < Rows/8; i++ {
		for j := 0; j < 8; j++ {
			row := i*8 + j

			if row >= Rows {
				break
			}

			for k := row + 1; k < Rows; k++ {
				mask := mat[row][i] ^ mat[k][i]
				mask >>= j
				mask &= 1
				mask = -mask

				for c := 0; c < constants.NBytes; c++ {
					mat[row][c] ^= mat[k][c] & mask
				}
			}

			if ((mat[row][i] >> j) & 1) == 0 {
				return PublicKey{}, errors.New("not systematic")
			}

			for k := 0; k < Rows; k++ {
				if k != row {
					mask := mat[k][i] >> j
					mask &= 1
					mask = -mask

					for c := 0; c < constants.NBytes; c++ {
						mat[k][c] ^= mat[row][c] & mask
					}

				}
			}
		}
	}

	var publicKey = PublicKey{}

	for i := 0; i < Rows; i++ { // to 768

		// 340 byte chunk (768 times)

		matRow := mat[i]

		// write the first 340 bytes to publicKey
		for j := 0; j < RowBytes; j++ { //to 340
			publicKey.pk[j+(i*340)] = matRow[96:][j]
		}

	}
	return publicKey, nil

}

func (pk PublicKey) Encapsulate(error_ [constants.NBytes]byte) ([constants.C0Bytes + constants.C1Bytes]byte, [constants.LBytes]byte) {
	var ciphertext [constants.C0Bytes + constants.C1Bytes]byte

	c0 := pk.Encode(error_)
	c1 := util.Hash2v(error_)

	copy(ciphertext[:constants.C0Bytes], c0[:])
	copy(ciphertext[constants.C0Bytes:], c1[:])

	var sessionKey [constants.LBytes]byte
	sessionKeyTmp := util.HashXvC(1, error_, ciphertext)

	copy(sessionKey[:], sessionKeyTmp[:])

	return ciphertext, sessionKey

}

// Encode has input public key, error_ vector
// Return: syndrome s
func (pk PublicKey) Encode(error_ [constants.NBytes]byte) [constants.C0Bytes]byte {
	var s [constants.C0Bytes]byte
	var b uint8
	var row [constants.NBytes]byte

	for i := 0; i < constants.C0Bytes; i++ {
		s[i] = 0
	}
	for i := 0; i < Rows; i++ {
		for j := 0; j < constants.NBytes; j++ {
			row[j] = 0
		}
		for j := 0; j < RowBytes; j++ {
			row[constants.NBytes-RowBytes+j] = pk.pk[(i*RowBytes)+j]
		}

		row[i/8] |= 1 << (i % 8)

		b = 0
		for j := 0; j < constants.NBytes; j++ {
			b ^= row[j] & error_[j]
		}
		b ^= b >> 4
		b ^= b >> 2
		b ^= b >> 1
		b &= 1

		s[i/8] |= b << (i % 8)
	}

	return s
}
