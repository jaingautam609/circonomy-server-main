package project

import (
	"circonomy-server/utils"
	"context"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	repository *Repository
}

func NewService(jobsRepository *Repository) *Service {
	return &Service{
		repository: jobsRepository,
	}
}

func (s *Service) getAllProjects() ([]project, error) {
	return s.repository.getAllProjects()
}

func (s *Service) getProjectsByStatus(status string, address string) ([]details, error) {
	g, _ := errgroup.WithContext(context.Background())

	projects := make([]project, 0)
	allCertificates := make([]certificate, 0)
	allContacts := make([]contact, 0)

	locFilter := locationFilter{}
	locFilter.Address = address
	if locFilter.Address != "" {
		locFilter.IsFiltered = true
	}
	g.Go(func() error {
		var err error
		projects, err = s.repository.getProjectsByStatus(status, locFilter)
		if err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		var err error
		allCertificates, err = s.repository.getAllCertificates()
		if err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		var err error
		allContacts, err = s.repository.getAllContacts()
		if err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	resProjects := make([]details, 0)

	for i := 0; i < len(projects); i++ {
		certificates := make([]certificate, 0)
		contacts := make([]contact, 0)

		for j := 0; j < len(allCertificates); j++ {
			if utils.ContainsUUID(projects[i].CertificatesIds, allCertificates[j].ID) {
				if allCertificates[i].ImagePath.Valid {
					url, urlErr := utils.GenerateSignedURL(allCertificates[j].ImagePath.String)
					if urlErr != nil {
						return nil, urlErr
					}
					allCertificates[j].ImageURL = url
				}
				certificates = append(certificates, allCertificates[j])
			}
		}

		for j := 0; j < len(allContacts); j++ {
			if utils.ContainsUUID(projects[i].ContactsIds, allContacts[j].ID) {
				if allContacts[i].ImagePath.Valid {
					url, urlErr := utils.GenerateSignedURL(allContacts[j].ImagePath.String)
					if urlErr != nil {
						return nil, urlErr
					}
					allContacts[j].ImageURL = url
				}
				contacts = append(contacts, allContacts[j])
			}
		}

		resProject := details{
			ID:            projects[i].ID,
			Name:          projects[i].Name,
			ProjectTime:   projects[i].ProjectTime,
			Capacity:      projects[i].Capacity,
			Address:       projects[i].Address,
			Lat:           projects[i].Lat,
			Long:          projects[i].Long,
			Continent:     projects[i].Continent,
			Country:       projects[i].Country,
			Available:     projects[i].Available,
			Rate:          projects[i].Rate,
			Method:        projects[i].Method,
			Description:   projects[i].Description,
			Certificates:  certificates,
			Contacts:      contacts,
			ImagePath:     projects[i].ImagePath,
			ProjectStatus: projects[i].ProjectStatus,
		}

		if resProject.ImagePath.Valid {
			url, urlErr := utils.GenerateSignedURL(resProject.ImagePath.String)
			if urlErr != nil {
				return nil, urlErr
			}
			resProject.ImageURL = url
		}

		resProjects = append(resProjects, resProject)
	}
	return resProjects, nil
}

func (s *Service) getProjectById(ID uuid.UUID) (*details, error) {
	project, getErr := s.repository.getProjectByID(ID)
	if getErr != nil {
		return nil, getErr
	}

	certificates, certErr := s.repository.getCertificatesByIds(project.CertificatesIds)
	if certErr != nil {
		return nil, certErr
	}
	for i := 0; i < len(certificates); i++ {
		if certificates[i].ImagePath.Valid {
			url, urlErr := utils.GenerateSignedURL(certificates[i].ImagePath.String)
			if urlErr != nil {
				return nil, urlErr
			}
			certificates[i].ImageURL = url
		}
	}

	projectDetails, err := s.repository.getProjectDetailsByProjectId(ID)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(projectDetails); i++ {
		if projectDetails[i].ImagePath.Valid {
			url, urlErr := utils.GenerateSignedURL(projectDetails[i].ImagePath.String)
			if urlErr != nil {
				return nil, urlErr
			}
			projectDetails[i].ImageURL = url
		}
	}

	contacts, contErr := s.repository.getContactsByIds(project.ContactsIds)
	if contErr != nil {
		return nil, contErr
	}

	for i := 0; i < len(contacts); i++ {
		if contacts[i].ImagePath.Valid {
			url, urlErr := utils.GenerateSignedURL(contacts[i].ImagePath.String)
			if urlErr != nil {
				return nil, urlErr
			}
			contacts[i].ImageURL = url
		}
	}
	resProject := details{
		ID:             project.ID,
		Name:           project.Name,
		ProjectTime:    project.ProjectTime,
		Capacity:       project.Capacity,
		Address:        project.Address,
		Lat:            project.Lat,
		Long:           project.Long,
		Continent:      project.Continent,
		Country:        project.Country,
		Available:      project.Available,
		Rate:           project.Rate,
		Method:         project.Method,
		Description:    project.Description,
		Certificates:   certificates,
		Contacts:       contacts,
		ImagePath:      project.ImagePath,
		ProjectStatus:  project.ProjectStatus,
		ProjectDetails: projectDetails,
		Methodology:    project.Methodology,
	}

	if resProject.ImagePath.Valid {
		url, urlErr := utils.GenerateSignedURL(project.ImagePath.String)
		if urlErr != nil {
			return nil, urlErr
		}
		resProject.ImageURL = url
	}
	return &resProject, nil
}

func (s *Service) getProjectLocations() ([]locationDetails, error) {
	return s.repository.getProjectLocations()
}

func (s *Service) createProject(request createProjectRequest, userID uuid.UUID) (uuid.UUID, error) {
	return s.repository.createProject(request, userID)
}
