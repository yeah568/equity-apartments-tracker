package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func main() {
	http.HandleFunc("/", handle)
	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/update" {
		update(w, r)
		return
	}
	fmt.Fprintln(w, "Hello, world!")
}

func update(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	client := urlfetch.Client(ctx)
	resp, err := client.Get("http://www.equityapartments.com/seattle/downtown-seattle/harbor-steps-apartments/")

	if err != nil {
		// handle error, probably
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	re := regexp.MustCompile(`\.unitAvailability = (.*);`)

	apartmentJSON := re.FindSubmatch(body)[1]

	var apartments ApartmentData
	err = json.Unmarshal(apartmentJSON, &apartments)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	// a1 := Apartment{
	// 	Unit:  "222",
	// 	Price: 1950,
	// }

	// key, _ := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "employee", nil), &a1)

	// var a2 Apartment
	// datastore.Get(ctx, key, &a2)

	// fmt.Fprintln(w, string(apartmentJSON[:]))
	fmt.Fprintln(w, apartments)
	return
}
