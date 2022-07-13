package cme348864

import (
	"cme/constants"
	"cme/fieldordering"
	"cme/polynomial"
	"cme/publickey"
	"cme/secretkey"
	"cme/seedtool"
	"cme/types"
	log "github.com/sirupsen/logrus"
)

type ClassicMcEliece struct {
}

// SeededKeyGen implements NIST Doc section 2.4.3
func (cme ClassicMcEliece) SeededKeyGen(seedPar *types.Seed) (secretkey.SecretKey, publickey.PublicKey) {

	var seed types.Seed = *seedPar

	var alpha fieldordering.FieldOrdering
	var g polynomial.Polynomial
	var pk publickey.PublicKey
	var s [constants.SBytes]byte
	var err error

	for {

		var alphaSeed [constants.Sigma2Bytes * constants.Q]byte
		var gSeed [constants.Sigma1Bytes * constants.T]byte

		c1 := seedtool.SeedBeginTypedReader(0, 64, &seed)

		c1.Read(s[:])
		c1.Read(alphaSeed[:])
		c1.Read(gSeed[:])

		// FieldOrdering algorith
		alpha, err = fieldordering.New(alphaSeed)
		if err != nil {
			log.Debug("field ord failed")
			c1.Read(seed[:])
			continue
		} else {
			log.Info("Success alpha")
		}

		// Irreducible algorithm
		g, err = polynomial.Irreducible(gSeed)
		if err != nil {
			log.Debug("irr poly failed")
			c1.Read(seed[:])
			continue
		} else {
			log.Info("Success irr poly")
		}

		// Public Key Gen
		pk, err = publickey.Generate(g, alpha)
		if err != nil {
			log.Debug("Pubkey gen failed")
			c1.Read(seed[:])
			continue
		} else {
			log.Info("Success pubkey gen")
		}

		break
	}

	sk := secretkey.New(seed, g, alpha, s)

	return sk, pk

}
