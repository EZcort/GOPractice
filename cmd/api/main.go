package main

import (
	"context"
	"fmt"
	"letsgo/config"
	"log"
	"net/http"

	handlers "letsgo/internal/app/delivery/http"
	db "letsgo/pkg/datasources"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.LoadConfig()
	client, err := db.ConnectMongoDB(cfg)
	if err != nil {
		log.Fatal("Ошибка подключения к MongoDB:", err)
	}
	defer client.Disconnect(context.Background())

	// db.InitMongoDB(client, cfg)

	go fmt.Printf("Ручки:\nhttp://localhost:8080/api/activate\nhttp://localhost:8080/api/process/\nhttp://localhost:8080/api/uuids\nhttp://localhost:8080/api/CSV/\nhttp://localhost:8080/api/metrics\n")

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/activate", handlers.LoadData(client, cfg.MongoDB.Database))
	apiMux.HandleFunc("/process/", handlers.GetUuid4)
	apiMux.HandleFunc("/uuids", handlers.GetUuids)
	apiMux.HandleFunc("/CSV/", handlers.GetCSV)
	apiMux.Handle("/metrics", promhttp.Handler())

	http.Handle("/api/", http.StripPrefix("/api", apiMux))

	log.Fatal(http.ListenAndServe(":"+cfg.Server.Port, nil))
}
