package main


import (
    "os"
    "fmt"
    "log"
    "time"
    "strconv"
    "net/http"
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/go-redis/redis"
)


// Коды статусов ответов HTTP
const httpBadRequest = 400
const httpOk         = 200

// Коды возвратов Redis
const keyDontExist   = -2
const keyNeverExpire = -1

// Другие настройки
const statKey    = "stats:all:keys"
const configPath = "config.json"
var config JsonConfig


func main() {
    if err := getConfig(configPath); err != nil {
        panic(err)
    }
    router := mux.NewRouter()
    router.HandleFunc("/", mainRequestHandler)
    router.HandleFunc("/stats", statisticsRequestHandler)
    srv := &http.Server{
        Handler: router,
        Addr:    ":8080",
        WriteTimeout: 1000 * time.Millisecond,
        ReadTimeout:  1000 * time.Millisecond,
    }
    log.Fatal(srv.ListenAndServe())
}


func getConfig(file string) error {
    configFile, err := os.Open(file)
    defer configFile.Close()
    if err != nil {
        return err
    }
    if err := json.NewDecoder(configFile).Decode(&config); err != nil {
        return err
    }
    return nil
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
        Addr:     config.RedisAddr,
        Password: config.RedisPassword,
        DB:       config.RedisDB,
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

    client.Close()

    return nil
}


func setStat(request *JsonMainRequest) error {

    client := redis.NewClient(&redis.Options{
        Addr:     config.RedisAddr,
        Password: config.RedisPassword,
        DB:       config.RedisDB,
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

    client.Close()

    return nil
}


func getCount(request *JsonMainRequest) string {

    timeOut         := time.Duration(config.TimeOut) * time.Second
    timeSameRequest := time.Duration(config.TimeSameRequest) * time.Second

    client := redis.NewClient(&redis.Options{
        Addr:     config.RedisAddr,
        Password: config.RedisPassword,
        DB:       config.RedisDB,
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

    client.Close()

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


type JsonConfig struct {
    TimeSameRequest int    `json:"time_same_request"`
    TimeOut         int    `json:"time_out"`
    RedisAddr       string `json:"redis_addr"`
    RedisPassword   string `json:"redis_password"`
    RedisDB         int    `json:"redis_db"`
}
