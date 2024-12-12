package clients

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"iam/config"
	"io"
	"net/http"
	"strings"

	logger "cloudmt.co.kr/mateLogger"
)

//go:embed html/*.html
var Templates embed.FS

type PostAccountUser struct {
	AccountId int64  `json:"accountId"`
	UserId    string `json:"userId"`
	IsUse     bool   `json:"isUse"`
}

type EmailRequest struct {
	Subject    string   `json:"subject"`    // 이메일 제목
	SenderName string   `json:"senderName"` // 발신자 이름
	To         []string `json:"to"`         // 수신자 이메일 주소 리스트
	Cc         []string `json:"cc"`         // 참조 수신자 이메일 주소 리스트
	Bcc        []string `json:"bcc"`        // 숨은 참조 수신자 이메일 주소 리스트
	ReplyTo    []string `json:"replyTo"`    // 답장 이메일 주소 리스트
	Body       string   `json:"body"`       // 이메일 본문
	IsBodyHtml bool     `json:"isBodyHtml"` // 본문이 HTML 형식인지 여부
}

type EmailItem struct {
	From string
	To   string
	URL  string
}

func (e EmailItem) MakeInviteEmailBody() (string, error) {
	data, err := Templates.ReadFile("html/InviteEmail.html")
	if err != nil {
		return "", err
	}

	body := e.replaceText(data)

	return body, nil
}

func (e EmailItem) MakeChangeEmailBody() (string, error) {
	data, err := Templates.ReadFile("html/ChangePasswordEmail.html")
	if err != nil {
		return "", err
	}

	body := e.replaceText(data)

	return body, nil
}

func (e EmailItem) replaceText(data []byte) string {
	body := string(data)

	body = strings.ReplaceAll(body, "{{from}}", e.From)
	body = strings.ReplaceAll(body, "{{to}}", e.To)
	body = strings.ReplaceAll(body, "{{url}}", e.URL)

	return body
}

func (e EmailRequest) SalesSendEmail(token, realm string) (string, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	return salesRequest(token, realm, "POST", "/email/send", body)
}

func SalesDeleteAccountUser(id, realm string, token string) (string, error) {
	return salesRequest(token, realm, "DELETE", "/accountUser/"+id, nil)
}

func SalesPostAccountUser(token, realm string, data PostAccountUser) (string, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return salesRequest(token, realm, "POST", "/accountUser", body)
}

func salesRequest(token, realm, method, url string, body []byte) (string, error) {
	conf := config.GetConfig()
	client := &http.Client{}

	req, err := http.NewRequest(method, conf.Sales_Reqeuest_Url+url, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("X-Target-Realm", realm)

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}

	defer resp.Body.Close()

	bytes, _ := io.ReadAll(resp.Body)
	str := string(bytes)

	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		logger.Error(str)
		return str, fmt.Errorf("sales status error[%d]", resp.StatusCode)
	}

	return str, nil
}
