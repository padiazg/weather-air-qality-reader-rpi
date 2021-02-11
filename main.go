// curl -X POST "http://rald-dev.greenbeep.com/api/v1/measurements" -H  "accept: application/json" -H  "X-API-Key: 10768da54e004c3d8d7c144c2a178291" -H  "Content-Type: application/json" -d "[{\"sensor\":\"SPS30\",\"source\":\"Barcequillo 1\",\"description\":\"Barcequillo 1\",\"pm1dot0\":1.015,\"pm2dot5\":4.245,\"pm10\":7.326,\"longitude\":-57.540101,\"latitude\":-25.363301,\"recorded\":\"2021-01-18T22:06:54.673Z\"}]"

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
	"github.com/joho/godotenv"
	sps30 "github.com/padiazg/go-sps30"
)

var (
	lg          = logger.NewPackageLogger("main", logger.InfoLevel)
	url         string
	apiKey      string
	sensor      string
	source      string
	description string
	latitude    float32
	longitude   float32
	sleep       time.Duration
)

// GetEnv get an env variable value
func GetEnv(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		panic("Env " + key + " not set")
	}
	return val
} // getEnv ...

func stringToFloat32(s string) float32 {
	f64, _ := strconv.ParseFloat(s, 32)
	return float32(f64)
} // stringToFloat32 ...

func init() {
	log.Print("Reading env file")
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("Error reading env file")
		log.Print("Error: ", err)
	}

	url = GetEnv("URL")
	apiKey = GetEnv("API_KEY")
	sensor = GetEnv("SENSOR")
	source = GetEnv("SOURCE")
	description = GetEnv("DESCRIPTION")
	latitude = stringToFloat32(GetEnv("LAT"))
	longitude = stringToFloat32(GetEnv("LON"))

	s0, _ := strconv.ParseInt(GetEnv("SLEEP"), 10, 64)
	sleep = time.Duration(s0)

} // initENV ...

func main() {
	defer logger.FinalizeLogger()

	// Create new connection to i2c-bus on 1 line with address 0x69.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x69, 1)
	if err != nil {
		lg.Error("Creating connection...", err)
	}
	defer i2c.Close()

	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	logger.ChangePackageLogLevel("sps30", logger.InfoLevel)

	sensor := sps30.NewSPS30(i2c)

	// Read serial
	serial, err := sensor.ReadSerial()
	lg.Infof("Serial: %s", serial)

	lg.Infof("Starting ticker every %d sec", sleep)
	var ticker *time.Ticker = time.NewTicker(sleep * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Started measurement ", t)
				_ = readMeasurement(sensor)

				// err = sensor.StartMeasurement()
				// time.Sleep(2 * time.Second)

				// // Read if there is new data available
				// var dataReady int = sensor.ReadDataReady()
				// lg.Infof("data-ready: %v", dataReady)

				// if dataReady == 1 {
				// 	// Read measurements
				// 	m, err := sensor.ReadMeasurement()
				// 	if err != nil {
				// 		lg.Errorf("read-measurement: %v", err)
				// 		return
				// 	}
				// 	// print to console
				// 	formatMeasurementHuman(m)
				// 	postMeasurement(m)
				// 	// lg.Infof("data: %v", data)
				// }
				// // Stop measurement, go to idle-mode again
				// err = sensor.StopMeasurement()
				// lg.Infof("Stopped measurement, sleeping...")
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	<-sigs
	log.Println("Service terminating...")
	ticker.Stop()
	done <- true
	fmt.Println("Ticker stopped")
} // main ...

// Formats data for human reading
func formatMeasurementHuman(m *sps30.AirQualityReading) {
	fmt.Printf("pm0.5 count: %8s\n", fmt.Sprintf("%4.3f", m.NumberPM05))
	fmt.Printf("pm1   count: %8s ug: %6s\n", fmt.Sprintf("%4.3f", m.NumberPM1), fmt.Sprintf("%2.3f", m.MassPM1))
	fmt.Printf("pm2.5 count: %8s ug: %6s\n", fmt.Sprintf("%4.3f", m.NumberPM25), fmt.Sprintf("%2.3f", m.MassPM25))
	fmt.Printf("pm4   count: %8s ug: %6s\n", fmt.Sprintf("%4.3f", m.NumberPM4), fmt.Sprintf("%2.3f", m.MassPM4))
	fmt.Printf("pm10  count: %8s ug: %6s\n", fmt.Sprintf("%4.3f", m.NumberPM10), fmt.Sprintf("%2.3f", m.MassPM10))
	fmt.Printf("pm_typ: %4.3f\n", m.TypicalParticleSize)
} // formatMeasurementHuman ...

// func formatMeasurement(m *sps30.AirQualityReading) *Reading {
// 	recorded := time.Now()
// 	lg.Infof("recorded: %s", recorded.Format(time.RFC3339Nano))
// 	return &Reading{{
// 		Sensor:      sensor,
// 		Source:      source,
// 		Description: description,
// 		Latitude:    latitude,
// 		Longitude:   longitude,
// 		Pm1Dot0:     int(m.MassPM10),
// 		Pm2Dot5:     int(m.MassPM25),
// 		Pm10:        int(m.MassPM10),
// 		Recorded:    recorded.Format(time.RFC3339Nano), // "2021-01-18T22:06:54.673Z"
// 	}}
// } // formatMeasurement ...

func readMeasurement(sensor *sps30.SPS30) error {
	// fmt.Println("Started measurement ", t)
	err := sensor.StartMeasurement()
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)

	// Read if there is new data available
	var dataReady int = sensor.ReadDataReady()
	lg.Infof("data-ready: %v", dataReady)

	if dataReady == 1 {
		// Read measurements
		m, err := sensor.ReadMeasurement()
		if err != nil {
			lg.Errorf("read-measurement: %v", err)
			return err
		}
		// print to console
		formatMeasurementHuman(m)
		postMeasurement(m)
		// lg.Infof("data: %v", data)
	}

	// Stop measurement, go to idle-mode again
	err = sensor.StopMeasurement()
	if err != nil {
		return err
	}
	lg.Infof("Stopped measurement, sleeping...")

	return nil
}

func postMeasurement(m *sps30.AirQualityReading) error {

	recorded := time.Now()
	r0, _ := json.Marshal([]struct {
		Sensor      string  `json:"sensor"`      // Model of the device
		Source      string  `json:"source"`      // Name used to identify the device
		Description string  `json:"description"` // User friendly name to identify the device
		Pm1Dot0     int     `json:"pm1dot0"`     // Concentration of PM1.0 inhalable particles per ug/m3
		Pm2Dot5     int     `json:"pm2dot5"`     // Concentration of PM2.5 inhalable particles per ug/m3
		Pm10        int     `json:"pm10"`        // Concentration of PM10 inhalable particles per ug/m3
		Longitude   float32 `json:"longitude"`   // Physical longitude coordinate of the device
		Latitude    float32 `json:"latitude"`    // Physical latitude coordinate of the device
		Recorded    string  `json:"recorded"`    // Date and time for when these values were measured
	}{{
		Sensor:      sensor,
		Source:      source,
		Description: description,
		Latitude:    latitude,
		Longitude:   longitude,
		Pm1Dot0:     int(m.MassPM10),
		Pm2Dot5:     int(m.MassPM25),
		Pm10:        int(m.MassPM10),
		Recorded:    recorded.Format(time.RFC3339Nano), // "2021-01-18T22:06:54.673
	}})

	lg.Info("r0 =>", string(r0))

	// method := "POST"
	payload := strings.NewReader(string(r0))

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		lg.Error(err)
		return err
	}

	req.Header.Add("X-API-Key", apiKey)
	req.Header.Add("Content-Type", "application/javascript")

	res, err := client.Do(req)
	if err != nil {
		lg.Error(err)
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	fmt.Println(string(body))

	return nil
} // postMeasurement ...
