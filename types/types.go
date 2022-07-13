package types

import "cme/constants"

type (
	Seed = [constants.LBytes]byte
	V    = [constants.NBytes]byte
	C    = [constants.CBytes]byte
)
