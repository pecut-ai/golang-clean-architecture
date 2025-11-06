package config

import (
	"golang-clean-architecture/internal/delivery/http"
	"golang-clean-architecture/internal/delivery/http/middleware"
	"golang-clean-architecture/internal/delivery/http/route"
	"golang-clean-architecture/internal/gateway/messaging"
	"golang-clean-architecture/internal/repository"
	"golang-clean-architecture/internal/usecase"
	"os"

	"github.com/IBM/sarama"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB       *gorm.DB
	App      *fiber.App
	Api      huma.API
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
	Producer sarama.SyncProducer
}

func Bootstrap(config *BootstrapConfig) {
	// Setup Huma API configuration
	humaConfig := huma.DefaultConfig(config.Config.GetString("APP_NAME"), "1.0.0")
	humaConfig.Servers = []*huma.Server{
		{URL: "http://localhost:" + config.Config.GetString("WEB_PORT")},
	}
	humaConfig.DocsPath = "/docs"
	humaConfig.OpenAPIPath = "/openapi.json"
	// Add security scheme for bearer token
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Bearer token authentication",
		},
	}

	// setup repositories
	userRepository := repository.NewUserRepository(config.Log)
	contactRepository := repository.NewContactRepository(config.Log)
	addressRepository := repository.NewAddressRepository(config.Log)

	// setup producer
	var userProducer *messaging.UserProducer
	var contactProducer *messaging.ContactProducer
	var addressProducer *messaging.AddressProducer

	if config.Producer != nil {
		userProducer = messaging.NewUserProducer(config.Producer, config.Log)
		contactProducer = messaging.NewContactProducer(config.Producer, config.Log)
		addressProducer = messaging.NewAddressProducer(config.Producer, config.Log)
	}

	// setup use cases
	userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository, userProducer)
	contactUseCase := usecase.NewContactUseCase(config.DB, config.Log, config.Validate, contactRepository, contactProducer)
	addressUseCase := usecase.NewAddressUseCase(config.DB, config.Log, config.Validate, contactRepository, addressRepository, addressProducer)

	// setup controller
	userController := http.NewUserController(userUseCase, config.Log)
	contactController := http.NewContactController(contactUseCase, config.Log)
	addressController := http.NewAddressController(addressUseCase, config.Log)

	// setup middleware
	authMiddleware := middleware.NewAuth(userUseCase)

	routeConfig := route.RouteConfig{
		App:               config.App,
		UserController:    userController,
		ContactController: contactController,
		AddressController: addressController,
		AuthMiddleware:    authMiddleware,
	}
	if !IsProduction() {
		api := humafiber.New(config.App, humaConfig)
		routeConfig.SetupDocs(api)
	}
	routeConfig.Setup()
}

func IsProduction() bool {
	env, exists := os.LookupEnv("APP_ENV")
	return exists && env == "production"
}
