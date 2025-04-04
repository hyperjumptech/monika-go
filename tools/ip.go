package tools

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type GeolocationIP struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Query       string  `json:"query"`
}

func GetGeolocationIP() (*GeolocationIP, error) {
	ip, err := getIp()
	if err != nil {
		return nil, err
	}

	return getGeolocation(*ip)
}

func getIp() (*string, error) {
	http.DefaultClient.Timeout = 3 * time.Second
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := string(ipBytes)
	return &result, nil
}

func getGeolocation(ip string) (*GeolocationIP, error) {
	http.DefaultClient.Timeout = 3 * time.Second
	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var geolocation GeolocationIP
	err = json.NewDecoder(resp.Body).Decode(&geolocation)
	if err != nil {
		return nil, nil
	}

	return &geolocation, nil
}
