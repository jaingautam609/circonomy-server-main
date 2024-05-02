package main

import (
	"circonomy-server/database"
	"circonomy-server/server"
	"circonomy-server/utils"
	"github.com/sirupsen/logrus"
	"os"
)

// TODO do not delete
//type projectStatus string
//
//const (
//	projectStatusActive   = "active"
//	projectStatusSoldOut  = "sold out"
//	projectStatusUpcoming = "upcoming"
//)
//
//type projectQuality string
//
//const (
//	projectQualityHigh = "high quality"
//	projectQualityLow  = "low quality"
//)
//
//type UUIDSlice []uuid.UUID
//
//type createProjectRequest struct {
//	Name          string               `json:"name" db:"name"`
//	ProjectTime   string               `json:"projectTime" db:"project_time"`
//	Capacity      string               `json:"capacity" db:"capacity"`
//	Address       string               `json:"address" db:"address"`
//	Lat           float64              `json:"lat" db:"lat"`
//	Long          float64              `json:"long" db:"long"`
//	Continent     null.String          `json:"continent" db:"continent"`
//	Country       null.String          `json:"country" db:"country"`
//	Available     int                  `json:"available" db:"available"`
//	Rate          int                  `json:"rate" db:"rate"`
//	Method        projectQuality       `json:"method" db:"method"`
//	Description   string               `json:"description" db:"description"`
//	Certificates  []certificateRequest `json:"certificates" db:"certificates_ids"`
//	Contacts      []contactRequest     `json:"contacts" db:"contacts_ids"`
//	ProjectStatus projectStatus        `json:"projectStatus" db:"project_status"`
//	ImageId       uuid.UUID            `json:"imageId"`
//}
//
//type contactRequest struct {
//	Name        string    `json:"name"`
//	ImageId     uuid.UUID `json:"imageId"`
//	Description string    `json:"description"`
//	Email       string    `json:"email"`
//}
//
//type certificateRequest struct {
//	Name    string    `json:"name" db:"name"`
//	ImageId uuid.UUID `json:"imageId"`
//}
//
//func main() {
//	emp := &createProjectRequest{
//		Name:        "test",
//		ProjectTime: "2 years",
//		Capacity:    "1000 mt",
//		Address:     "Noida",
//		Lat:         28.8,
//		Long:        78,
//		Continent:   null.StringFrom("asia"),
//		Country:     null.StringFrom("india"),
//		Available:   600,
//		Rate:        100,
//		Method:      projectQualityHigh,
//		Description: "This is test",
//		Certificates: []certificateRequest{
//			{
//				Name:    "test certificate",
//				ImageId: uuid.UUID{},
//			},
//		},
//		Contacts: []contactRequest{
//			{
//				Name:        "test contact",
//				ImageId:     uuid.UUID{},
//				Description: "this is tet contact",
//				Email:       "rahul@remotestate.com",
//			},
//		},
//		ProjectStatus: projectStatusActive,
//		ImageId:       uuid.UUID{},
//	}
//	e, err := json.Marshal(emp)~
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(string(e))
//}

func main() {
	if err := database.ConnectAndMigrate(os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		database.SSLModeDisable); err != nil {
		logrus.WithError(err).Panic("Failed to initialize and migrate database")
	}
	logrus.Info("database connection and migration successful!!")
	utils.CreateAWSStorageClient()
	logrus.Info("AWS storage client created successfully!!")
	srv := server.SetupRoutes()
	if err := srv.Run(); err != nil {
		logrus.Errorln("ListenAndServe Errors:", err)
	}
}
