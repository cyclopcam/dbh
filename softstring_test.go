package dbh

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type SoftStringTester struct {
	ID   int64      `gorm:"primaryKey"`
	Str  SoftString `json:"str"`
	Str2 SoftString `json:"str2"`
	Str3 SoftString `json:"str3,omitempty"`
}

func TestSoftStringGorm(t *testing.T) {
	db := OpenSqliteTestDB(t)
	require.NoError(t, db.Exec("CREATE TABLE soft_string_tester (id INTEGER PRIMARY KEY, str TEXT, str2 TEXT, str3 TEXT)").Error)

	row := SoftStringTester{
		ID:   1,
		Str:  "hello",
		Str2: "", // becomes NULL in SQL
		// Str3 remains zero value and should also become NULL in SQL
	}
	require.NoError(t, db.Create(&row).Error)

	var str1, str2, str3 sql.NullString
	require.NoError(t, db.Raw("SELECT str, str2, str3 FROM soft_string_tester WHERE id = 1").Row().Scan(&str1, &str2, &str3))
	require.True(t, str1.Valid)
	require.Equal(t, "hello", str1.String)
	require.False(t, str2.Valid)
	require.Equal(t, "", str2.String)
	require.False(t, str3.Valid)

	read := SoftStringTester{}
	require.NoError(t, db.First(&read, 1).Error)
	require.Equal(t, row.Str, read.Str)
	require.Equal(t, row.Str2, read.Str2)
	require.Equal(t, row.Str3, read.Str3)

	// Update state transition: non-empty -> empty should serialize back to NULL on save.
	read.Str2 = "world"
	require.NoError(t, db.Save(&read).Error)
	require.NoError(t, db.Raw("SELECT str2 FROM soft_string_tester WHERE id = 1").Row().Scan(&str1))
	require.True(t, str1.Valid)
	require.Equal(t, "world", str1.String)

	read.Str2 = ""
	require.NoError(t, db.Save(&read).Error)
	require.NoError(t, db.Raw("SELECT str2 FROM soft_string_tester WHERE id = 1").Row().Scan(&str1))
	require.False(t, str1.Valid)

	// Ensure database NULL is decoded to empty string.
	require.NoError(t, db.Exec("INSERT INTO soft_string_tester (id, str, str2, str3) VALUES (?, ?, ?, ?)", 2, nil, "x", nil).Error)
	read2 := SoftStringTester{}
	require.NoError(t, db.First(&read2, 2).Error)
	require.Equal(t, SoftString(""), read2.Str)
	require.Equal(t, SoftString("x"), read2.Str2)
	require.Equal(t, SoftString(""), read2.Str3)
}

func TestSoftStringJSON(t *testing.T) {
	empty := SoftString("")
	nonEmpty := SoftString("hello")

	j, err := json.Marshal(&empty)
	require.NoError(t, err)
	require.Equal(t, `""`, string(j))

	j, err = json.Marshal(&nonEmpty)
	require.NoError(t, err)
	require.Equal(t, `"hello"`, string(j))

	var unmarshaled SoftString
	err = json.Unmarshal([]byte(`"world"`), &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, SoftString("world"), unmarshaled)

	err = json.Unmarshal([]byte(`null`), &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, SoftString(""), unmarshaled)

	err = json.Unmarshal([]byte(`true`), &unmarshaled)
	require.Error(t, err)
}
