package main

import (
    "fmt"
    "html"
    "log"
    "io/ioutil"
    "net/http"
    "github.com/go-redis/redis"
)


func main() {
    http.HandleFunc("/", GetSomething)
    http.HandleFunc("/stats", GetStatistics)

    log.Fatal(http.ListenAndServe(":8080", nil))
}


func GetSomething(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q\n", html.EscapeString(r.URL.Path))
    //fmt.Fprintf(w, "%q\n",  r.Method)
    //fmt.Fprintf(w, "%q\n",  r.Header)
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("Error reading body %v", err)
        http.Error(w, "can't read body", http.StatusBadRequest)
    }
    fmt.Fprintf(w, "%q\n",  body)
}


func GetStatistics(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, take statistics")
}


func ExampleNewClient() {
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0, // use default DB
    })

    pong, err := client.Ping().Result()
    fmt.Println(pong, err)
}
