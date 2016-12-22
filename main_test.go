// Copyright 2016 Calum MacRae. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Get push notifications detailing what Xûr -- the mysterious travelling salesman
// in Bungie's game: Destiny -- is selling.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/jsonq"
)

const testResponse = "test_response.json"

// readJSONFromFile reads a JSON formatted file and returns its raw contents in an array of byte
func readJSONFromFile(file string) []byte {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return raw
}

// exposeJSONFromFile unmarshals the data returned from 'readJSONFromFile'
// into a map[string]interface{} object and returns it.
func exposeJSONFromFile(file string) map[string]interface{} {
	data := map[string]interface{}{}
	byt := readJSONFromFile(file)
	if err := json.Unmarshal(byt, &data); err != nil {
		panic(err)
	}

	dec := json.NewDecoder(strings.NewReader(string(byt)))
	if decErr := dec.Decode(&data); decErr != nil {
		fmt.Println(decErr)
		os.Exit(1)
	}
	return data
}

// Open a new JSON query on the dummy data stored in ./test_response.json
// This is for testing purposes only. The contents of this file were produced
// from a past query to Xûr's 'Advisors' endpoint
func TestLookup(t *testing.T) {
	testPushoverToken := os.Getenv("TEST_PUSHOVER_TOKEN")
	if testPushoverToken == "" {
		fmt.Println("The TEST_PUSHOVER_TOKEN environment variable is empty!")
		os.Exit(1)
	}
	testPushoverUserKey := os.Getenv("TEST_PUSHOVER_USER_KEY")
	if testPushoverUserKey == "" {
		fmt.Println("The TEST_PUSHOVER_USER_KEY environment variable is empty!")
		os.Exit(1)
	}

	testAPIKey := os.Getenv("TEST_BNET_API_KEY")
	if testAPIKey == "" {
		fmt.Println("The TEST_BNET_API_KEY environment variable is empty!")
		os.Exit(1)
	}

	testdata := exposeJSONFromFile(testResponse)
	jq := jsonq.NewQuery(testdata)
	testItemCategories, err := jq.ArrayOfObjects("Response", "data", "saleItemCategories")
	if err != nil {
		panic(err)
	}
	// A buffer to collect generated content
	var content bytes.Buffer
	content = generateInvTemplate(content, testItemCategories, testAPIKey)
	notify(testPushoverToken, testPushoverUserKey, "Tests Passed!", content.String())
}
