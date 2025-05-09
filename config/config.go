package config

import (
	"fmt"
	"foundation-go/utility/envutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

var v = viper.New()

var (
	MainConfigName = "config.yaml"
	MainConfigDir  = "config"
)

var (
	defaultConfigPath = "./" + filepath.Join(MainConfigDir, envutil.Production, MainConfigName)
)

func Init() {
	log.Println("moduleMainConfigPath", moduleMainConfigPath())
	appendConfig(moduleMainConfigPath())
}

func moduleMainConfigPath() string {
	currentPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// 往上级目录寻找config/，直到找到 go.mod
	for {
		configFile, err := filepath.Abs(filepath.Join(currentPath, MainConfigDir, envutil.Env(), MainConfigName))
		if err != nil {
			panic(err)
		}

		if _, err := os.Stat(configFile); err == nil {
			return configFile
		}
		// 当前目录包含go.mod，不再往上级寻找
		if _, err := os.Stat(filepath.Join(currentPath, "go.mod")); err == nil {
			break
		}

		parentDir := filepath.Dir(currentPath)
		if parentDir == currentPath {
			break
		}
		currentPath = parentDir
	}

	return defaultConfigPath
}

func appendConfig(configFilePath string) {
	log.Printf("load config %s.", configFilePath)
	v.SetConfigFile(configFilePath)
	v.AutomaticEnv()
	if err := v.MergeInConfig(); err != nil {
		panic(err.Error())
	}
}

func AllKeys() []string                   { return v.AllKeys() }
func AllSettings() map[string]interface{} { return v.AllSettings() }
func Get(key string) interface{}          { return v.Get(key) }
func IsSet(key string) bool               { return v.IsSet(key) }
func GetInt(key string) int               { return v.GetInt(key) }
func GetUint(key string) uint             { return v.GetUint(key) }
func GetInt32(key string) int32           { return v.GetInt32(key) }
func GetUint32(key string) uint32         { return v.GetUint32(key) }
func GetInt64(key string) int64           { return v.GetInt64(key) }
func GetUint64(key string) uint64         { return v.GetUint64(key) }
func GetFloat64(key string) float64       { return v.GetFloat64(key) }
func GetBool(key string) bool             { return v.GetBool(key) }
func GetString(key string) string         { return v.GetString(key) }
func GetIntSlice(key string) []int        { return v.GetIntSlice(key) }
func GetStringSlice(key string) []string  { return v.GetStringSlice(key) }
func GetTime(key string) time.Time        { return v.GetTime(key) }
func GetStringMap(key string) map[string]interface{} {
	return v.GetStringMap(key)
}
func GetStringMapString(key string) map[string]string {
	return v.GetStringMapString(key)
}
func GetStringMapStringSlice(key string) map[string][]string {
	return v.GetStringMapStringSlice(key)
}
func GetDurationInSecond(key string) time.Duration {
	return v.GetDuration(key) * time.Second
}
func GetDurationInMillSecond(key string) time.Duration {
	return v.GetDuration(key) * time.Millisecond
}

func Sub(key string) *viper.Viper { return v.Sub(key) }

// MustSet panic if config is not existed.
// return full path if existed.
func MustSet(key string) string {
	if v.IsSet(key) {
		return key
	}
	panic(fmt.Sprintf("The configuration item %s is missing.", key))
}

func MustGetInt(key string) int         { return v.GetInt(MustSet(key)) }
func MustGetUint(key string) uint       { return v.GetUint(MustSet(key)) }
func MustGetInt32(key string) int32     { return v.GetInt32(MustSet(key)) }
func MustGetUint32(key string) uint32   { return v.GetUint32(MustSet(key)) }
func MustGetInt64(key string) int64     { return v.GetInt64(MustSet(key)) }
func MustGetUint64(key string) uint64   { return v.GetUint64(MustSet(key)) }
func MustGetFloat64(key string) float64 { return v.GetFloat64(MustSet(key)) }
func MustGetBool(key string) bool       { return v.GetBool(MustSet(key)) }
func MustGetString(key string) string   { return v.GetString(MustSet(key)) }
func MustGetIntSlice(key string) []int  { return v.GetIntSlice(MustSet(key)) }
func MustGetTime(key string) time.Time  { return v.GetTime(MustSet(key)) }
func MustGetStringSlice(key string) []string {
	return v.GetStringSlice(MustSet(key))
}
func MustGetStringMap(key string) map[string]interface{} {
	return v.GetStringMap(MustSet(key))
}
func MustGetStringMapString(key string) map[string]string {
	return v.GetStringMapString(MustSet(key))
}
func MustGetStringMapStringSlice(key string) map[string][]string {
	return v.GetStringMapStringSlice(MustSet(key))
}
func MustGetDurationInSecond(key string) time.Duration {
	return v.GetDuration(MustSet(key)) * time.Second
}
func MustGetDurationInMillSecond(key string) time.Duration {
	return v.GetDuration(MustSet(key)) * time.Millisecond
}

func GetIntOrDefault(or int, key string) int {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetInt(configPath)
}

func GetUintOrDefault(or uint, key string) uint {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetUint(configPath)
}

func GetInt32OrDefault(or int32, key string) int32 {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetInt32(configPath)
}

func GetUint32OrDefault(or uint32, key string) uint32 {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetUint32(configPath)
}

func GetInt64OrDefault(or int64, key string) int64 {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetInt64(configPath)
}

func GetUint64OrDefault(or uint64, key string) uint64 {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetUint64(configPath)
}

func GetFloat64OrDefault(or float64, key string) float64 {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetFloat64(configPath)
}

func GetBoolOrDefault(or bool, key string) bool {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetBool(configPath)
}

func GetStringOrDefault(or string, key string) string {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetString(configPath)
}

func GetIntSliceOrDefault(or []int, key string) []int {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetIntSlice(configPath)
}

func GetTimeOrDefault(or time.Time, key string) time.Time {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetTime(configPath)
}

func GetStringSliceOrDefault(or []string, key string) []string {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetStringSlice(configPath)
}

func GetStringMapOrDefault(or map[string]interface{}, key string) map[string]interface{} {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetStringMap(configPath)
}

func GetStringMapStringOrDefault(or map[string]string, key string) map[string]string {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetStringMapString(configPath)
}

func GetStringMapStringSliceOrDefault(or map[string][]string, key string) map[string][]string {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetStringMapStringSlice(configPath)
}

func GetDurationInSecondOrDefault(or time.Duration, key string) time.Duration {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetDuration(configPath) * time.Second
}

func GetDurationInMillSecondOrDefault(or time.Duration, key string) time.Duration {
	configPath := key
	if !v.IsSet(configPath) {
		return or
	}
	return v.GetDuration(configPath) * time.Millisecond
}
