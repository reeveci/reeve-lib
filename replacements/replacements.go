package replacements

import (
	"fmt"
	"regexp"
	"strings"
)

func Apply(value string, expressions []string) (result string, err error) {
	result = value
	for _, expression := range expressions {
		result, err = RegexpReplace(result, expression)
		if err != nil {
			return
		}
	}
	return
}

func RegexpReplace(value string, expression string) (string, error) {
	runes := []rune(expression)

	if len(runes) == 0 || runes[0] == '\\' {
		return "", invalidExpressionError(expression)
	}

	parts := splitUnescaped(expression, runes[0], '\\')
	if len(parts) != 4 || parts[0] != "" || parts[3] != "" {
		return "", invalidExpressionError(expression)
	}

	regex, err := regexp.Compile(parts[1])
	if err != nil {
		return "", fmt.Errorf(`error compiling regexp for replace "%s" - %s`, parts[1], err)
	}

	return regex.ReplaceAllString(value, parts[2]), nil
}

func splitUnescaped(s string, sep, esc rune) []string {
	seps := string(sep)
	escs := string(esc)

	result := make([]string, 0, strings.Count(s, seps)+1)
	current := ""

	for i := strings.IndexRune(s, sep); i >= 0; i = strings.IndexRune(s, sep) {
		escaped := countBack(s[:i], esc)%2 == 1
		if escaped {
			current += s[:i-len(escs)] + seps
		} else {
			result = append(result, current+s[:i])
			current = ""
		}
		s = s[i+len(seps):]
	}

	result = append(result, current+s)
	return result
}

func countBack(s string, r rune) int {
	result := 0
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i -= 1 {
		if runes[i] != r {
			break
		}
		result += 1
	}
	return result
}

func invalidExpressionError(input string) error {
	return fmt.Errorf(`invalid regexp substitution expression - expected "/regex/substitution/" (where '/' may be replaced by any unicode character except the escape sequence '\') but got "%s"`, input)
}
