package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HTTPGetAndParse(url, bearer string, target interface{}) error {
	// create a new http client
	client := &http.Client{}

	// create a new http request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// set headers
	req.Header.Add("Authorization", "Bearer "+bearer)
	req.Header.Add("User-Agent", UserAgent())

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check the response code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	// parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// unmarshal the response
	err = json.Unmarshal(body, &target)
	if err != nil {
		return err
	}

	return nil
}
