package main

import (
	"context"
	"os"

	cpsdk "github.com/rudderlabs/rudder-control-plane-sdk"
	"github.com/rudderlabs/rudder-control-plane-sdk/identity"
	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

func main() {
	apiUrl := os.Getenv("CPSDK_API_URL")
	workspaceToken := os.Getenv("CPSDK_WORKSPACE_TOKEN")
	namespace := os.Getenv("CPSDK_NAMESPACE")
	hostedSecret := os.Getenv("CPSDK_HOSTED_SECRET")
	adminUsername := os.Getenv("CPSDK_ADMIN_USERNAME")
	adminPassword := os.Getenv("CPSDK_ADMIN_PASSWORD")

	c := config.New()
	loggerFactory := logger.NewFactory(c)
	var log logger.Logger = loggerFactory.NewLogger()

	var options []cpsdk.Option = []cpsdk.Option{
		cpsdk.WithBaseUrl(apiUrl),
		cpsdk.WithLogger(log),
	}

	if namespace != "" {
		options = append(options, cpsdk.WithNamespaceIdentity(namespace, hostedSecret))
	} else {
		options = append(options, cpsdk.WithWorkspaceIdentity(workspaceToken))
	}

	if adminUsername != "" && adminPassword != "" {
		adminCredentials := &identity.AdminCredentials{
			AdminUsername: adminUsername,
			AdminPassword: adminPassword,
		}
		options = append(options, cpsdk.WithAdminCredentials(adminCredentials))
	}

	sdk, err := cpsdk.New(options...)
	if err != nil {
		panic(err)
	}

	log.Info("Getting workspace configs")

	_, err = sdk.Client.GetWorkspaceConfigs(context.Background())
	if err != nil {
		panic(err)
	}
}
