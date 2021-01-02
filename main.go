package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Weather struct {
	Main struct {
		Temp float64
	}
}

func main() {
	apiKey := os.Getenv("OWM_API_KEY")
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=Hong Kong&appid=%s&units=metric", apiKey)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var weather Weather
		err = json.Unmarshal(body, &weather)
		if err != nil {
			log.Fatal(err)
		}
		t := fmt.Sprintf("%.f°C\n", weather.Main.Temp)
		w.Write([]byte(t))
	})
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
