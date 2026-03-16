package dbh

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type NullBoolTester struct {
	ID int64    `gorm:"primaryKey"`
	B1 NullBool `json:"b1"`
	B2 NullBool `json:"b2"`
	B3 NullBool `json:"b3"`
	B4 NullBool `json:"b4,omitempty"`
}

func TestNullBoolGorm(t *testing.T) {
	db := OpenSqliteTestDB(t)
	require.NoError(t, db.Exec("CREATE TABLE null_bool_tester (id INTEGER PRIMARY KEY, b1 BOOLEAN, b2 BOOLEAN, b3 BOOLEAN, b4 BOOLEAN)").Error)

	row := NullBoolTester{
		ID: 1,
		B1: NullBool{Bool: true, Valid: true},
		B2: NullBool{Bool: false, Valid: true},
		B3: NullBool{},
		// B4 is the zero-value NullBool (invalid), so it should be stored as NULL.
	}
	require.NoError(t, db.Create(&row).Error)

	var b1, b2, b3, b4 sql.NullBool
	require.NoError(t, db.Raw("SELECT b1, b2, b3, b4 FROM null_bool_tester WHERE id = 1").Row().Scan(&b1, &b2, &b3, &b4))
	require.True(t, b1.Valid)
	require.True(t, b1.Bool)
	require.True(t, b2.Valid)
	require.False(t, b2.Bool)
	require.False(t, b3.Valid)
	require.False(t, b4.Valid)

	read := NullBoolTester{}
	require.NoError(t, db.First(&read, 1).Error)
	require.Equal(t, row, read)

	// Update state transition: false/valid -> true/valid should persist as true.
	read.B2 = NullBool{Bool: true, Valid: true}
	require.NoError(t, db.Save(&read).Error)
	require.NoError(t, db.Raw("SELECT b2 FROM null_bool_tester WHERE id = 1").Row().Scan(&b1))
	require.True(t, b1.Valid)
	require.True(t, b1.Bool)

	// Update state transition: valid -> invalid should persist NULL.
	read.B2 = NullBool{}
	require.NoError(t, db.Save(&read).Error)
	require.NoError(t, db.Raw("SELECT b2 FROM null_bool_tester WHERE id = 1").Row().Scan(&b1))
	require.False(t, b1.Valid)
}

func TestNullBoolJSON(t *testing.T) {
	tru := NullBool{Bool: true, Valid: true}
	fal := NullBool{Bool: false, Valid: true}
	invalid := NullBool{}

	j, err := json.Marshal(&tru)
	require.NoError(t, err)
	require.Equal(t, `true`, string(j))

	j, err = json.Marshal(&fal)
	require.NoError(t, err)
	require.Equal(t, `false`, string(j))

	j, err = json.Marshal(&invalid)
	require.NoError(t, err)
	require.Equal(t, `null`, string(j))

	var unmarshaled NullBool
	require.NoError(t, json.Unmarshal([]byte(`true`), &unmarshaled))
	require.Equal(t, tru, unmarshaled)

	require.NoError(t, json.Unmarshal([]byte(`false`), &unmarshaled))
	require.Equal(t, fal, unmarshaled)

	require.NoError(t, json.Unmarshal([]byte(`null`), &unmarshaled))
	require.False(t, unmarshaled.Valid)
	require.False(t, unmarshaled.Bool)

	require.Error(t, json.Unmarshal([]byte(`1`), &unmarshaled))
}

func TestNullBoolIsZero(t *testing.T) {
	require.True(t, NullBool{}.IsZero())
	require.True(t, NullBool{Bool: false, Valid: false}.IsZero())
	require.False(t, NullBool{Bool: true, Valid: true}.IsZero())
	require.False(t, NullBool{Bool: false, Valid: true}.IsZero())
}

func TestNullBoolJSONOmitZero(t *testing.T) {
	type payload struct {
		ID int64    `json:"id"`
		B  NullBool `json:"b,omitzero"`
	}

	var p payload
	j, err := json.Marshal(&p)
	require.NoError(t, err)
	require.Equal(t, `{"id":0}`, string(j))

	p = payload{
		ID: 1,
		B:  NullBool{Bool: false, Valid: true},
	}
	j, err = json.Marshal(&p)
	require.NoError(t, err)
	require.Equal(t, `{"id":1,"b":false}`, string(j))
}
