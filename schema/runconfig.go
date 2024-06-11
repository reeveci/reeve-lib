package schema

import (
	"fmt"

	"github.com/google/shlex"
	"github.com/reeveci/reeve-lib/replacements"
)

type RunConfig struct {
	Task      string              `json:"task" yaml:"task"`
	Command   RawCommand          `json:"command" yaml:"command"`
	Input     RawParam            `json:"input" yaml:"input"`
	Directory RawParam            `json:"directory" yaml:"directory"`
	User      RawParam            `json:"user" yaml:"user"`
	Params    map[string]RawParam `json:"params" yaml:"params"`
}

func (config RunConfig) GetEnv() []string {
	keys := make([]string, 0, 4+len(config.Params))

	if key, ok := getEnv(config.Command); ok {
		keys = append(keys, key)
	}
	if key, ok := getEnv(config.Input); ok {
		keys = append(keys, key)
	}
	if key, ok := getEnv(config.Directory); ok {
		keys = append(keys, key)
	}
	if key, ok := getEnv(config.User); ok {
		keys = append(keys, key)
	}
	for _, v := range config.Params {
		if key, ok := getEnv(v); ok {
			keys = append(keys, key)
		}
	}

	return keys
}

func getEnv(param RawParam) (string, bool) {
	switch value := param.(type) {
	case map[string]any:
		if envKey, ok := value["env"].(string); ok && envKey != "" {
			return envKey, true
		}
	case EnvParam:
		if value.Env != "" {
			return value.Env, true
		}
	}
	return "", false
}

func (config RunConfig) Resolve(env map[string]Env, vars map[string]Var) (result ResolvedRunConfig, unresolvedEnv, unresolvedVars []string, err error) {
	resolver := resolver{Env: env, Vars: vars}
	return resolver.Resolve(config)
}

type ResolvedRunConfig struct {
	Command   []string
	Input     string
	Directory string
	User      string
	Params    map[string]string
}

type resolver struct {
	Env  map[string]Env
	Vars map[string]Var

	unresolvedEnv, unresolvedVars []string
}

func (r *resolver) Resolve(config RunConfig) (result ResolvedRunConfig, unresolvedEnv, unresolvedVars []string, err error) {
	r.unresolvedEnv = make([]string, 0, 4+len(config.Params))
	r.unresolvedVars = make([]string, 0, 4+len(config.Params))

	result.Params = make(map[string]string, len(config.Params))

	if result.Command, _, err = r.resolveCommand(config.Command); err != nil {
		err = fmt.Errorf("invalid command - %s", err)
		return
	}
	if result.Input, _, err = r.resolve(config.Input); err != nil {
		err = fmt.Errorf("invalid input - %s", err)
		return
	}
	if result.Directory, _, err = r.resolve(config.Directory); err != nil {
		err = fmt.Errorf("invalid directory - %s", err)
		return
	}
	if result.User, _, err = r.resolve(config.User); err != nil {
		err = fmt.Errorf("invalid user - %s", err)
		return
	}

	for k, v := range config.Params {
		if k == "" {
			continue
		}

		var value string
		var found bool
		value, found, err = r.resolve(v)
		if err != nil {
			err = fmt.Errorf("invalid param \"%s\" - %s", k, err)
			return
		}
		if !found {
			continue
		}
		result.Params[k] = value
	}

	return result, r.unresolvedEnv, r.unresolvedVars, nil
}

func (r *resolver) resolveCommand(command RawCommand) (result []string, found bool, err error) {
	switch value := command.(type) {
	case []string:
		return value, true, nil

	case LiteralCommand:
		return value, true, nil

	case []any:
		args := make([]string, len(value))
		for i, rawArg := range value {
			var ok bool
			args[i], ok = rawArg.(string)
			if !ok {
				err = fmt.Errorf("command may only contain strings but contains %T (%v)", rawArg, rawArg)
				return
			}
		}
		return args, true, nil

	default:
		var resolvedCommand string
		if resolvedCommand, found, err = r.resolve(command); !found || err != nil {
			return
		}
		result, err = shlex.Split(resolvedCommand)
		return
	}
}

func (r *resolver) resolve(param RawParam) (result string, found bool, err error) {
	switch value := param.(type) {
	case nil:
		return "", true, nil

	case string:
		return value, true, nil

	case LiteralParam:
		return string(value), true, nil

	case map[string]any:
		if rawEnv := value["env"]; rawEnv != nil {
			envKey, ok := rawEnv.(string)
			if !ok {
				err = fmt.Errorf("env must be a string but is %T (%v)", rawEnv, rawEnv)
				return
			}
			var envVal Env
			envVal, found = r.Env[envKey]
			if !found {
				r.unresolvedEnv = append(r.unresolvedEnv, envKey)
				return
			}
			rawExpressions, _ := value["replace"].([]any)
			expressions := make([]string, len(rawExpressions))
			for i, rawReplace := range rawExpressions {
				expressions[i], ok = rawReplace.(string)
				if !ok {
					err = fmt.Errorf("replace may only contain strings but contains %T (%v)", rawReplace, rawReplace)
					return
				}
			}
			result, err = replacements.Apply(envVal.Value, expressions)
			return
		}

		if rawVar := value["var"]; rawVar != nil {
			varKey, ok := value["var"].(string)
			if !ok {
				err = fmt.Errorf("var must be a string but is %T (%v)", rawVar, rawVar)
				return
			}
			var varVal Var
			varVal, found = r.Vars[varKey]
			if !found {
				r.unresolvedVars = append(r.unresolvedVars, varKey)
				return
			}
			rawExpressions, _ := value["replace"].([]any)
			expressions := make([]string, len(rawExpressions))
			for i, rawReplace := range rawExpressions {
				expressions[i], ok = rawReplace.(string)
				if !ok {
					err = fmt.Errorf("replace may only contain strings but contains %T (%v)", rawReplace, rawReplace)
					return
				}
			}
			result, err = replacements.Apply(string(varVal), expressions)
			return
		}

		err = fmt.Errorf("unexpected value %v", value)
		return

	case EnvParam:
		var envVal Env
		envVal, found = r.Env[value.Env]
		if !found {
			r.unresolvedEnv = append(r.unresolvedEnv, value.Env)
			return
		}
		result, err = replacements.Apply(envVal.Value, value.Replace)
		return

	case VarParam:
		var varVal Var
		varVal, found = r.Vars[value.Var]
		if !found {
			r.unresolvedVars = append(r.unresolvedVars, value.Var)
			return
		}
		result, err = replacements.Apply(string(varVal), value.Replace)
		return

	default:
		err = fmt.Errorf("unexpected value %v of type %T", value, value)
		return
	}
}
