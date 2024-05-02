package dbhelpers

import (
	"circonomy-server/database"

	"github.com/google/uuid"
)

func AddImage(path, imgType string) (uuid.UUID, error) {
	// language = SQL
	SQL := `INSERT INTO uploads(path, type) VALUES ($1,$2) RETURNING id
`
	var id uuid.UUID
	err := database.CirconomyDB.Get(&id, SQL, path, imgType)
	return id, err
}
