package auth

import "github.com/golang-jwt/jwt/v5"

type JWTAuthenticator struct {
	secret   string
	audience string
	issue    string
}

func NewJWTAuthenticator(secret, audience, issue string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret,
		audience,
		issue,
	}
}

func (a *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return nil, nil
}
