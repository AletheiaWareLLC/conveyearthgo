package conveyearthgo

import (
	"aletheiaware.com/netgo"
	"os"
)

func Scheme() string {
	if netgo.IsSecure() {
		return "https"
	}
	return "http"
}

func Host() string {
	return os.Getenv("HOST")
}
