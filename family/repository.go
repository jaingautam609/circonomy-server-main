package family

import (
	"circonomy-server/database"
	"circonomy-server/dbutil"
	"circonomy-server/repobase"
	"circonomy-server/utils"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	repobase.Base
}

func NewRepository(sqlDB *sqlx.DB) *Repository {
	return &Repository{
		repobase.NewBase(sqlDB),
	}
}

func (r *Repository) txx(fn func(txRepo *Repository) error) error {
	return dbutil.WithTransaction(r.DB(), func(tx *sqlx.Tx) error {
		repoCopy := *r
		repoCopy.Base = r.Base.CopyWithTX(tx)
		return fn(&repoCopy)
	})
}

func (r *Repository) createFamily(request createFamilyRequest, userID uuid.UUID) (*family, error) {
	var id uuid.UUID
	err := r.txx(func(txRepo *Repository) error {
		err := txRepo.removeFromOtherFamily(userID)
		if err != nil {
			return err
		}

		// language = SQL
		SQL := `INSERT INTO family(name, created_by) VALUES ($1, $2) RETURNING id`
		err = txRepo.Get(&id, SQL, request.Name, userID)
		if err != nil {
			return err
		}

		// language = SQL
		SQL = `INSERT INTO family_users(family_id, user_id) VALUES ($1, $2)`
		_, err = txRepo.Exec(SQL, id, userID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.getFamily(id, userID)
}

func (r *Repository) removeFromOtherFamily(userID uuid.UUID) error {
	// Ideally this should return single entity but an array is used for just in case scenario
	SQL := `UPDATE family_users set archived_at = now() where user_id = $1 and archived_at is null returning family_id`
	var familyIds []uuid.UUID
	err := r.Select(&familyIds, SQL, userID)
	if err != nil {
		return err
	}

	for _, familyId := range familyIds {
		SQL := `select created_by from family where id = $1`
		var familyCreatorId uuid.UUID
		err := r.Get(&familyCreatorId, SQL, familyId)
		if err != nil {
			return err
		}
		// only if the family from which this user is being removed is created/owned by this user
		if familyCreatorId == userID {
			SQL = `select user_id from family_users where family_id = $1 and archived_at is null order by created_at asc limit 1`
			var oldestMemberID uuid.UUID
			err = r.Get(&oldestMemberID, SQL, familyId)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					SQL = `UPDATE family set archived_at = now() where id = $1`
					_, err = r.Exec(SQL, userID)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				SQL = `UPDATE family set created_by = $1 where id = $2`
				_, err = r.Exec(SQL, oldestMemberID, familyId)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *Repository) updateFamily(request updateFamilyRequest, familyID uuid.UUID, userID uuid.UUID) (*family, error) {
	// language = SQL
	SQL := `UPDATE family set name = $1 where created_by = $2 and id = $3
`
	_, err := r.Exec(SQL, request.Name, userID, familyID)
	if err != nil {
		return nil, err
	}
	return r.getFamily(familyID, userID)
}

func (r *Repository) getFamily(familyID uuid.UUID, userID uuid.UUID) (*family, error) {
	familyDetails := &family{}
	// language = SQL
	SQL := `
		SELECT id, name, created_by from family where id = $1
`
	err := r.Get(familyDetails, SQL, familyID.String())
	if err != nil {
		return nil, err
	}

	familyDetails.Members, err = r.getMembers(familyID)
	if err != nil {
		return nil, err
	}
	familyDetails.Invitations, err = r.getInvitations(familyID)
	if err != nil {
		return nil, err
	}
	if familyDetails.CreatedBy == userID {
		familyDetails.CanSendInvitation = true
		return familyDetails, nil
	}
	for _, familyMember := range familyDetails.Members {
		if familyMember.Id == userID {
			return familyDetails, nil
		}
	}

	return nil, errors.New("This user is not an owner or a family member")
}

func (r *Repository) getMembers(familyID uuid.UUID) ([]*familyMember, error) {
	var familyMembers []*familyMember
	// language = SQL
	SQL := `
		SELECT 
		    users.id, 
		    users.name, 
		    users.email, 
		    users.number, 
		    uploads.path, 
		    project_details.project_count, 
		    project_details.total_credits
		FROM users 
			join family_users on users.id = family_users.user_id 
			left join uploads on users.upload_id = uploads.id
			left join (
					select projects_bought.user_id, count(distinct (p_id)) project_count, sum(credits) total_credits
					from projects_bought
					group by projects_bought.user_id
				) project_details on users.id = project_details.user_id
		WHERE family_users.family_id = $1
			and family_users.archived_at is null
			and uploads.archived_at is null
			and users.archived_at is null
`
	err := r.Select(&familyMembers, SQL, familyID)
	return familyMembers, err
}

func (r *Repository) getInvitations(familyID uuid.UUID) ([]*familyInvitation, error) {
	var invitations []*familyInvitation
	// language = SQL
	SQL := `
		SELECT email, created_at
		FROM family_invitations 
		WHERE family_invitations.family_id = $1
			and family_invitations.archived_at is null
`
	err := r.Select(&invitations, SQL, familyID)
	return invitations, err
}

func (r *Repository) checkAlreadyInvitedEmail(familyID uuid.UUID, email string) (string, error) {
	var existingEmail string
	SQL := `select email from family_invitations where email = $1 and family_id = $2 and archived_at is null`
	err := r.Get(&existingEmail, SQL, email, familyID)
	return existingEmail, err
}

func (r *Repository) checkAlreadyExistingMemberEmail(familyID uuid.UUID, email string) (string, error) {
	var existingEmail string
	SQL := `
	select users.email
	from users
			 join family_users on users.id = family_users.user_id
	where family_users.archived_at is null
	  and users.archived_at is null
	  and users.email = $1
	  and family_users.family_id = $2
`
	err := r.Get(&existingEmail, SQL, email, familyID)
	return existingEmail, err
}

func (r *Repository) invite(familyID uuid.UUID, email string) (uuid.UUID, error) {
	// language = SQL
	SQL := `INSERT INTO family_invitations(email, family_id) VALUES ($1, $2) returning id`
	var invitationID uuid.UUID
	err := r.Get(&invitationID, SQL, email, familyID)
	return invitationID, err
}

func (r *Repository) getFamilyNameOwnerName(familyID uuid.UUID) (string, string) {
	// language = SQL
	SQL := `
		select family.name, users.name
		from family
				 join users on family.created_by = users.id
		where family.id = $1
`
	var familyName, ownerName string
	err := database.CirconomyDB.QueryRowx(SQL, familyID).Scan(&familyName, &ownerName)
	if err != nil {
		logrus.Errorf("error while reading family details for family %v error %v", familyID, err)
	}
	return familyName, ownerName
}

func (r *Repository) getOwnFamily(userID uuid.UUID) (*family, error) {
	familyDetails := &family{}
	// language = SQL
	SQL := `
		SELECT id, name, created_by from family where id = (select family_id from family_users where user_id = $1 and archived_at is null)
`
	err := r.Get(familyDetails, SQL, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	familyDetails.Members, err = r.getMembers(familyDetails.Id)
	if err != nil {
		return nil, err
	}
	familyDetails.Invitations, err = r.getInvitations(familyDetails.Id)
	if err != nil {
		return nil, err
	}
	for i := range familyDetails.Members {
		if familyDetails.Members[i].ImagePath.Valid {
			url, urlErr := utils.GenerateSignedURL(familyDetails.Members[i].ImagePath.String)
			if urlErr != nil {
				return nil, err
			}
			familyDetails.Members[i].ImageUrl = url
		}
	}
	if familyDetails.CreatedBy == userID {
		familyDetails.CanSendInvitation = true
		return familyDetails, nil
	}

	for _, familyMember := range familyDetails.Members {
		if familyMember.Id == userID {
			return familyDetails, nil
		}
	}

	return nil, errors.New("This user is not an owner or a family member")
}

func (r *Repository) getInvitationDetails(invitationID uuid.UUID) (*OwnerDetails, error) {
	details := &OwnerDetails{}
	// language = SQL
	SQL := `
		select 
		       family.id as family_id,
		       family.name as family_name,
			   users.id as user_id,
			   users.name,
			   users.email,
			   users.number,
			   uploads.path
		from family_invitations
				 join family on family_invitations.family_id = family.id
				 join users on family.created_by = users.id
				 left join uploads on users.upload_id = uploads.id
		where users.archived_at is null
		  and family_invitations.archived_at is null
		  and users.account_type = 'individual'
		  and family.archived_at is null
		  and family_invitations.id = $1
`
	err := r.Get(details, SQL, invitationID)
	return details, err
}

func (r *Repository) getCurrentFamilyDetails(userID uuid.UUID) (*OwnerDetails, error) {
	details := &OwnerDetails{}
	// language = SQL
	SQL := `
		select 
		       family.id as family_id,
		       family.name as family_name,
			   family_owner.id as user_id,
			   family_owner.name,
			   family_owner.email,
			   family_owner.number,
			   uploads.path
		from users
				 join family_users on users.id = family_users.user_id
				 join family on family_users.family_id = family.id
				 join users family_owner on family.created_by = family_owner.id
				 left join uploads on family_owner.upload_id = uploads.id
		where users.archived_at is null
		  and family_users.archived_at is null
		  and users.account_type = 'individual'
		  and family.archived_at is null
		  and family_owner.archived_at is null
		  and users.id = $1
`
	err := r.Get(details, SQL, userID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return details, err
}

func (r *Repository) getCurrentInvitations(userID uuid.UUID) ([]*OwnerDetails, error) {
	var details []*OwnerDetails
	// language = SQL
	SQL := `
		select family.name as family_name,
			   family_owner.id as user_id,
			   family_owner.name,
			   family_owner.email,
			   family_owner.number,
			   uploads.path,
			   family_invitations.id as invitation_id
		from users
				 join family_invitations on family_invitations.email = users.email
				 join family on family_invitations.family_id = family.id
				 join users family_owner on family.created_by = family_owner.id
				 left join uploads on family_owner.upload_id = uploads.id
		where users.archived_at is null
		  and family_invitations.archived_at is null
		  and users.account_type = 'individual'
		  and family.archived_at is null
		  and family_owner.archived_at is null
		  and users.id = $1
`
	err := r.Select(&details, SQL, userID)
	return details, err
}

func (r *Repository) joinFamily(invitationID, userID uuid.UUID) error {
	return r.txx(func(txRepo *Repository) error {
		SQL := `select family_id from family_invitations where id = $1 and archived_at is null`
		var familyID uuid.UUID
		err := txRepo.Get(&familyID, SQL, invitationID)
		if err != nil {
			return err
		}

		SQL = `update family_invitations set archived_at = now() where id = $1`
		_, err = txRepo.Exec(SQL, invitationID)
		if err != nil {
			return err
		}

		err = txRepo.removeFromOtherFamily(userID)
		if err != nil {
			return err
		}

		// language = SQL
		SQL = `INSERT INTO family_users(family_id, user_id) VALUES ($1, $2)`
		_, err = txRepo.Exec(SQL, familyID, userID)
		return err
	})

}
