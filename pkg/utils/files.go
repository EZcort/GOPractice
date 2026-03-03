package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	db "letsgo/pkg/datasources"
	"letsgo/store/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ReadXLSXandMatch(filePath string, client *mongo.Client, dbName string, dateFrom, dateTo time.Time) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("Ошибка открытия XLSX: %v\n", err)
		return err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	chan_rows, err := f.Rows(sheetName)
	if err != nil {
		log.Printf("Ошибка чтения строк XLSX: %v\n", err)
		return err
	}
	defer chan_rows.Close()

	collection := client.Database(dbName).Collection("docs")
	for chan_rows.Next() {
		fiscalDriveNumber, err := chan_rows.Columns()
		if err != nil {
			continue
		}

		if !isValidFiscalDriveNumber(fiscalDriveNumber[0]) {
			log.Printf("Неверный формат fiscalDriveNumber: %s (должен быть 16 цифр)", fiscalDriveNumber)
			continue
		}

		docs, err := db.FindInMongo(collection, fiscalDriveNumber[0], dateFrom, dateTo)
		if err != nil {
			log.Printf("Ошибка поиска в MongoDB: %v", err)
			continue
		}

		for _, doc := range docs {
			err := SaveToCSVandJSON(doc)
			if err != nil {
				log.Printf("Ошибка сохранения в CSV или JSON: %v\n", err)
				continue
			}
		}
	}
	log.Printf("Создание CSV завершено")
	return nil
}

func SaveToCSVandJSON(doc bson.M) error {
	err := os.MkdirAll("TMP", 0755)
	if err != nil {
		log.Printf("Ошибка создания директории: %v\n", err)
		return err
	}

	fileUuid4 := uuid.New().String()

	csvFileName := "TMP/" + fileUuid4 + ".csv"
	csvFile, err := os.Create(csvFileName)
	if err != nil {
		log.Printf("Ошибка создания CSV: %v\n", err)
		return err
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	docData, _ := doc["doc"].(bson.M)
	fiscalDriveNumber := convertToString(docData["fiscalDriveNumber"])
	fiscalDocumentNumber := convertToString(docData["fiscalDocumentNumber"])
	itemsName := checkItems(docData)

	csvRecord := []string{
		fiscalDriveNumber,
		fiscalDocumentNumber,
		itemsName,
	}

	err = writer.Write(csvRecord)
	if err != nil {
		log.Printf("Ошибка записи в CSV: %v\n", err)
		return err
	}

	jsonFileName := "TMP/" + fileUuid4 + ".json"
	jsonRecord := models.JSONRecord{
		UUID:              fileUuid4,
		FiscalDriveNumber: fiscalDriveNumber,
		Status:            "processed",
	}

	jsonData, err := json.MarshalIndent(jsonRecord, "", "  ")
	if err != nil {
		log.Printf("Ошибка маршалинга JSON: %v\n", err)
		return err
	}

	err = os.WriteFile(jsonFileName, jsonData, 0644)
	if err != nil {
		log.Printf("Ошибка записи JSON файла: %v\n", err)
		return err
	}

	fmt.Printf("Документ %s сохранён\n", csvFileName)
	fmt.Printf("Документ %s сохранён\n", jsonFileName)
	fmt.Printf("---Содержимое CSV файла---\n")
	fmt.Printf("FiscalDriveNumber: %s\n", fiscalDriveNumber)
	fmt.Printf("FiscalDocumentNumber: %s\n", fiscalDocumentNumber)
	fmt.Printf("Items: %s\n", itemsName)
	fmt.Printf("---Содержимое JSON файла---\n")
	fmt.Printf("UUID: %s\n", jsonRecord.UUID)
	fmt.Printf("FiscalDriveNumber: %s\n", jsonRecord.FiscalDriveNumber)
	fmt.Printf("Status: %s\n", jsonRecord.Status)

	return nil
}

func checkItems(doc bson.M) string {
	items, ok := doc["items"]
	if !ok {
		return ""
	}

	itemsSlice, ok := items.(bson.A)
	if !ok || len(itemsSlice) == 0 {
		return ""
	}

	var itemNames strings.Builder
	for i, item := range itemsSlice {
		itemMap, ok := item.(bson.M)
		if !ok {
			continue
		}

		name, ok := itemMap["name"].(string)
		if !ok {
			continue
		}

		if i > 0 {
			itemNames.WriteString(", ")
		}
		itemNames.WriteString(name)
	}

	return itemNames.String()
}
