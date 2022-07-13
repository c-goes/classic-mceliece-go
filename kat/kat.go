package kat

import (
	"cme/cme348864"
	"cme/constants"
	"cme/types"
	"encoding/hex"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
)

type Kat struct {
	Count int    `json:"count"`
	Seed  string `json:"seed"`
	Pk    string `json:"pk"`
	Sk    string `json:"sk"`
	Ct    string `json:"ct"`
	Ss    string `json:"ss"` // key
	E     string `json:"e"`  // expected e bytes
}

func KatRun() {
	file, err := os.Open("kats.json")
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	kats := []Kat{}
	err = decoder.Decode(&kats)
	if err != nil {
		panic(err)
	}
	for _, kat := range kats {

		seedTmp, err := hex.DecodeString(kat.Seed)

		if err != nil {
			panic(err)
		}

		pk, err := hex.DecodeString(kat.Pk)
		if err != nil {
			panic(err)
		}
		sk, err := hex.DecodeString(kat.Sk)
		if err != nil {
			panic(err)
		}
		ct, err := hex.DecodeString(kat.Ct)
		if err != nil {
			panic(err)
		}
		ss, err := hex.DecodeString(kat.Ss)
		if err != nil {
			panic(err)
		}
		e, err := hex.DecodeString(kat.E)
		if err != nil {
			panic(err)
		}

		log.Debug("NIST KAT", kat.Count)
		log.Debug("len seed", len(seedTmp[:]))
		log.Debug("len sk", len(sk))
		log.Debug("len pk", len(pk))
		log.Debug("len ct", len(ct))
		log.Debug("len ss", len(ss))
		log.Debug("len e", len(e))

		if constants.N/8 == len(e) {
			log.Infof("e has length %d bit, correct for cme348864", len(e)*8)
		}

		cme := cme348864.ClassicMcEliece{}

		// convert the seed slice to array pointer:
		secretKey, publicKey := cme.SeededKeyGen((*types.Seed)(seedTmp))
		_ = secretKey

		var eTest [constants.NBytes]byte
		copy(eTest[:], e)
		ciphertext, sessionKey := publicKey.Encapsulate(eTest)
		_ = ciphertext
		log.Info("ciphertext", ciphertext)
		log.Info("session key", sessionKey)
		//log.Info("e", eTest)
		log.Info("expected ss", ss)

		var ssArray [32]byte
		copy(ssArray[:], ss)
		if sessionKey != ssArray {
			log.Fatal("Encryption of e failed")
		} else {
			log.Info("Encryption successful")
		}

		/*		// Test decryption 1
				var ctArray [128]byte
				copy(ctArray[:], ct)
				decryptedKey1 := secretKey.Decapsulate(ctArray)
				log.Info("decrypted 1:", decryptedKey1)

				if decryptedKey1 != ssArray {
					log.Fatal("Decryption 1 of ciphertext failed")
				} else {
					log.Info("Decryption 1 successful")
				}*/

		// Test decryption 2
		decryptedKey := secretKey.Decapsulate(ciphertext)
		log.Info("decrypted 2:", decryptedKey)

		if decryptedKey != ssArray {
			log.Fatal("Decryption 2 of ciphertext failed")
		} else {
			log.Info("Decryption 2 successful")
		}
	}
}
