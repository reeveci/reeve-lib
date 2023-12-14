package schema

import (
	"fmt"
	"regexp"
	"strings"
)

type Condition struct {
	// Text
	Include []string `json:"include" yaml:"include"`
	Exclude []string `json:"exclude" yaml:"exclude"`

	// Env
	IncludeEnv []string `json:"include env" yaml:"include env"`
	ExcludeEnv []string `json:"exclude env" yaml:"exclude env"`

	// Vars
	IncludeVar []string `json:"include var" yaml:"include var"`
	ExcludeVar []string `json:"exclude var" yaml:"exclude var"`

	// Regex
	Match    []string `json:"match" yaml:"match"`
	Mismatch []string `json:"mismatch" yaml:"mismatch"`
}

func (c Condition) Empty() bool {
	return len(c.Include) == 0 &&
		len(c.Exclude) == 0 &&

		len(c.IncludeEnv) == 0 &&
		len(c.ExcludeEnv) == 0 &&

		len(c.IncludeVar) == 0 &&
		len(c.ExcludeVar) == 0 &&

		len(c.Match) == 0 &&
		len(c.Mismatch) == 0
}

func (c Condition) Check(key string, facts map[string]Fact, env map[string]Env, vars map[string]Var) (bool, error) {
	// rule logic:
	// rule is empty OR fact is empty OR ((includes are empty OR (include matches OR ...)) AND (excludes are empty OR (exclude matches AND ...)))

	if c.Empty() {
		return true, nil
	}

	if strings.HasPrefix(key, ENV_PREFIX) {
		envKey := strings.TrimPrefix(key, ENV_PREFIX)
		if envKey != "" {
			return c.checkFact(Fact{env[envKey].Value}, env, vars)
		}
	}

	if strings.HasPrefix(key, VAR_PREFIX) {
		varKey := strings.TrimPrefix(key, VAR_PREFIX)
		if varKey != "" {
			return c.checkFact(Fact{string(vars[varKey])}, env, vars)
		}
	}

	return c.checkFact(facts[key], env, vars)
}

func (c Condition) checkFact(fact Fact, env map[string]Env, vars map[string]Var) (bool, error) {
	if len(fact) == 0 {
		return true, nil
	}

	factMap := make(map[string]bool, len(fact))
	for _, fact := range fact {
		factMap[fact] = true
	}

	// check includes

	hasRules := false
	found := false

	if len(c.Include) > 0 && !found {
		hasRules = true
		for _, value := range c.Include {
			if factMap[value] {
				found = true
				break
			}
		}
	}

	if len(c.IncludeEnv) > 0 && !found {
		hasRules = true
		for _, key := range c.IncludeEnv {
			value, ok := env[key]

			if ok && factMap[value.Value] {
				found = true
				break
			}
		}
	}

	if len(c.IncludeVar) > 0 && !found {
		hasRules = true
		for _, key := range c.IncludeVar {
			value, ok := vars[key]

			if ok && factMap[string(value)] {
				found = true
				break
			}
		}
	}

	if len(c.Match) > 0 && !found {
		hasRules = true
	L:
		for _, value := range c.Match {
			regexp, err := regexp.Compile(value)
			if err != nil {
				return false, fmt.Errorf(`error compiling regexp for condition [match "%s"] - %s`, value, err)
			}
			for _, f := range fact {
				if regexp.MatchString(f) {
					found = true
					break L
				}
			}
		}
	}

	if hasRules && !found {
		return false, nil
	}

	// check excludes

	if len(c.Exclude) > 0 && len(fact) > 0 {
		for _, value := range c.Exclude {
			if factMap[value] {
				return false, nil
			}
		}
	}

	if len(c.ExcludeEnv) > 0 && len(fact) > 0 {
		for _, key := range c.ExcludeEnv {
			value, ok := env[key]

			if ok && factMap[value.Value] {
				return false, nil
			}
		}
	}

	if len(c.ExcludeVar) > 0 && len(fact) > 0 {
		for _, key := range c.ExcludeVar {
			value, ok := vars[key]

			if ok && factMap[string(value)] {
				return false, nil
			}
		}
	}

	if len(c.Mismatch) > 0 && len(fact) > 0 {
		for _, value := range c.Mismatch {
			regexp, err := regexp.Compile(value)
			if err != nil {
				return false, fmt.Errorf(`error compiling regexp for condition [match not "%s"]. - %s`, value, err)
			}
			for _, f := range fact {
				if regexp.MatchString(f) {
					return false, nil
				}
			}
		}
	}

	return true, nil
}
