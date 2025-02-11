package crypto

import "github.com/alexedwards/argon2id"

func Argon2Hash(value string) (string, error) {
	return argon2id.CreateHash(value, argon2id.DefaultParams)
}

func Argon2ValidateHash(hash string) error {
	_, _, _, err := argon2id.DecodeHash(hash)
	return err
}

func Argon2CompareToHash(value, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(value, hash)
}
