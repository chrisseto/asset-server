package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"upper.io/db.v3"
)

var VALID_FILTERS = map[string]string{
	">":   ">",
	">=":  ">=",
	"<":   "<",
	"<=":  "<=",
	"=":   "=",
	"!=":  "!=",
	"":    "=",
	"eq":  "=",
	"ne":  "!=",
	"gt":  ">",
	"lt":  "<",
	"gte": ">=",
	"lte": "<=",
}

func columnName(field reflect.StructField) (string, error) {
	tag, ok := field.Tag.Lookup("db")
	if !ok {
		return "", errors.New("Column does not exist")
	}
	return strings.Split(tag, ",")[0], nil
}

func parseCondition(value string) (string, error) {
	filter, ok := VALID_FILTERS[value]

	if !ok {
		return "", errors.New("Invalid filter")
	}

	return filter, nil
}

func parseValue(field reflect.StructField, value string) (interface{}, error) {
	switch field.Type.Kind() {
	case reflect.String:
		return value, nil
	case reflect.Int:
		return strconv.ParseInt(value, 10, 0)
	case reflect.Bool:
		return strconv.ParseBool(value)
	case reflect.ValueOf(time.Time{}).Kind():
		return time.Parse(time.RFC3339, value)
	default:
		return nil, errors.New("Could not parse value")
	}

}

func ParseFilters(params map[string][]string) db.Cond {
	cond := db.Cond{}
	re := regexp.MustCompile(`^filter\[(\w+)\](?:\[(\S+)\]|())$`)
	typ := reflect.TypeOf(Asset{})

	for k, vs := range params {
		for _, v := range vs {
			matches := re.FindStringSubmatch(k)

			if len(matches) < 1 {
				continue
			}

			field, ok := typ.FieldByName(matches[1])

			if !ok {
				continue
			}

			filter, err := parseCondition(matches[2])
			if err != nil {
				continue
			}

			column, err := columnName(field)
			if err != nil {
				continue
			}

			value, err := parseValue(field, v)
			if err != nil {
				continue
			}

			cond[fmt.Sprintf("%s %s", column, filter)] = value
		}
	}

	return cond
}
