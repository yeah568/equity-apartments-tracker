// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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

type Term struct {
	Length int
	Price  int
}

type Special struct {
	Active  bool
	Title   string
	Expires string
}

type Amenity struct {
	Name string
	Icon string
}

type Matterport struct {
	Key          string
	Id           int
	MediaTags    string
	Caption      string
	DisplayOrder int
	MediaId      int
}

type AvailableUnit struct {
	LedgerID      string `json:"LedgerId"`
	UnitID        string `json:"UnitId"`
	BuildingID    string `json:"BuildingId"`
	AvailableDate string
	BestTerm      Term
	Terms         []Term
	SqFt          int
	Bed           int
	Bath          int
	FloorplanID   string `json:"FloorplanId"`
	FloorplanName string
	Floor         string
	Description   string
	Amenities     []Amenity
	Special       Special
	Floorplan     string
	Photos        []string
	Videos        []string
	Matterports   []Matterport
}

type BedroomType struct {
	ID             int `json:"Id"`
	DisplayName    string
	BedroomCount   int
	AvailableUnits []AvailableUnit
}

type TileOptions struct {
	DisplaySqFt          bool
	DisplayFloorPlanName bool
}

type TileInfo struct {
	Order     int
	IsVisible bool
}

type ApartmentData struct {
	BedroomTypes            []BedroomType `json:"BedroomTypes"`
	PremiumUnits            []BedroomType
	DefaultView             string
	UnitDisplayCount        int
	PrimaryNeighborhoodUrl  string
	MainPhone               string
	IlsPhones               []string
	PaidSearchPhone         string
	TileOptions             TileOptions
	HeaderPricingDisclaimer string
	TileInfo                TileInfo
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
	resp, err := client.Get("http://www.equityapartments.com/seattle/south-lake-union/cascade-apartments")

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
