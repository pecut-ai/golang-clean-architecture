package bootstrap

import (
	"context"
	"errors"
	"fmt"

	"golang-clean-architecture/internal/config"
	httproute "golang-clean-architecture/internal/delivery/http/route"
	"golang-clean-architecture/internal/gateway/messaging"
	"golang-clean-architecture/internal/logging"
	appmiddleware "golang-clean-architecture/internal/middleware"
	"golang-clean-architecture/internal/repository"
	"golang-clean-architecture/internal/usecase"

	"github.com/IBM/sarama"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	authservice "github.com/pecut-ai/auth-service/pkg/v2/auth"
	authclient "github.com/pecut-ai/auth-service/pkg/v2/client"
	"gorm.io/gorm"
)

const ComponentAPI = "api"

type Application struct {
	Ctx      context.Context
	Cfg      *config.Config
	Log      *logging.Logger
	App      *fiber.App
	API      huma.API
	DB       *gorm.DB
	Producer sarama.SyncProducer
	Auth     *authservice.SDK
}

func NewApplication(ctx context.Context, cfg *config.Config, log *logging.Logger) (_ *Application, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg == nil {
		return nil, errors.New("configuration is required")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate configuration: %w", err)
	}
	if log == nil {
		log = logging.NewNop()
	}

	app := &Application{Ctx: ctx, Cfg: cfg, Log: log.Component(ComponentAPI)}
	defer func() {
		if err != nil {
			_ = app.closeResources()
		}
	}()

	app.App = fiber.New(config.FiberConfig(cfg, app.Log))
	app.registerBaseMiddleware()

	app.DB, err = config.OpenDatabase(cfg.Database, log)
	if err != nil {
		return nil, err
	}
	app.Producer, err = config.OpenKafkaProducer(cfg.Kafka)
	if err != nil {
		return nil, err
	}
	if err = app.configureAuth(); err != nil {
		return nil, err
	}
	app.registerAuthMiddleware()

	app.API = humafiber.New(app.App, cfg.HumaConfig())
	app.registerRoutes()
	app.Log.Infow("application_bootstrap_ready")
	return app, nil
}

func (a *Application) registerBaseMiddleware() {
	a.App.Use(requestid.New(requestid.Config{Header: fiber.HeaderXRequestID}))
	a.App.Use(appmiddleware.RequestContext(ComponentAPI))
	a.App.Use(appmiddleware.RequestLogger(a.Log))
	a.App.Use(appmiddleware.Recover(a.Log))
	a.App.Use(appmiddleware.CORS(a.Cfg.HTTP))
}

func (a *Application) configureAuth() error {
	if !a.Cfg.Auth.Enabled {
		a.Log.Warnw("auth_service_disabled")
		return nil
	}
	authLog := a.Log.Component("auth_service")
	client, err := authclient.New(authclient.Config{
		Target:         a.Cfg.Auth.Target,
		ServiceID:      a.Cfg.Auth.ServiceID,
		InternalSecret: a.Cfg.Auth.InternalSecret,
	})
	if err != nil {
		return fmt.Errorf("create auth-service client: %w", err)
	}
	a.Auth = authservice.New(client, authservice.WithLogger(func(ctx context.Context) func(string, ...any) {
		return func(message string, args ...any) {
			authLog.FromContext(ctx).With(args...).Debug(message)
		}
	}))
	return nil
}

func (a *Application) registerAuthMiddleware() {
	a.App.Use("/api", func(c *fiber.Ctx) error {
		if c.Path() == "/api/health" {
			return c.Next()
		}
		if a.Auth == nil {
			return fiber.ErrServiceUnavailable
		}
		return a.Auth.Middleware.AuthMiddleware()(c)
	})
	a.App.Use("/api", appmiddleware.AuthContext())
}

func (a *Application) registerRoutes() {
	validate := validator.New()
	repositoryLog := a.Log.Component("repository")
	producerLog := a.Log.Component("kafka_producer")
	usecaseLog := a.Log.Component("usecase")

	contactRepository := repository.NewContactRepository(repositoryLog)
	addressRepository := repository.NewAddressRepository(repositoryLog)

	var contactProducer *messaging.ContactProducer
	var addressProducer *messaging.AddressProducer
	if a.Producer != nil {
		contactProducer = messaging.NewContactProducer(a.Producer, producerLog)
		addressProducer = messaging.NewAddressProducer(a.Producer, producerLog)
	}

	httproute.Register(a.API, httproute.Dependencies{
		Contact: usecase.NewContactUseCase(a.DB, usecaseLog, validate, contactRepository, contactProducer),
		Address: usecase.NewAddressUseCase(a.DB, usecaseLog, validate, contactRepository, addressRepository, addressProducer),
	})
}
