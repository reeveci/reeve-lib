package exe

import (
	"os"
	"strconv"
	"strings"
)

func GetEnvDef(name, def string) (result string) {
	result = os.Getenv(name)
	if result == "" {
		result = def
	}
	return
}

func GetBoolEnvDef(name string, def bool) (result bool) {
	value := strings.ToLower(os.Getenv(name))
	if value == "" {
		return def
	}
	result, _ = strconv.ParseBool(value)
	return
}

func GetEnvFieldMap(name, def string) (result map[string]bool) {
	fields := strings.Fields(GetEnvDef(name, def))
	result = make(map[string]bool, len(fields))
	for _, field := range fields {
		result[field] = true
	}
	return
}
