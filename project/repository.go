package project

import (
	"circonomy-server/database"
	"circonomy-server/dbutil"
	"circonomy-server/repobase"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

func (r *Repository) getAllProjects() ([]project, error) {
	// language = SQL
	SQL := `SELECT projects.id, 
			name, 
			project_time, 
			capacity,
			address,
			lat,
			long,
			continent,
			country,
			available, 
			rate, 
			method, 
			description, 
			certificates_ids, 
			contacts_ids, 
			project_status,
			u.path
			FROM projects
			LEFT JOIN uploads u on u.id = projects.upload_id
			WHERE projects.archived_at IS NULL
			AND u.archived_at IS NULL
`
	projects := make([]project, 0)
	err := r.Select(&projects, SQL)
	return projects, err
}

func (r *Repository) getProjectsByStatus(status string, filter locationFilter) ([]project, error) {
	// language = SQL
	SQL := `SELECT projects.id, 
			name, 
			project_time, 
			capacity,  
			address,
			lat,
			long,
			continent,
			country,
			available, 
			rate, 
			method, 
			description, 
			certificates_ids, 
			contacts_ids, 
			project_status,
			u.path
			FROM projects
			LEFT JOIN uploads u on u.id = projects.upload_id
			WHERE project_status=$1
			AND ($2 OR address=$3)
			AND projects.archived_at IS NULL
			AND u.archived_at IS NULL
`
	projects := make([]project, 0)
	err := r.Select(&projects, SQL, status, !filter.IsFiltered, filter.Address)
	return projects, err
}

func (r *Repository) getAllCertificates() ([]certificate, error) {
	// language = SQL
	SQL := `SELECT c.id, 
			c.name,  
			u.path
			FROM certificates c
			LEFT JOIN uploads u on u.id = c.upload_id
			WHERE c.archived_at IS NULL
			AND u.archived_at IS NULL
`
	certificates := make([]certificate, 0)
	err := r.Select(&certificates, SQL)
	return certificates, err
}

func (r *Repository) getAllContacts() ([]contact, error) {
	// language = SQL
	SQL := `SELECT c.id, 
			c.name,  
			u.path,
			c.description,
			c.email,
			c.linkedin_link,
			c.phone,
			c.designation
			FROM contacts c 
			LEFT JOIN uploads u on c.upload_id = u.id
			WHERE c.archived_at IS NULL
			AND u.archived_at IS NULL
`
	contacts := make([]contact, 0)
	err := r.Select(&contacts, SQL)
	return contacts, err
}

func (r *Repository) getContactsByIds(ids []uuid.UUID) ([]contact, error) {
	// language = SQL
	SQL := `SELECT c.id, 
			c.name,  
			u.path,
			c.description,
			c.email,
			c.designation,
			c.phone,
			c.linkedin_link
			FROM contacts c 
			LEFT JOIN uploads u on u.id = c.upload_id
			WHERE c.id = ANY ($1)
			AND c.archived_at IS NULL 
			AND u.archived_at IS NULL
`
	contacts := make([]contact, 0)
	err := r.Select(&contacts, SQL, pq.Array(ids))
	return contacts, err
}

func (r *Repository) getCertificatesByIds(ids []uuid.UUID) ([]certificate, error) {
	// language = SQL
	SQL := `SELECT c.id, 
			c.name,  
			u.path,
			c.status
			FROM certificates c
			LEFT JOIN uploads u on u.id = c.upload_id
			WHERE c.id = ANY ($1)
			AND c.archived_at IS NULL
			AND u.archived_at IS NULL
`
	certificates := make([]certificate, 0)
	err := r.Select(&certificates, SQL, pq.Array(ids))
	return certificates, err
}

func (r *Repository) getProjectDetailsByProjectId(iD uuid.UUID) ([]projectDetails, error) {
	// language = SQL
	SQL := `SELECT
			project_details_upload.name,  
			u.path 
			FROM project_details_upload
			LEFT JOIN uploads u on u.id = project_details_upload.upload_id
			WHERE project_details_upload.project_id = $1
			AND project_details_upload.archived_at IS NULL
			AND u.archived_at IS NULL
`
	details := make([]projectDetails, 0)
	err := r.Select(&details, SQL, iD)
	return details, err
}

func (r *Repository) getProjectByID(id uuid.UUID) (project, error) {
	// language = SQL
	SQL := `SELECT projects.id, 
			name, 
			project_time, 
			capacity,  
			address,
			lat,
			long,
			continent,
			country,
			available, 
			rate, 
			method, 
			description, 
			certificates_ids, 
			contacts_ids, 
			project_status,
			u.path,
			projects.methodology
			FROM projects
			LEFT JOIN uploads u on u.id = projects.upload_id
			WHERE projects.id=$1
			AND projects.archived_at IS NULL
			AND u.archived_at IS NULL 
`
	project := project{}
	err := r.Get(&project, SQL, id)
	return project, err
}

func (r *Repository) getProjectLocations() ([]locationDetails, error) {
	// language = SQL
	SQL := `SELECT DISTINCT address,lat,long,continent,country FROM projects WHERE archived_at IS NULL GROUP BY address,lat,long,continent,country;`
	states := make([]locationDetails, 0)
	err := r.Select(&states, SQL)
	return states, err
}

func (r *Repository) createCertification(request certificateRequest) (uuid.UUID, error) {
	// language = SQL
	SQL := `INSERT INTO certificates(name, upload_id, status) values ($1, $2, $3) returning id`
	var id uuid.UUID
	err := database.CirconomyDB.Get(&id, SQL, request.Name, request.ImageId, request.Status)
	return id, err
}

func (r *Repository) createContact(request contactRequest) (uuid.UUID, error) {
	// language = SQL
	SQL := `INSERT INTO contacts(name, description, email, upload_id, designation, phone, linkedin_link) VALUES ($1,$2,$3,$4,$5,$6,&$7) returning id`
	var id uuid.UUID
	err := database.CirconomyDB.Get(&id, SQL, request.Name, request.Description, request.Email, request.ImageId, request.Designation, request.Phone, request.LinkedinLink)
	return id, err
}

func (r *Repository) createProject(request createProjectRequest, userID uuid.UUID) (uuid.UUID, error) {
	var projectID uuid.UUID
	err := r.txx(func(txRepo *Repository) error {
		certificateIDs := make(UUIDSlice, 0)
		for _, certificate := range request.Certificates {
			id, err := txRepo.createCertification(certificate)
			if err != nil {
				return err
			}
			certificateIDs = append(certificateIDs, id)
		}

		contactIDs := make(UUIDSlice, 0)
		for _, contact := range request.Contacts {
			id, err := txRepo.createContact(contact)
			if err != nil {
				return err
			}
			contactIDs = append(contactIDs, id)
		}

		// language = SQL
		SQL := `INSERT INTO projects(created_by, name, project_time, capacity, address, available, rate, method, description, 
                     certificates_ids, contacts_ids, project_status, upload_id, lat, long, continent, country, methodology) 
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15, $16, $17, $18) returning id`
		err := txRepo.Get(&projectID, SQL, userID, request.Name, request.ProjectTime, request.Capacity, request.Address,
			request.Available, request.Rate, request.Method, request.Description, pq.Array(certificateIDs), pq.Array(contactIDs), request.ProjectStatus,
			request.ImageId, request.Lat, request.Long, request.Continent, request.Country, request.Methodology)
		if err != nil {
			return err
		}

		// language = SQL
		SQL = `INSERT INTO project_credits_operation(created_by, project_id, operation, amount, after_operation_amount, message) VALUES ($1,$2,$3,$4,$5, $6)`
		_, err = txRepo.Exec(SQL, userID, projectID, projectCreditOperationAddition, request.Available, request.Available, "Initial project creation")
		if err != nil {
			return err
		}

		// language = SQL
		SQL = `INSERT INTO project_details_upload(name, upload_id, project_id) VALUES ($1,$2,$3)`
		for _, details := range request.ProjectDetails {
			_, err = txRepo.Exec(SQL, details.Name, details.ImageId, projectID)
			if err != nil {
				return err
			}
		}
		return nil

	})
	return projectID, err
}
