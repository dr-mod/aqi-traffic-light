package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

const CheckFile = "/var/tmp/check.bin"
const CacheFile = "/var/tmp/output.json"

const AvUrl  = "https://api.airvisual.com/v2/nearest_city?lat=50.4501&lon=30.5234&key=" + YOUR_KEY
const OwmUrl  = "https://api.openweathermap.org/data/2.5/weather?id=703448&appid=" + YOUR_ID

func main() {
	now := time.Now()
	lastTime := ReadLastTime()

	if lastTime.Before(now) {
		aqi, temp, humidity := GetData()
		outside, err := json.Marshal(Outside{temp, humidity, aqi})
		if err == nil {
			saveLastTime(now.Add(20 * time.Minute))
			saveCache(outside)
			fmt.Print(string(outside))
			saveToDB(now, temp, humidity, aqi)
		} else {
			fmt.Print("{}")
		}
	} else {
		fmt.Print(string(readCache()))
	}
}

func saveToDB(now time.Time, temp float64, humidity float64, aqi float64) {
	db, _ := sql.Open("mysql", "USER:PASSWORD@tcp(127.0.0.1:3306)/measurements")
	stmt, _ := db.Prepare("INSERT INTO weather(temp, humidity, aqi) values(?,?,?)")
	stmt.Exec(temp, humidity, aqi)
	db.Close()
}

func GetData() (float64, float64, float64) {
	aqiChanel := make(chan float64)
	tempHumidityChanel := make(chan float64)
	go GetAqi(aqiChanel)
	go GetTempHumidity(tempHumidityChanel)
	aqi := <- aqiChanel
	temp := <- tempHumidityChanel
	humidity := <- tempHumidityChanel
	return aqi, temp, humidity
}

func GetTempHumidity(channel chan<- float64) {
	var temp, humidity float64 = math.NaN(), math.NaN()
	if owm, err := FetchBody(OwmUrl); err == nil {
		owmInterface := ByteToInterface(owm)
		temp = owmInterface["main"].
		(map[string]interface{})["temp"].(float64)
		humidity = owmInterface["main"].
		(map[string]interface{})["humidity"].(float64)
	}
	channel <- temp
	channel <- humidity
}

func GetAqi(channel chan<- float64) {
	var aqi float64 = math.NaN()
	if av, err := FetchBody(AvUrl); err == nil {
		avInterface := ByteToInterface(av)
		aqi = avInterface["data"].
		(map[string]interface{})["current"].
		(map[string]interface{})["pollution"].
		(map[string]interface{})["aqius"].(float64)
	}
	channel <- aqi
}

func readCache() []byte {
	file, err := ioutil.ReadFile(CacheFile)
	if err != nil {
		marshal, _ := json.Marshal(Outside{0, 0, 0})
		return marshal
	}
	return file
}

func saveCache(outside []byte) {
	_ = ioutil.WriteFile(CacheFile, outside, 0644)
}

func saveLastTime(now time.Time) {
	timestampBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(now.Unix()))
	_ = ioutil.WriteFile(CheckFile, timestampBytes, 0644)
}

func ReadLastTime() time.Time {
	oldTimestamp, err := ioutil.ReadFile(CheckFile)
	if err != nil {
		return time.Unix(0, 0)
	}

	var ret int64
	buf := bytes.NewBuffer(oldTimestamp)
	err = binary.Read(buf, binary.LittleEndian, &ret)
	if err != nil {
		return time.Unix(0, 0)
	}

	return time.Unix(ret, 0)
}

func ByteToInterface(body []byte) map[string]interface{} {
	var result map[string]interface{}
	_ = json.Unmarshal([]byte(body), &result)
	return result
}

func FetchBody(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil || response.StatusCode != 200 {
		return nil, errors.New("bad fetch" + url)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil || response.StatusCode != 200 {
		return nil, errors.New("сan't get body " + url)
	}
	return body, nil
}

type Outside struct {
	Temp float64 `json:"temp"`
	Humidity float64 `json:"humidity"`
	Aqi float64 `json:"aqi"`
}
