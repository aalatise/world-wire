package stringutil

import (
	"github.com/op/go-logging"
	"math/rand"
	"os/exec"
	"time"
)

var LOGGER = logging.MustGetLogger("utilities")
var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
var letterRunesSpecial = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+<>?,./:[]{};'")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int, includeSpecialCharacters bool) string {
	length := len(letterRunes)

	if includeSpecialCharacters {
		length = len(letterRunesSpecial)
	}

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(length)]
	}

	return string(b)
}

func ToString(i interface{}) string {
	if i != nil {
		return i.(string)
	} else {
		return ""
	}
}

func GenerateUUID() string {
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		LOGGER.Errorf(err.Error())
		return ""
	}
	return string(out)
}

