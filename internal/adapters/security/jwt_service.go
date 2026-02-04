package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService implementa a interface TokenService.
type JWTService struct {
	secretKey []byte
	issuer    string
}

// NewJWTService cria uma nova instância de JWTService.
func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secretKey: []byte(secret),
		issuer:    "rankit-api",
	}
}

// GenerateToken gera um token JWT para o usuário.
func (s *JWTService) GenerateToken(userID string) (string, int64, error) {
	expiresIn := time.Duration(15) * time.Minute // 15 minutos conforme requisitos
	expirationTime := time.Now().Add(expiresIn).Unix()

	claims := jwt.MapClaims{
		"sub": userID,
		"iss": s.issuer,
		"exp": expirationTime,
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", 0, err
	}

	return signedToken, expiresIn.Milliseconds() / 1000, nil // Retorna segundos
}

// ValidateToken valida o token JWT e retorna o ID do usuário.
func (s *JWTService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Valida o método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verifica expiração (já verificado pelo jwt.Parse, mas reforçando)
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return "", errors.New("token expirado")
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			return "", errors.New("token sem ID de usuário (sub)")
		}

		return userID, nil
	}

	return "", errors.New("token inválido")
}
