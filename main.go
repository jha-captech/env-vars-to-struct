package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
)

type ConfigCustom struct {
	Env  string `env:"ENV"`
	Text struct {
		TextValue string `env:"TEXT_VALUE"`
		BoolValue bool   `env:"BOOL_VALUE"`
		IntValue  int    `env:"INT_VALUE"`
	}
}

func main() {
	_ = os.Setenv("ENV", "dev")
	_ = os.Setenv("TEXT_VALUE", "this is text")
	_ = os.Setenv("BOOL_VALUE", strconv.FormatBool(true))
	_ = os.Setenv("INT_VALUE", strconv.FormatInt(50, 10))

	config := ConfigCustom{}
	if err := ParseStructFromEnv(&config, true); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(fmt.Sprintf("%+v", config))
}

// ParseStructFromEnv takes a struct as an input and recursively loops tough all fields on the
// struct. If a field is not another struct and has a `env` tag, the environment variable associated
// with that tag will be retrieved and added to the struct.
//
// If the `errOnMissingValue` flag is set to `true`, any tag that is missing an environment variable
// will result in an error being returned.
func ParseStructFromEnv(obj any, errOnMissingValue bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("in ParseStructFromEnv: %w", err)
		}
	}()
	val := reflect.ValueOf(obj)

	// if pointer, get value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Iterate through the struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// Check if the field is a struct
		if field.Kind() == reflect.Struct {
			if err := ParseStructFromEnv(field.Addr().Interface(), errOnMissingValue); err != nil {
				return err
			}
			continue
		}

		// Get and then set env value based on tag if present
		fieldType := val.Type().Field(i)
		envTag := fieldType.Tag.Get("env")

		if field.CanSet() && envTag != "" {
			switch field.Kind() {
			case reflect.String:
				value, err := getEnvString(envTag, errOnMissingValue)
				if err != nil {
					return err
				}
				field.SetString(value)
			case reflect.Int:
				value, err := getEnvInt64(envTag, errOnMissingValue)
				if err != nil {
					return err
				}
				field.SetInt(value)
			case reflect.Bool:
				value, err := getEnvBool(envTag, errOnMissingValue)
				if err != nil {
					return err
				}
				field.SetBool(value)
			default:
				continue
			}
		}
	}
	return nil
}

func newEnvVarMissingErr[T any](key string) (T, error) {
	var blank T
	errMsg := fmt.Sprintf("enviroment variable '%s' is missing or blank", key)
	return blank, errors.New(errMsg)
}

func newEnvVarParsingErr[T any](key string, err error) (T, error) {
	var blank T
	errMsg := fmt.Sprintf(
		"error parsing enviroment variable '%s' to type '%T': %v",
		key,
		blank,
		err,
	)
	return blank, errors.New(errMsg)
}

func getEnvString(key string, errIfMissing bool) (string, error) {
	value := os.Getenv(key)
	if errIfMissing && value == "" {
		return newEnvVarMissingErr[string](key)
	}
	return value, nil
}

func getEnvInt64(key string, errIfMissing bool) (int64, error) {
	value := os.Getenv(key)
	if errIfMissing && value == "" {
		return newEnvVarMissingErr[int64](key)
	}
	convertedInt, err := strconv.Atoi(value)
	if err != nil {
		return newEnvVarParsingErr[int64](key, err)
	}
	return int64(convertedInt), nil
}

func getEnvBool(key string, errIfMissing bool) (bool, error) {
	value := os.Getenv(key)
	if errIfMissing && value == "" {
		return newEnvVarMissingErr[bool](key)
	}
	convertedBool, err := strconv.ParseBool(value)
	if err != nil {
		return newEnvVarParsingErr[bool](key, err)
	}
	return convertedBool, nil
}
