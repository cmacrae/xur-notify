// Copyright 2016 Calum MacRae. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gregdel/pushover"
	"github.com/jmoiron/jsonq"
)

const bnetBaseURL = "https://www.bungie.net/Platform/Destiny/"
const invTemplate = `<u><b>{{.Category}}</b></u>
{{range .Items}}{{.Name}}
{{end}}
`

// item is used for storing item information retrieved from the Bungie API.
type item struct {
	Name string
	Tier string
	Type string
	Icon string
}

// inventory is used to store items categorically (determined by 'itemCategory').
type inventory struct {
	Category string
	Items    []item
}

// getJSON performs an HTTP GET request on the given URL, using the given API key
// as the value of the header 'X-API-Key', whilst adhering to  the given timeout.
// It returns the body in an array of byte.
func getJSON(u string, key string, t int) []byte {
	timeout := time.Duration(t) * time.Second
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req.Header.Set("x-api-key", key)
	req.Header.Set("cache-control", "no-cache")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			fmt.Println(cErr)
			os.Exit(1)
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return body
}

// readJSONFromFile reads a JSON formatted file and returns its raw contents in an array of byte
func readJSONFromFile(file string) []byte {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return raw
}

// exposeJSON unmarshals the json body returned from 'getJSON' using the given URL and API key
// into a map[string]interface{} object and returns it.
func exposeJSON(url string, key string) map[string]interface{} {
	data := map[string]interface{}{}
	byt := getJSON(url, key, 5)
	if err := json.Unmarshal(byt, &data); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dec := json.NewDecoder(strings.NewReader(string(byt)))
	if decErr := dec.Decode(&data); decErr != nil {
		fmt.Println(decErr)
		os.Exit(1)
	}
	return data
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

// notify sends a push notification using Pushover, using the given application & user token, with
// the value of the 1st given string as the message title, and the 2nd as the body.
func notify(t string, u string, h string, m string) {
	message := &pushover.Message{
		Title:     h,
		Message:   m,
		Timestamp: time.Now().Unix(),
		HTML:      true,
	}
	app := pushover.New(t)
	recipient := pushover.NewRecipient(u)
	response, err := app.SendMessage(message, recipient)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Pushover Response:\n%v\n", response)
}

// itemLookup performs an iteration of API queries over an array of maps against the Bungie API.
// On each iteration, certain field data is stored into an item struct, which is then appended to
// an array of items in the given inventory. Finally, the given inventory is returned, with its new
// collection of items.
func itemLookup(inv inventory, a []map[string]interface{}, k string) inventory {
	for i := 0; i < len(a); i++ {
		siQuery := jsonq.NewQuery(a[i])
		itemHash, siErr := siQuery.Int("item", "itemHash")
		if siErr != nil {
			fmt.Println(siErr)
			os.Exit(1)
		}

		// http://bungienetplatform.wikia.com/wiki/DestinyDefinitionType
		hashType := "6"
		itemHashString := fmt.Sprint(itemHash)
		hashReqURL := bnetBaseURL + "Manifest/" + hashType + "/" + itemHashString + "/"

		itemData := exposeJSON(hashReqURL, k)
		idQuery := jsonq.NewQuery(itemData)

		itemName, itemNameErr := idQuery.String("Response", "data", "inventoryItem", "itemName")
		if itemNameErr != nil {
			fmt.Println(itemNameErr)
			os.Exit(1)
		}
		itemType, itemTypeErr := idQuery.String("Response", "data", "inventoryItem", "itemTypeName")
		if itemTypeErr != nil {
			fmt.Println(itemTypeErr)
			os.Exit(1)
		}
		itemTier, itemTierErr := idQuery.String("Response", "data", "inventoryItem", "tierTypeName")
		if itemTierErr != nil {
			fmt.Println(itemTierErr)
			os.Exit(1)
		}
		itemIcon, itemIconErr := idQuery.String("Response", "data", "inventoryItem", "icon")
		if itemIconErr != nil {
			fmt.Println(itemIconErr)
			os.Exit(1)
		}

		thisItem := item{
			Name: itemName,
			Tier: itemTier,
			Type: itemType,
			Icon: itemIcon,
		}

		inv.Items = append(inv.Items, thisItem)
	}
	return inv
}

// generateInvTemplate generates a text template from the given array of map, into the given bytes.Buffer,
// which is then returned. It makes use of the 'itemLookup' function, so an API key should be provided as
// the 3rd parameter.
func generateInvTemplate(b bytes.Buffer, a []map[string]interface{}, k string) bytes.Buffer {
	for i := 0; i < len(a); i++ {
		sicQuery := jsonq.NewQuery(a[i])
		categoryTitle, err := sicQuery.String("categoryTitle")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Just run for dynamic items - no use getting notifications about static stock!
		if categoryTitle == "Exotic Gear" || categoryTitle == "Weapon Ornaments" {
			saleItems, siErr := sicQuery.ArrayOfObjects("saleItems")
			if siErr != nil {
				fmt.Println(siErr)
				os.Exit(1)
			}

			inv := inventory{
				Category: categoryTitle,
			}

			inv = itemLookup(inv, saleItems, k)

			t := template.Must(template.New(inv.Category).Parse(invTemplate))
			err = t.Execute(&b, inv)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

	return b
}

func main() {
	pushoverToken := os.Getenv("PUSHOVER_TOKEN")
	if pushoverToken == "" {
		fmt.Println("The PUSHOVER_TOKEN environment variable is empty!")
		os.Exit(1)
	}
	pushoverUserKey := os.Getenv("PUSHOVER_USER_KEY")
	if pushoverUserKey == "" {
		fmt.Println("The PUSHOVER_USER_KEY environment variable is empty!")
		os.Exit(1)
	}

	t := time.Now()
	today := int(t.Weekday())
	if today < 5 {
		notify(pushoverToken, pushoverUserKey, "Xûr ain't here yet!", "Check your scheduling")
		os.Exit(3)
	}

	apiKey := os.Getenv("BNET_API_KEY")
	if apiKey == "" {
		fmt.Println("The BNET_API_KEY environment variable is empty!")
		os.Exit(1)
	}

	// Open a new JSON query on the data returned from Xûr's 'Advisors' endpoint
	// FROM BNET
	xurURL := bnetBaseURL + "Advisors/Xur/"
	data := exposeJSON(xurURL, apiKey)
	jq := jsonq.NewQuery(data)

	// Open a new JSON query on the dummy data stored in ./dummy_response.json
	// This is for testing purposes only. The contents of this file were produced
	// from a past query to Xûr's 'Advisors' endpoint
	// FROM FILE
	//data := exposeJSONFromFile("dummy_response.json")
	//jq := jsonq.NewQuery(data)

	// Pull the array of objects from $.response.data.saleItemCategories into a var to iterate over
	saleItemCategories, err := jq.ArrayOfObjects("Response", "data", "saleItemCategories")
	if err != nil {
		panic(err)
	}

	// A buffer to collect generated content
	var content bytes.Buffer

	content = generateInvTemplate(content, saleItemCategories, apiKey)

	notify(pushoverToken, pushoverUserKey, "Xûr's in town!", content.String())
}
