package domain

import (
	"fmt"
	"net/http"
	"strings"
)

const ClientIDKey = "client_id"

const ErrorLevelWarn = "warn"
const ErrorLevelError = "error"
const ErrorLevelCritical = "critical"

type AppError struct {
	Err   error
	Code  int
	Level string
}


func GetRequestData(r *http.Request) (map[string][]string, error) {
	data := map[string][]string{}

	if r.Method == "GET" {
		return r.URL.Query(), nil
	}

	if r.Method == "POST" || r.Method == "PUT" {
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			return data, fmt.Errorf("error getting POST data: %v", err.Error())
		}

		data = r.PostForm
	}

	return data, nil
}

func GetFirstStringFromSlice(strSlice []string) string {
	if len(strSlice) > 0 {
		return strSlice[0]
	}

	return ""
}

func GetClientIDFromRequest(r *http.Request) (string, error) {
	return GetRequestParam(ClientIDKey, r)
}

func GetBearerTokenFromRequest(r *http.Request) string {
	bearerKey := "Bearer"

	return GetFirstStringFromSlice(r.Header[bearerKey])
}

func GetRequestParam(paramName string, r *http.Request) (string, error) {
	requestData, err := GetRequestData(r)

	if err != nil {
		return "", err
	}

	return GetFirstStringFromSlice(requestData[paramName]), nil
}

func GetSubPartKeyValues(inString, outerDelimiter, innerDelimiter string) map[string]string {
	keyValues := map[string]string{}
	allPairs := strings.Split(inString, outerDelimiter)

	for _, p := range allPairs {
		pParts := strings.Split(p, innerDelimiter)
		if len(pParts) == 2 {
			keyValues[pParts[0]] = pParts[1]
		}
	}

	return keyValues
}
