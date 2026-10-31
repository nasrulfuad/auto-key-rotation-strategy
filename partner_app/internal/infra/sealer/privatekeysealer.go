package sealer

// PrivateKeySealer encrypts private keys at rest.
type PrivateKeySealer interface {
	// Seal encrypts plaintext with a fresh random nonce and returns the ciphertext and that nonce.
	Seal(plaintext []byte) (ciphertext, nonce []byte, err error)

	// Open decrypts ciphertext sealed with nonce. It fails when the nonce has the wrong size or the ciphertext was tampered with.
	Open(ciphertext, nonce []byte) ([]byte, error)
}
