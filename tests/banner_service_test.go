package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const getUrl = "/banner-user?tag_id=1001&feature_id=864"
const host = "http://localhost:8080"
const adminToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MTU2NzQ2NTV9.SIIgpJrvyQDti-jT1uJMIntKFDUhtQ7SVsZLkSsJEPY"
const userToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoidXNlciIsImV4cCI6MTcxNTY3NDY1NX0.xlLDM8mRW0LrvrUMBkadeuB0cI2Abj99syc_P0xVEPc"

var (
	c = http.Client{}
)

func getJson(resp *http.Response) (string, error) {
	var data json.RawMessage
	err := json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		return "", err
	}

	strData, err := data.MarshalJSON()

	if err != nil {
		return "", err
	}

	return string(strData), nil
}

func TestGetNoAuthBanner(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		host+getUrl,
		nil,
	)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}

	resp, err := c.Do(req)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}

	expect := "401"
	status := strings.Split(resp.Status, " ")[0]
	if status != expect {
		fmt.Println(status)
		t.Fail()
	}
}

func TestGetAuthUserBanner(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		host+getUrl,
		nil,
	)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}

	req.Header.Add("token", userToken)
	resp, err := c.Do(req)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}

	expect := "200"
	status := strings.Split(resp.Status, " ")[0]
	if status != expect {
		fmt.Println(status)
		t.Fail()
	}
}

func TestGetBannerContentUser(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		host+getUrl,
		nil,
	)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}

	req.Header.Add("token", userToken)
	resp, err := c.Do(req)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}
	data, err := getJson(resp)
	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}
	expect := "{\"title\":\"some_title\",\"text\":\"some_text\",\"url\":\"some_url\"}"

	if data != expect {
		t.Fail()
	}
}
func TestGetBannerContentAdmin(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		host+getUrl,
		nil,
	)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}

	req.Header.Add("token", adminToken)
	resp, err := c.Do(req)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}
	data, err := getJson(resp)
	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}
	expect := "{\"title\":\"some_title\",\"text\":\"some_text\",\"url\":\"some_url\"}"

	if data != expect {
		t.Fail()
	}
}

func TestGetBannerContentAdminUseLastRevision(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		host+getUrl+"&user_last_revision=true",
		nil,
	)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}

	req.Header.Add("token", adminToken)
	resp, err := c.Do(req)

	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}
	data, err := getJson(resp)
	if err != nil {
		fmt.Println("ERROR")
		t.Fail()
	}
	expect := "{\"title\":\"some_title\",\"text\":\"some_text\",\"url\":\"some_url\"}"

	if data != expect {
		t.Fail()
	}
}
