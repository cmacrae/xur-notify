package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bdenning/go-pushover"
	"github.com/jmoiron/jsonq"
)

const bnetBaseUrl = "https://www.bungie.net/Platform/Destiny/"

type Item struct {
	Name string
	Tier string
	Type string
	Icon string
}

type Inventory struct {
	Category string
	Items    []Item
}

// Perform an HTTP GET request on the given URL (1st param), using the given API key
// (2nd param) as the value of the header 'X-API-Key', using the given timeout (3rd param) and return the body
func getJson(u string, key string, t int) []byte {
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

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return body
}

// Read a file and return its raw contents in an array of byte
func readJsonFromFile(file string) []byte {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return raw
}

// Unmarshal the json body returned from 'getJson' using the provided URL and API key
// into a map[string]interface{} object and return it.
func exposeJson(url string, key string) map[string]interface{} {
	data := map[string]interface{}{}
	byt := getJson(url, key, 5)
	if err := json.Unmarshal(byt, &data); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dec := json.NewDecoder(strings.NewReader(string(byt)))
	dec.Decode(&data)
	return data
}

// Unmarshal the json body returned from 'readJsonFromFile' into a map[string]interface{} object and return it.
func exposeJsonFromFile(file string) map[string]interface{} {
	data := map[string]interface{}{}
	byt := readJsonFromFile(file)
	if err := json.Unmarshal(byt, &data); err != nil {
		panic(err)
	}

	dec := json.NewDecoder(strings.NewReader(string(byt)))
	dec.Decode(&data)
	return data
}

func notify(token string, user string, message string) {
	msg := pushover.NewMessage(token, user)
	msg.Push(message)
}

func main() {
	t := time.Now()
	today := int(t.Weekday())
	if today < 4 {
		fmt.Println("Xûr ain't here yet!")
		os.Exit(1)
	}

	apiKey := os.Getenv("BNET_API_KEY")
	if apiKey == "" {
		fmt.Println("The BNET_API_KEY environment variable is empty!")
		os.Exit(1)
	}

	// Open a new JSON query on the data returned from Xûr's 'Advisors' endpoint
	// FROM BNET
	xurUrl := bnetBaseUrl + "Advisors/Xur/"
	data := exposeJson(xurUrl, apiKey)
	jq := jsonq.NewQuery(data)

	// Open a new JSON query on the dummy data stored in ./dummy_response.json
	// This is for testing purposes only. The contents of this file were produced
	// from a past query to Xûr's 'Advisors' endpoint
	// FROM FILE
	//data := exposeJsonFromFile("dummy_response.json")
	//jq := jsonq.NewQuery(data)

	// Pull the array of objects from $.response.data.saleItemCategories into a var to iterate over
	saleItemCategories, err := jq.ArrayOfObjects("Response", "data", "saleItemCategories")
	if err != nil {
		panic(err)
	}

	// Iterate over the array of objects in 'saleItemCategories' and perform queries using returned properties
	for i := 0; i < len(saleItemCategories); i++ {
		sicQuery := jsonq.NewQuery(saleItemCategories[i])
		categoryTitle, err := sicQuery.String("categoryTitle")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		saleItems, err := sicQuery.ArrayOfObjects("saleItems")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		inv := Inventory{
			Category: categoryTitle,
		}

		for i := 0; i < len(saleItems); i++ {
			siQuery := jsonq.NewQuery(saleItems[i])
			itemHash, err := siQuery.Int("item", "itemHash")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// http://bungienetplatform.wikia.com/wiki/DestinyDefinitionType
			hashType := "6"
			itemHashString := fmt.Sprint(itemHash)
			hashReqUrl := bnetBaseUrl + "Manifest/" + hashType + "/" + itemHashString
			fmt.Println(hashReqUrl)

			itemData := exposeJson(hashReqUrl, apiKey)
			idQuery := jsonq.NewQuery(itemData)

			itemName, _ := idQuery.String("Response", "data", "inventoryItem", "itemName")
			itemType, _ := idQuery.String("Response", "data", "inventoryItem", "itemTypeName")
			itemTier, _ := idQuery.String("Response", "data", "inventoryItem", "tierTypeName")
			itemIcon, _ := idQuery.String("Response", "data", "inventoryItem", "icon")

			item := Item{
				Name: itemName,
				Tier: itemTier,
				Type: itemType,
				Icon: itemIcon,
			}

			inv.Items = append(inv.Items, item)
		}

		fmt.Printf("*%v*\n", inv.Category)
		for _, v := range inv.Items {
			fmt.Printf("%v - %v - %v\n", v.Name, v.Type, v.Tier)
		}
	}
}
