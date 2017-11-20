package main


import (
    "encoding/json"
    "fmt"
    "log"
    "time"
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
    fmt.Fprintf(w, "Hello, take statistics")
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


func setStat(request *JsonMainRequest) error {

    client := redis.NewClient(&redis.Options{
        Addr:     RedisAddr,
        Password: RedisPassword,
        DB:       RedisDB, 
    })

    var statKey string = "stats:all:keys"
    var key     string = MakeKeyForStatistics(request)

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

    //fmt.Println(client.HGetAll(key).Result())
    //fmt.Println(client.SMembers(statKey).Result())

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

    fmt.Println(key, timeExpire, count)

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
    Statistics []struct {
        Country  string `json:"country"`
        App      string `json:"app"`
        Platform string `json:"platform"`
        Count    int    `json:"count"`
    } `json:"Statistics"`
}