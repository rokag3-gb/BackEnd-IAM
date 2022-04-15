# CLOUDMATE IAM Middleware Server

## 환경변수

- `KEYCLOAK_CLIENT_ID`: 연동할 Keycloak 서버에 등록한 Client의 ID
- `KEYCLOAK_CLIENT_SECRET`: 연동할 Keycloak 서버가 발급한 Client Secret
- `KEYCLOAK_REALM`: 연동할 Keycloak 서버에 등록한 Realm 이름
- `KEYCLOAK_ENDPOINT`: 연동할 Keycloak 서버의 주소

## 실행

실행 전 환경변수 설정 필요

```bash
go get -u

# 빌드 후 실행
go build .
./iam

# 바로 빌드 및 실행
go run .
```
