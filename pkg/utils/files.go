package files

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
		docs, err := findInMongo(collection, strValue)
		if err != nil {
			continue
		}

		for _, doc := range docs {
			err = saveToCSV(strValue, doc)
			if err != nil {
				log.Printf("Ошибка сохранения в CSV: %v\n", err)
			}
		}

	}
	log.Printf("Создание CSV завершено")
	return nil
}

func findInMongo(collection *mongo.Collection, fdNumber string) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pattern := "^" + fdNumber[:len(fdNumber)-2]
	filter := bson.M{
		"$and": []bson.M{
			{"doc.fiscalDriveNumber": bson.M{
				"$regex": pattern}},
			// {"doc.DateTime": дата},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	err = cursor.All(ctx, &docs)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func parseDate(doc bson.M) (time.Time, error) {
	docField := doc["doc"].(bson.M)
	dateTimeField := docField["dateTime"].(bson.M)
	dateStr := dateTimeField["$date"].(string)
	return time.Parse(time.RFC3339, dateStr)
}

// "dateTime": time.Now().UTC() - обратно в объект

func saveToCSV(fdNumber string, doc bson.M) error {
	workDir := "TMP/" + fdNumber
	err := os.MkdirAll(workDir, 0755)
	if err != nil {
		log.Printf("Ошибка создания директории: %v\n", err)
		return err
	}

	fileUuid4 := uuid.New().String()
	fileName := workDir + "/" + fileUuid4 + ".csv"
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Ошибка создания CSV: %v\n", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	docData, ok := doc["doc"].(bson.M)
	if !ok {
		return fmt.Errorf("Ошибка извлечения данных из doc\n")
	}

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
		return err
	}

	fmt.Printf("Документ %s сохранён\n", fileName)
	fmt.Printf("---Содержимое CSV файла---\n")
	fmt.Printf("FiscalDriveNumber: %s\n", fiscalDriveNumber)
	fmt.Printf("FiscalDocumentNumber: %s\n", fiscalDocumentNumber)
	fmt.Printf("Items: %s\n", itemsName)
	fmt.Println("-----------------------------------")
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

func convertToString(value interface{}) string {
	switch v := value.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		if v == float64(int(v)) {
			return strconv.Itoa(int(v))
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}
