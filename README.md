# Ahnlab Cloudmate Merlin API Server

## 개요
Merlin API Server 는 Keycloak, Vault 와 연동하여 사용자 정보 및 권한정보를 관리할 수 있는 API 를 제공합니다.

## Keycloak 연동 기능
Keycloak 연동기능으로는 User, Group, Service Account 등을 관리, 조회 할 수 있습니다.

## Vault 연동 기능
Vault 기능으로는 Secret 등을 관리, 조회 할 수 있습니다.

## Sale Backend 연동 기능
Sale Backend 시스템과 연동하여 Account 를 관리하는 기능이 있습니다.

## 자체 기능
Keycloak, Vault 에 속하지 않는 Merlin의 자체 기능으로는 Role, Authority 등이 있습니다. 
Role 기능은 Keycloak 의 Role 과는 다른 기능이며 Database 상에서도 데이터가 분리되어있습니다.