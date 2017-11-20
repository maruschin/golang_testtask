package main


import (
    "encoding/json"
    "fmt"
    "log"
    "time"
    "strconv"
    "net/http"
    "github.com/go-redis/redis"
    "crypto/md5"
    "encoding/hex"
)


// Время, в рамках которого, все запросы будем принимать за один
const timeSameRequest = time.Second * 5
// Время, после которого необходимо обнулять счетчик
const timeOut = time.Minute * 10
// Коды статусов ответов HTTP
const httpBadRequest = 400
const httpOk         = 200

// Коды возвратов Redis
const keyDontExist   = -2
const keyNeverExpire = -1

// Настройки Redis
const RedisAddr     = "localhost:6379"
const RedisPassword = "" // no password set
const RedisDB       = 0  // use default DB

// Другие настройки
const statKey = "stats:all:keys"



func main() {
    http.HandleFunc("/", mainRequestHandler)
    http.HandleFunc("/stats", statisticsRequestHandler)

    log.Fatal(http.ListenAndServe(":8080", nil))
}


func mainRequestHandler(w http.ResponseWriter, r *http.Request) {

    var req JsonMainRequest
    var res JsonMainResponse
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(httpBadRequest)
        return
    }

    res.Pos = getCount(&req)
    setStat(&req)
    valueB, _ := json.Marshal(&res)

    fmt.Fprintf(w, "%s\n", string(valueB))
}


func statisticsRequestHandler(w http.ResponseWriter, r *http.Request) {

	var res JsonStatResponse

	if err := getStat(&res); err != nil {
		panic(err)
	}

	valueB, _ := json.Marshal(&res)

	fmt.Fprintf(w, "%s\n", string(valueB))
}


func MakeKeyForStatistics(request *JsonMainRequest) string {
    // Из реквизитов запроса делаем md5 сумму для хранения
    // статистики в БД
    str := fmt.Sprintf(
        "Country-%sPlatform-%sApplication%s",
        request.Device.Geo.Country,
        request.Device.Os,
        request.App.Bundle)
    hasher := md5.New()
    hasher.Write([]byte(str))
    return fmt.Sprintf("stat:%s",hex.EncodeToString(hasher.Sum(nil)))
}


func getStat(response *JsonStatResponse) error {

    client := redis.NewClient(&redis.Options{
        Addr:     RedisAddr,
        Password: RedisPassword,
        DB:       RedisDB,
    })

    keys, _ := client.SMembers(statKey).Result()

    for _, key := range keys {
        entry, _ := client.HGetAll(key).Result()
        count, _ := strconv.Atoi(entry["count"])
        structEntry := JsonStatStatistics{
            Platform: entry["platform"],
            App:      entry["app"],
            Country:  entry["country"],
            Count:    count,
        }
        response.Statistics = append(response.Statistics, structEntry)
    }

    return nil
}


func setStat(request *JsonMainRequest) error {

    client := redis.NewClient(&redis.Options{
        Addr:     RedisAddr,
        Password: RedisPassword,
        DB:       RedisDB, 
    })

    var key string = MakeKeyForStatistics(request)

    if err := client.SAdd(statKey, key).Err(); err != nil {
    	panic(err)
    }

    // Проверяем ключ на наличие и в случае отсутствия, вносим данные
    if val, _ := client.Exists(key).Result(); val == 0 {
        client.HSet(key, "country",  request.Device.Geo.Country).Err()
        client.HSet(key, "platform", request.Device.Os).Err()
        client.HSet(key, "app",      request.App.Bundle).Err()
    }

    if err := client.HIncrBy(key, "count", 1).Err(); err != nil {
    	panic(err)
    }

    return nil
}


func getCount(request *JsonMainRequest) string {

    client := redis.NewClient(&redis.Options{
        Addr:     RedisAddr,
        Password: RedisPassword, 
        DB:       RedisDB,
    })

    var key string = request.Device.Ifa

    timeExpire, _ := client.TTL(key).Result()
    if timeExpire == keyDontExist {
        if err := client.Incr(key).Err(); err != nil {
            panic(err)
        }
        if err := client.Expire(key, timeOut).Err(); err != nil {
            panic(err)
        }
    } else {
        if timeOut - timeExpire > timeSameRequest {
            if err := client.Incr(key).Err(); err != nil {
                panic(err)
            }
        }
        if err := client.Expire(key, timeOut).Err(); err != nil {
            panic(err)
        }
    }

    count, _ := client.Get(key).Result()

    return count
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
    Pos string `json:"pos"`
}


type JsonStatResponse struct {
    Statistics []JsonStatStatistics `json:"statistics"`
}

type JsonStatStatistics struct {
    Country  string `json:"country"`
    App      string `json:"app"`
    Platform string `json:"platform"`
    Count    int    `json:"count"`
}