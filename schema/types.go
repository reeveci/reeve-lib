package schema

const ENV_PREFIX = "env "
const VAR_PREFIX = "var "

const DEFAULT_WORKER_GROUP = "default"
const DEFAULT_STAGE = "default"

type PipelineDefinition struct {
	Name        string               `json:"name" yaml:"name"`
	Headline    string               `json:"headline" yaml:"headline"`
	Description string               `json:"description" yaml:"description"`
	When        map[string]Condition `json:"when" yaml:"when"`
	Steps       []Step               `json:"steps" yaml:"steps"`
}

type Pipeline struct {
	PipelineDefinition `yaml:",inline"`

	Env            map[string]Env    `json:"env" yaml:"env"`
	Facts          map[string]Fact   `json:"facts" yaml:"facts"`
	TaskDomains    map[string]string `json:"taskDomains" yaml:"taskDomains"`
	TrustedDomains []string          `json:"trustedDomains" yaml:"trustedDomains"`
	TrustedTasks   []string          `json:"trustedTasks" yaml:"trustedTasks"`

	Setup Setup `json:"setup" yaml:"setup"`
}

type Env struct {
	Value    string `json:"value" yaml:"value"`
	Priority uint32 `json:"priority" yaml:"priority"`
	Secret   bool   `json:"secret" yaml:"secret"`
}

type Var string

type Setup struct {
	RunConfig `yaml:",inline"`
}

type Step struct {
	RunConfig `yaml:",inline"`

	Name          string               `json:"name" yaml:"name"`
	Stage         string               `json:"stage" yaml:"stage"`
	When          map[string]Condition `json:"when" yaml:"when"`
	IgnoreFailure bool                 `json:"ignoreFailure" yaml:"ignoreFailure"`
}

type RawParam interface{}
type LiteralParam string
type EnvParam struct {
	Env     string   `json:"env" yaml:"env"`
	Replace []string `json:"replace" yaml:"replace"`
}
type VarParam struct {
	Var     string   `json:"var" yaml:"var"`
	Replace []string `json:"replace" yaml:"replace"`
}

type Fact []string
