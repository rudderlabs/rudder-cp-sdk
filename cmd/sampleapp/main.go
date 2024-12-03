package main

import (
	"context"
	"fmt"
	"os"
	"time"

	cpsdk "github.com/rudderlabs/rudder-cp-sdk"
	"github.com/rudderlabs/rudder-cp-sdk/identity"
	"github.com/rudderlabs/rudder-cp-sdk/modelv2"
	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

var log logger.Logger

func main() {
	// setup a logger using the rudder-go-kit package
	c := config.New()
	loggerFactory := logger.NewFactory(c)
	log = loggerFactory.NewLogger()

	if err := run(); err != nil {
		log.Fatalf("main error: %v", err)
	}
}

// setupControlPlaneSDK creates a new SDK instance using the environment variables
func setupControlPlaneSDK() (*cpsdk.ControlPlane, error) {
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

// run is the main function that uses the SDK
func run() error {
	sdk, err := setupControlPlaneSDK()
	if err != nil {
		return fmt.Errorf("error setting up control plane sdk: %v", err)
	}
	defer sdk.Close()

	var wcs modelv2.WorkspaceConfigs
	err = sdk.Client.GetWorkspaceConfigs(context.Background(), &wcs, time.Time{})
	if err != nil {
		return fmt.Errorf("error getting workspace configs: %v", err)
	}

	// Some time passes and we want to get the latest changes only, without re-fetching the entire config
	err = sdk.Client.GetWorkspaceConfigs(context.Background(), &wcs, wcs.UpdatedAt())
	if err != nil {
		return fmt.Errorf("error getting workspace configs: %v", err)
	}

	return nil
}
