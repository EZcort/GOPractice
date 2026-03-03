package utils

import (
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

func parseDate(doc bson.M) (time.Time, error) {
	docField := doc["doc"].(bson.M)
	dateTimeField := docField["dateTime"].(bson.M)
	dateStr := dateTimeField["$date"].(string)
	return time.Parse(time.RFC3339, dateStr)
}

// "dateTime": time.Now().UTC() - обратно в объект
