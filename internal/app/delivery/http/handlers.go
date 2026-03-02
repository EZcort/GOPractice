package http

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type CSVRecord struct {
	FiscalDriveNumber    string `json:"fiscalDriveNumber"`
	FiscalDocumentNumber string `json:"fiscalDocumentNumber"`
	Items                string `json:"items"`
}

func GetUuid4(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Не указан FiscalDriveNumber",
		})
		return
	}

	fiscalDriveNumber := pathParts[len(pathParts)-1]
	if fiscalDriveNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Не указан FiscalDriveNumber",
		})
		return
	}

	dirPath := "TMP/" + fiscalDriveNumber
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "FiscalDriveNumber не найден",
		})
		return
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	var uuids []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
			uuidStr := strings.TrimSuffix(file.Name(), ".csv")
			if _, err := uuid.Parse(uuidStr); err == nil {
				uuids = append(uuids, uuidStr)
			}
		}
	}

	json.NewEncoder(w).Encode(uuids)
}

func GetUuids(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if _, err := os.Stat("TMP"); os.IsNotExist(err) {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	dirs, err := os.ReadDir("TMP")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	var allFileNames []string
	for _, dir := range dirs {
		if dir.IsDir() {
			dirPath := "TMP/" + dir.Name()
			files, err := os.ReadDir(dirPath)
			if err != nil {
				continue
			}

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
					fileName := strings.TrimSuffix(file.Name(), ".csv")
					allFileNames = append(allFileNames, fileName)
				}
			}
		}
	}

	json.NewEncoder(w).Encode(allFileNames)
}

func GetCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Нет Uuid",
		})
		return
	}

	fileName := pathParts[len(pathParts)-1]
	if fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Нет Uuid",
		})
		return
	}

	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		fileName = fileName + ".csv"
	}

	var foundFile string
	err := filepath.Walk("TMP",

		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if !info.IsDir() && info.Name() == fileName {
				foundFile = path
				return filepath.SkipAll
			}
			return nil
		})

	if err != nil || foundFile == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Uuid не найден",
		})
		return
	}

	file, err := os.Open(foundFile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	var data []CSVRecord
	for _, record := range records {
		if len(record) >= 3 {
			data = append(data, CSVRecord{
				FiscalDriveNumber:    record[0],
				FiscalDocumentNumber: record[1],
				Items:                record[2],
			})
		}
	}

	json.NewEncoder(w).Encode(data)
}
