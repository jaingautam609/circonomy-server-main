package server

import (
	"circonomy-server/admin"
	"circonomy-server/database"
	"circonomy-server/family"
	"circonomy-server/farmer"
	"circonomy-server/handlers"
	"circonomy-server/kilnOperator"
	"circonomy-server/middlewares"
	"circonomy-server/project"
	"circonomy-server/providers"
	"circonomy-server/providers/emailprovider"
	"circonomy-server/providers/smsprovider"
	"circonomy-server/utils"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Server struct {
	chi.Router
	emailProvider       providers.EmailProvider
	server              *http.Server
	FamilyHandler       *family.Handler
	ProjectHandler      *project.Handler
	FarmerHandler       *farmer.Handler
	AdminHandler        *admin.Handler
	kilnOperatorHandler *kilnOperator.Handler
}

const (
	ReadHeaderTimeout = time.Second * 10
)

func SetupRoutes() *Server {

	server := &Server{
		emailProvider: emailprovider.NewSendGridEmailProvider(os.Getenv("SENDGRID_KEY")),
	}
	router := chi.NewRouter()
	// family provides handler for family
	familyService := family.NewService(family.NewRepository(database.CirconomyDB), server.emailProvider)
	server.FamilyHandler = family.NewHandler(familyService)
	projectService := project.NewService(project.NewRepository(database.CirconomyDB))
	server.ProjectHandler = project.NewHandler(projectService)
	handlers.FamilyService = familyService
	handlers.EmailProvider = server.emailProvider

	smsProvider := smsprovider.NewSMSProvider()
	farmerService := farmer.NewService(farmer.NewRepository(database.CirconomyDB), smsProvider)
	server.FarmerHandler = farmer.NewHandler(farmerService)
	server.AdminHandler = admin.NewHandler(admin.NewService(admin.NewRepository(database.CirconomyDB)), farmerService)

	kilnOperatorService := kilnOperator.NewService(kilnOperator.NewRepository(database.CirconomyDB), smsProvider)
	server.kilnOperatorHandler = kilnOperator.NewHandler(kilnOperatorService, farmerService)

	router.Route("/", func(r chi.Router) {
		r.Use(middlewares.CommonMiddlewares()...)
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.RespondJSON(w, http.StatusOK, struct {
				Status string `json:"status"`
			}{Status: "server is running"})
		})
		r.Post("/register", handlers.SignUp)
		r.Post("/login", handlers.Login)
		r.Post("/enquiry", handlers.ContactUs)
		r.Post("/contactUs", handlers.ContactUs)
		r.Post("/subscribe", handlers.Subscribe)
		r.Get("/logout", handlers.Logout)
		r.Post("/check-otp", handlers.CheckOTP)
		r.Post("/send-otp", handlers.SendOTP)
		r.Post("/check-password", handlers.CheckPassword)
		r.Put("/reset-password", handlers.ResetPassword)
		r.Get("/email-exist/{email}", handlers.DoesEmailExist)
		r.Post("/number-exist", handlers.DoesNumberExist)
		r.Post("/upload-image", handlers.UploadImage)
		r.Route("/public", func(home chi.Router) {
			home.Route("/clients", func(cl chi.Router) {
				cl.Get("/org", handlers.GetOrgClientsList)
				cl.Get("/individual", handlers.GetIndividualClientsList)
			})
			home.Get("/org-list", handlers.GetOrgList)
			home.Get("/individual-list", handlers.GetIndividualList)
		})
		r.Route("/profile", func(pr chi.Router) {
			pr.Use(middlewares.AuthMiddleware)
			pr.Get("/id/{id}", handlers.GetProfileByID)
			pr.Get("/org-people/{id}", handlers.GetPeopleInOrg)
			pr.Get("/person-history/{id}", handlers.GetPeopleCreditsHistory)
			pr.Get("/org-history/{id}", handlers.GetOrgCreditsHistory)
			pr.Put("/org/{id}", handlers.EditOrgProfile)
			pr.Put("/person/{id}", handlers.EditPersonProfile)
		})
		r.Route("/project", server.ProjectHandler.Serve)
		r.Route("/family", server.FamilyHandler.Serve)
		r.Route("/farmer", server.FarmerHandler.Serve)
		r.Route("/admin", server.AdminHandler.Serve)
		r.Route("/kiln-operator", server.kilnOperatorHandler.Serve)
	})
	server.Router = router
	return server
}

func (svc *Server) Run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	svc.server = &http.Server{
		Addr:              ":" + port,
		Handler:           svc.Router,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}
	logrus.Infof("server starting at port %s", port)
	return svc.server.ListenAndServe()
}

func (svc *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return svc.server.Shutdown(ctx)
}
