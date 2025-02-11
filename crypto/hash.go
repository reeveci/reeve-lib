package crypto

func Hash(value string) (string, error) {
	return Argon2Hash(value)
}

func ValidateHash(hash string) error {
	return Argon2ValidateHash(hash)
}

func CompareToHash(value, hash string) (bool, error) {
	return Argon2CompareToHash(value, hash)
}
