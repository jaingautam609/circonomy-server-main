package dbhelpers

import (
	"circonomy-server/database"
	"circonomy-server/dbutil"
	"circonomy-server/models"
	"circonomy-server/providers"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"

	"github.com/google/uuid"
)

func CreateIndividual(user *models.User) (uuid.UUID, error) {
	// language = SQL
	SQL := `INSERT INTO users(name ,email ,password ,number ,address ,account_type, country_code) VALUES ($1,$2,$3,$4,$5,$6, $7) returning id`
	var id uuid.UUID
	err := database.CirconomyDB.Get(&id, SQL, user.Name, user.Email, user.Password, user.Number, user.Address, user.AccountType, user.CountryCode)
	return id, err
}

func CreateBusiness(user *models.RegisterUserRequest) (uuid.UUID, uuid.UUID, error) {
	// language = SQL
	var orgId, id uuid.UUID
	SQL := `INSERT INTO organization(name) VALUES ($1) returning id`
	err := database.CirconomyDB.Get(&orgId, SQL, user.Name)
	if err != nil {
		return orgId, id, err
	}
	// language = SQL
	SQL = `INSERT INTO users(name ,email ,password ,number ,address, account_type ,size, organisation_id, country_code) VALUES ($1,$2,$3,$4,$5,$6,$7, $8, $9) returning id`
	err = database.CirconomyDB.Get(&id, SQL, user.Name, user.Email, user.Password, user.Number, user.Address, user.AccountType, user.OrgDetails, orgId, user.CountryCode)
	if err != nil {
		return orgId, id, err
	}
	SQL = `update organization set user_id = $1 where id = $2`
	_, err = database.CirconomyDB.Exec(SQL, id, orgId)
	return orgId, id, err
}

func CreateOrganization(name string) (uuid.UUID, error) {
	// language = SQL
	var orgId uuid.UUID
	SQL := `INSERT INTO organization(name) VALUES ($1) returning id`
	err := database.CirconomyDB.Get(&orgId, SQL, name)
	return orgId, err

}

func CreateUserOrganizationLink(userID uuid.UUID, orgID uuid.NullUUID) error {
	return dbutil.WithTransaction(database.CirconomyDB, func(tx *sqlx.Tx) error {
		// language = SQL
		SQL := `UPDATE users_organisations set archived_at = now() where user_id = $1 and archived_at is null`
		_, err := tx.Exec(SQL, userID)
		if err != nil {
			return err
		}
		if orgID.Valid {
			SQL = `INSERT INTO users_organisations(user_id, organization_id) VALUES ($1,$2)`
			_, err = tx.Exec(SQL, userID, orgID)
			return err
		}
		return nil
	})

}

func CreateUserSession(userID uuid.UUID, token string, expiry time.Time) error {
	// language = SQL
	SQL := `INSERT INTO sessions(user_id,session_token,expiry_time) VALUES ($1,$2,$3)`
	_, err := database.CirconomyDB.Exec(SQL, userID, token, expiry)
	return err
}

func GetOrganizationID(org string) (uuid.UUID, error) {
	// language = SQL
	SQL := `SELECT id from organization where name = $1 AND archived_at IS NULL`
	var id uuid.UUID
	err := database.CirconomyDB.Get(&id, SQL, org)
	return id, err
}

func DoesEmailExist(email string) (bool, error) {
	// language = SQL
	SQL := `SELECT exists (SELECT 1 FROM users where email = $1 AND archived_at IS NULL) `
	var res bool
	err := database.CirconomyDB.Get(&res, SQL, email)
	return res, err
}

func DoesNumberExist(number, countryCode string) (bool, error) {
	// language = SQL
	SQL := `SELECT exists (SELECT 1 FROM users where number = $1 AND country_code = $2 AND archived_at IS NULL)`
	var res bool
	err := database.CirconomyDB.Get(&res, SQL, number, countryCode)

	return res, err
}

func GetPasswordByEmail(email string) (string, error) {
	// language = SQL
	SQL := `SELECT password from users where email = $1 AND archived_at IS NULL`
	var pwd string
	err := database.CirconomyDB.Get(&pwd, SQL, email)
	return pwd, err
}

func InsertEnquiry(request models.EnquiryRequest, emailType providers.EmailType) error {
	//language = SQL
	SQL := `INSERT INTO enquiry(first_name, last_name, email, query, email_type) values ($1, $2, $3, $4, $5)`
	_, err := database.CirconomyDB.Exec(SQL, request.FirstName, request.LastName, request.Email, request.QueryString, emailType)
	return err
}

func GetUserByEmailID(email string) (models.UserInfo, error) {
	// language = SQL
	SQL := `SELECT users.id as id,account_type, u.path FROM users left join uploads u on u.id = users.upload_id WHERE email = $1 AND users.archived_at IS NULL`
	var id models.UserInfo
	err := database.CirconomyDB.Get(&id, SQL, email)

	return id, err
}

func GetUserByUserID(userID uuid.UUID) (models.UserSessionInfo, error) {
	// language = SQL
	SQL := `
		SELECT 
		       id,
		       account_type, 
		       email, 
		       number 
		FROM 
		    users 
		WHERE 
		    id = $1 
		  AND users.archived_at IS NULL`
	var userSessionInfo models.UserSessionInfo
	err := database.CirconomyDB.Get(&userSessionInfo, SQL, userID)

	return userSessionInfo, err
}

func GetSession(sessionToken string) (models.Session, error) {
	// language = SQL
	SQL := `SELECT user_id,expiry_time FROM sessions WHERE session_token=$1 AND archived_at IS NULL `
	var sess models.Session
	err := database.CirconomyDB.Get(&sess, SQL, sessionToken)
	return sess, err
}

func DelSession(sessionToken string) error {
	// language = SQL
	SQL := `UPDATE sessions SET archived_at=now() WHERE session_token=$1`
	_, err := database.CirconomyDB.Exec(SQL, sessionToken)
	return err
}

func GetUserOrgName(userID uuid.UUID) (string, uuid.NullUUID, error) {
	// language = SQL
	SQL := `SELECT name,id from organization WHERE id=(SELECT organization_id FROM users_organisations WHERE user_id = $1 AND archived_at IS NULL ORDER BY created_at desc LIMIT 1) AND archived_at IS NULL `
	var name string
	var orgID uuid.NullUUID
	err := database.CirconomyDB.QueryRowx(SQL, userID).Scan(&name, &orgID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return name, orgID, err
	}
	return name, orgID, nil
}

func GetAllOrgNames() ([]models.Organization, error) {
	// language = SQL
	SQL := `SELECT id,name FROM organization where archived_at IS NULL`
	orgList := make([]models.Organization, 0)
	err := database.CirconomyDB.Select(&orgList, SQL)
	return orgList, err
}

func GetAllIndividualNames() ([]models.UserBasicInfo, error) {
	// language = SQL
	SQL := `SELECT id,name FROM users WHERE account_type=$1 AND archived_at IS NULL`
	orgList := make([]models.UserBasicInfo, 0)
	err := database.CirconomyDB.Select(&orgList, SQL, models.UserAccountTypeIndividual)
	return orgList, err
}

func UpdateUserImage(userID, uploadID uuid.UUID) error {
	// language = SQL
	SQL := `UPDATE users SET upload_id=$2 WHERE id=$1 AND archived_at IS NULL`
	_, err := database.CirconomyDB.Exec(SQL, userID, uploadID)
	return err
}

func UpdatePassword(user models.LoginCredentials) error {
	// language = SQL
	SQL := `UPDATE users SET password=$2 WHERE email=$1 AND archived_at IS NULL`
	_, err := database.CirconomyDB.Exec(SQL, user.Email, user.Password)
	return err
}

func StoreOTP(otp models.OTP) error {
	// language = SQL
	SQL := `INSERT INTO user_otp(input,otp,type,expiry) VALUES($1,$2,$3,$4) `
	_, err := database.CirconomyDB.Exec(SQL, otp.Input, otp.OTP, otp.Type, otp.Expiry)
	return err
}

func GetOTP(otp models.OTP) (string, error) {
	// language = SQL
	SQL := `SELECT otp FROM user_otp WHERE input=$1 AND type=$2 AND expiry > now() AND archived_at IS NULL ORDER BY expiry desc LIMIT 1`
	var fetchedOTP string
	err := database.CirconomyDB.Get(&fetchedOTP, SQL, otp.Input, otp.Type)
	return fetchedOTP, err
}
