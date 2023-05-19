package clients

import (
	"iam/config"
	"io/ioutil"
	"net/http"

	logger "cloudmt.co.kr/mateLogger"
)

func SalesDeleteAccountUser(id string, token string) (string, error) {
	conf := config.GetConfig()

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", conf.Sales_Reqeuest_Url+"/accountUser/"+id, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes)

	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		logger.Error(str)
	}

	return str, nil
}
