package data

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/zclconf/go-cty/cty"
	"os"
)

type PromptProvider struct {
	initialized bool

	Prompt  hcl.Expression `hcl:"prompt" json:"prompt"`
	Default hcl.Expression `hcl:"default" json:"default"`

	promptParsed string
	def          string
	ctx          context.Context
}

func (e *PromptProvider) Name() string {
	return "prompt"
}

func (e *PromptProvider) SetContext(context context.Context) {
	e.ctx = context
}

func (e *PromptProvider) Version() string {
	return "1"
}

func (e *PromptProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/prompt"
}

func (e *PromptProvider) DecodeBody(body hcl.Body, opts ...ProviderOption) hcl.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags hcl.Diagnostics

	hclContext := global.HclEvalContext()

	schema := e.Schema()
	content, d := body.Content(schema)
	diags = append(diags, d...)

	attr := content.Attributes["prompt"]
	var key cty.Value

	global.EvalContextMutex.RLock()
	key, d = attr.Expr.Value(hclContext)
	global.EvalContextMutex.RUnlock()
	diags = append(diags, d...)

	e.promptParsed = key.AsString()

	attr = content.Attributes["default"]
	global.EvalContextMutex.RLock()
	key, d = attr.Expr.Value(hclContext)
	global.EvalContextMutex.RUnlock()
	diags = append(diags, d...)

	e.def = key.AsString()

	return diags

}

func (e *PromptProvider) New() Provider {
	return &PromptProvider{
		initialized: true,
	}
}

func (e *PromptProvider) Schema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "prompt",
				Required: false,
			},
			{
				Name:     "default",
				Required: false,
			},
		},
	}
}

func (e *PromptProvider) Attributes(ctx context.Context, id string, opts ...ProviderOption) (map[string]cty.Value, hcl.Diagnostics) {
	return map[string]cty.Value{
		"prompt":  cty.StringVal(e.promptParsed),
		"default": cty.StringVal(e.def),
	}, nil
}

func (e *PromptProvider) Initialized() bool {
	return e.initialized
}

func (e *PromptProvider) Logger() *logrus.Entry {
	return global.Logger().WithField("provider", e.Name())
}

func (e *PromptProvider) Value(ctx context.Context, id string, opts ...ProviderOption) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if !e.initialized {
		panic("provider not initialized")
	}

	cfg := NewProviderConfig(opts...)

	logger := e.Logger()

	envVarName := fmt.Sprintf("%s%s__%s", meta.EnvVarPrefix, e.Name(), id)
	logger.Tracef("checking for environment variable %s", envVarName)
	envExists, ok := os.LookupEnv(envVarName)
	if ok {
		logger.Debug("environment variable found, using that")
		return envExists, nil
	}
	if cfg.Behavior.Unattended {
		logger.Warn("--unattended/--ci mode enabled, falling back to default")
		return e.def, nil
	}

	prompt := e.promptParsed
	if prompt == "" {
		prompt = fmt.Sprintf("Enter a value for %s:", e.Name())
	}

	input := survey.Input{
		Renderer: survey.Renderer{},
		Message:  prompt,
		Default:  e.def,
		Help:     "",
		Suggest:  nil,
	}
	var resp string
	err := survey.AskOne(&input, &resp)
	if err != nil {
		logger.Warn("unable to get value from prompt: ", err)
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "unable to get value from prompt",
			Detail:   err.Error(),
		})
		return e.def, diags
	}

	return resp, diags
}
