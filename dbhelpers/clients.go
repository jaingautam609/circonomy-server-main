package dbhelpers

import (
	"circonomy-server/database"
	"circonomy-server/models"
)

func GetAllOrgClients() ([]models.Clients, error) {
	// language = SQL
	SQL := `SELECT u.id, u.name, u2.path, sum(pb.credits) as total_credits
       		FROM users u
       		LEFT JOIN projects_bought pb on u.id = pb.user_id
       		LEFT JOIN uploads u2 on u.upload_id = u2.id
           	WHERE u.account_type = $1 
           	AND u.archived_at IS NULL
			AND u2.archived_at IS NULL
			GROUP BY u.id, u.name, u2.path
`
	clientList := make([]models.Clients, 0)
	err := database.CirconomyDB.Select(&clientList, SQL, models.UserAccountTypeCorporate)
	return clientList, err
}

func GetAllIndividualClients() ([]models.Clients, error) {
	// language = SQL
	SQL := `SELECT u.id, u.name, uploads.path, sum(pb.credits) as total_credits
       		FROM users u
       		LEFT JOIN projects_bought pb on u.id = pb.user_id
       		LEFT JOIN uploads on u.upload_id = uploads.id
           	WHERE u.account_type=$1 
           	AND u.archived_at IS NULL
			AND uploads.archived_at IS NULL
			GROUP BY u.id, u.name, uploads.path
`
	clientList := make([]models.Clients, 0)
	err := database.CirconomyDB.Select(&clientList, SQL, models.UserAccountTypeIndividual)
	return clientList, err
}
