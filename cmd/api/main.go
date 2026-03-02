package main

import (
	"context"
	// "fmt"
	"letsgo/config"
	"log"
	"net/http"

	handlers "letsgo/internal/app/delivery/http"
	db "letsgo/pkg/datasources"
	files "letsgo/pkg/utils"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.LoadConfig()
	client, err := db.ConnectMongoDB(cfg)
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}
	defer client.Disconnect(context.Background())

	db.InitMongoDB(client, cfg)

	// fmt.Printf("Ручки:\nhttp://localhost:8080/get/Active/\nhttp://localhost:8080/get/process/\nhttp://localhost:8080/get/uuids\nhttp://localhost:8080/get/CSV/\nhttp://localhost:8080/get/metrics\n")

	getGroup := http.NewServeMux()
	getGroup.HandleFunc("/Activate", func(w http.ResponseWriter, r *http.Request) {
		go files.ReadXLSXandMatch("docs/book.xlsx", client, cfg.MongoDB.Database)
	})
	getGroup.HandleFunc("/process/", handlers.GetUuid4)
	getGroup.HandleFunc("/uuids", handlers.GetUuids)
	getGroup.HandleFunc("/CSV/", handlers.GetCSV)
	getGroup.Handle("/metrics", promhttp.Handler())
	getGroup.Handle("/get/", http.StripPrefix("/get", getGroup))
	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, getGroup))
}
