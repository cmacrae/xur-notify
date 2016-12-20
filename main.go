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

type item struct {
	Name string
	Tier string
	Type string
	Icon string
}

type inventory struct {
	Category string
	Items    []item
}

// Perform an HTTP GET request on the given URL (1st param), using the given API key
// (2nd param) as the value of the header 'X-API-Key', using the given timeout (3rd param) and return the body
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

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return body
}

// Read a file and return its raw contents in an array of byte
func readJSONFromFile(file string) []byte {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return raw
}

// Unmarshal the json body returned from 'getJSON' using the provided URL and API key
// into a map[string]interface{} object and return it.
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

// Unmarshal the json body returned from 'readJSONFromFile' into a map[string]interface{} object and return it.
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

// Send a push notification using Pushover with the contents of Xûr's inventory
func notify(t string, u string, m string) {
	message := &pushover.Message{
		Title:     "Xûr's in town!",
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

	// Iterate over the array of objects in 'saleItemCategories' and perform queries using returned properties
	for i := 0; i < len(saleItemCategories); i++ {
		sicQuery := jsonq.NewQuery(saleItemCategories[i])
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

			inv = itemLookup(inv, saleItems, apiKey)

			t := template.Must(template.New(inv.Category).Parse(invTemplate))
			err = t.Execute(&content, inv)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}

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

	notify(pushoverToken, pushoverUserKey, content.String())
}
