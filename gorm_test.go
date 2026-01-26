package dbh

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

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

type TestNull3 struct {
	ID   int64      `gorm:"primaryKey"`
	Str  SoftString `json:"str"`
	Str2 SoftString `json:"str2"`
	Str3 SoftString `json:"str3,omitempty"`
	Int  int64      `json:"int"`
}

func TestGormNulls(t *testing.T) {
	db := OpenSqliteTestDB(t)
	require.NoError(t, db.Exec("CREATE TABLE test_null1 (id INTEGER PRIMARY KEY, str TEXT, int INT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE test_null2 (id INTEGER PRIMARY KEY, str TEXT, int INT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE test_null3 (id INTEGER PRIMARY KEY, str TEXT, str2 TEXT, str3 TEXT, int INT)").Error)

	v1 := TestNull1{}
	v2 := TestNull2{}
	v3 := TestNull3{
		Str:  "hello",
		Str2: "", // becomes null
		// Str3 is null	by default
		Int: 123,
	}

	require.NoError(t, db.Create(&v1).Error)
	require.NoError(t, db.Create(&v2).Error)
	require.NoError(t, db.Create(&v3).Error)

	// TestNull1, using gorm:"default:null"
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

	// Validate our own NullString
	var s1, s2, s3 sql.NullString
	require.NoError(t, db.Raw("SELECT str, str2, str3 FROM test_null3 WHERE id = 1").Row().Scan(&s1, &s2, &s3))
	require.Equal(t, true, s1.Valid)
	require.Equal(t, "hello", s1.String)
	require.Equal(t, false, s2.Valid)
	require.Equal(t, "", s2.String)
	require.Equal(t, false, s3.Valid)
}

func TestNullToJSON(t *testing.T) {
	t1 := TestNull1{
		Str: "",
		Int: 123,
	}
	j, err := json.Marshal(&t1)
	require.NoError(t, err)
	t.Logf("json: %v", string(j))

	t2 := TestNull2{
		ID: 1,
		Str: sql.NullString{
			String: "hello",
			Valid:  true,
		},
		Int: sql.NullInt64{
			Int64: 123,
			Valid: true,
		},
	}
	j, err = json.Marshal(&t2)
	require.NoError(t, err)
	t.Logf("json: %v", string(j))
	// hmm.. yeah NullString doesn't go well into JSON

	t3 := TestNull3{
		ID:   1,
		Str:  "hello",
		Str2: "",
		// Str3 is omitted from JSON via omitempty (or omitzero)
		Int: 123,
	}
	j, err = json.Marshal(&t3)
	require.NoError(t, err)
	t.Logf("json: %v", string(j))

	t3s := TestNull3{}
	err = json.Unmarshal(j, &t3s)
	require.NoError(t, err)
	require.Equal(t, t3, t3s)
}

type TestUpdates struct {
	ID        int64 `gorm:"primaryKey"`
	Value     string
	CreatedAt IntTime `gorm:"autoCreateTime:milli"`
	UpdatedAt IntTime `gorm:"autoUpdateTime:milli"`
}

func TestGormUpdatedAndCreateAt(t *testing.T) {
	db := OpenSqliteTestDB(t)
	require.NoError(t, db.Exec("CREATE TABLE test_updates (id INTEGER PRIMARY KEY, value TEXT, created_at INT, updated_at INT)").Error)
	v := TestUpdates{
		Value: "abc",
	}
	now := time.Now()
	t1 := now.Add(-time.Millisecond)
	t2 := now.Add(time.Millisecond)
	require.NoError(t, db.Create(&v).Error)
	createdAt1 := v.CreatedAt
	updatedAt1 := v.UpdatedAt
	require.WithinRange(t, createdAt1.Get(), t1, t2)
	require.WithinRange(t, updatedAt1.Get(), t1, t2)
	time.Sleep(10 * time.Millisecond)

	// update - make sure UpdatedAt changes with a call to Update(). I would expect this to happen,
	// but have never really looked into it, which is why I'm writing this test.
	require.Equal(t, v.UpdatedAt.Get().UnixMilli(), v.CreatedAt.Get().UnixMilli())
	require.NoError(t, db.Model(&v).Update("value", "123").Error)
	require.Greater(t, v.UpdatedAt.Get().UnixMilli(), v.CreatedAt.Get().UnixMilli())
}
