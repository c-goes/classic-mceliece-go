package constants

const (
	M      = 12
	N      = 3488
	NBytes = 436
	SBytes = 436
	T      = 64
	Tau    = 2 * T // doc: 2.4.4
	Q      = 1 << M
	K      = N - M*T
	KBytes = K / 8

	L           = 256
	LBytes      = 32
	Sigma1      = 16
	Sigma1Bytes = 2
	Sigma2      = 32
	Sigma2Bytes = 4

	C0Bytes = (M * T) / 8
	C1Bytes = LBytes
	CBytes  = C0Bytes + C1Bytes
)
