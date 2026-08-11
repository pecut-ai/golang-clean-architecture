package bootstrap

import (
	"context"
	"errors"
	"fmt"
)

func (a *Application) Start() error {
	a.Log.Infow("http_server_starting", "port", a.Cfg.HTTP.Port, "url", a.Cfg.App.URL)
	if err := a.App.Listen(fmt.Sprintf(":%d", a.Cfg.HTTP.Port)); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}
	var errs []error
	if a.App != nil {
		if err := a.App.ShutdownWithContext(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown HTTP server: %w", err))
		}
	}
	if err := a.closeResources(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (a *Application) closeResources() error {
	var errs []error
	if a.Auth != nil && a.Auth.Client != nil {
		if err := a.Auth.Client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close auth-service client: %w", err))
		}
		a.Auth = nil
	}
	if a.Producer != nil {
		if err := a.Producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close Kafka producer: %w", err))
		}
		a.Producer = nil
	}
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("get database pool for close: %w", err))
		} else if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close database pool: %w", err))
		}
		a.DB = nil
	}
	return errors.Join(errs...)
}
