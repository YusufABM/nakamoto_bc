// Description: RSA implementation in Go
package rsa

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
)

type SecretKey struct {
	N *big.Int
	D *big.Int
}

type PublicKey struct {
	N *big.Int
	E *big.Int
}

//pk = (n, e) : n = p*q, e is the public exponent

// generate rsa key pair
// k is the number of bits
func Keygen(k int) (PublicKey, SecretKey, error) {
	e := big.NewInt(3) // Public exponent

	var p, q *big.Int
	var err error
	for {
		p, err = rand.Prime(rand.Reader, k/2)
		if err != nil {
			return PublicKey{}, SecretKey{}, err
		}

		q, err = rand.Prime(rand.Reader, k/2)
		if err != nil {
			return PublicKey{}, SecretKey{}, err
		}

		// Ensure gcd(3, p-1) == 1 and gcd(3, q-1) == 1
		if new(big.Int).GCD(nil, nil, e, new(big.Int).Sub(p, big.NewInt(1))).Cmp(big.NewInt(1)) == 0 &&
			new(big.Int).GCD(nil, nil, e, new(big.Int).Sub(q, big.NewInt(1))).Cmp(big.NewInt(1)) == 0 {
			break
		}
	}
	//Calculations
	n := new(big.Int).Mul(p, q)
	ph := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))

	d := new(big.Int).ModInverse(e, ph)
	if d == nil {
		return PublicKey{}, SecretKey{}, err
	}

	//returns the keys
	return PublicKey{N: n, E: e}, SecretKey{N: n, D: d}, nil
}

// Enrypts a message m using the public key pk
func encrypt(m big.Int, pk PublicKey) *big.Int {

	c := new(big.Int).Exp(&m, pk.E, pk.N)

	return c
}

func decrypt(c big.Int, sk SecretKey) *big.Int {
	m := new(big.Int).Exp(&c, sk.D, sk.N)
	return m
}

func SignMessage(message []byte, sk SecretKey) []byte {
	hash := sha256.New()
	hash.Write(message)
	hashedMessage := hash.Sum(nil)
	intHashedMessage := new(big.Int).SetBytes(hashedMessage)
	signature := big.NewInt(0)
	signature = signature.Exp(intHashedMessage, sk.D, sk.N)
	fmt.Println("From RSASign Signature: ", signature)
	fmt.Println("From RSASign Hashed message: ", hashedMessage)

	return signature.Bytes()
}

func VerifySignature(message []byte, signature []byte, pk PublicKey) bool {
	hash := sha256.New()
	hash.Write(message)
	intSignature := new(big.Int).SetBytes(signature)
	hashedMessage := hash.Sum(nil)
	verificationMessage := intSignature.Exp(intSignature, pk.E, pk.N)
	fmt.Println("From RSAVerify Verification message: ", verificationMessage)
	fmt.Println("From RSAVerify Hashed message: ", hashedMessage)
	hm := new(big.Int).SetBytes(hashedMessage)

	if verificationMessage.Cmp(hm) == 0 {
		fmt.Println("Signature is valid")
		return true
	}

	return false
}

func EncodePublicKey(key PublicKey) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(key.E.String() + ":" + key.N.String()))
	return encoded
}

func DecodePublicKey(encodedKey string) PublicKey {
	decode, _ := base64.StdEncoding.DecodeString(encodedKey)
	// Split the encoded key into exponent and modulus
	parts := strings.Split(string(decode), ":")
	if len(parts) != 2 {
		panic(fmt.Errorf("invalid encoded key format"))
	}

	// Convert the exponent (E) back to an int
	e := new(big.Int)
	e.SetString(parts[0], 10)

	// Convert the modulus (N) back to a big.Int
	n := new(big.Int)
	n.SetString(parts[1], 10) // b	ase 10

	// Construct the public key
	pubKey := PublicKey{
		N: n,
		E: e,
	}
	return pubKey
}
