package dbhelpers

import (
	"circonomy-server/database"
	"circonomy-server/models"
	"github.com/volatiletech/null"

	"github.com/google/uuid"
)

func GetProfileByID(id uuid.UUID) (models.UserProfile, error) {
	// language = SQL
	SQL := `SELECT u.name ,u.email ,u.number, u.country_code ,u.address,u.account_type, u2.path ,u.size, u.upload_id, sum(pb.credits) as total_credits
			FROM users u 
			LEFT JOIN uploads u2 on u.upload_id = u2.id
			LEFT JOIN projects_bought pb on u.id = pb.user_id
			WHERE u.id=$1
			AND u.archived_at IS NULL 
			AND u2.archived_at IS NULL
            GROUP BY u.name, u.email, u.number, u.country_code, u.address, u.account_type, u2.path, u.size, u.upload_id
`
	var user models.UserProfile
	err := database.CirconomyDB.Get(&user, SQL, id)
	return user, err
}

func GetPeopleInOrg(organizationID uuid.UUID) ([]models.OrgPeople, error) {
	// language = SQL
	SQL := `
			SELECT u.id, u.name, u.address, uploads.path, organization.name as organization_name
			FROM users u
					 JOIN users_organisations uo on u.id = uo.user_id
					 LEFT JOIN uploads on u.upload_id = uploads.id
					 JOIN organization on uo.organization_id = organization.id
			WHERE uo.organization_id = $1
			  AND u.archived_at IS NULL
			  AND uploads.archived_at IS NULL
`
	users := make([]models.OrgPeople, 0)
	err := database.CirconomyDB.Select(&users, SQL, organizationID)
	return users, err
}

func GetPersonCreditsHistory(id uuid.UUID) ([]models.CreditsHistory, error) {
	// language = SQL
	SQL := `SELECT p.id, p.name, pb.credits, pb.bought_at_rate, pb.bought_at_cost, pb.created_at, u.path
			FROM projects_bought pb 
			JOIN projects p on pb.p_id = p.id
        	LEFT JOIN uploads u on p.upload_id = u.id
			WHERE pb.user_id=$1
  			AND pb.archived_at IS NULL
  			AND u.archived_at IS NULL
			AND p.archived_at IS NULL 
`
	history := make([]models.CreditsHistory, 0)
	err := database.CirconomyDB.Select(&history, SQL, id)
	return history, err
}

func GetOrgCreditsHistory(id uuid.UUID) ([]models.CreditsHistory, error) {
	// language = SQL
	SQL := `SELECT p.id, p.name, pb.credits, pb.bought_at_rate, pb.bought_at_cost, pb.created_at, u.path, us.name as bought_by_name
			FROM projects_bought pb 
			JOIN projects p on pb.p_id = p.id
        	JOIN users us on us.id = pb.bought_by
			LEFT JOIN uploads u on p.upload_id = u.id
			WHERE pb.user_id=$1
  			AND pb.archived_at IS NULL
  			AND u.archived_at IS NULL
			AND p.archived_at IS NULL 
`
	history := make([]models.CreditsHistory, 0)
	err := database.CirconomyDB.Select(&history, SQL, id)
	return history, err
}

func UpdateOrgProfileByID(id uuid.UUID, info *models.EditOrgProfile) error {
	// language = SQL
	SQL := `UPDATE users SET name=$2, address=$3, upload_id=$4::UUID, size=$5 WHERE id=$1`

	_, err := database.CirconomyDB.Exec(SQL, id, info.Name, info.Address, info.UploadID, info.Size)

	SQL = `Update organization set name = $1 where user_id = $2`
	_, err = database.CirconomyDB.Exec(SQL, info.Name, id)
	return err
}

func UpdatePersonProfileByID(id uuid.UUID, info *models.EditPersonProfile) error {
	// language = SQL
	SQL := `UPDATE users SET name=$2, address=$3, upload_id=$4::UUID WHERE id=$1
`
	uploadId := null.NewString(info.UploadID, true)
	if info.UploadID == "" {
		uploadId.Valid = false
	}
	_, err := database.CirconomyDB.Exec(SQL, id, info.Name, info.Address, uploadId)

	return err
}
