package webdrops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/cima-lexis/lexisdn/config"
)

type Session struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string
}

func parseTemplate(tmplS string, args interface{}) (io.Reader, error) {
	tmpl := template.New("query")
	_, err := tmpl.Parse("client_id={{.ClientID}}&grant_type=password&password={{.Password}}&username={{.User}}")
	if err != nil {
		return nil, fmt.Errorf("Error parsing querystring template: %w", err)
	}

	var tmplResult bytes.Buffer

	err = tmpl.Execute(&tmplResult, config.Config)
	if err != nil {
		return nil, fmt.Errorf("Error building querystring: %w", err)
	}

	return strings.NewReader(tmplResult.String()), nil
}

func (sess *Session) Login() error {

	data := url.Values{}
	data.Set("client_id", config.Config.ClientID)
	data.Set("grant_type", "password")
	data.Set("password", config.Config.Password)
	data.Set("username", config.Config.User)

	client := &http.Client{}
	req, err := http.NewRequest("POST", config.Config.AuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error submitting HTTP request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", res.Status)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	//fmt.Println(string(body))
	if err != nil {
		return fmt.Errorf("Error downloading HTTP response: %w", err)
	}
	sess.ClientID = config.Config.ClientID
	return json.Unmarshal(body, sess)
}

func (sess *Session) Refresh() error {

	data := url.Values{}
	data.Set("client_id", config.Config.ClientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", sess.RefreshToken)

	client := &http.Client{}
	req, err := http.NewRequest("POST", config.Config.AuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("Error creating HTTP request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error submitting HTTP request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", res.Status)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Error downloading HTTP response: %w", err)
	}

	return json.Unmarshal(body, sess)

}
