package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	cpsdk "github.com/rudderlabs/rudder-cp-sdk"
	"github.com/rudderlabs/rudder-cp-sdk/diff"
	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-cp-sdk/poller"
	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
)

/**
* TODO use the diff package to update the cache with the new workspace configs
**/

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	c := config.New()
	loggerFactory := logger.NewFactory(c)
	log := loggerFactory.NewLogger()

	if err := run(ctx, log); err != nil {
		log.Fataln("main error", obskit.Error(err))
		os.Exit(1)
	}
}

// setupControlPlaneSDK creates a new SDK instance using the environment variables
func setupControlPlaneSDK(log logger.Logger) (*cpsdk.ControlPlane, error) {
	apiUrl := os.Getenv("CPSDK_API_URL")
	workspaceToken := os.Getenv("CPSDK_WORKSPACE_TOKEN")
	namespace := os.Getenv("CPSDK_NAMESPACE")
	hostedSecret := os.Getenv("CPSDK_HOSTED_SECRET")
	adminUsername := os.Getenv("CPSDK_ADMIN_USERNAME")
	adminPassword := os.Getenv("CPSDK_ADMIN_PASSWORD")

	options := []cpsdk.Option{
		cpsdk.WithBaseUrl(apiUrl),
		cpsdk.WithLogger(log),
	}

	if namespace != "" {
		options = append(options, cpsdk.WithNamespaceIdentity(namespace, hostedSecret))
	} else {
		options = append(options, cpsdk.WithWorkspaceIdentity(workspaceToken))
	}

	if adminUsername != "" || adminPassword != "" {
		if adminUsername == "" || adminPassword == "" {
			return nil, fmt.Errorf("both admin username and password must be provided")
		}

		adminCredentials := &identity.AdminCredentials{
			AdminUsername: adminUsername,
			AdminPassword: adminPassword,
		}
		options = append(options, cpsdk.WithAdminCredentials(adminCredentials))
	}

	return cpsdk.New(options...)
}

func setupWorkspaceConfigsPoller[K comparable, T diff.UpdateableElement](
	getter poller.WorkspaceConfigsGetter[K, T],
	handler poller.WorkspaceConfigsHandler[K, T],
	constructor func() diff.UpdateableList[K, T],
	log logger.Logger,
) (*poller.WorkspaceConfigsPoller[K, T], error) {
	return poller.NewWorkspaceConfigsPoller(getter, handler, constructor,
		poller.WithLogger[K, T](log.Child("poller")),
		poller.WithPollingInterval[K, T](1*time.Second),
		poller.WithPollingBackoffInitialInterval[K, T](1*time.Second),
		poller.WithPollingBackoffMaxInterval[K, T](1*time.Minute),
		poller.WithPollingBackoffMultiplier[K, T](1.5),
		poller.WithOnResponse[K, T](func(err error) {
			if err != nil {
				// Bump metric on failure
			} else {
				// Bump metric on success
			}
		}),
	)
}

func setupClientWithPoller(log logger.Logger) (func(context.Context), error) {
	sdk, err := setupControlPlaneSDK(log)
	if err != nil {
		return nil, fmt.Errorf("error setting up control plane sdk: %v", err)
	}

	cache := &modelv2.WorkspaceConfigs[string, *modelv2.WorkspaceConfig]{}
	updater := diff.Updater[string, *modelv2.WorkspaceConfig]{}

	p, err := setupWorkspaceConfigsPoller[string, *modelv2.WorkspaceConfig](
		func(ctx context.Context, l diff.UpdateableList[string, *modelv2.WorkspaceConfig], updatedAfter time.Time) error {
			return sdk.GetWorkspaceConfigs(ctx, l, updatedAfter)
		},
		func(list diff.UpdateableList[string, *modelv2.WorkspaceConfig]) (time.Time, error) {
			return updater.UpdateCache(list, cache)
		},
		func() diff.UpdateableList[string, *modelv2.WorkspaceConfig] {
			return &modelv2.WorkspaceConfigs[string, *modelv2.WorkspaceConfig]{}
		},
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up poller: %v", err)
	}

	return p.Run, nil
}

// run is the main function that uses the SDK
func run(ctx context.Context, log logger.Logger) error {
	poll, err := setupClientWithPoller(log)
	if err != nil {
		return fmt.Errorf("error setting up client with poller: %v", err)
	}

	poll(ctx) // blocking call, runs until context is cancelled

	return nil
}
