package table

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	letters = "ABCDEFGHIJKLMONQRSTUVWXYZ"
)

type cellrange struct {
	ISerializable
	StartRow    int
	StopRow     int
	StartColumn int
	StopColumn  int
	TableId     string
}

func (cr *cellrange) ToBytes() []byte {
	res, err := json.Marshal(cr)
	if err != nil {
		return nil
	}

	return res
}

func MakeRange(xrange string) *cellrange {
	cr := new(cellrange)
	rangeParts := strings.Split(xrange, "!")

	if len(rangeParts) == 2 {
		cr.TableId = rangeParts[0]
		rangeParts = strings.Split(rangeParts[1], ":")
	} else {
		rangeParts = strings.Split(rangeParts[0], ":")
	}
	startParts := getStringPartsFromAlphaNumeric(rangeParts[0])
	startRow, startColumn := parseAlphaNumericParts(startParts)
	stopParts := getStringPartsFromAlphaNumeric(rangeParts[1])
	stopRow, stopColumn := parseAlphaNumericParts(stopParts)
	cr.StartRow = startRow
	cr.StartColumn = startColumn
	cr.StopRow = stopRow + 1
	cr.StopColumn = stopColumn + 1
	return cr
}

func MakeRangeFromBytes(bytes []byte) *cellrange {
	var m cellrange
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil
	}
	return &m
}

func MakeRangeFromMap(rng map[string]interface {}) *cellrange {
	cr := new(cellrange)
	cr.TableId = rng["table_id"].(string)
	cr.StartRow = int(rng["start_row"].(float64))
	cr.StopRow = int(rng["stop_row"].(float64))
	cr.StartColumn = int(rng["start_column"].(float64))
	cr.StopColumn = int(rng["stop_column"].(float64))
	return cr
}

func parseFormula(value string) []string {
	funcCall := value[1:len(value)]
	funcParts := strings.Split(funcCall, "(")
	funcParts[1] = funcParts[1][:len(funcParts[1])-1] // remove the ')'
	return funcParts
}

func getNumberFromAlpha(alpha string) int {
	sum := 0
	upperAlpha := strings.ToUpper(alpha)
	la := utf8.RuneCountInString(upperAlpha)
	for i := 0; i < la; i++ {
		index := strings.Index(letters, string([]rune(upperAlpha)[i])) + 1
		sum += index * int(math.Pow(26, float64((la-(i+1)))))
	}
	return sum - 1
}

func parseAlphaNumericParts(parts []string) (int, int) {
	row, column := 0, -1
	if len(parts) == 2 {
		column = getNumberFromAlpha(parts[0])
		row64, _ := strconv.ParseInt(parts[1], 0, 32)
		row = int(row64)
	} else {
		row64, err := strconv.ParseInt(parts[0], 0, 32)
		if err != nil {
			column = getNumberFromAlpha(parts[0])
		} else {
			row = int(row64)
		}
	}
	return row - 1, column
}

func getStringPartsFromAlphaNumeric(alpha string) []string {
	re := regexp.MustCompile("[a-zA-Z]+|\\d+")
	return re.FindAllString(alpha, -1)
}
