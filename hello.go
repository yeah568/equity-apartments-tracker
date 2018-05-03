package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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

type buildingSite struct {
	name string
	URL  string
}

var siteURLs = []buildingSite{
	{"HarborSteps", "http://www.equityapartments.com/seattle/downtown-seattle/harbor-steps-apartments"},
	{"Olympus", "http://www.equityapartments.com/seattle/belltown/olympus-apartments"},
	{"Cascade", "http://www.equityapartments.com/seattle/south-lake-union/cascade-apartments"},
}

func getSiteData(ctx context.Context, bs buildingSite, ch chan<- *ApartmentData) (ApartmentData, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(bs.URL)
	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	re := regexp.MustCompile(`\.unitAvailability = (.*);`)

	apartmentJSON := re.FindSubmatch(body)[1]

	var apartments ApartmentData
	err = json.Unmarshal(apartmentJSON, &apartments)

	fmt.Printf("%+v\n", apartments)

	apartments.BuildingName = bs.name

	ch <- &apartments

	return apartments, err
}

func contains(keys []*datastore.Key, testKey *datastore.Key) (bool, int) {
	for i, key := range keys {
		if key.Equal(testKey) {
			return true, i
		}
	}
	return false, -1
}

func update(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	ch := make(chan *ApartmentData)

	for _, URL := range siteURLs {
		go getSiteData(ctx, URL, ch)
	}

	retrievalTime := time.Now()

	for range siteURLs {
		apartments := *<-ch
		fmt.Fprintln(w, apartments.BuildingName)

		buildingKey := datastore.NewKey(ctx, "Building", apartments.BuildingName, 0, nil)
		building := NewBuilding(apartments)

		var empty Building
		if err := datastore.Get(ctx, buildingKey, &empty); err == datastore.ErrNoSuchEntity {
			datastore.Put(ctx, buildingKey, building)
		}

		for _, bedroomType := range apartments.BedroomTypes {

			unitTypeKey := datastore.NewKey(ctx, "UnitType", bedroomType.DisplayName, 0, buildingKey)

			var empty UnitType
			if err := datastore.Get(ctx, unitTypeKey, &empty); err == datastore.ErrNoSuchEntity {
				unitType := &UnitType{Name: bedroomType.DisplayName}
				datastore.Put(ctx, unitTypeKey, unitType)
			}

			var existingUnits []Unit

			existingUnitsKeys, _ := datastore.NewQuery("Unit").Ancestor(unitTypeKey).GetAll(ctx, &existingUnits)

			for _, unit := range bedroomType.AvailableUnits {
				fmt.Fprintf(w, "%s - %s, %d sqft, $%d\r\n", unit.UnitID, bedroomType.DisplayName, unit.SqFt, unit.BestTerm.Price)

				unitKey := datastore.NewKey(ctx, "Unit", unit.UnitID, 0, unitTypeKey)

				// if in existing units:
				exists, index := contains(existingUnitsKeys, unitKey)
				if exists {
					// can keep as available
					// remove from existing units array
					// add new price
					existingUnitsKeys = append(existingUnitsKeys[:index], existingUnitsKeys[index+1:]...)
					existingUnits = append(existingUnits[:index], existingUnits[index+1:]...)
					fmt.Fprintln(w, "exists!")
				} else {
					// insert new entity
					// add new price
					unitData := new(Unit)
					unitData.IsAvailable = true
					FillUnitData(unitData, &unit)
					fmt.Fprintln(w, "new!")
					fmt.Fprintf(w, "%+v\n", unitData)
					datastore.Put(ctx, unitKey, unitData)
				}

				price := &Price{
					Price:         unit.BestTerm.Price,
					TermLength:    unit.BestTerm.Length,
					DateRetrieved: retrievalTime,
				}

				fmt.Fprintf(w, "Inserting %+v\n", *price)

				_, err := datastore.Put(ctx, datastore.NewKey(ctx, "Price", "", price.DateRetrieved.Unix(), unitKey), price)
				if err != nil {
					fmt.Fprintln(w, err)
				}
			}

			// mark all existing units remaining as unavailable
			for i, u := range existingUnitsKeys {
				unitData := existingUnits[i]
				unitData.IsAvailable = false
				datastore.Put(ctx, u, unitData)
			}
		}
	}

	return
}
