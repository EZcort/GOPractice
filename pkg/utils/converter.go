package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func convertToString(value any) string {
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

func isValidFiscalDriveNumber(fdn string) bool {
	if len(fdn) != 16 {
		return false
	}

	for _, char := range fdn {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"02.01.2006",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"02.01.2006 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}

	if timestamp, err := strconv.ParseInt(dateStr, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}

	return time.Time{}, fmt.Errorf("не удалось распарсить дату '%s': %v", dateStr, lastErr)
}

func ParseDateRange(dateFromStr, dateToStr string) (time.Time, time.Time, error) {
	if dateFromStr == "" {
		return time.Time{}, time.Time{}, errors.New("поле date_from обязательно")
	}
	if dateToStr == "" {
		return time.Time{}, time.Time{}, errors.New("поле date_to обязательно")
	}

	dateFrom, err := ParseDate(dateFromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("ошибка парсинга date_from: %v", err)
	}

	dateTo, err := ParseDate(dateToStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("ошибка парсинга date_to: %v", err)
	}

	dateFrom = time.Date(dateFrom.Year(), dateFrom.Month(), dateFrom.Day(), 0, 0, 0, 0, time.UTC)
	dateTo = time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, time.UTC)

	if dateFrom.After(dateTo) {
		return time.Time{}, time.Time{}, errors.New("date_from не может быть больше date_to")
	}

	return dateFrom, dateTo, nil
}
