package main

import (
	"ancient-wood-monitor/config"
	"ancient-wood-monitor/internal/handlers"
	"ancient-wood-monitor/internal/services"
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := getConfigPath()
	if err := config.LoadConfig(configPath); err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		log.Println("Using default configuration")
	}

	gin.SetMode(config.AppConfig.Server.Mode)

	influxDBService, err := services.NewInfluxDBService()
	if err != nil {
		log.Printf("Warning: Failed to connect to InfluxDB: %v", err)
		log.Println("Running in mock mode - some features may not work properly")
	} else {
		defer influxDBService.Close()
		log.Println("Successfully connected to InfluxDB")
	}

	alertService := services.NewAlertService(influxDBService)
	sensorService := services.NewSensorService()
	handler := handlers.NewHandler(influxDBService, alertService, sensorService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		api.POST("/lora/data", handler.ReceiveLoRaData)

		api.GET("/sensors", handler.GetSensors)
		api.GET("/sensors/:id", handler.GetSensor)
		api.GET("/buildings", handler.GetBuildings)

		api.GET("/data/acoustic", handler.GetAcousticData)
		api.GET("/data/moisture", handler.GetMoistureData)

		api.GET("/alerts", handler.GetAlerts)

		api.GET("/risk-zones", handler.GetRiskZones)

		api.GET("/predict/termite", handler.PredictTermiteActivity)

		api.POST("/simulate/fumigation", handler.SimulateFumigation)

		api.GET("/analysis/wavelet", handler.GetWaveletAnalysis)
	}

	frontendPath := getFrontendPath()
	r.StaticFile("/", filepath.Join(frontendPath, "index.html"))
	r.StaticFile("/app.js", filepath.Join(frontendPath, "app.js"))
	r.Static("/static", frontendPath)

	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, "index.html"))
	})

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("========================================")
	log.Printf("古代木结构建筑虫蛀监测系统")
	log.Printf("服务器启动在 http://localhost%s", addr)
	log.Printf("前端页面: http://localhost%s/", addr)
	log.Printf("API文档:  http://localhost%s/api/v1/", addr)
	log.Printf("========================================")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getConfigPath() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	
	configPath := filepath.Join(basepath, "..", "..", "config", "config.yaml")
	return configPath
}

func getFrontendPath() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	
	frontendPath := filepath.Join(basepath, "..", "..", "..", "frontend", "public")
	absPath, err := filepath.Abs(frontendPath)
	if err == nil {
		return absPath
	}
	return frontendPath
}
