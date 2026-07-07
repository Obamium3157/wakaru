package jmdict

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// TODO: написать тесты для этого файла

func GetStr(rawJSON []json.RawMessage, pos int) (string, error) {
	var s string
	if err := json.Unmarshal(rawJSON[pos], &s); err == nil {
		return s, nil
	}

	var num float64
	if err := json.Unmarshal(rawJSON[pos], &num); err == nil {
		return strconv.FormatFloat(num, 'f', -1, 64), nil
	}
	return "", fmt.Errorf("element %d is neither string nor number", pos)
}

// GetInt returns (number, isInt, err)
func GetInt(rawJSON []json.RawMessage, pos int) (int, bool, error) {
	var num float64
	if err := json.Unmarshal(rawJSON[pos], &num); err == nil {
		return int(num), true, nil
	}

	var s string
	if err := json.Unmarshal(rawJSON[pos], &s); err == nil {
		v, err := strconv.Atoi(s)
		if err != nil {
			return 0, false, nil
		}
		return v, true, nil
	}
	return 0, false, fmt.Errorf("element %d is neither string nor number", pos)
}
