package util

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDString converts a pgtype.UUID to its canonical string form.
// Returns "" if the UUID is not valid.
func UUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	u, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return ""
	}
	return u.String()
}

// TsTime converts a pgtype.Timestamptz to a time.Time.
// Returns zero time if not valid.
func TsTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

// ParseUUID parses a raw UUID string into pgtype.UUID.
func ParseUUID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(raw)
	return id, err
}
