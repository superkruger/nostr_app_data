package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetStringOrDefault returns the value of the named environment variable, or the default value if the environment variable
// is not set.
func GetStringOrDefault(name, defaultValue string) string {
	res := os.Getenv(name)
	if res == "" {
		return defaultValue
	}
	return res
}

// GetStringsOrDefault returns an array of strings seperated by comma from the environment variable,
// or the default value if the environment variable is not set.
func GetStringsOrDefault(name string, defaultValue []string) []string {
	res := os.Getenv(name)
	if res == "" {
		return defaultValue
	}
	return strings.Split(res, ",")
}

// GetBoolOrDefault returns the value of the environment variable or the default value if not present.
// Errors if an invalid value is provided
func GetBoolOrDefault(name string, defaultValue bool) (bool, error) {
	res := os.Getenv(name)
	if res == "" {
		return defaultValue, nil
	}
	b, err := strconv.ParseBool(res)
	if err != nil {
		return false, fmt.Errorf("failed to parse %s=%s: %w", name, res, err)
	}
	return b, nil
}

// MustGetBoolOrDefault returns the value of the environment variable or the default value if not present.
// Panics if an invalid value is provided
func MustGetBoolOrDefault(name string, defaultValue bool) bool {
	res := os.Getenv(name)
	if res == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(res)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s=%s: %w", name, res, err))
	}
	return b
}

// MustGetBool returns the value of the environment variable.
// Panics if the value is missing or an invalid value is provided.
func MustGetBool(name string) bool {
	res := os.Getenv(name)
	if res == "" {
		panic(fmt.Errorf("missing env variable: %s", name))
	}
	b, err := strconv.ParseBool(res)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s=%s: %w", name, res, err))
	}
	return b
}

// MustGetString returns the value of the environment variable.
// Panics if the value is missing or an empty value is provided.
func MustGetString(name string) string {
	res := os.Getenv(name)
	if res == "" {
		panic(fmt.Errorf("missing env variable: %s", name))
	}
	return res
}

// MustGetStrings returns an array of strings seperated by comma from the environment variable.
// Panics if the value is missing or an empty value is provided.
func MustGetStrings(name string) []string {
	res := os.Getenv(name)
	if res == "" {
		panic(fmt.Errorf("missing env variable: %s", name))
	}
	return strings.Split(res, ",")
}

// GetIntOrDefault returns the value of the environment variable or the default value if not present.
// Errors if an invalid value is provided
func GetIntOrDefault(name string, defaultValue int) (int, error) {
	res := os.Getenv(name)
	if res == "" {
		return defaultValue, nil
	}
	b, err := strconv.Atoi(res)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s=%s: %w", name, res, err)
	}
	return b, nil
}

// MustGetIntOrDefault returns the value of the environment variable or the default value if not present.
// Panics if an invalid value is provided
func MustGetIntOrDefault(name string, defaultValue int) int {
	res := os.Getenv(name)
	if res == "" {
		return defaultValue
	}
	b, err := strconv.Atoi(res)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s=%s: %w", name, res, err))
	}
	return b
}

// MustGetInt returns the value of the environment variable.
// Panics if the value is missing or an invalid value is provided.
func MustGetInt(name string) int {
	res := os.Getenv(name)
	if res == "" {
		panic(fmt.Errorf("missing env variable: %s", name))
	}
	b, err := strconv.Atoi(res)
	if err != nil {
		panic(fmt.Errorf("failed to parse %s=%s: %w", name, res, err))
	}
	return b
}
