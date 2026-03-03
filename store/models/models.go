package models

type JSONLRecord struct {
	UUID              string `json:"uuid"`
	FiscalDriveNumber string `json:"fiscalDriveNumber"`
}

type CSVRecord struct {
	FiscalDriveNumber    string `json:"fiscalDriveNumber"`
	FiscalDocumentNumber string `json:"fiscalDocumentNumber"`
	Items                string `json:"items"`
}
