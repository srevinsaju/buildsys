package data

import (
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/setting"
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/zclconf/go-cty/cty"
	"net/url"
	"os"
	"path/filepath"
)

type gitProviderAuthConfig struct {
	username string
	password string

	isSsh         bool
	sshPassword   string
	sshPrivateKey string
}

type gitProviderConfig struct {
	repo        string
	tag         string
	branch      string
	ref         string
	destination string
	commit      string

	depth    int
	caBundle []byte

	auth  gitProviderAuthConfig
	files []string
}

const (
	GitBlockArgumentUrl         = "url"
	GitBlockArgumentTag         = "tag"
	GitBlockArgumentBranch      = "branch"
	GitBlockArgumentRef         = "ref"
	GitBlockArgumentDestination = "destination"
	GitBlockArgumentCommit      = "commit"
	GitBlockArgumentDepth       = "depth"
	GitBlockArgumentCaBundle    = "ca_bundle"
	GitBlockArgumentAuth        = "auth"
	GitBlockArgumentFiles       = "files"

	GitBlockAttrLastTag             = "last_tag"
	GitBlockAttrCommitsSinceLastTag = "commits_since_last_tag"
	GitBlockAttrSha                 = "sha"

	GitBlockAttrIsTag = "is_tag"
	GitBlockAttrRef   = "ref"
	GitBlockAttrFiles = "files"
)

type GitProvider struct {
	initialized bool
	Default     hcl.Expression `hcl:"default" json:"default"`

	ctx context.Context
	cfg gitProviderConfig
}

func (e *GitProvider) Logger() *logrus.Entry {
	return global.Logger().WithField("provider", e.Name())
}

func (e *GitProvider) Name() string {
	return "git"
}

func (e *GitProvider) Identifier() string {
	return "data.git"
}

func (e *GitProvider) SetContext(context context.Context) {
	e.ctx = context
}

func (e *GitProvider) Version() string {
	return "1"
}

func (e *GitProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/git"
}

func (e *GitProvider) DecodeBody(body hcl.Body, opts ...ProviderOption) hcl.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags hcl.Diagnostics
	evalContext := global.HclEvalContext()

	schema := e.Schema()
	content, d := body.Content(schema)
	diags = diags.Extend(d)

	global.EvalContextMutex.RLock()
	repo, d := content.Attributes[GitBlockArgumentUrl].Expr.Value(evalContext)
	global.EvalContextMutex.RUnlock()
	diags = diags.Extend(d)

	tagAttr, ok := content.Attributes[GitBlockArgumentTag]
	tag := cty.StringVal("")
	if ok {
		global.EvalContextMutex.RLock()
		tag, d = tagAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	branchAttr, ok := content.Attributes[GitBlockArgumentBranch]
	branch := cty.StringVal("")
	if ok {
		global.EvalContextMutex.RLock()
		branch, d = branchAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	refAttr, ok := content.Attributes[GitBlockArgumentRef]
	ref := cty.StringVal("")
	if ok {
		global.EvalContextMutex.RLock()
		ref, d = refAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	commitAttr, ok := content.Attributes[GitBlockArgumentCommit]
	commit := cty.StringVal("")
	if ok {
		global.EvalContextMutex.RLock()
		commit, d = commitAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	destinationAttr, ok := content.Attributes[GitBlockArgumentDestination]
	destination := cty.StringVal("")
	if ok {
		global.EvalContextMutex.RLock()
		destination, d = destinationAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	depthAttr, ok := content.Attributes[GitBlockArgumentDepth]
	depth := cty.NumberIntVal(0)
	if ok {
		global.EvalContextMutex.RLock()
		depth, d = depthAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	caBundleAttr, ok := content.Attributes[GitBlockArgumentCaBundle]
	caBundle := cty.StringVal("")
	if ok {
		global.EvalContextMutex.RLock()
		caBundle, d = caBundleAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	filesAttr, ok := content.Attributes[GitBlockArgumentFiles]
	files := cty.ListValEmpty(cty.String)
	if ok {
		global.EvalContextMutex.RLock()
		files, d = filesAttr.Expr.Value(evalContext)
		global.EvalContextMutex.RUnlock()
		diags = diags.Extend(d)
	}

	authBlock := content.Blocks.OfType(GitBlockArgumentAuth)
	var authConfig gitProviderAuthConfig
	if len(authBlock) == 1 {
		auth, d := content.Blocks.OfType(GitBlockArgumentAuth)[0].Body.Content(GitProviderAuthSchema())
		diags = diags.Extend(d)

		global.EvalContextMutex.RLock()
		authUsername, d := auth.Attributes["username"].Expr.Value(evalContext)
		diags = diags.Extend(d)

		authPassword, d := auth.Attributes["password"].Expr.Value(evalContext)
		diags = diags.Extend(d)

		authSshPassword, d := auth.Attributes["ssh_password"].Expr.Value(evalContext)
		diags = diags.Extend(d)

		authSshPrivateKey, d := auth.Attributes["ssh_private_key"].Expr.Value(evalContext)
		diags = diags.Extend(d)
		global.EvalContextMutex.RUnlock()

		authConfig = gitProviderAuthConfig{
			username:      authUsername.AsString(),
			password:      authPassword.AsString(),
			sshPassword:   authSshPassword.AsString(),
			sshPrivateKey: authSshPrivateKey.AsString(),
			isSsh:         authSshPassword.AsString() != "" || authSshPrivateKey.AsString() != "",
		}
	}

	depthInt, _ := depth.AsBigFloat().Int64()
	var f []string
	for _, file := range files.AsValueSlice() {
		f = append(f, file.AsString())
	}

	e.cfg = gitProviderConfig{
		repo:        repo.AsString(),
		tag:         tag.AsString(),
		branch:      branch.AsString(),
		commit:      commit.AsString(),
		ref:         ref.AsString(),
		destination: destination.AsString(),
		depth:       int(depthInt),
		caBundle:    []byte(caBundle.AsString()),
		auth:        authConfig,
		files:       f,
	}

	return diags
}

func init() {
	h, _ := os.UserHomeDir()
	setting.Git.HomePath = h
}

func (e *GitProvider) New() Provider {
	return &GitProvider{
		initialized: true,
	}
}

func GitProviderAuthSchema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "username",
				Required: false,
			},
			{
				Name:     "password",
				Required: false,
			},
			{
				Name:     "ssh_password",
				Required: false,
			},
			{
				Name:     "ssh_private_key",
				Required: false,
			},
		},
	}
}

func (e *GitProvider) Schema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: GitBlockArgumentAuth,
			},
		},
		Attributes: []hcl.AttributeSchema{
			{
				Name:     GitBlockArgumentUrl,
				Required: true,
			},
			{
				Name:     GitBlockArgumentTag,
				Required: false,
			},
			{
				Name:     GitBlockArgumentBranch,
				Required: false,
			},
			{
				Name:     GitBlockArgumentCommit,
				Required: false,
			},
			{
				Name:     GitBlockArgumentDestination,
				Required: false,
			},
			{
				Name:     GitBlockArgumentDepth,
				Required: false,
			},
			{
				Name:     GitBlockArgumentFiles,
				Required: false,
			},
		},
	}
}

func (e *GitProvider) Initialized() bool {
	return e.initialized
}

func (e *GitProvider) Value(ctx context.Context, id string, opts ...ProviderOption) (string, hcl.Diagnostics) {
	if !e.initialized {
		panic("provider not initialized")
	}
	return "", nil
}

func (e *GitProvider) Attributes(ctx context.Context, id string, opts ...ProviderOption) (map[string]cty.Value, hcl.Diagnostics) {
	logger := e.Logger()
	var diags hcl.Diagnostics
	if !e.initialized {
		panic("provider not initialized")
	}

	var attrs = make(map[string]cty.Value)

	ppb := &ui.PassiveProgressBar{
		Logger:  logger,
		Message: fmt.Sprintf("pulling git repo %s", e.Identifier()),
	}
	ppb.Init()

	ref := e.cfg.ref
	if e.cfg.tag != "" {
		ref = e.cfg.tag
	} else {
		ref = e.cfg.branch
	}
	gitOpts := git.CloneRepoOptions{
		Depth:  e.cfg.depth,
		Branch: ref,
		Bare:   false,
		// TODO: SkipTLSVerify: e.cfg.skipTLSVerify,
		// TODO: make it configurable
		Quiet: true,
	}
	// TODO: implement git submodules

	destination, d := e.resolveDestination(ctx, id)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return nil, diags
	}

	repoUrl, err := url.Parse(e.cfg.repo)
	if err != nil {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to parse git repo url",
			Detail:   err.Error(),
		})
	}

	logger.Debugf("cloning git repo to %s", destination)
	err = git.CloneWithArgs(ctx, nil, repoUrl.String(), destination, gitOpts)
	ppb.Done()

	if err != nil {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to clone git repo",
			Detail:   err.Error(),
		})
	}
	repo, closer, err := git.RepositoryFromContextOrOpen(ctx, destination)
	if err != nil {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to open git repo",
			Detail:   err.Error(),
		})
	}

	gitBranch, err := repo.GetHEADBranch()
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "failed to get git branch",
			Detail:   err.Error(),
		})
	}
	branch := ""
	if gitBranch != nil {
		branch = gitBranch.Name
	}

	lastTag := ""
	tags, err := repo.GetTags(0, 1)
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "failed to get git tags",
			Detail:   err.Error(),
		})
	}
	noTags := false
	if len(tags) == 0 {
		noTags = true
	} else if len(tags) > 1 {
		panic("more than 1 tag returned when only one was supposed to be returned")
	}
	commitsSinceLastTag := cty.NullVal(cty.Number)
	if !noTags {
		lastTag = tags[0]

		commitCount, err := repo.CommitsCountBetween(lastTag, "HEAD")
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "failed to get commits since last tag",
				Detail:   err.Error(),
			})
		} else {
			commitsSinceLastTag = cty.NumberIntVal(commitCount)
		}

	}

	sha, err := repo.ConvertToSHA1("HEAD")
	if err != nil {
		return nil, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to get git sha",
			Detail:   err.Error(),
		})
	}

	err = closer.Close()
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "failed to close git repo",
			Detail:   err.Error(),
		})
	}

	ref, err = repo.ResolveReference("HEAD")
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "failed to resolve git reference",
			Detail:   err.Error(),
		})
	}

	_, err = repo.GetTagNameBySHA(sha.String())
	isTagResolved := cty.NullVal(cty.Bool)
	if git.IsErrNotExist(err) {
		isTagResolved = cty.False
	} else if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "failed to resolve git tag",
			Detail:   err.Error(),
		})
	} else {
		isTagResolved = cty.True
	}

	// read files and store them in the map if whitelisted
	files := make(map[string]cty.Value)
	for _, file := range e.cfg.files {
		f := filepath.Join(destination, file)
		if _, err := os.Stat(f); err == nil {
			content, err := os.ReadFile(f)
			if err != nil {
				return nil, diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "failed to read file",
					Detail:   err.Error(),
				})
			}
			files[file] = cty.StringVal(string(content))
		} else if !os.IsNotExist(err) {
			return nil, diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "failed to read file",
				Detail:   err.Error(),
			})
		}
	}

	var filesCty cty.Value
	if len(files) > 0 {
		filesCty = cty.MapVal(files)
	} else {
		filesCty = cty.NullVal(cty.Map(cty.String))
	}

	attrs[GitBlockArgumentBranch] = cty.StringVal(branch)
	attrs[GitBlockArgumentTag] = cty.StringVal(e.cfg.tag)
	attrs[GitBlockAttrIsTag] = isTagResolved
	attrs[GitBlockAttrRef] = cty.StringVal(ref)
	attrs[GitBlockArgumentUrl] = cty.StringVal(e.cfg.repo)
	attrs[GitBlockAttrSha] = cty.StringVal(sha.String())
	attrs[GitBlockAttrLastTag] = cty.StringVal(lastTag)
	attrs[GitBlockAttrCommitsSinceLastTag] = commitsSinceLastTag
	attrs[GitBlockAttrFiles] = filesCty
	attrs[GitBlockArgumentDestination] = cty.StringVal(destination)

	// get the commit
	return attrs, diags
}

func (e *GitProvider) resolveDestination(ctx context.Context, id string) (string, hcl.Diagnostics) {
	logger := e.Logger()
	tmpDir := global.TempDir()

	var diags hcl.Diagnostics
	destination := e.cfg.destination
	if destination == "" || destination == "memory" {
		if e.cfg.destination == "memory" {
			// we deprecate this mode, warn the users
			logger.Warn("git provider destination is set to memory, this mode is deprecated, currently it writes to a temporary directory")
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "deprecated git destination",
				Detail:   "git provider destination is set to memory, this mode is deprecated, currently it writes to a temporary directory",
			})
		}
		destination = filepath.Join(tmpDir, e.Identifier(), id)
	}
	return destination, diags
}
