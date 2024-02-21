package trade

import "strings"

func IsEmptyPrice(price string) bool {
	if price == "" {
		return true
	}
	if strings.ReplaceAll(price, " ", "") == "" {
		return true
	}

	return false
}
