package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak"
	"github.com/golang-jwt/jwt"
)

var ctx = context.Background()

var clientID = "test_service"
var clientSecret = "d7c2424e-7dfc-4a74-a6c5-bd6588ba2d73"
var realm = "test_realm"
var key_uri = "https://iam.cloudmt.co.kr"

func main() {
	http.Handle("/groups", introspect_middleware(http.HandlerFunc(group)))
	http.Handle("/users", introspect_middleware(http.HandlerFunc(users)))
	http.Handle("/secret_groups", introspect_middleware(http.HandlerFunc(secret_groups)))
	http.Handle("/secrets", introspect_middleware(http.HandlerFunc(secrets)))
	http.HandleFunc("/", badRequeset)

	err := http.ListenAndServe(":9255", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func badRequeset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func introspect_middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !keycloak_introspect(w, r) {
			badRequeset(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func group(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("group list"))
		//그룹 목록을 보내주면 됩니다...
		//각 그룹의 유저 수를 보내줘야 합니다.
		//각 그룹별로 상세정보를 뽑아낸 다음에 유저의 수를 직접 세서 추가해서 같이 보내주어야 합니다...
	} else if r.Method == "POST" {
		w.Write([]byte("add group"))
		//그룹 추가를 해주면 됩니다...
	} else if r.Method == "PUT" {
		w.Write([]byte("edit Group"))
		//그룹 을 수정해주면 됩니다...
	} else if r.Method == "DELETE" {
		w.Write([]byte("delete Group"))
		//그룹 을 삭제해주면 됩니다...
	} else {
		badRequeset(w, r)
	}
}

func users(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("user list"))
		//유저 목록을 보내주면 됩니다...
	} else if r.Method == "POST" {
		w.Write([]byte("add user"))
		//유저를 추가해주면 됩니다...
	} else if r.Method == "PUT" {
		w.Write([]byte("edit user"))
		//유저 목록을 보내주면 됩니다...
	} else if r.Method == "DELETE" {
		w.Write([]byte("delete user"))
		//유저 목록을 보내주면 됩니다...
	} else {
		badRequeset(w, r)
	}
}

func secret_groups(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("secret_groups list"))
	} else if r.Method == "POST" {
		w.Write([]byte("add secret_groups"))
	} else if r.Method == "DELETE" {
		w.Write([]byte("delete secret_groups"))
	} else {
		badRequeset(w, r)
	}
}

func secrets(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Write([]byte("secrets list"))
	} else if r.Method == "POST" {
		w.Write([]byte("add secrets"))
	} else if r.Method == "PUT" {
		w.Write([]byte("edit secrets"))
	} else if r.Method == "DELETE" {
		w.Write([]byte("delete secrets"))
	} else {
		badRequeset(w, r)
	}
}

func keycloak_introspect(w http.ResponseWriter, r *http.Request) bool {
	token := r.Header.Get("token")
	if token == "" {
		return false
	}

	username := getUsernameJWT(token)
	if username == "" {
		return false
	}
	// 여기서 구한 username 으로 권한 체크를 해야합니다.
	// 다만 keycloak 설정에 따라 토큰의 내용이 변경될 수도 있으므로 이후 테스트 필요...

	var client = gocloak.NewClient(key_uri)
	rptResult, err := client.RetrospectToken(token, clientID, clientSecret, realm)
	if err != nil {
		return false
	}

	if !rptResult.Active {
		return false
	}

	return true
}

func getUsernameJWT(token string) string {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(""), nil
	})
	if err != nil {
		return ""
	}

	if !t.Valid {
		return ""
	}

	claims := t.Claims.(jwt.MapClaims)
	tmp := claims["preferred_username"]
	username := fmt.Sprintf("%v", tmp)

	return username
}
