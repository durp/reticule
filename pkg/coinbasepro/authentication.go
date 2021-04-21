package coinbasepro

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func NewAuth(key string, passphrase string, secret string) *Auth {
	return &Auth{
		Key:        key,
		Passphrase: passphrase,
		Secret:     secret,
	}
}

// Auth is required to sign requests to all private coinbasepro endpoints. Before being able to sign any requests,
// an API key must be created via the Coinbase Pro website.
// The API key will be scoped to a specific profile and assigned specific permissions at creation time.
//  - View - Allows a key read permissions. This includes all GET endpoints.
//  - Transfer - Allows a key to transfer currency on behalf of an account, including deposits and withdrawals.
//    !! Enable Transfer with caution - API key transfers WILL BYPASS two-factor authentication. !!
//  - Trade - Allows a key to enter orders, as well as retrieve trade data. This includes POST /orders and several GET endpoints.
// Upon creating a key and providing a  passphrase, the Secret will be generated.
//
// !! All 3 pieces of information are required to authenticate requests to coinbasepro.
type Auth struct {
	Key        string
	Passphrase string
	Secret     string
}

// Sign creates a sha256 HMAC using the base64-decoded Secret key on the pre-hash string
// `timestamp + method + requestPath + body` (where + represents string concatenation) and then base64 encoding the output.
func (a *Auth) Sign(message string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(a.Secret)
	if err != nil {
		return "", err
	}

	signature := hmac.New(sha256.New, key)
	_, err = signature.Write([]byte(message))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature.Sum(nil)), nil
}
