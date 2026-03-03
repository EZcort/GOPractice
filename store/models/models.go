package models

type JSONRecord struct {
	UUID              string `json:"uuid"`
	FiscalDriveNumber string `json:"fiscalDriveNumber"`
	Status            string `json:"status"`
}

type CSVRecord struct {
	FiscalDriveNumber    string `json:"fiscalDriveNumber"`
	FiscalDocumentNumber string `json:"fiscalDocumentNumber"`
	Items                string `json:"items"`
}
