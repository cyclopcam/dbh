package dbh

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDBNotExist(t *testing.T) {
	require.False(t, DBNotExistRegex.MatchString(`does not exist`))
	require.True(t, DBNotExistRegex.MatchString(`database "foobar" does not exist`))
	require.False(t, DBNotExistRegex.MatchString(`table "foobar" does not exist`))
	require.False(t, DBNotExistRegex.MatchString(`"foobar" does not exist`))
}

func TestErrorIdentification(t *testing.T) {
	require.True(t, IsKeyViolation(errors.New("pq: duplicate key value violates unique constraint \"users_pkey\"")))
	require.True(t, IsKeyViolationOnIndex(errors.New("pq: duplicate key value violates unique constraint \"users_pkey\""), "users_pkey"))
	require.True(t, IsKeyViolation(errors.New("UNIQUE constraint failed: user.username_normalized")))
	// grrr.. sqlite mentions field name, not index name, so we can't actually tell which index it is
	require.True(t, IsKeyViolationOnIndex(errors.New("UNIQUE constraint failed: user.username_normalized"), "username_normalized"))
}
