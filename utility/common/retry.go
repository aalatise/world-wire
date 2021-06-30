package common

import (
	"reflect"
	"runtime"
	"time"

	"github.com/stellar/go/support/errors"
)

// For retrying invoking function until time is up or function successfully return result.
// Can pass math.Inf(1) as retry_times for infinite retries.
func Retry(retry_times float64, interval time.Duration, fn func() error) error {
	originalFuncName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	LOGGER.Infof("Retrying calling function %v", originalFuncName)
	LOGGER.Infof("----- Calling interval: %v, Calling times: %v -----", interval, retry_times)
	for retry_times > 0 {
		LOGGER.Infof("Attempt invoking function: %v", originalFuncName)
		if err := fn(); err == nil {
			return nil
		}
		LOGGER.Warningf("Failed invoking function, %v retrying attempts remains", retry_times)
		time.Sleep(interval)
		retry_times--
	}

	return errors.New("Failed calling function " + originalFuncName)
}
