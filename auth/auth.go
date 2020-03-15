package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"github.com/viktoriaschule/management-server/log"
)

func CheckUser(username, password string) (bool, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", "https://ldap.vs-ac.de/login", nil)
	if err != nil {
		return false, err
	}
	request.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	response, err := client.Do(request)
	if err != nil {
		return false, errors.Wrap(err, "failed requesting ldap API")
	}
	//noinspection GoUnhandledErrorResult
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		var ldapResponse LdapResponse
		err = json.NewDecoder(response.Body).Decode(&ldapResponse)
		if err != nil {
			log.Errorf("Error parsing json: %v", err)
			os.Exit(1)
		}
		return ldapResponse.Status, nil
	}
	if response.StatusCode == http.StatusUnauthorized {
		return false, nil
	}
	return false, errors.New(fmt.Sprintf("requesting ldap authentication failed with status code %d", response.StatusCode))
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

type LdapResponse struct {
	Status    bool
	Grade     string
	IsTeacher bool
}
