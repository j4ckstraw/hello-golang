package main
import (
    "fmt"
    "encoding/json"
    "net/http"
    "strings"
    "log"
    "time"
)

type weatherData struct {
    Name string `json:"name"`
    Main struct {
        Temp float64 `json:"temp"`
    } `json:"main"`

}
// PLZ CHANGE THIS APIKEY
const APIKEY string = "cd292dd05b715a55396929b1b717c487"

func query(city string) (weatherData, error) {
    url := "http://api.openweathermap.org/data/2.5/weather?APPID=" + APIKEY + "&units=metric&q=" + city
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

func hello(w http.ResponseWriter, r *http.Request){
    w.Write([]byte("hello!\n"))
}

type weatherProvider interface {
    temperature(city string) (float64, error)
}

type openWeatherMap struct {}
func (w openWeatherMap) temperature(city string) (float64, error) {
    resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + APIKEY + "&q=" + city)
    if err != nil {
        return 0,err
    }
    defer resp.Body.Close()

    var d struct {
        Main struct {
            Kelvin float64  `josn:"temp"`
        } `json:"main"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
        return 0, err
    }
    log.Printf("openweathermap: %s: %.2f",city,d.Main.Kelvin)
    return d.Main.Kelvin, nil
}

type weatherUnderground struct {
    apiKey string
}

func (w weatherUnderground) temperature(city string) (float64, error) {
    resp, err := http.Get("http://api.wunderground.com/api/" + w.apiKey + "/conditions/q/" + city + ".json")
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    var d struct {
        Observation struct {
            Celsius float64 `json:"temp_c"`
        } `json:"current_observation"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
        return 0, err
    }
    Kelvin := d.Observation.Celsius + 273.15
    log.Printf("weatherUnderground: %s: %.2f", city, Kelvin)
    return Kelvin, err
}

func temperature(city string, providers ...weatherProvider) (float64, error){
    sum := 0.0

    for _, privider := range providers {
        k, err := privider.temperature(city)
        if err != nil {
            return 0, err
        }

        sum += k
    }
    return sum / float64(len(providers)), nil
}

type multiWeatherProvider []weatherProvider
func (w multiWeatherProvider) temperature(city string) (float64, error){
    sum := 0.0

    for _, privider := range w{
        k, err := privider.temperature(city)
        if err != nil {
            return 0, nil
        }
        sum += k
    }
    return sum / float64(len(w)), nil
}

func main() {
	http.HandleFunc("/weather0/", func(w http.ResponseWriter, r *http.Request){
        city := strings.SplitN(r.URL.Path,"/",3)[2]
        data, err := query(city)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        json.NewEncoder(w).Encode(data)
    })

    mw := multiWeatherProvider {
        openWeatherMap{},
        weatherUnderground{apiKey: APIKEY},
    }

    http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request){
        begin := time.Now()
        city := strings.SplitN(r.URL.Path,"/",3)[2]

        temp, err := mw.temperature(city)
        if err != nil {
            http.Error(w, err.Error(),http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type","application/json; charset=utf-8")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "city": city,
            "temp": temp,
            "took": time.Since(begin).String(),
        })
    })
	http.HandleFunc("/",hello)
    http.ListenAndServe(":8080",nil)
}

