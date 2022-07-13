package seedtool

import (
	"cme/types"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"io"
	"io/ioutil"
)

// SeedBegin has optional skip (>0),
// out must pointing to a make([]byte, 32)
// seed is hexstring
func SeedBegin(skip int64, domain uint8, seed string, out *[]byte) {
	katsSeed := seed

	decodedKatsSeed, err := hex.DecodeString(katsSeed)
	if err != nil {
		panic(err)
	}
	SeedBeginTyped(skip, domain, (*types.Seed)(decodedKatsSeed), out)
}

func SeedBeginTypedReader(skip int64, domain uint8, seed *types.Seed) sha3.ShakeHash {
	c1 := sha3.NewShake256()
	_, err := c1.Write([]byte{domain})
	if err != nil {
		panic(err)
	}
	_, err = c1.Write(seed[:])
	if err != nil {
		panic(err)
	}

	if skip < 0 {
		log.Fatal("no negative skip")
	}
	if skip > 0 {
		_, err = io.CopyN(ioutil.Discard, c1, skip)
		if err != nil {
			panic(err)
		}
	}

	return c1

}

func SeedBeginTyped(skip int64, domain uint8, seed *types.Seed, out *[]byte) {
	c1 := SeedBeginTypedReader(skip, domain, seed)

	_, err := c1.Read(*out)
	if err != nil {
		panic(err)
	}
}
