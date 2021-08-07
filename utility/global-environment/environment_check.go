package global_environment

import (
	"os"
)

func VariableCheck() {
	domainId := os.Getenv(ENV_KEY_HOME_DOMAIN_NAME)
	svcName := os.Getenv(ENV_KEY_SERVICE_NAME)
	envVersion := os.Getenv(ENV_KEY_ENVIRONMENT_VERSION)
	secretLocation := os.Getenv(ENV_KEY_SECRET_STORAGE_LOCATION)

	if domainId == "" || svcName == "" || envVersion == "" || secretLocation == "" {
		panic("Initializing failed, Require the following environment variables to start up the service.\nHOME_DOMAIN_NAME: " + domainId + "\nSERVICE_NAME: " + svcName + "\nENVIRONMENT_VERSION: " + envVersion + "\nSECRET_STORAGE_LOCATION: " + secretLocation)
	}
}
