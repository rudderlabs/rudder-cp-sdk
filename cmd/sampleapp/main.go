package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
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

func setupWorkspaceConfigsPoller[K comparable](
	getter poller.WorkspaceConfigsGetter[K],
	handler poller.WorkspaceConfigsHandler[K],
	constructor func() diff.UpdateableObject[K],
	log logger.Logger,
) (*poller.WorkspaceConfigsPoller[K], error) {
	return poller.NewWorkspaceConfigsPoller(getter, handler, constructor,
		poller.WithLogger[K](log.Child("poller")),
		poller.WithPollingInterval[K](1*time.Second),
		poller.WithPollingBackoffInitialInterval[K](1*time.Second),
		poller.WithPollingBackoffMaxInterval[K](1*time.Minute),
		poller.WithPollingBackoffMultiplier[K](1.5),
		poller.WithOnResponse[K](func(updated bool, err error) {
			if err != nil {
				// e.g. bump metric on failure
				log.Errorn("failed to poll workspace configs", obskit.Error(err))
			} else {
				// e.g. bump metric on success
				log.Debugn("successfully polled workspace configs", logger.NewBoolField("updated", updated))
			}
		}),
	)
}

func setupClientWithPoller(
	cache diff.UpdateableObject[string],
	cacheMu *sync.RWMutex,
	log logger.Logger,
) (func(context.Context), error) {
	sdk, err := setupControlPlaneSDK(log)
	if err != nil {
		return nil, fmt.Errorf("error setting up control plane sdk: %v", err)
	}

	updater := diff.Updater[string]{}

	p, err := setupWorkspaceConfigsPoller[string](
		func(ctx context.Context, l diff.UpdateableObject[string], updatedAfter time.Time) error {
			return sdk.GetWorkspaceConfigs(ctx, l, updatedAfter)
		},
		func(obj diff.UpdateableObject[string]) (time.Time, bool, error) {
			cacheMu.Lock()
			defer cacheMu.Unlock()
			return updater.UpdateCache(obj, cache)
		},
		func() diff.UpdateableObject[string] {
			return &modelv2.WorkspaceConfigs{}
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
	var (
		// WARNING: if you don't want to use modelv2.WorkspaceConfigs because you're interested in a smaller subset of
		// the data, then have a look at diff_test.go for an example of how to implement a custom UpdateableList.
		cache   = &modelv2.WorkspaceConfigs{}
		cacheMu = &sync.RWMutex{}
	)

	poll, err := setupClientWithPoller(cache, cacheMu, log)
	if err != nil {
		return fmt.Errorf("error setting up client with poller: %v", err)
	}

	pollingDone := make(chan struct{})
	go func() {
		poll(ctx) // blocking call, runs until context is cancelled
		close(pollingDone)
	}()

	cacheMu.RLock()
	// do something with "cache" here
	cacheMu.RUnlock() // nolint:staticcheck

	return nil
}
