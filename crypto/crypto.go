package crypto

import (
	"log"
	"golang.org/x/crypto/bcrypt"	
)

// Hash and salt function
func HashAndSalt(password []byte) string {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

// Compare password with stored hash
func ComparePasswords(hash string, password []byte) bool {
	byteHash := []byte(hash)
	err := bcrypt.CompareHashAndPassword(byteHash, password)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

