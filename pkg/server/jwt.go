package server

import (
	"fmt"

	jwtlib "github.com/dgrijalva/jwt-go"
)

// TokenEncodeString is the byte string used for encoding/decoding JWT tokens.
// It's recommended to change this value for every use of the package
var TokenEncodeString = []byte("gfe895tu359hjteijte4hjaurhtuh59yjh5e")

// Issuer is the value used as a JWT Claim issuer.
var Issuer = "MyOrganisation"

// Claims contains the JWT StandardClaims and a data map for custom data
type Claims struct {
	Data map[string]interface{}
	jwtlib.StandardClaims
}

// CreateToken will create new JWT token with the provided data
func CreateToken(data map[string]interface{}, expires int64) (string, error) {
	claims := &Claims{
		Data: make(map[string]interface{}),
		StandardClaims: jwtlib.StandardClaims{
			ExpiresAt: expires,
			Issuer:    Issuer,
		},
	}
	for k, v := range data {
		claims.Data[k] = v
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	token.Claims = claims

	return token.SignedString(TokenEncodeString)
}

// ParseToken parses a JWT token and returns the custom data of the token
func ParseToken(t string) (map[string]interface{}, error) {
	token, err := jwtlib.ParseWithClaims(t, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method %v", token.Method)
		}
		return TokenEncodeString, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims := token.Claims.(*Claims)
	return claims.Data, nil
}
