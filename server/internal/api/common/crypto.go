package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/rotisserie/eris"
	"golang.org/x/crypto/bcrypt"
)

// func ComputeHMAC256FromInterface(toSign interface{}, secretKey string) (signature string, err error) {

// 	jsonData, err := json.Marshal(toSign)

// 	if err != nil {
// 		return signature, eris.Wrap(err, "ComputeHMAC256FromInterface")
// 	}

// 	return ComputeHMAC256(string(jsonData), secretKey), nil
// }

func ComputeHMAC256(toSign []byte, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(toSign)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// takes a string and secretkey to sign, and compares it to the provided signature
// it can also verify x first characters, that's enough entropy for userID+HMAC verification
func VerifyHMAC(secretKey string, toSign []byte, providedSign string, compareOnlyFirstCharacters int) (isValid bool) {

	signed := ComputeHMAC256(toSign, secretKey)

	// compare all, or if text to sign is smaller than the limit of chars
	if compareOnlyFirstCharacters == 0 || len(toSign) < compareOnlyFirstCharacters {
		return signed == providedSign
	}

	// too much characters to compare

	if len(providedSign) < compareOnlyFirstCharacters {
		return false
	}

	signed = signed[0:8]
	providedSign = providedSign[0:8]

	return signed == providedSign
}

func HashPassword(password string) (hashedPassword string, err error) {

	pwd, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", eris.Wrap(err, "HashPassword error")
	}

	return string(pwd), nil
}

func CheckPasswordHash(password string, hash string) (isValid bool) {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return false
	}
	return true
}

func Sha256Hash(str string) []byte {
	hash := sha256.Sum256([]byte(str))
	return hash[:]
}

// https://golang.org/src/crypto/cipher/example_test.go
func EncryptString(str string, passphrase string) (string, error) {

	data := []byte(str)

	block, _ := aes.NewCipher(Sha256Hash(passphrase))

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return "", eris.Wrap(err, "EncryptString error")
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", eris.Wrap(err, "EncryptString reader error")
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return fmt.Sprintf("%x", ciphertext), nil
}

func Decrypt(data []byte, passphrase string) ([]byte, error) {

	block, err := aes.NewCipher(Sha256Hash(passphrase))

	if err != nil {
		return nil, eris.Wrap(err, "Decrypt new cipher error")
	}

	gcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, eris.Wrap(err, "Decrypt new gcm error")
	}

	nonceSize := gcm.NonceSize()

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return nil, eris.Wrap(err, "Decrypt open gcm error")
	}

	return plaintext, nil
}

func DecryptFromHexString(str string, passphrase string) (string, error) {

	if str == "" {
		return "", eris.New("DecryptFromHexString empty string")
	}

	data, err := hex.DecodeString(str)

	if err != nil {
		return "", err
	}

	decodedBytes, errDec := Decrypt(data, passphrase)

	if errDec != nil {
		return "", errDec
	}

	return string(decodedBytes), nil
}
