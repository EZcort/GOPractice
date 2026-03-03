package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	constants "letsgo/pkg/constants"
	db "letsgo/pkg/datasources"
	models "letsgo/store/models"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ReadXLSXandMatch(filePath string, client *mongo.Client, dbName string) error {
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
		row, err := chan_rows.Columns()
		if err != nil {
			continue
		}
		rowValue, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			log.Printf("Ошибка чтения преобразования строки XLSX: %v", err)
			continue
		}
		value := uint64(rowValue)
		strValue := strconv.FormatUint(value, 10)
		docs, err := db.FindInMongo(collection, strValue)
		if err != nil {
			continue
		}

		for _, doc := range docs {
			fileUuid4, err := saveToCSV(doc)
			if err != nil {
				log.Printf("Ошибка сохранения в CSV: %v\n", err)
			}

			docData, _ := doc["doc"].(bson.M)
			err = appendToJSONL(fileUuid4, convertToString(docData["fiscalDriveNumber"]))
			if err != nil {
				log.Printf("Ошибка добавления в jsonL: %v", err)
			}
		}

	}
	log.Printf("Создание CSV завершено")
	return nil
}

func saveToCSV(doc bson.M) (string, error) {
	err := os.MkdirAll("TMP", 0755)
	if err != nil {
		log.Printf("Ошибка создания директории: %v\n", err)
		return "", err
	}

	fileUuid4 := uuid.New().String()
	fileName := "TMP/" + fileUuid4 + ".csv"
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Ошибка создания CSV: %v\n", err)
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	docData, _ := doc["doc"].(bson.M)
	fiscalDriveNumber := convertToString(docData["fiscalDriveNumber"])
	fiscalDocumentNumber := convertToString(docData["fiscalDocumentNumber"])
	itemsName := checkItems(docData)

	record := []string{
		fiscalDriveNumber,
		fiscalDocumentNumber,
		itemsName,
	}

	err = writer.Write(record)
	if err != nil {
		log.Printf("Ошибка записи в CSV: %v\n", err)
		return "", err
	}

	fmt.Printf("Документ %s сохранён\n", fileName)
	fmt.Printf("---Содержимое CSV файла---\n")
	fmt.Printf("FiscalDriveNumber: %s\n", fiscalDriveNumber)
	fmt.Printf("FiscalDocumentNumber: %s\n", fiscalDocumentNumber)
	fmt.Printf("Items: %s\n", itemsName)
	return fileUuid4, nil
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

func appendToJSONL(uuid string, fiscalDriveNumber string) error {
	constants.JSONLMutex.Lock()
	defer constants.JSONLMutex.Unlock()

	record := models.JSONLRecord{
		UUID:              uuid,
		FiscalDriveNumber: fiscalDriveNumber,
	}

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(constants.JSONLFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(append(recordJSON, '\n'))
	if err != nil {
		return err
	}

	fmt.Printf("Запись %s добавлена в JSONL\n", uuid)
	fmt.Println("-----------------------------------")
	return nil
}
