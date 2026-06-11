package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type LoRaDataPacket struct {
	PacketID        string                 `json:"packet_id"`
	DeviceType      string                 `json:"device_type"`
	DeviceID        string                 `json:"device_id"`
	Timestamp       time.Time              `json:"timestamp"`
	Sequence        uint64                 `json:"sequence"`
	Data            map[string]interface{} `json:"data"`
	RSSI            float64                `json:"rssi"`
	SNR             float64                `json:"snr"`
	SpreadingFactor int                    `json:"spreading_factor"`
}

var globalSequence uint64

func generatePacketID(deviceID string, timestamp time.Time, sequence uint64) string {
	if sequence == 0 {
		sequence = uint64(time.Now().UnixNano())
	}

	h := fnv.New64a()
	h.Write([]byte(deviceID))
	h.Write([]byte(fmt.Sprintf("%d", timestamp.UnixNano())))
	h.Write([]byte(fmt.Sprintf("%d", sequence)))

	return fmt.Sprintf("%s-%d-%x", deviceID, timestamp.Unix(), h.Sum64())
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

	reportInterval := 1 * time.Hour
	simInterval := 3 * time.Second

	simStartTime := time.Now().Truncate(reportInterval)
	simHourOffset := 0

	ticker := time.NewTicker(simInterval)
	defer ticker.Stop()

	for range ticker.C {
		reportTime := simStartTime.Add(time.Duration(simHourOffset) * reportInterval)
		hour := reportTime.Hour()

		for _, sensor := range acousticSensors {
			data := generateAcousticData(sensor, hour)
			seq := atomic.AddUint64(&globalSequence, 1)

			packet := LoRaDataPacket{
				PacketID:        generatePacketID(sensor.ID, reportTime, seq),
				DeviceType:      sensor.Type,
				DeviceID:        sensor.ID,
				Timestamp:       reportTime,
				Sequence:        seq,
				Data:            data,
				RSSI:            -70 - rand.Float64()*40,
				SNR:             5 + rand.Float64()*15,
				SpreadingFactor: 7 + rand.Intn(5),
			}

			if err := sendPacket(packet); err != nil {
				fmt.Printf("Error sending acoustic data from %s: %v\n", sensor.ID, err)
			} else {
				fmt.Printf("[%s] Sent acoustic: %s, events: %v\n",
					reportTime.Format("2006-01-02 15:04"), sensor.ID, data["event_count"])
			}

			if rand.Float64() < 0.15 {
				time.Sleep(50 * time.Millisecond)
				if err := sendPacket(packet); err != nil {
					fmt.Printf("Duplicate acoustic packet (expected): %s\n", sensor.ID)
				}
			}
		}

		for _, sensor := range moistureSensors {
			data := generateMoistureData(sensor, hour)
			seq := atomic.AddUint64(&globalSequence, 1)

			packet := LoRaDataPacket{
				PacketID:        generatePacketID(sensor.ID, reportTime, seq),
				DeviceType:      sensor.Type,
				DeviceID:        sensor.ID,
				Timestamp:       reportTime,
				Sequence:        seq,
				Data:            data,
				RSSI:            -65 - rand.Float64()*35,
				SNR:             8 + rand.Float64()*12,
				SpreadingFactor: 7 + rand.Intn(5),
			}

			if err := sendPacket(packet); err != nil {
				fmt.Printf("Error sending moisture data from %s: %v\n", sensor.ID, err)
			} else {
				fmt.Printf("[%s] Sent moisture: %s, moisture: %.1f%%\n",
					reportTime.Format("2006-01-02 15:04"), sensor.ID, data["moisture"])
			}

			if rand.Float64() < 0.1 {
				time.Sleep(30 * time.Millisecond)
				if err := sendPacket(packet); err != nil {
					fmt.Printf("Duplicate moisture packet (expected): %s\n", sensor.ID)
				}
			}
		}

		simHourOffset = (simHourOffset + 1) % 24
		fmt.Println("---")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initSensors()
	runSimulation()
}
