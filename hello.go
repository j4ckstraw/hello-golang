package main
import (
    "fmt"
    "encoding/json"
    "net/http"
    "strings"
)

type weatherData struct {
    Name string `json:"name"`
    Main struct {
        Temp float64 `json:"temp"`
    } `json:"main"`

}
// PLZ CHANGE THIS APIKEY
const APIKEY string = "xxxxxxxxxxxxxxxxxxxxx"

func query(city string) (weatherData, error) {
    url := "http://api.openweathermap.org/data/2.5/weather?APPID=" + APIKEY + "&q=" + city
    fmt.Println(url)
    response, err := http.Get(url)
    if err != nil {
        return weatherData{}, err
    }
    defer response.Body.Close()

    var d weatherData

    if err := json.NewDecoder(response.Body).Decode(&d); err != nil {
        return weatherData{}, err
    }
    return d, nil
}

func main() {
	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request){
        city := strings.SplitN(r.URL.Path,"/",3)[2]
        data, err := query(city)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        json.NewEncoder(w).Encode(data)
    })
	http.HandleFunc("/",hello)
    http.ListenAndServe(":8080",nil)
}

func hello(w http.ResponseWriter, r *http.Request){
    w.Write([]byte("hello!\n"))
}
