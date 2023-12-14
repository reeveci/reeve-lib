package conditions

import "github.com/reeveci/reeve-lib/schema"

func ApplyDefaults(conditions *map[string]schema.Condition, defaults map[string]schema.Condition) {
	if *conditions == nil {
		*conditions = make(map[string]schema.Condition, len(defaults))
	}

	for key, condition := range defaults {
		if key != "" {
			if (*conditions)[key].Empty() {
				(*conditions)[key] = condition
			}
		}
	}
}

func Check(facts map[string]schema.Fact, conditions map[string]schema.Condition, env map[string]schema.Env, vars map[string]schema.Var) (bool, error) {
	for key, condition := range conditions {
		if key != "" {
			ok, err := condition.Check(key, facts, env, vars)
			if !ok || err != nil {
				return false, err
			}
		}
	}

	return true, nil
}
