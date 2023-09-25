package conductor

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	yaml "github.com/zclconf/go-cty-yaml"

	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/third-party/hashicorp/terraform/lang/funcs"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"os"
	"os/exec"
	"time"
)

func CreateEvalContext(cfg Config, process Process) *hcl.EvalContext {
	// --> set up HCL context
	paths := cfg.Paths
	behavior := cfg.Behavior
	hclContext := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"abs":              stdlib.AbsoluteFunc,
			"abspath":          funcs.AbsPathFunc,
			"alltrue":          funcs.AllTrueFunc,
			"anytrue":          funcs.AnyTrueFunc,
			"basename":         funcs.BasenameFunc,
			"base64decode":     funcs.Base64DecodeFunc,
			"base64encode":     funcs.Base64EncodeFunc,
			"base64gzip":       funcs.Base64GzipFunc,
			"base64sha256":     funcs.Base64Sha256Func,
			"base64sha512":     funcs.Base64Sha512Func,
			"bcrypt":           funcs.BcryptFunc,
			"can":              tryfunc.CanFunc,
			"ceil":             stdlib.CeilFunc,
			"chomp":            stdlib.ChompFunc,
			"coalesce":         funcs.CoalesceFunc,
			"coalescelist":     stdlib.CoalesceListFunc,
			"compact":          stdlib.CompactFunc,
			"concat":           stdlib.ConcatFunc,
			"contains":         stdlib.ContainsFunc,
			"csvdecode":        stdlib.CSVDecodeFunc,
			"dirname":          funcs.DirnameFunc,
			"distinct":         stdlib.DistinctFunc,
			"element":          stdlib.ElementFunc,
			"endswith":         funcs.EndsWithFunc,
			"chunklist":        stdlib.ChunklistFunc,
			"file":             funcs.MakeFileFunc(paths.Cwd, false),
			"fileexists":       funcs.MakeFileExistsFunc(paths.Cwd),
			"fileset":          funcs.MakeFileSetFunc(paths.Cwd),
			"filebase64":       funcs.MakeFileFunc(paths.Cwd, true),
			"filebase64sha256": funcs.MakeFileBase64Sha256Func(paths.Cwd),
			"filebase64sha512": funcs.MakeFileBase64Sha512Func(paths.Cwd),
			"filemd5":          funcs.MakeFileMd5Func(paths.Cwd),
			"filesha1":         funcs.MakeFileSha1Func(paths.Cwd),
			"filesha256":       funcs.MakeFileSha256Func(paths.Cwd),
			"filesha512":       funcs.MakeFileSha512Func(paths.Cwd),
			"flatten":          stdlib.FlattenFunc,
			"floor":            stdlib.FloorFunc,
			"format":           stdlib.FormatFunc,
			"formatdate":       stdlib.FormatDateFunc,
			"formatlist":       stdlib.FormatListFunc,
			"indent":           stdlib.IndentFunc,
			"index":            funcs.IndexFunc, // stdlib.IndexFunc is not compatible
			"join":             stdlib.JoinFunc,
			"jsondecode":       stdlib.JSONDecodeFunc,
			"jsonencode":       stdlib.JSONEncodeFunc,
			"keys":             stdlib.KeysFunc,
			"length":           funcs.LengthFunc,
			"list":             funcs.ListFunc,
			"log":              stdlib.LogFunc,
			"lookup":           funcs.LookupFunc,
			"lower":            stdlib.LowerFunc,
			"map":              funcs.MapFunc,
			"matchkeys":        funcs.MatchkeysFunc,
			"max":              stdlib.MaxFunc,
			"md5":              funcs.Md5Func,
			"merge":            stdlib.MergeFunc,
			"min":              stdlib.MinFunc,
			"one":              funcs.OneFunc,
			"parseint":         stdlib.ParseIntFunc,
			"pathexpand":       funcs.PathExpandFunc,
			"pow":              stdlib.PowFunc,
			"range":            stdlib.RangeFunc,
			"regex":            stdlib.RegexFunc,
			"regexall":         stdlib.RegexAllFunc,
			"replace":          funcs.ReplaceFunc,
			"reverse":          stdlib.ReverseListFunc,
			"rsadecrypt":       funcs.RsaDecryptFunc,
			"sensitive":        funcs.SensitiveFunc,
			"nonsensitive":     funcs.NonsensitiveFunc,
			"setintersection":  stdlib.SetIntersectionFunc,
			"setproduct":       stdlib.SetProductFunc,
			"setsubtract":      stdlib.SetSubtractFunc,
			"setunion":         stdlib.SetUnionFunc,
			"sha1":             funcs.Sha1Func,
			"sha256":           funcs.Sha256Func,
			"sha512":           funcs.Sha512Func,
			"signum":           stdlib.SignumFunc,
			"slice":            stdlib.SliceFunc,
			"sort":             stdlib.SortFunc,
			"split":            stdlib.SplitFunc,
			"startswith":       funcs.StartsWithFunc,
			"strcontains":      funcs.StrContainsFunc,
			"strrev":           stdlib.ReverseFunc,
			"substr":           stdlib.SubstrFunc,
			"sum":              funcs.SumFunc,
			"textdecodebase64": funcs.TextDecodeBase64Func,
			"textencodebase64": funcs.TextEncodeBase64Func,
			"timestamp":        funcs.TimestampFunc,
			"timeadd":          stdlib.TimeAddFunc,
			"timecmp":          funcs.TimeCmpFunc,
			"title":            stdlib.TitleFunc,
			"tostring":         funcs.MakeToFunc(cty.String),
			"tonumber":         funcs.MakeToFunc(cty.Number),
			"tobool":           funcs.MakeToFunc(cty.Bool),
			"toset":            funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
			"tolist":           funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
			"tomap":            funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
			"transpose":        funcs.TransposeFunc,
			"trim":             stdlib.TrimFunc,
			"trimprefix":       stdlib.TrimPrefixFunc,
			"trimspace":        stdlib.TrimSpaceFunc,
			"trimsuffix":       stdlib.TrimSuffixFunc,
			"try":              tryfunc.TryFunc,
			"upper":            stdlib.UpperFunc,
			"urlencode":        funcs.URLEncodeFunc,
			"uuid":             funcs.UUIDFunc,
			"uuidv5":           funcs.UUIDV5Func,
			"values":           stdlib.ValuesFunc,
			"which": function.New(&function.Spec{
				Params: []function.Parameter{
					{
						Name:             "executable",
						AllowDynamicType: true,
						Type:             cty.String,
					},
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					path, err := exec.LookPath(args[0].AsString())
					if err != nil {
						return cty.StringVal(""), err
					}
					return cty.StringVal(path), nil
				},
				Description: "Returns the absolute path to an executable in the current PATH.",
			}),
			"yamldecode": yaml.YAMLDecodeFunc,
			"yamlencode": yaml.YAMLEncodeFunc,
			"zipmap":     stdlib.ZipmapFunc,

			"ansifmt": ui.AnsiFunc,
			"env": function.New(&function.Spec{
				Params: []function.Parameter{
					{
						Name:             "Key of the environment variable",
						AllowDynamicType: true,
						Type:             cty.String,
					},
				},
				VarParam: &function.Parameter{
					Name:        "lists",
					Description: "One or more lists of strings to join.",
					Type:        cty.String,
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					v, ok := os.LookupEnv(args[0].AsString())
					if ok {
						return cty.StringVal(v), nil
					}
					def := args[1]
					return def, nil
				},
				Description: "Returns the value of the environment variable, returns the default value if environment variable is empty, else returns empty string.",
			}),
		},

		Variables: map[string]cty.Value{
			"true":  cty.True,
			"false": cty.False,
			"null":  cty.NullVal(cty.DynamicPseudoType),

			"owd":      cty.StringVal(paths.Owd),
			"cwd":      cty.StringVal(paths.Cwd),
			"hostname": cty.StringVal(cfg.Hostname),
			"hostuser": cty.StringVal(cfg.User),

			"pipeline": cty.ObjectVal(map[string]cty.Value{
				"id":      cty.StringVal(process.Id.String()),
				"path":    cty.StringVal(paths.Pipeline),
				"tempDir": cty.StringVal(process.TempDir),
			}),

			"togomak": cty.ObjectVal(map[string]cty.Value{
				"version":        cty.StringVal(meta.AppVersion),
				"boot_time":      cty.StringVal(time.Now().Format(time.RFC3339)),
				"boot_time_unix": cty.NumberIntVal(time.Now().Unix()),
				"pipeline_id":    cty.StringVal(process.Id.String()),
				"ci":             cty.BoolVal(behavior.Ci),
				"unattended":     cty.BoolVal(behavior.Unattended),
			}),

			// introduced in v1.5.0
			"ansi": cty.ObjectVal(map[string]cty.Value{
				"bg": cty.ObjectVal(map[string]cty.Value{
					"red":    cty.StringVal("\033[41m"),
					"green":  cty.StringVal("\033[42m"),
					"yellow": cty.StringVal("\033[43m"),
					"blue":   cty.StringVal("\033[44m"),
					"purple": cty.StringVal("\033[45m"),
					"cyan":   cty.StringVal("\033[46m"),
					"white":  cty.StringVal("\033[47m"),
					"grey":   cty.StringVal("\033[100m"),
				}),
				"fg": cty.ObjectVal(map[string]cty.Value{
					"red":       cty.StringVal("\033[31m"),
					"green":     cty.StringVal("\033[32m"),
					"yellow":    cty.StringVal("\033[33m"),
					"blue":      cty.StringVal("\033[34m"),
					"purple":    cty.StringVal("\033[35m"),
					"cyan":      cty.StringVal("\033[36m"),
					"white":     cty.StringVal("\033[37m"),
					"grey":      cty.StringVal("\033[90m"),
					"bold":      cty.StringVal("\033[1m"),
					"italic":    cty.StringVal("\033[3m"),
					"underline": cty.StringVal("\033[4m"),
				}),

				"reset": cty.StringVal("\033[0m"),
			}),
		},
	}
	global.SetHclEvalContext(hclContext)
	return hclContext
}
