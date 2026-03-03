package db_utils

import (
	"context"
	"encoding/json"
	"fmt"
	"letsgo/config"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB(cfg *config.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoDB.URI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Подключен к MongoDB!")
	return client, nil
}

func InitMongoDB(client *mongo.Client, cfg *config.Config) error {
	coll := client.Database(cfg.MongoDB.Database).Collection("docs")

	count, _ := coll.CountDocuments(context.Background(), bson.M{})
	if count > 0 {
		log.Println("Данные уже есть в БД")
		return nil
	}

	log.Println("Загрузка данных...")
	data, err := os.ReadFile("docs/test.documents.json")
	if err != nil {
		log.Printf("Ошибка чтения данных: %v", err)
		return err
	}

	var docs []any
	json.Unmarshal(data, &docs)

	_, err = coll.InsertMany(context.Background(), docs)
	if err != nil {
		log.Printf("Ошибка после вставки данных: %v", err)
		return err
	}
	log.Println("Данные загружены")
	return nil
}

func FindInMongo(collection *mongo.Collection, fiscalDriveNumber string, dateFrom, dateTo time.Time) ([]bson.M, error) {
	fmt.Printf("Поиск по датам: %v, %v\n", dateFrom, dateTo)

	filter := bson.M{
		"doc.fiscalDriveNumber": fiscalDriveNumber,
		"doc.dateTime": bson.M{
			"$gte": dateFrom,
			"$lte": dateTo,
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
