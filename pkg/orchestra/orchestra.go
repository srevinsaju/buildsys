package orchestra

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/conductor"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/graph"
	"github.com/srevinsaju/togomak/v1/pkg/handler"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"sync"
)

func ExpandGlobalParams(t *Togomak, cfg conductor.Config) {
	paramsGo := make(map[string]cty.Value)
	if cfg.Behavior.Child.Enabled {
		m := make(map[string]string)
		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				if strings.HasPrefix(e[:i], ci.TogomakParamEnvVarPrefix) {
					m[e[:i]] = e[i+1:]
				}
			}
		}
		for k, v := range m {
			if ci.TogomakParamEnvVarRegex.MatchString(k) {
				paramsGo[ci.TogomakParamEnvVarRegex.FindStringSubmatch(k)[1]] = cty.StringVal(v)
			}
		}
	}
	global.EvalContextMutex.Lock()
	t.ectx.Variables[ci.ParamBlock] = cty.ObjectVal(paramsGo)
	global.EvalContextMutex.Unlock()
}

func Perform(togomak *conductor.Togomak) int {
	cfg := togomak.Config

	t, ctx := NewContextWithTogomak(cfg)
	ctx, cancel := context.WithCancel(ctx)

	logger := togomak.Logger
	logger.Debugf("starting watchdogs and signal handlers")
	h := StartHandlers(ctx, togomak.DiagWriter)

	defer cancel()
	defer h.WriteDiagnostics()

	// region: external parameters
	ExpandGlobalParams(&t, cfg)
	// endregion

	// --> parse the config file
	// we will now read the pipeline from togomak.hcl
	pipe, hclDiags := ci.Read(ctx, t.parser)
	if hclDiags.HasErrors() {
		logger.Fatal(t.hclDiagWriter.WriteDiagnostics(hclDiags))
	}

	// whitelist all stages if unspecified
	filterList := cfg.Pipeline.Filtered

	// write the pipeline to the temporary directory
	pipelineFilePath := filepath.Join(t.cwd, t.tempDir, meta.ConfigFileName)
	var pipelineData []byte
	for _, f := range t.parser.Files() {
		pipelineData = append(pipelineData, f.Bytes...)
	}

	err := os.WriteFile(pipelineFilePath, pipelineData, 0644)
	if err != nil {
		return h.Fatal()
	}
	var d hcl.Diagnostics

	pipe, d = ExpandImports(ctx, pipe, togomak.Parser)
	h.Diags.Extend(d)
	if h.Diags.HasErrors() {
		return h.Fatal()
	}

	/// we will first expand all local blocks
	logger.Debugf("expanding local blocks")
	locals, d := pipe.Locals.Expand()
	h.Diags.Extend(d)
	if d.HasErrors() {
		return h.Fatal()
	}
	pipe.Local = locals

	// store the pipe in the context
	ctx = context.WithValue(ctx, c.TogomakContextPipeline, pipe)

	// --> validate the pipeline
	// TODO: validate the pipeline

	// --> generate a dependency graph
	// we will now generate a dependency graph from the pipeline
	// this will be used to generate the pipeline
	logger.Debugf("generating dependency graph")
	depGraph, d := graph.TopoSort(ctx, pipe)
	h.Diags.Extend(d)
	if h.Diags.HasErrors() {
		return h.Fatal()
	}

	// endregion: interrupt h

	var diagsMutex sync.Mutex

	logger.Debugf("starting runnables")
	for _, layer := range depGraph.TopoSortedLayers() {
		// we parse the TOGOMAK_ENV file at the beginning of every layer
		// this allows us to have different environments for different layers

		d = ExpandOutputs(t, logger)
		h.Diags.Extend(d)
		if h.Diags.HasErrors() {
			break
		}

		for _, runnableId := range layer {

			runnable, skip, d := pipe.Resolve(runnableId)
			if skip {
				continue
			}
			if d.HasErrors() {
				diagsMutex.Lock()
				h.Diags.Extend(d)
				diagsMutex.Unlock()
				break
			}

			ok, d, overridden := CanRun(runnable, ctx, filterList, runnableId, depGraph)
			diagsMutex.Lock()
			h.Diags.Extend(d)
			diagsMutex.Unlock()
			if d.HasErrors() {
				break
			}

			// prepare step needs to run before the runnable is run
			// we will also need to prompt the user with the information saying that it has been skipped
			d = runnable.Prepare(ctx, !ok, overridden)
			diagsMutex.Lock()
			h.Diags.Extend(d)
			diagsMutex.Unlock()
			if d.HasErrors() {
				break
			}

			if !ok {
				logger.Debugf("skipping runnable %s, condition evaluated to false", runnableId)
				continue
			}

			logger.Debugf("runnable %s is %T", runnableId, runnable)

			if runnable.IsDaemon() {
				h.Tracker.AppendDaemon(runnable)
			} else {
				h.Tracker.AppendRunnable(runnable)
			}

			go RunWithRetries(runnableId, runnable, ctx, h, logger)

			if cfg.Pipeline.DryRun {
				// TODO: implement --concurrency option
				// wait for the runnable to finish
				// disable concurrency
				h.Tracker.RunnableWait()
				h.Tracker.DaemonWait()
			}
		}
		h.Tracker.RunnableWait()

		if h.Diags.HasErrors() {
			if h.Tracker.HasDaemons() && !cfg.Pipeline.DryRun && !cfg.Behavior.Unattended {
				logger.Info("pipeline failed, waiting for daemons to shut down")
				logger.Info("hit Ctrl+C to force stop them")
				// wait for daemons to stop
				h.Tracker.DaemonWait()
			} else if h.Tracker.HasDaemons() && !cfg.Pipeline.DryRun {
				logger.Info("pipeline failed, waiting for daemons to shut down...")
				// wait for daemons to stop
				cancel()
			}
			break
		}
	}

	h.Tracker.DaemonWait()
	if h.Diags.HasErrors() {
		return h.Fatal()
	}
	return h.Ok()
}

func StartHandlers(ctx context.Context, diagWriter hcl.DiagnosticWriter) *handler.Handler {
	h := handler.NewHandler(ctx, diagWriter)
	go h.Interrupt()
	go h.Kill()
	go h.Daemons()
	return h
}
