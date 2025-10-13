package dbh

import (
	"database/sql/driver"
	"fmt"
)

// SoftString makes NULL and an empty string equivalent.
// It's "soft" because there is no "hard line" between NULL and "".
// When empty, it marshals to JSON as "".
// When empty, it marshals to SQL as NULL.
// The idea is that in Go and Typescript, it's nicer to work with "" than with null,
// but in the database, NULL is imperative, for things like unique keys.
type SoftString string

// Scan implements the [Scanner] interface.
func (s *SoftString) Scan(value any) error {
	if value == nil {
		*s = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = SoftString(v)
	case []byte:
		*s = SoftString(string(v))
	default:
		panic("unexpected type for MuteString")
	}
	return nil
}

// Value implements the [driver.Valuer] interface.
func (s SoftString) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	return string(s), nil
}

// MarshalJSON implements the [json.Marshaler] interface.
func (s SoftString) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s + `"`), nil
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (s *SoftString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid JSON string for MuteString: %s", string(data))
	}
	*s = SoftString(data[1 : len(data)-1])
	return nil
}

// IsZero returns true if this MuteString is empty.
func (s SoftString) IsZero() bool {
	return s == ""
}
