package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"letsgo/pkg/utils"
	"letsgo/store/models"

	"go.mongodb.org/mongo-driver/mongo"
)

func LoadData(client *mongo.Client, dbName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Ошибка при парсинге формы: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Файл не найден в запросе", http.StatusBadRequest)
			return
		}
		defer file.Close()

		dateFromStr := r.FormValue("date_from")
		dateToStr := r.FormValue("date_to")

		dateFrom, dateTo, err := utils.ParseDateRange(dateFromStr, dateToStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := os.MkdirAll("docs", os.ModePerm); err != nil {
			http.Error(w, "Ошибка при создании директории: "+err.Error(), http.StatusInternalServerError)
			return
		}

		timestamp := time.Now().Format("20060102_150405")
		filename := "docs/" + timestamp + "_" + header.Filename

		dst, err := os.Create(filename)
		if err != nil {
			http.Error(w, "Ошибка при создании файла: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Ошибка при сохранении файла: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := utils.ReadXLSXandMatch(filename, client, dbName, dateFrom, dateTo); err != nil {
			http.Error(w, "Ошибка при обработке файла: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := os.Remove(filename); err != nil {
			log.Printf("Не удалось удалить файл %s: %v", filename, err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Файл успешно загружен и обработан",
			"filename":  header.Filename,
			"size":      header.Size,
			"date_from": dateFrom.Format(time.RFC3339),
			"date_to":   dateTo.Format(time.RFC3339),
		})
	}
}

func GetUuid4(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		json.NewEncoder(w).Encode([]string{})
		return
	}
	fiscalDriveNumber := parts[len(parts)-1]

	entries, err := os.ReadDir("TMP")
	if err != nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	var result []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := "TMP/" + entry.Name()
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var record models.JSONRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		if record.FiscalDriveNumber == fiscalDriveNumber {
			result = append(result, record.UUID)
		}
	}

	json.NewEncoder(w).Encode(result)
}

func GetUuids(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	entries, err := os.ReadDir("TMP")
	if err != nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	uniq := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := "TMP" + entry.Name()
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var record models.JSONRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		if record.UUID != "" {
			uniq[record.UUID] = true
		}
	}

	uuids := make([]string, 0, len(uniq))
	for uuid := range uniq {
		uuids = append(uuids, uuid)
	}

	json.NewEncoder(w).Encode(uuids)
}

func GetCSV(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		http.Error(w, "UUID не указан", http.StatusBadRequest)
		return
	}

	uuid := parts[len(parts)-1]
	csvPath := "TMP/" + uuid + ".csv"

	file, err := os.Open(csvPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Файл не найден", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка при открытии файла", http.StatusInternalServerError)
		}
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Ошибка при получении информации о файле", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+uuid+".csv")
	http.ServeContent(w, r, uuid+".csv", fileInfo.ModTime(), file)
}
