package http

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	constants "letsgo/pkg/constants"
	models "letsgo/store/models"
)

func GetUuid4(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		json.NewEncoder(w).Encode([]string{})
		return
	}
	fiscalDriveNumber := parts[len(parts)-1]

	data, err := os.ReadFile(constants.JSONLFilePath)
	if err != nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	lines := strings.Split(string(data), "\n")
	var result []string

	for _, line := range lines {
		if line == "" {
			continue
		}

		var record models.JSONLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
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

	data, err := os.ReadFile(constants.JSONLFilePath)
	if err != nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	lines := strings.Split(string(data), "\n")
	uniq := make(map[string]bool)

	for _, line := range lines {
		if line == "" {
			continue
		}

		var record models.JSONLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
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
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 {
		json.NewEncoder(w).Encode([]models.CSVRecord{})
		return
	}

	uuid := parts[len(parts)-1]
	csvPath := "TMP/" + uuid + ".csv"

	file, err := os.Open(csvPath)
	if err != nil {
		json.NewEncoder(w).Encode([]models.CSVRecord{})
		return
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		json.NewEncoder(w).Encode([]models.CSVRecord{})
		return
	}

	var result []models.CSVRecord
	for _, r := range records {
		if len(r) >= 3 {
			result = append(result, models.CSVRecord{
				FiscalDriveNumber:    r[0],
				FiscalDocumentNumber: r[1],
				Items:                r[2],
			})
		}
	}

	json.NewEncoder(w).Encode(result)
}
