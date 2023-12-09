package mongoquerier

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

func StructToM(source interface{}) (bson.M, error) {
	// Marshal source to JSON
	jsonSource, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	// Unmarshal JSON into a map
	var data map[string]interface{}
	if err := json.Unmarshal(jsonSource, &data); err != nil {
		return nil, err
	}

	result := bson.M{}
	structValues := reflect.ValueOf(source)
	structTypes := reflect.TypeOf(source)

	for i := 0; i < structTypes.NumField(); i++ {
		fieldType := structTypes.Field(i)
		fieldValue := structValues.Field(i)
		tagValue := fieldType.Tag.Get("json")
		jsonKey := strings.Split(tagValue, ",")[0]

		if jsonKey == "" {
			jsonKey = fieldType.Name
		}

		// Check if the field exists in the JSON data
		if _, ok := data[jsonKey]; ok {
			zeroValue := reflect.Zero(fieldType.Type)
			// Omit fields that default JSONMarshaler hasn't omitted
			if reflect.DeepEqual(zeroValue.Interface(), fieldValue.Interface()) {
				continue
			}

			if fieldType.Type.Kind() == reflect.Struct {
				valueMap, err := StructToM(fieldValue.Interface())
				if err != nil {
					return nil, err
				}

				for valueKey, valueValue := range valueMap {
					result[fmt.Sprintf("%s.%s", jsonKey, valueKey)] = valueValue
				}

				continue
			}

			result[jsonKey] = fieldValue.Interface()
		}

	}

	return result, nil
}

func CastStruct[S any, D any](source S) (destination D, err error) {
	// Convert struct to JSON string
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		return
	}

	// Unmarshal JSON into new struct
	if err = json.Unmarshal(sourceJSON, &destination); err != nil {
		return
	}

	return
}

func CastInto[S any, D any](source S, destination D) error {
	// Convert struct to JSON string
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		return err
	}

	// Unmarshal JSON into new struct
	err = json.Unmarshal(sourceJSON, destination)
	return err
}
