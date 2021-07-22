package marketing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/silinternational/wecarry-api/models"
)

const (
	ApiTimeout                = 10 * time.Second
	MailChimpStatusSubscribed = "subscribed"
)

type ApiRequest struct {
	Method      string
	URL         string
	Body        string
	ContentType string
	Username    string
	Password    string
	Headers     map[string]string
}

type MailChimpListMember struct {
	EmailAddress string            `json:"email_address"`
	Status       string            `json:"status"`
	MergeFields  map[string]string `json:"merge_fields"`
	Tags         []string          `json:"tags"`
}

type MailChimpError struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func AddUserToList(u models.User, apiBaseURL, listId, username, password string) error {
	apiURL := fmt.Sprintf("%s/lists/%s/members", apiBaseURL, listId)
	member := MailChimpListMember{
		EmailAddress: u.Email,
		Status:       MailChimpStatusSubscribed,
		MergeFields: map[string]string{
			"FNAME": u.FirstName,
			"LNAME": u.LastName,
		},
		Tags: []string{"Newsletter"},
	}

	reqBody, err := json.Marshal(member)
	if err != nil {
		return err
	}

	resp, err := callApi(ApiRequest{
		Method:      http.MethodPost,
		URL:         apiURL,
		Body:        string(reqBody),
		ContentType: "application/json",
		Username:    username,
		Password:    password,
		Headers:     nil,
	})
	if err != nil {
		return fmt.Errorf("error calling MailChimp API: %s. Response: %s", err.Error(), resp)
	}

	return nil
}

func callApi(apiRequest ApiRequest) (string, error) {
	reqBody := strings.NewReader(apiRequest.Body)
	req, err := http.NewRequest(apiRequest.Method, apiRequest.URL, reqBody)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(apiRequest.Username, apiRequest.Password)
	req.Header.Set("content-type", apiRequest.ContentType)
	req.Header.Set("accept", "application/json")

	for key, val := range apiRequest.Headers {
		req.Header.Set(key, val)
	}

	client := &http.Client{
		Timeout: ApiTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body calling %s %s: %s",
			apiRequest.Method, apiRequest.URL, err.Error())
	}

	if resp.StatusCode > http.StatusNoContent {
		return string(bodyText), fmt.Errorf(
			"unexpected api response status code (%v) calling %s %s",
			resp.StatusCode, apiRequest.Method, apiRequest.URL)
	}

	return string(bodyText), nil
}
