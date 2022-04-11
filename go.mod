module iam

go 1.17

require (
	cloudmt.co.kr/mateLogger v0.0.0
	github.com/Nerzal/gocloak v1.0.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1 // indirect
	gopkg.in/resty.v1 v1.10.3 // indirect
)

replace cloudmt.co.kr/mateLogger v0.0.0 => ./mateLogger
