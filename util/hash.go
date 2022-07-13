package util

import (
	"cme/constants"
	"cme/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

// Hash2v is a function for hashing
// v is the e error
// domain is hardcoded to 2
func Hash2v(v types.V) [constants.LBytes]byte {
	const domain uint8 = 2

	c1 := sha3.NewShake256()
	_, err := c1.Write([]byte{domain})
	if err != nil {
		log.Panic(err)
	}
	_, err = c1.Write(v[:])
	if err != nil {
		panic(err)
	}

	var result [32]byte

	_, err = c1.Read(result[:])
	if err != nil {
		log.Panic(err)
	}

	return result

}

// HashXvC is for the types of hashing (1,v,C) and (0,v,C)
// with x in [0,1]
// C = (C_0, C_1)
func HashXvC(domain uint8, v types.V, c types.C) [constants.LBytes]byte {

	if domain != 0 && domain != 1 {
		log.Panic("HashXvC has domains 0 and 1")
	}
	c1 := sha3.NewShake256()
	_, err := c1.Write([]byte{domain})
	if err != nil {
		panic(err)
	}
	_, err = c1.Write(v[:])
	if err != nil {
		panic(err)
	}
	_, err = c1.Write(c[:])
	if err != nil {
		panic(err)
	}

	var result [32]byte

	_, err = c1.Read(result[:])
	if err != nil {
		log.Panic(err)
	}
	return result

}
