package signing

// SignerByName set a signer by algorithm name.
//
// Deprecated: use SignerByAlgo.
func SignerByName(algo string) Option {
	return SignerByAlgo(algo)
}
