package envutil

import (
	"fmt"
	"os"
	"strconv"
)

const (
	Dev        = "dev"
	Production = "production"
)

func IsDev() bool {
	for _, env := range []string{"IS_LOCAL", "LOCAL", "LOCAL_ENV"} {
		if _, exist := os.LookupEnv(env); exist {
			return true
		}
	}
	return false
}

func TryToGet(key string) (value, set bool) {
	if _, set = os.LookupEnv(key); !set {
		return false, false
	}
	v, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return false, true
	}
	return v, true
}

func MustGetNotEmpty(key string) string {
	if _, exist := os.LookupEnv(key); !exist {
		panic(fmt.Sprintf("env '%s' is not set", key))
	}
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("env '%s' is empty", key))
	}
	return v
}

func Env() string {
	if IsDev() {
		return Dev
	}
	return Production
}
