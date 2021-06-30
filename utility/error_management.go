package utility

import (
	"github.com/op/go-logging"
	"os"
)

func ExitOnErr(LOGGER *logging.Logger, err error, errorMsg string) {

	if err != nil {
		LOGGER.Errorf(errorMsg+":  %v", err)
		os.Exit(1)
	}

}
