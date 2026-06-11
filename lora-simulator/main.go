package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type LoRaDataPacket struct {
	DeviceType      string                 `json:"device_type"`
	DeviceID        string                 `json:"device_id"`
	Timestamp       time.Time              `json:"timestamp"`
	Data            map[string]interface{} `json:"data"`
	RSSI            float64                `json:"rssi"`
	SNR             float64                `json:"snr"`
	SpreadingFactor int                    `json:"spreading_factor"`
}

type SensorConfig struct {
	ID        string
	Type      string
	Building  string
	Location  string
	BaseValue float64
	Variance  float64
}

var apiURL = "http://localhost:8080/api/v1/lora/data"

var acousticSensors []SensorConfig
var moistureSensors []SensorConfig

func initSensors() {
	buildings := []string{"应县木塔", "佛光寺"}

	for _, building := range buildings {
		acousticCount := 25
		moistureCount := 20

		if building == "应县木塔" {
			acousticCount = 30
			moistureCount = 25
		}

		for i := 0; i < acousticCount; i++ {
			sensorID := generateSensorID("AC", building, i+1)
			acousticSensors = append(acousticSensors, SensorConfig{
				ID:        sensorID,
				Type:      "acoustic_emission",
				Building:  building,
				Location:  getLocation(building, i),
				BaseValue: 30 + rand.Float64()*50,
				Variance:  20,
			})
		}

		for i := 0; i < moistureCount; i++ {
			sensorID := generateSensorID("MS", building, i+1)
			moistureSensors = append(moistureSensors, SensorConfig{
				ID:        sensorID,
				Type:      "wood_moisture",
				Building:  building,
				Location:  getLocation(building, i),
				BaseValue: 15 + rand.Float64()*8,
				Variance:  3,
			})
		}
	}
}

func generateSensorID(prefix, building string, num int) string {
	buildingCode := "YMT"
	if building == "佛光寺" {
		buildingCode = "FGS"
	}
	return fmt.Sprintf("%s-%s-%03d", prefix, buildingCode, num)
}

func getLocation(building string, index int) string {
	locations := []string{
		"一层斗拱", "二层斗拱", "三层斗拱", "四层斗拱", "五层斗拱",
		"东立柱", "西立柱", "南立柱", "北立柱",
		"主梁", "次梁", "横梁",
		"东侧墙", "西侧墙", "南侧墙", "北侧墙",
		"塔顶", "塔基",
	}

	if building == "佛光寺" {
		locations = []string{
			"东大殿斗拱", "东大殿立柱", "东大殿主梁",
			"文殊殿斗拱", "文殊殿立柱", "文殊殿主梁",
			"山门", "钟楼", "鼓楼",
			"配殿", "藏经楼",
		}
	}

	return locations[index%len(locations)]
}

func generateAcousticData(sensor SensorConfig, hour int) map[string]interface{} {
	hourFactor := 1.0 + 0.5*math.Sin(float64(hour)*math.Pi/12)
	eventCount := int(sensor.BaseValue*hourFactor + rand.Float64()*sensor.Variance)

	highRisk := rand.Float64() < 0.1
	if highRisk {
		eventCount = int(sensor.BaseValue * 3)
	}

	data := map[string]interface{}{
		"building":       sensor.Building,
		"location":       sensor.Location,
		"event_count":    eventCount,
		"energy":         100 + rand.Float64()*900,
		"amplitude":      40 + rand.Float64()*60,
		"duration":       1 + rand.Float64()*10,
		"rise_time":      0.1 + rand.Float64()*2,
		"counts":         10 + rand.Intn(100),
		"frequency_peak": 1000 + rand.Float64()*8000,
	}

	return data
}

func generateMoistureData(sensor SensorConfig, hour int) map[string]interface{} {
	seasonFactor := 1.0 + 0.3*math.Sin(float64(time.Now().Month())*math.Pi/6)
	diurnalFactor := 1.0 + 0.1*math.Sin(float64(hour)*math.Pi/12)

	moisture := sensor.BaseValue * seasonFactor * diurnalFactor
	moisture += (rand.Float64() - 0.5) * sensor.Variance

	highMoisture := rand.Float64() < 0.05
	if highMoisture {
		moisture = 28 + rand.Float64()*5
	}

	data := map[string]interface{}{
		"building":    sensor.Building,
		"location":    sensor.Location,
		"moisture":    math.Max(5, math.Min(40, moisture)),
		"temperature": 15 + rand.Float64()*15,
	}

	return data
}

func sendPacket(packet LoRaDataPacket) error {
	jsonData, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to marshal packet: %w", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send packet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func runSimulation() {
	fmt.Println("Starting LoRa sensor simulator...")
	fmt.Printf("API URL: %s\n", apiURL)
	fmt.Printf("Acoustic sensors: %d\n", len(acousticSensors))
	fmt.Printf("Moisture sensors: %d\n", len(moistureSensors))
	fmt.Println("Press Ctrl+C to stop")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	hour := 0
	for range ticker.C {
		now := time.Now()

		for _, sensor := range acousticSensors {
			data := generateAcousticData(sensor, hour)

			packet := LoRaDataPacket{
				DeviceType:      sensor.Type,
				DeviceID:        sensor.ID,
				Timestamp:       now,
				Data:            data,
				RSSI:            -70 - rand.Float64()*40,
				SNR:             5 + rand.Float64()*15,
				SpreadingFactor: 7 + rand.Intn(5),
			}

			if err := sendPacket(packet); err != nil {
				fmt.Printf("Error sending acoustic data from %s: %v\n", sensor.ID, err)
			} else {
				fmt.Printf("Sent acoustic data: %s, events: %v\n", sensor.ID, data["event_count"])
			}
		}

		for _, sensor := range moistureSensors {
			data := generateMoistureData(sensor, hour)

			packet := LoRaDataPacket{
				DeviceType:      sensor.Type,
				DeviceID:        sensor.ID,
				Timestamp:       now,
				Data:            data,
				RSSI:            -65 - rand.Float64()*35,
				SNR:             8 + rand.Float64()*12,
				SpreadingFactor: 7 + rand.Intn(5),
			}

			if err := sendPacket(packet); err != nil {
				fmt.Printf("Error sending moisture data from %s: %v\n", sensor.ID, err)
			} else {
				fmt.Printf("Sent moisture data: %s, moisture: %.1f%%\n", sensor.ID, data["moisture"])
			}
		}

		hour = (hour + 1) % 24
		fmt.Println("---")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initSensors()
	runSimulation()
}
