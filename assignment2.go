package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"math/rand"
)

type APIResult struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type NewLocReq struct {
	Address string `json:"address"`
	City    string `json:"city"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type NewLocResp struct {
	ID         int    `json:"id" bson:"_id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type GetLocResp struct {
	ID         int    `json:"id" bson:"_id"`
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type PutLocReq struct {
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

//CREATE
func createLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	
	nlreq := NewLocReq{}
	
	json.NewDecoder(req.Body).Decode(&nlreq)
	
	lat, lng := getLatLong(getUrl(nlreq))
	
	nlres := NewLocResp{}
	nlres.Address = nlreq.Address
	nlres.City = nlreq.City
	nlres.State = nlreq.State
	nlres.ID = getCount()
	nlres.Name = nlreq.Name
	nlres.Zip = nlreq.Zip
	nlres.Coordinate.Lat = lat
	nlres.Coordinate.Lng = lng
	
	resJson, _ := json.Marshal(nlres)
	
	addLocToDB(nlres)
	
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(201)
	fmt.Fprintf(rw, "%s", resJson)
}

//GET
func getLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	
	id := p.ByName("location_id")
	locationId, _ := strconv.Atoi(id)
	
	glres := GetLocResp{}

	glres = getLocFromDB(locationId)
	
	resJson, _ := json.Marshal(glres)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(200)
	fmt.Fprintf(rw, "%s", resJson)

}

//PUT
func putLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	
	id := p.ByName("location_id")
	locationId, _ := strconv.Atoi(id)
	
	preq := PutLocReq{}
	
	json.NewDecoder(req.Body).Decode(&preq)

	pres := updateLocInDB(locationId, preq)
	
	resJson, _ := json.Marshal(pres)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(201)
	fmt.Fprintf(rw, "%s", resJson)

}

//DELETE
func deleteLoc(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	
	id := p.ByName("location_id")
	locationId, _ := strconv.Atoi(id)
	
	session, err1 := mgo.Dial(getdbUrl())
	
	if err1 != nil {
		fmt.Println("Error while connecting to database")
		os.Exit(1)
	} else {
		fmt.Println("Session is Created")
	}

	
	err2 := session.DB("tripplan").C("location").RemoveId(locationId)
	if err2 != nil {
		panic(err2)
	} else {
		fmt.Println("Location deleted is: " + id)
	}

	rw.WriteHeader(200)
}

func addLocToDB(nlres NewLocResp) {
	
	session, err1 := mgo.Dial(getdbUrl())
	
	if err1 != nil {
		fmt.Println("Error while connecting to database")
		os.Exit(1)
	} else {
		fmt.Println("Session is Created")
	}

	
	session.DB("tripplan").C("location").Insert(nlres)
	
	session.Close()
	

}

func getLocFromDB(locationId int) GetLocResp {

	
	session, err1 := mgo.Dial(getdbUrl())
	
	if err1 != nil {
		fmt.Println("Error while connecting to database")
		os.Exit(1)
	} else {
		fmt.Println("Session is Created")
	}

	glres := GetLocResp{}

	err2 := session.DB("tripplan").C("location").FindId(locationId).One(&glres)

	if err2 != nil {
		panic(err2)
	} else {
		fmt.Println("Location retrieved from DB")
	}
	
	session.Close()
	

	return glres

}

func updateLocInDB(locationId int, plr PutLocReq) GetLocResp {

	plres := GetLocResp{}

	
	session, err1 := mgo.Dial(getdbUrl())

	
	if err1 != nil {
		fmt.Println("Error while connecting to database")
		os.Exit(1)
	} else {
		fmt.Println("Session is Created")
	}

	err2 := session.DB("tripplan").C("location").FindId(locationId).One(&plres)

	if err2 != nil {
		panic(err2)
	} else {
		fmt.Println("Location retrieved")
	}

	plres.Address = plr.Address
	plres.City = plr.City
	plres.State = plr.State
	plres.Zip = plr.Zip

	nltemp := NewLocReq{}
	
	nltemp.Address = plres.Address
	nltemp.City = plres.City
	nltemp.State = plres.State
	nltemp.Zip = plres.Zip
	nltemp.Name = plres.Name


	lat, lng := getLatLong(getUrl(nltemp))

	plres.Coordinate.Lat = lat
	plres.Coordinate.Lng = lng

	err3 := session.DB("tripplan").C("location").UpdateId(locationId, plres)

	if err3 != nil {
		panic(err3)
	} else {
		fmt.Println("Location is updated")
	}

	session.Close()
	

	return plres

}

func getUrl(nlreq NewLocReq) string {

	var address string = nlreq.Address
	address = strings.Replace(address, " ", "+", -1)

	var city string = nlreq.City
	city = strings.Replace(city, " ", "+", -1)
	city = ",+" + city

	var state string = nlreq.State
	state = strings.Replace(state, " ", "+", -1)
	state = ",+" + state

	var zip string = nlreq.Zip
	zip = strings.Replace(zip, " ", "+", -1)
	zip = "+" + zip

	
	var Part1 string = "http://maps.google.com/maps/api/geocode/json?address="
	var Part2 string = address + city + state + zip
	var Part3 string = "&sensor=false"

	var url string = Part1 + Part2 + Part3
	

	return url
}

func getLatLong(url string) (float64, float64) {

	result := APIResult{}

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while getting response ", err.Error())
		os.Exit(1)
	}

	
	json.NewDecoder(response.Body).Decode(&result)


	lat := result.Results[0].Geometry.Location.Lat
	lng := result.Results[0].Geometry.Location.Lng

	return lat, lng

}

func getCount() int {
	count := rand.Intn(10000)
	return count
}

func getdbUrl() string {
	var url string = "mongodb://aditya42:aditya42$@ds043694.mongolab.com:43694/tripplan"
	return url
}

func main() {
	mux := httprouter.New()
	mux.GET("/locations/:location_id", getLoc)
	mux.POST("/locations", createLoc)
	mux.PUT("/locations/:location_id", putLoc)
	mux.DELETE("/locations/:location_id", deleteLoc)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
