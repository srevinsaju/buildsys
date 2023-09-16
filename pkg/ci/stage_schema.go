package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"os/exec"
)

const StageBlock = "stage"

type StageContainerVolume struct {
	Source      hcl.Expression `hcl:"source" json:"source"`
	Destination hcl.Expression `hcl:"destination" json:"destination"`
}

type StageContainerVolumes []StageContainerVolume

type StageContainerPort struct {
	Hostname      hcl.Expression `hcl:"host,optional" json:"host"`
	ContainerPort hcl.Expression `hcl:"container_port" json:"container_port"`
	Port          hcl.Expression `hcl:"port" json:"port"`
}

type StageContainerPorts []StageContainerPort

type StageContainer struct {
	Image      hcl.Expression        `hcl:"image" json:"image"`
	Volumes    StageContainerVolumes `hcl:"volume,block" json:"volumes"`
	Ports      StageContainerPorts   `hcl:"ports,optional" json:"ports"`
	Host       hcl.Expression        `hcl:"host,optional" json:"host"`
	Entrypoint hcl.Expression        `hcl:"entrypoint,optional" json:"entrypoint"`
	Stdin      bool                  `hcl:"stdin,optional" json:"stdin"`
}

type Stages []Stage

type StageEnvironment struct {
	Name  string         `hcl:"name" json:"name"`
	Value hcl.Expression `hcl:"value" json:"value"`
}

type StageRetry struct {
	Enabled            bool `hcl:"enabled" json:"enabled"`
	Attempts           int  `hcl:"attempts" json:"attempts"`
	ExponentialBackoff bool `hcl:"exponential_backoff" json:"exponential_backoff"`
	MinBackoff         int  `hcl:"min_backoff" json:"min_backoff"`
	MaxBackoff         int  `hcl:"max_backoff" json:"max_backoff"`
}

type StageUse struct {
	Macro      hcl.Expression `hcl:"macro" json:"macro"`
	Parameters hcl.Expression `hcl:"parameters,optional" json:"parameters"`
}
type StageDaemon struct {
	Enabled bool `hcl:"enabled" json:"enabled"`
	Timeout int  `hcl:"timeout,optional" json:"timeout"`

	Lifecycle *Lifecycle `hcl:"lifecycle,block" json:"lifecycle"`
}

type StagePostHook struct {
	Stage CoreStage `hcl:"stage,block" json:"stage"`
}

type StagePreHook struct {
	Stage CoreStage `hcl:"stage,block" json:"stage"`
}

type Stage struct {
	Id        string `hcl:"id,label" json:"id"`
	CoreStage `hcl:",remain"`
}

type CoreStage struct {
	ctx            context.Context
	ctxInitialised bool
	terminated     bool

	DependsOn hcl.Expression `hcl:"depends_on,optional" json:"depends_on"`

	Condition hcl.Expression `hcl:"if,optional" json:"if"`
	ForEach   hcl.Expression `hcl:"for_each,optional" json:"for_each"`
	Use       *StageUse      `hcl:"use,block" json:"use"`

	Daemon *StageDaemon `hcl:"daemon,block" json:"daemon"`
	Retry  *StageRetry  `hcl:"retry,block" json:"retry"`

	Name        string              `hcl:"name,optional" json:"name"`
	Dir         hcl.Expression      `hcl:"dir,optional" json:"dir"`
	Script      hcl.Expression      `hcl:"script,optional" json:"script"`
	Shell       hcl.Expression      `hcl:"shell,optional" json:"shell"`
	Args        hcl.Expression      `hcl:"args,optional" json:"args"`
	Container   *StageContainer     `hcl:"container,block" json:"container"`
	Environment []*StageEnvironment `hcl:"env,block" json:"environment"`

	PreHook  []*StagePreHook  `hcl:"pre_hook,block" json:"pre_hook"`
	PostHook []*StagePostHook `hcl:"post_hook,block" json:"post_hook"`

	process                 *exec.Cmd
	macroWhitelistedStages  []string
	dependsOnVariablesMacro []hcl.Traversal
	ContainerId             string
}

type PreStage struct {
	CoreStage `hcl:",remain"`
}

type PostStage struct {
	CoreStage `hcl:",remain"`
}
