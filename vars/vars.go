package vars

import (
	"fmt"
	"strings"

	"github.com/reeveci/reeve-lib/schema"
)

type PipelineEnvBundle struct {
	Env                       []string
	PipelineEnv, RemainingEnv []string
}

func FindAllEnv(pipeline schema.Pipeline) (result PipelineEnvBundle) {
	pipelineResults := make(map[string]bool)
	remainingResults := make(map[string]bool)

	for key, condition := range pipeline.When {
		if key == "" {
			continue
		}

		if strings.HasPrefix(key, schema.ENV_PREFIX) {
			envKey := strings.TrimPrefix(key, schema.ENV_PREFIX)
			if envKey != "" {
				pipelineResults[envKey] = true
			}
		}

		for _, key := range condition.IncludeEnv {
			if key != "" {
				pipelineResults[key] = true
			}
		}
		for _, key := range condition.ExcludeEnv {
			if key != "" {
				pipelineResults[key] = true
			}
		}
	}

	for _, key := range pipeline.Setup.GetEnv() {
		remainingResults[key] = true
	}

	for _, step := range pipeline.Steps {
		for key, condition := range step.When {
			if key == "" {
				continue
			}

			if strings.HasPrefix(key, schema.ENV_PREFIX) {
				envKey := strings.TrimPrefix(key, schema.ENV_PREFIX)
				if envKey != "" {
					remainingResults[envKey] = true
				}
			}

			for _, key := range condition.IncludeEnv {
				if key != "" {
					remainingResults[key] = true
				}
			}
			for _, key := range condition.ExcludeEnv {
				if key != "" {
					remainingResults[key] = true
				}
			}
		}

		for _, key := range step.GetEnv() {
			remainingResults[key] = true
		}
	}

	for key := range pipelineResults {
		delete(remainingResults, key)
	}

	result.Env = make([]string, 0, len(pipelineResults)+len(remainingResults))
	result.PipelineEnv = make([]string, 0, len(pipelineResults))
	result.RemainingEnv = make([]string, 0, len(remainingResults))
	for key := range pipelineResults {
		result.Env = append(result.Env, key)
		result.PipelineEnv = append(result.PipelineEnv, key)
	}
	for key := range remainingResults {
		result.Env = append(result.Env, key)
		result.RemainingEnv = append(result.RemainingEnv, key)
	}

	return
}

func MergeEnv(keys []string, envs ...map[string]schema.Env) (result map[string]schema.Env, err error) {
	result = make(map[string]schema.Env, len(keys))

	for _, env := range envs {
		if len(env) > 0 {
			for _, key := range keys {
				if key != "" {
					existing, existingOk := result[key]
					value, ok := env[key]
					if ok && (!existingOk || value.Priority < existing.Priority) {
						result[key] = value
					}
				}
			}
		}
	}

	missing := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ok := result[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		err = fmt.Errorf("missing env %s", strings.Join(missing, ", "))
		return
	}

	return
}
