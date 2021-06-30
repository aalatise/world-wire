package utility

import "strings"

func Contains(a []string, x string) int {
	for key, n := range a {
		if x == strings.ToUpper(n) {
			return key
		}
	}
	return -1
}
