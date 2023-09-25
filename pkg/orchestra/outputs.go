package orchestra

import (
	"github.com/hashicorp/go-envparse"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/conductor"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
)

func ExpandOutputs(togomak *conductor.Togomak) hcl.Diagnostics {
	var diags hcl.Diagnostics
	logger := togomak.Logger
	togomakEnvFile := filepath.Join(togomak.Process.TempDir, meta.OutputEnvFile)
	logger.Tracef("%s will be stored and exported here: %s", meta.OutputEnvVar, togomakEnvFile)
	envFile, err := os.OpenFile(togomakEnvFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err == nil {
		e, err := envparse.Parse(envFile)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "could not parse TOGOMAK_ENV file",
				Detail:   err.Error(),
			})
			return diags
		}
		x.Must(envFile.Close())
		ee := make(map[string]cty.Value)
		for k, v := range e {
			ee[k] = cty.StringVal(v)
		}
		global.EvalContextMutex.Lock()
		togomak.EvalContext.Variables[ci.OutputBlock] = cty.ObjectVal(ee)
		global.EvalContextMutex.Unlock()
	} else {
		logger.Warnf("could not open %s file, ignoring... :%s", meta.OutputEnvVar, err)
	}
	return diags
}
