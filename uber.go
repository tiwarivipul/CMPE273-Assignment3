package main

// Creating the Trip Planner using Uber API.........

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

//Defining structs for geo details and fare details......

type EstimateFare struct {
	StartLatitude  float64
	StartLongitude float64
	EndLatitude    float64
	EndLongitude   float64
	Prices         []FareDetail `json:"prices"`
}

type FareDetail struct {
	ProductId       string  `json:"product_id"`
	CurrencyCode    string  `json:"currency_code"`
	DisplayName     string  `json:"display_name"`
	Estimate        string  `json:"estimate"`
	LowEstimate     int     `json:"low_estimate"`
	HighEstimate    int     `json:"high_estimate"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	Duration        int     `json:"duration"`
	Distance        float64 `json:"distance"`
}

// Uber Authorization Details.....

type DetailRequests struct {
	ServerToken    string
	ClientId       string
	ClientSecret   string
	AppName        string
	AuthorizeUrl   string
	AccessTokenUrl string
	AccessToken    string
	BaseUrl        string
}

// Internal method that implements the Getter interface
func (pricer *EstimateFare) get(c *Client) error {
	fareDetailParam := map[string]string{
		"start_latitude":  strconv.FormatFloat(pricer.StartLatitude, 'f', 2, 32),
		"start_longitude": strconv.FormatFloat(pricer.StartLongitude, 'f', 2, 32),
		"end_latitude":    strconv.FormatFloat(pricer.EndLatitude, 'f', 2, 32),
		"end_longitude":   strconv.FormatFloat(pricer.EndLongitude, 'f', 2, 32),
	}

	data := c.getRequest("estimates/price", fareDetailParam)
	if e := json.Unmarshal(data, &pricer); e != nil {
		return e
	}
	return nil
}

const (
	// Uber API endpoint
	APIUrl string = "https://sandbox-api.uber.com/v1/%s%s"
)

// follow we defines the  interface which is there all HTTP Get requests
type GetInterface interface {
	get(c *Client) error
}

// Defining the Client Struct to manage connections with API.......

type Client struct {
	Options *DetailRequests
}

func Create(options *DetailRequests) *Client {
	return &Client{options}
}

func (c *Client) Get(getter GetInterface) error {
	if e := getter.get(c); e != nil {
		return e
	}

	return nil
}

// Defining the getRequest function.....
func (c *Client) getRequest(endpoint string, params map[string]string) []byte {
	urlParams := "?"
	params["server_token"] = c.Options.ServerToken
	for k, v := range params {
		if len(urlParams) > 1 {
			urlParams += "&"
		}
		urlParams += fmt.Sprintf("%s=%s", k, v)
	}

	url := fmt.Sprintf(APIUrl, endpoint, urlParams)

	res, err := http.Get(url)
	if err != nil {
		//log.Fatal(err)
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	return data
}

type Products struct {
	Latitude  float64
	Longitude float64
	Products  []Product `json:"products"`
}

type Product struct {
	ProductId   string `json:"product_id"`
	Description string `json:"description"`
	DisplayName string `json:"display_name"`
	Capacity    int    `json:"capacity"`
	Image       string `json:"image"`
}

func (pl *Products) get(c *Client) error {
	productParams := map[string]string{
		"latitude":  strconv.FormatFloat(pl.Latitude, 'f', 2, 32),
		"longitude": strconv.FormatFloat(pl.Longitude, 'f', 2, 32),
	}

	data := c.getRequest("products", productParams)
	if e := json.Unmarshal(data, &pl); e != nil {
		return e
	}
	return nil
}

type requestObject struct {
	Id          int
	Name        string `json:"Name"`
	Address     string `json:"Address"`
	City        string `json:"City"`
	State       string `json:"State"`
	Zip         string `json:"Zip"`
	Coordinates struct {
		Lat float64
		Lng float64
	}
}

// Struct for definrd geo locations......

var id int
var tripId int

type ResultInfo struct {
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
		PartialMatch bool     `json:"partial_match"`
		PlaceID      string   `json:"place_id"`
		Types        []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type TripsResponseInfo struct {
	ID                     string   `json:"id"`
	Status                 string   `json:"status"`
	StartingFromLocationID string   `json:"starting_from_location_id"`
	BestRouteLocationIds   []string `json:"best_route_location_ids"`
	TotalUberCosts         int      `json:"total_uber_costs"`
	TotalUberDuration      int      `json:"total_uber_duration"`
	TotalDistance          float64  `json:"total_distance"`
}

type RideDetailRequest struct {
	EndLatitude    string `json:"end_latitude"`
	EndLongitude   string `json:"end_longitude"`
	ProductID      string `json:"product_id"`
	StartLatitude  string `json:"start_latitude"`
	StartLongitude string `json:"start_longitude"`
}

type CurrentTripDetail struct {
	BestRouteLocationIds      []string `json:"best_route_location_ids"`
	ID                        string   `json:"id"`
	NextDestinationLocationID string   `json:"next_destination_location_id"`
	StartingFromLocationID    string   `json:"starting_from_location_id"`
	Status                    string   `json:"status"`
	TotalDistance             float64  `json:"total_distance"`
	TotalUberCosts            int      `json:"total_uber_costs"`
	TotalUberDuration         int      `json:"total_uber_duration"`
	UberWaitTimeEta           int      `json:"uber_wait_time_eta"`
}

// Struct for all the details related to ride......

type RequestDetail struct {
	Driver          interface{} `json:"driver"`
	Eta             int         `json:"eta"`
	Location        interface{} `json:"location"`
	RequestID       string      `json:"request_id"`
	Status          string      `json:"status"`
	SurgeMultiplier int         `json:"surge_multiplier"`
	Vehicle         interface{} `json:"vehicle"`
}

type resObj struct {
	Greeting string
}

func createLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	id = id + 1

	decoder := json.NewDecoder(req.Body)
	var t requestObject
	t.Id = id
	err := decoder.Decode(&t)
	if err != nil {
		fmt.Println("Error")
	}

	st := strings.Join(strings.Split(t.Address, " "), "+")
	fmt.Println(st)
	constr := []string{strings.Join(strings.Split(t.Address, " "), "+"), strings.Join(strings.Split(t.City, " "), "+"), t.State}
	lstringplus := strings.Join(constr, "+")
	locstr := []string{"http://maps.google.com/maps/api/geocode/json?address=", lstringplus}
	resp, err := http.Get(strings.Join(locstr, ""))

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Please Check the address, not valid address......")
	}
	var data ResultInfo
	err = json.Unmarshal(body, &data)
	fmt.Println(data.Status)

	// getting the geo Coordinates........

	t.Coordinates.Lat = data.Results[0].Geometry.Location.Lat
	t.Coordinates.Lng = data.Results[0].Geometry.Location.Lng

	conn, err := mgo.Dial("mongodb://cmpe273:1234@ds057954.mongolab.com:57954/cmpe273")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.SetMode(mgo.Monotonic, true)
	c := conn.DB("cmpe273").C("people")
	err = c.Insert(t)

	js, err := json.Marshal(t)
	if err != nil {
		fmt.Println("Error")
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

func getLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	fmt.Println(p.ByName("locid"))
	id, err1 := strconv.Atoi(p.ByName("locid"))
	if err1 != nil {
		panic(err1)
	}
	conn, err := mgo.Dial("mongodb://cmpe273:1234@ds057954.mongolab.com:57954/cmpe273")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.SetMode(mgo.Monotonic, true)
	c := conn.DB("cmpe273").C("people")
	result := requestObject{}
	err = c.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		fmt.Println(err)
	}

	js, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error")
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

type newRequest struct {
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

func updateLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id, err1 := strconv.Atoi(p.ByName("locid"))

	if err1 != nil {
		panic(err1)
	}
	conn, err := mgo.Dial("mongodb://cmpe273:1234@ds057954.mongolab.com:57954/cmpe273")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.SetMode(mgo.Monotonic, true)
	c := conn.DB("cmpe273").C("people")

	decoder := json.NewDecoder(req.Body)
	var t newRequest
	err = decoder.Decode(&t)
	if err != nil {
		fmt.Println("Error")
	}

	colQuerier := bson.M{"id": id}
	change := bson.M{"$set": bson.M{"address": t.Address, "city": t.City, "state": t.State, "zip": t.Zip}}
	err = c.Update(colQuerier, change)
	if err != nil {
		panic(err)
	}

}

func deleteLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	id, err1 := strconv.Atoi(p.ByName("locid"))

	if err1 != nil {
		panic(err1)
	}
	conn, err := mgo.Dial("mongodb://cmpe273:1234@ds057954.mongolab.com:57954/cmpe273")
	conn.SetMode(mgo.Monotonic, true)
	c := conn.DB("cmpe273").C("people")

	if err != nil {
		panic(err)
	}
	defer conn.Close()
	err = c.Remove(bson.M{"id": id})
	if err != nil {
		fmt.Printf("Could not find kitten %s to delete", id)
	}
	rw.WriteHeader(http.StatusNoContent)
}

type customerUber struct {
	LocationIds            []string `json:"location_ids"`
	StartingFromLocationID string   `json:"starting_from_location_id"`
}

func planTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	decoder := json.NewDecoder(req.Body)
	var uUD customerUber
	err := decoder.Decode(&uUD)
	if err != nil {
		log.Println("Error")
	}

	log.Println(uUD.StartingFromLocationID)

	var options DetailRequests
	options.ServerToken = "Sv4_MIQJELdgc6oS_95puEBbHDSFkTpUS5QUeUfI"
	options.ClientId = "ANGa5yKlM8JAJkjg1Xbszb9yjFCIoOrj"
	options.ClientSecret = "B6ZrI8EJKGQw0zKVZdSaGrO11pUTG5t2Ib8NE4mO"
	options.AppName = "CMPE-273-Assignment3"
	options.BaseUrl = "https://sandbox-api.uber.com/v1/"

	client := Create(&options)

	sid, err1 := strconv.Atoi(uUD.StartingFromLocationID)
	fmt.Println(uUD.StartingFromLocationID)
	fmt.Println(sid)
	if err1 != nil {
		panic(err1)
	}

	conn, err := mgo.Dial("mongodb://cmpe273:1234@ds057954.mongolab.com:57954/cmpe273")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.SetMode(mgo.Monotonic, true)
	c := conn.DB("cmpe273").C("people")
	result := requestObject{}
	err = c.Find(bson.M{"id": sid}).One(&result)
	if err != nil {
		fmt.Println(err)
	}

	index := 0
	totalPrice := 0
	totalDistance := 0.0
	totalDuration := 0
	bestroute := make([]float64, len(uUD.LocationIds))
	m := make(map[float64]string)

	for _, ids := range uUD.LocationIds {

		lid, err1 := strconv.Atoi(ids)

		if err1 != nil {
			panic(err1)
		}

		resultLID := requestObject{}
		err = c.Find(bson.M{"id": lid}).One(&resultLID)
		if err != nil {
			fmt.Println(err)
		}
		objfare := &EstimateFare{}
		objfare.StartLatitude = result.Coordinates.Lat
		objfare.StartLongitude = result.Coordinates.Lng
		objfare.EndLatitude = resultLID.Coordinates.Lat
		objfare.EndLongitude = resultLID.Coordinates.Lng

		if e := client.Get(objfare); e != nil {
			fmt.Println(e)
		}
		totalDistance = totalDistance + objfare.Prices[0].Distance
		totalDuration = totalDuration + objfare.Prices[0].Duration
		totalPrice = totalPrice + objfare.Prices[0].LowEstimate
		bestroute[index] = objfare.Prices[0].Distance
		m[objfare.Prices[0].Distance] = ids
		index = index + 1
	}

	sort.Float64s(bestroute)

	var tripresinfo TripsResponseInfo

	tripId = tripId + 1

	//  Getting all the values related to Trip Response

	tripresinfo.ID = strconv.Itoa(tripId)
	tripresinfo.TotalDistance = totalDistance
	tripresinfo.TotalUberCosts = totalPrice
	tripresinfo.TotalUberDuration = totalDuration
	tripresinfo.Status = "Planning"
	tripresinfo.StartingFromLocationID = strconv.Itoa(sid)
	tripresinfo.BestRouteLocationIds = make([]string, len(uUD.LocationIds))
	index = 0
	for _, ind := range bestroute {
		tripresinfo.BestRouteLocationIds[index] = m[ind]
		index = index + 1
	}
	fmt.Println(tripresinfo.BestRouteLocationIds[1])

	c1 := conn.DB("cmpe273").C("trips")
	err = c1.Insert(tripresinfo)

	js, err := json.Marshal(tripresinfo)
	if err != nil {
		fmt.Println("Error")
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)

}

func getTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	conn, err := mgo.Dial("mongodb://cmpe273:1234@ds057954.mongolab.com:57954/cmpe273")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.SetMode(mgo.Monotonic, true)
	c := conn.DB("cmpe273").C("trips")
	result := TripsResponseInfo{}
	err = c.Find(bson.M{"id": p.ByName("tripid")}).One(&result)
	if err != nil {
		fmt.Println(err)
	}

	js, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error")
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)
}

var currentPos int
var ogtID int

func requestTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	kid, err1 := strconv.Atoi(p.ByName("tripid"))
	var siD int

	if err1 != nil {
		panic(err1)
	}
	var currenttrip CurrentTripDetail
	result1 := requestObject{}
	result2 := requestObject{}
	conn, err := mgo.Dial("mongodb://cmpe273:1234@ds057954.mongolab.com:57954/cmpe273")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.SetMode(mgo.Monotonic, true)
	c := conn.DB("cmpe273").C("trips")
	result := TripsResponseInfo{}

	err = c.Find(bson.M{"id": strconv.Itoa(kid)}).One(&result)
	if err != nil {
		fmt.Println(err)
	} else {

		var iD int

		c1 := conn.DB("cmpe273").C("people")
		if currentPos == 0 {
			iD, err = strconv.Atoi(result.StartingFromLocationID)
			siD = iD
			if err != nil {

				fmt.Println(err)
			}
		} else {
			iD, err = strconv.Atoi(result.BestRouteLocationIds[currentPos-1])
			siD = iD
			if err != nil {

				fmt.Println(err)
			}
		}

		err = c1.Find(bson.M{"id": iD}).One(&result1)
		if err != nil {
			fmt.Println(err)
		}
		iD, err = strconv.Atoi(result.BestRouteLocationIds[currentPos])
		if err != nil {

			fmt.Println(err)
		}
		err = c1.Find(bson.M{"id": iD}).One(&result2)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(result2.Coordinates.Lat)
	}

	currenttrip.ID = strconv.Itoa(ogtID)
	currenttrip.BestRouteLocationIds = result.BestRouteLocationIds
	currenttrip.StartingFromLocationID = strconv.Itoa(siD)
	currenttrip.NextDestinationLocationID = result.BestRouteLocationIds[currentPos]
	currenttrip.TotalDistance = result.TotalDistance
	currenttrip.TotalUberCosts = result.TotalUberCosts
	currenttrip.TotalUberDuration = result.TotalUberDuration
	currenttrip.Status = "requesting"

	var options DetailRequests
	options.ServerToken = "Sv4_MIQJELdgc6oS_95puEBbHDSFkTpUS5QUeUfI"
	options.ClientId = "ANGa5yKlM8JAJkjg1Xbszb9yjFCIoOrj"
	options.ClientSecret = "B6ZrI8EJKGQw0zKVZdSaGrO11pUTG5t2Ib8NE4mO"
	options.AppName = "CMPE-273-Assignment3"
	options.BaseUrl = "https://sandbox-api.uber.com/v1/"
	client := Create(&options)

	pl := Products{}
	pl.Latitude = result1.Coordinates.Lat
	pl.Longitude = result1.Coordinates.Lng
	if e := pl.get(client); e != nil {
		fmt.Println(e)
	}
	var keyid string
	i := 0
	for _, product := range pl.Products {
		if i == 0 {
			keyid = product.ProductId
		}
	}

	var ridedetail RideDetailRequest
	// Geeting all the details for Uber ride....

	ridedetail.StartLatitude = strconv.FormatFloat(result1.Coordinates.Lat, 'f', 6, 64)
	ridedetail.StartLongitude = strconv.FormatFloat(result1.Coordinates.Lng, 'f', 6, 64)
	ridedetail.EndLatitude = strconv.FormatFloat(result2.Coordinates.Lat, 'f', 6, 64)
	ridedetail.EndLongitude = strconv.FormatFloat(result2.Coordinates.Lng, 'f', 6, 64)
	ridedetail.ProductID = keyid
	buf, _ := json.Marshal(ridedetail)
	body := bytes.NewBuffer(buf)
	url := fmt.Sprintf(APIUrl, "requests?", "access_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicmVxdWVzdCJdLCJzdWIiOiJlYWZlZjc2Ny1iOGY1LTQzNGEtYWE4MS00MzhkZWMxZTU4YzYiLCJpc3MiOiJ1YmVyLXVzMSIsImp0aSI6IjY2ZjVlNTUwLWNkZjMtNGUwMS04ZGFkLWI5MzUzODQ2NzhiNSIsImV4cCI6MTQ1MDc1Nzc4OCwiaWF0IjoxNDQ4MTY1Nzg4LCJ1YWN0IjoiU2E3bEkxQ1UxRGFrYzB3alUyUWM2MEdSTHU1eWNJIiwibmJmIjoxNDQ4MTY1Njk4LCJhdWQiOiJBTkdhNXlLbE04SkFKa2pnMVhic3piOXlqRkNJb09yaiJ9.MJK0Xyz0ti5RLhGBwcFqo-st5P6PPJchs0VlLxS5moLr2gvvOyFLjeMNkee_7DKdnzHYj5nnrjd0XgxoaeKYawVOq4NVBMbK3JghDJzrhgddokWSk3iO05H-sSY-PuUV9qDdnTK1zzjpzrL5X9nna-Ly5wwlyZhQWK5KuT6znABxy5W2mwU8YDpCTwcUA0tHpZ_mTlfKYDmpp1ZiG-Sr0NkjcIavaotznr8ejuLZa9bkQy6PexLaMpQ5SBVlpEB3voGMsUjDuUiewxwtX9trsr9yCDOpBopyKQq33y58kzFUczLbXXqVpWFDIvEE2La1Kr-C9Z6Z8ztZ7C8nYEF8aw")
	res, err := http.Post(url, "application/json", body)
	if err != nil {
		fmt.Println(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	var rRes RequestDetail
	err = json.Unmarshal(data, &rRes)
	currenttrip.UberWaitTimeEta = rRes.Eta

	js, err := json.Marshal(currenttrip)
	if err != nil {
		fmt.Println("Error")
		return
	}
	ogtID = ogtID + 1
	currentPos = currentPos + 1
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(js)

}

func main() {
	mux := httprouter.New()

	id = 0
	tripId = 0
	currentPos = 0
	ogtID = 0
	mux.POST("/locations", createLocation)
	mux.POST("/trips", planTrip)
	mux.GET("/locations/:locid", getLocation)
	mux.GET("/trips/:tripid", getTrip)
	mux.PUT("/locations/:locid", updateLocation)
	mux.PUT("/trips/:tripid/request", requestTrip)
	mux.DELETE("/locations/:locid", deleteLocation)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
