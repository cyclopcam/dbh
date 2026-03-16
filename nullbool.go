package dbh

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

// NullBool mirrors [sql.NullBool], but also supports JSON marshalling and unmarshalling.
type NullBool struct {
	Bool  bool
	Valid bool
}

func (n NullBool) IsZero() bool {
	return !n.Valid
}

// Scan implements the [Scanner] interface.
func (n *NullBool) Scan(value any) error {
	var v sql.NullBool
	if err := v.Scan(value); err != nil {
		return err
	}
	*n = NullBool(v)
	return nil
}

// Value implements the [driver.Valuer] interface.
func (n NullBool) Value() (driver.Value, error) {
	return sql.NullBool{
		Bool:  n.Bool,
		Valid: n.Valid,
	}.Value()
}

// MarshalJSON implements the [json.Marshaler] interface.
func (n NullBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Bool)
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (n *NullBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = NullBool{}
		return nil
	}
	var v bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*n = NullBool{
		Bool:  v,
		Valid: true,
	}
	return nil
}
