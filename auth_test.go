package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	testDBPath = "/tmp"
	testUrl    = "http://localhost:8081"
)

/*
curl -v "http://localhost:8080/signin"   -X POST   -d "{\"email\":\"tom\",\"password\":\"test\"}"   -H "Content-Type: application/json"
*/
func TestLogin(t *testing.T) {
	s, c := server(&Opts{Port: 8081, DBPath: testDBPath, Url: testUrl + "/send/email/{email}/{token}"})
	resp := doSignin("tom@test.ch", "test")

	assert.Equal(t, 200, resp.StatusCode)

	defer resp.Body.Close()
	s.Shutdown(context.Background())
	<-c
}

func TestLoginWrongEmail(t *testing.T) {
	s, c := server(&Opts{Port: 8081, DBPath: testDBPath, Url: testUrl + "/send/email/{email}/{token}"})
	resp := doSignin("tomtest.ch", "test")

	assert.Equal(t, 400, resp.StatusCode)
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	assert.Equal(t, 0, strings.Index(bodyString, "SI02"))

	defer resp.Body.Close()
	s.Shutdown(context.Background())
	<-c
}

func TestLoginTwice(t *testing.T) {
	s, c := server(&Opts{Port: 8081, DBPath: testDBPath, Url: testUrl + "/send/email/{email}/{token}"})
	resp := doSignin("tom@test.ch", "test")
	resp.Body.Close()
	resp = doSignin("tom@test.ch", "test")

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, 0, strings.Index(bodyString, "SI05"))

	s.Shutdown(context.Background())
	<-c
}

func TestLoginWrong(t *testing.T) {
	s, c := server(&Opts{Port: 8081, DBPath: testDBPath, Url: testUrl + "/send/email/{email}/{token}"})
	resp := doSignin("tom@test.ch", "test")
	resp.Body.Close()
	resp = doSignin("tom@test.ch", "test")

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)

	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, 0, strings.Index(bodyString, "SI05"))

	s.Shutdown(context.Background())
	<-c
}

func TestConfirm(t *testing.T) {
	s, c := server(&Opts{Port: 8081, DBPath: testDBPath, Url: testUrl + "/send/email/{email}/{token}"})
	resp := doSignin("tom@test.ch", "test")
	assert.Equal(t, 200, resp.StatusCode)

	token := token("tom@test.ch")
	resp = doConfirm("tom@test.ch", token)
	assert.Equal(t, 200, resp.StatusCode)

	s.Shutdown(context.Background())
	<-c
}

func doSignin(email string, pass string) *http.Response {
	type Payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	data := Payload{
		email,
		pass,
	}
	payloadBytes, _ := json.Marshal(data)
	body := bytes.NewReader(payloadBytes)

	req, _ := http.NewRequest("POST", testUrl+"/signin", body)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	return resp
}

func doConfirm(email string, token string) *http.Response {
	c := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, _ := c.Get(testUrl + "/send/email/" + email + "/" + token)
	return resp
}

func token(email string) string {
	r, _ := getEmailToken(email)
	return string(r)
}

func getEmailToken(email string) (string, error) {
	var emailToken string
	err := db.QueryRow("SELECT emailToken from users where email = ?", email).Scan(&emailToken)
	if err != nil {
		return "", err
	}
	return emailToken, nil
}