package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"iam/config"
	"iam/iamdb"
	"iam/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type MerlinClaims struct {
	Exp   int64    `json:"exp"`   // (Expiration Time)
	Iat   int64    `json:"iat"`   // (Issued At)
	Iss   string   `json:"iss"`   // (Issuer)
	Sub   string   `json:"sub"`   // (Subject)
	Jti   string   `json:"jti"`   // (JWT ID)
	Uid   string   `json:"uid"`   // (User ID)
	Scope []string `json:"scope"` // (Scope)
	Tid   string   `json:"tid"`   // (Tenant ID)
}

func (MerlinClaims) Valid() error {
	return nil
}

func GetToken(uid, tenantID, sub, code string, scope []string) (string, error) {
	conf := config.GetConfig()

	key, err := loadPrivateKey(conf.Https_keyfile)
	if err != nil {
		return "", err
	}

	now := time.Now()
	jti := uuid.New()
	exp := now.Add(time.Duration(conf.TokenExpirationMinute) * time.Minute)

	claims := MerlinClaims{
		Exp:   exp.Unix(),
		Iat:   now.Unix(),
		Iss:   "Merlin",
		Sub:   sub,
		Jti:   jti.String(),
		Uid:   uid,
		Scope: scope,
		Tid:   tenantID,
	}

	token, err := createJWT(key, claims)
	if err != nil {
		return "", err
	}

	db, err := iamdb.DBClient()
	if err != nil {
		return "", err
	}
	defer db.Close()

	data := models.TokenData{
		TokenId:       claims.Jti,
		TokenTypeCode: code,
		TenantId:      tenantID,
		IssuedAtUTC:   now.Format(time.DateTime),
		Iat:           now.Unix(),
		Issuer:        claims.Iss,
		IssuedUserId:  claims.Uid,
		ExpiredAtUTC:  exp.Format(time.DateTime),
		Exp:           claims.Exp,
		SubjectUserId: claims.Sub,
		ScopeCSV:      scope,
		Token:         token,
	}

	err = iamdb.InsertToken(db, data)
	if err != nil {
		return "", err
	}

	return token, nil
}

func TokenIntrospect(token string) (bool, error) {
	conf := config.GetConfig()

	key, err := loadPublicKeyFromCert(conf.Https_certfile)
	if err != nil {
		return false, err
	}

	err = verifyJWT(token, key)
	if err != nil {
		return false, err
	}

	return true, nil
}

func loadPublicKeyFromCert(certFile string) (*rsa.PublicKey, error) {
	if certFile == "" {
		_, key := generateDummyRSAKey()
		return key, nil
	}

	certData, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("인증서 파일 읽기 실패: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("유효하지 않은 인증서 파일")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("인증서 파싱 실패: %v", err)
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("RSA 공개 키를 찾을 수 없음")
	}

	return publicKey, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		key, _ := generateDummyRSAKey()
		return key, nil
	}

	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPrivateKeyFromPEM(keyData)
}

func createJWT(privateKey *rsa.PrivateKey, claims MerlinClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func verifyJWT(token string, publicKey *rsa.PublicKey) error {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	})

	return err
}

func generateDummyRSAKey() (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("디버깅용 RSA 키 생성 실패: %v", err))
	}

	publicKey := &privateKey.PublicKey
	return privateKey, publicKey
}

func TokenParse(token string) (string, error) {
	tokenId := ""

	t, _ := jwt.Parse(token, nil)
	if t == nil {
		return tokenId, errors.New("invalid authorization")
	}

	claims, _ := t.Claims.(jwt.MapClaims)
	if claims == nil {
		return tokenId, errors.New("invalid token")
	}

	tokenId = fmt.Sprintf("%v", claims["jti"])
	if tokenId == "" {
		return tokenId, errors.New("invalid token")
	}

	return tokenId, nil
}
