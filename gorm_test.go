package dbh

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

// This isn't really a test - just exploration.
// The thing I learned here is this:
// I initially thought that gorm:"default:null"` would give me the magical desired
// behaviour of treating zero/empty values as null, when Creating and Saving.
// It just so happens to do this on Create(), because the default:null metadata is invoked
// when creating a record, and the field is the default value. However, it does not
// have any effect when using Save(). So the bottom line is that we must either use
// pointers or sql.NullX if we want control over null-ness.

type TestNull1 struct {
	ID  int64  `gorm:"primaryKey"`
	Str string `gorm:"default:null"`
	Int int64  `gorm:"default:null"`
}

type TestNull2 struct {
	ID  int64 `gorm:"primaryKey"`
	Str sql.NullString
	Int sql.NullInt64
}

func TestGormNulls(t *testing.T) {
	db := OpenSqliteTestDB(t)
	require.NoError(t, db.Exec("CREATE TABLE test_null1 (id INTEGER PRIMARY KEY, str TEXT, int INT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE test_null2 (id INTEGER PRIMARY KEY, str TEXT, int INT)").Error)

	v1 := TestNull1{}
	v2 := TestNull2{}

	require.NoError(t, db.Create(&v1).Error)
	require.NoError(t, db.Create(&v2).Error)

	// At this point, everything is null (i.e. after Create, the default:null metadata works)
	nstr := sql.NullString{}
	nint := sql.NullInt64{}
	require.NoError(t, db.Raw("SELECT str, int FROM test_null1 WHERE id = 1").Row().Scan(&nstr, &nint))
	require.Equal(t, false, nstr.Valid)
	require.Equal(t, false, nint.Valid)

	require.NoError(t, db.Save(&v1).Error)
	require.NoError(t, db.Save(&v2).Error)

	// BUT!!!
	// After a Save(), the fields are no longer null
	require.NoError(t, db.Raw("SELECT str, int FROM test_null1 WHERE id = 1").Row().Scan(&nstr, &nint))
	require.Equal(t, true, nstr.Valid)
	require.Equal(t, true, nint.Valid)
}
