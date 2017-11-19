package main


import (
	"encoding/json"
    "fmt"
    "log"
    "time"
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
		Bundle string `json:"bundle"`
	} `json:"app"`
	Device struct {
		Ifa string `json:"ifa"`
		Os  string `json:"os"`
		Geo struct {
			Country  string `json:"country"`
		} `json:"geo"`
	} `json:"device"`
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