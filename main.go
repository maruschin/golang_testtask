package main

import (
	"encoding/json"
	//"string"
    "fmt"
    //"html"
    "log"
    "time"
    //"os"
    //"io/ioutil"
    "net/http"
    "github.com/go-redis/redis"
)


// Время, в рамках которого, все запросы будем принимать за один
const timeSameRequest = 5
// Время, после которого необходимо обнулять счетчик
const timeOut = 600

// Коды статусов ответов HTTP
const httpBadRequest = 400
const httpOk         = 200


func main() {
    http.HandleFunc("/", GetSomething)
    http.HandleFunc("/stats", GetStatistics)

    log.Fatal(http.ListenAndServe(":8080", nil))
}


func GetSomething(w http.ResponseWriter, r *http.Request) {

    var req JsonMainRequest
    var res JsonMainResponse
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
    	w.WriteHeader(httpBadRequest)
    	return
    }

    //fmt.Fprintf(w, "Ifa: %q\n", req.Device.Ifa)
    //fmt.Fprintf(w, "Country: %q\n", req.Device.Geo.Country)
    //fmt.Fprintf(w, "App: %q\n", req.App.Bundle)
    //fmt.Fprintf(w, "Platform: %q\n", req.Device.Os)

    count := ExampleNewClient(&req)
    res.Pos = count
    valueB, _ := json.Marshal(&res)
    fmt.Fprintf(w, "%s\n", string(valueB))
}


func GetStatistics(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, take statistics")
}


func InitValue(ifa *IfaData) string {
	ifa.Count = 1
	ifa.Time  = int(time.Now().Unix())
	valueB, _ := json.Marshal(&ifa)
	return string(valueB)
}


func ExampleNewClient(request *JsonMainRequest) int {
	var ifa IfaData

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0, // use default DB
    })
    
    key := request.Device.Ifa

    // Если ключа в БД нет, то заносим первичные данные
    val, err := client.Get(key).Result()
    if err == redis.Nil {
    	val = InitValue(&ifa)
    	err := client.Set(key, val, 0).Err()
    	if err != nil {
    		panic(err)
    	}
    }

    // Разбираем запрос в структуру
    json.Unmarshal([]byte(val), &ifa)

    currentTime := int(time.Now().Unix())
    timeDelta := currentTime - ifa.Time

    if timeSameRequest < timeDelta {
    	ifa.Count++
    	ifa.Time = currentTime
    }
	if timeOut < timeDelta {
		ifa.Count = 0
	}

	// Кладем данные обратно в Redis
    valueB, _ := json.Marshal(&ifa)
    val = string(valueB)
    err = client.Set(key, val, 0).Err()
    if err != nil {
    	panic(err)
    }

    fmt.Println(key, val)

    return ifa.Count
}



type IfaData struct {
	Count int `json:"count"`
	Time  int `json:"time"`
}



type JsonMainRequest struct {
	App struct {
		Bundle    string   `json:"bundle"`
		Cat       []string `json:"cat"`
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		Publisher struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"publisher"`
		Ver string `json:"ver"`
	} `json:"app"`
	At     int      `json:"at"`
	Bcat   []string `json:"bcat"`
	Cur    []string `json:"cur"`
	Device struct {
		Carrier        string `json:"carrier"`
		ConnectionType int    `json:"connectiontype"`
		Geo            struct {
			Accuracy int    `json:"accuracy"`
			City     string `json:"city"`
			Country  string `json:"country"`
			Ext      struct {
				OldGeo struct {
					City    string `json:"city"`
					Country string `json:"country"`
					Region  string `json:"region"`
					Zip     string `json:"zip"`
				} `json:"old_geo"`
			} `json:"ext"`
			IpService int     `json:"ipservice"`
			Lat       float64 `json:"lat"`
			Lon       float64 `json:"lon"`
			Region    string  `json:"region"`
			Type      int     `json:"type"`
			Utcoffset int     `json:"utcoffset"`
			Zip       string  `json:"zip"`
		} `json:"geo"`
		H        int    `json:"h"`
		Ifa      string `json:"ifa"`
		IP       string `json:"ip"`
		Js       int    `json:"js"`
		Language string `json:"language"`
		Make     string `json:"make"`
		Model    string `json:"model"`
		Os       string `json:"os"`
		Osv      string `json:"osv"`
		Pxratio  int    `json:"pxratio"`
		Ua       string `json:"ua"`
		W        int    `json:"w"`
	} `json:"device"`
	Ext struct {
		Envisionx struct {
			Ssp int `json:"ssp"`
		} `json:"envisionx"`
	} `json:"ext"`
	ID  string `json:"id"`
	Imp []struct {
		Banner struct {
			API   []int `json:"api"`
			Battr []int `json:"battr"`
			Btype []int `json:"btype"`
			H     int   `json:"h"`
			Pos   int   `json:"pos"`
			W     int   `json:"w"`
		} `json:"banner"`
		BidFloor          float64 `json:"bidfloor"`
		DisplayManager    string  `json:"displaymanager"`
		DisplayManagerVer string  `json:"displaymanagerver"`
		Ext               struct {
			Brsrclk int `json:"brsrclk"`
			Dlp     int `json:"dlp"`
		} `json:"ext"`
		ID    string `json:"id"`
		Instl int    `json:"instl"`
		Tagid string `json:"tagid"`
		Video struct {
			API           []int    `json:"api"`
			Battr         []int    `json:"battr"`
			CompanionType []int    `json:"companiontype"`
			H             int      `json:"h"`
			Linearity     int      `json:"linearity"`
			MaxDuration   int      `json:"maxduration"`
			Mimes         []string `json:"mimes"`
			MinDuration   int      `json:"minduration"`
			Protocols     []int    `json:"protocols"`
			Sequence      int      `json:"sequence"`
			W             int      `json:"w"`
		} `json:"video"`
	} `json:"imp"`
}


type JsonMainResponse struct {
	Pos int `json:"pos"`
}


type JsonStatResponse struct {
	Statistics []struct {
		Country  string `json:"country"`
		App      string `json:"app"`
		Platform string `json:"platform"`
		Count    int    `json:"count"`
	} `json:"Statistics"`
}