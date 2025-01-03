package models

import (
	"database/sql/driver"
	"strings"
	"time"
)

type TimestampArray []time.Time

func (a *TimestampArray) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return a.scanString(string(v))
	case string:
		return a.scanString(v)
	case nil:
		*a = nil
		return nil
	default:
		return driver.ErrSkip
	}
}

func (a *TimestampArray) scanString(str string) error {
	if str == "{}" {
		*a = TimestampArray{}
		return nil
	}

	str = strings.Trim(str, "{}")
	parts := strings.Split(str, ",")

	times := make([]time.Time, len(parts))
	for i, part := range parts {
		part = strings.Trim(strings.TrimSpace(part), "\"")
		t, err := time.Parse("2006-01-02 15:04:05.999999", part)
		if err != nil {
			return err
		}
		times[i] = t
	}

	*a = times
	return nil
}

func (a TimestampArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	strs := make([]string, len(a))
	for i, t := range a {
		strs[i] = "\"" + t.Format("2006-01-02 15:04:05.999999") + "\""
	}

	return "{" + strings.Join(strs, ",") + "}", nil
}
