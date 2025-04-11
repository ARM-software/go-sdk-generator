package generator

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"

	"github.com/Arm-Debug/go-sdk-generator/generator/config"
)

//go:generate go-jsonschema -p generator ../schema.json -o ../generator/types.gen.go --struct-name-from-title --capitalization SDK

const (
	implementationFile     = "templates/implementation.go.tmpl"
	interfaceFile          = "templates/interface.go.tmpl"
	configFile             = "templates/config.go.tmpl"
	cliFile                = "templates/cli.go.tmpl"
	readmeFile             = "templates/readme.md.tmpl"
	rootCliFile            = "templates/root.go.tmpl"
	mainCliFile            = "templates/main.go.tmpl"
	typesFile              = "templates/types.go.tmpl"
	goModSdkFile           = "templates/go.mod.sdk.tmpl"
	goModCliFile           = "templates/go.mod.cli.tmpl"
	codeOfConductFile      = "templates/CODE_OF_CONDUCT.md.tmpl"
	dockerFile             = "templates/Dockerfile.tmpl"
	pyprojectFile          = "templates/pyproject.toml.tmpl"
	modulePropertiesFile   = "templates/module.properties.tmpl"
	knownIssuesFile        = "templates/KNOWN_ISSUES.md.tmpl"
	developmentFile        = "templates/DEVELOPMENT.md.tmpl"
	contribFile            = "templates/CONTRIBUTING.md.tmpl"
	codeOwnersFile         = "templates/CODEOWNERS.tmpl"
	dependabotFile         = "templates/dependabot.yml.tmpl"
	prTemplateFile         = "templates/PULL_REQUEST_TEMPLATE.md.tmpl"
	ciWorkflowFile         = "templates/ci.yml.tmpl"
	dependabotWorkflowFile = "templates/dependabot.workflow.yml.tmpl"
	releaseWorkflowFile    = "templates/release.yml.tmpl"
	goreleaserFile         = "templates/goreleaser.yaml.tmpl"
	goreleaserLibraryFile  = "templates/goreleaser-library.yaml.tmpl"
	golangciFile           = "templates/golangci.yaml.tmpl"
	golangciPrecommitFile  = "templates/golangci.pre-commit.yaml.tmpl"
)

//go:embed templates
var fs embed.FS

type templateToPopulate struct {
	templateFile string
	output       string
}

func (t *templateToPopulate) Populate(ctx context.Context, data SDK) (err error) {
	err = populateTemplate(ctx, data, t.templateFile, t.output)
	return
}

func NewTemplateToPopulate(templateFile, output string) templateToPopulate {
	return templateToPopulate{
		templateFile: templateFile,
		output:       output,
	}
}

type templates []templateToPopulate

func (t *templates) Populate(ctx context.Context, data SDK) (err error) {
	for _, tp := range *t {
		err = tp.Populate(ctx, data)
		if err != nil {
			return
		}
	}
	return
}

func formatImports(ctx context.Context, path string) (bytes []byte, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	src, err := filesystem.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("%w: file '%v' doesn't exist %v", commonerrors.ErrNotFound, path, err.Error())
		return
	}

	cleanedSrc, err := imports.Process(path, src, &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   4,
		Fragment:   true,
		FormatOnly: false,
	})
	if err != nil {
		err = fmt.Errorf("%w: could not cleanup imports for file '%v': %v", commonerrors.ErrUnexpected, path, err.Error())
		return
	}

	bytes = cleanedSrc
	return
}

var rootCmds map[string]any = make(map[string]any)

var templateFuncMap = template.FuncMap{
	"title": func(s string) string {
		if len(s) == 1 {
			return strings.ToUpper(s)
		}
		return fmt.Sprintf("%s%s", strings.ToUpper(string(s[0])), s[1:])
	},
	"replaceUnderscores": func(s string) string {
		return strings.ReplaceAll(s, "_", "-")
	},
	"is_func": func(s any) bool {
		return strings.Contains(fmt.Sprint(s), "func()")
	},
	"lower": strings.ToLower,
	"notSeenRootCmd": func(s string) bool {
		_, ok := rootCmds[s]
		if !ok {
			rootCmds[s] = true
		}
		return !ok
	},
	"resetNotSeenRootCmd": func() string {
		rootCmds = make(map[string]any)
		return ""
	},
	"uuid": func() string {
		uuid, err := idgen.GenerateUUID4()
		for err != nil {
			uuid, err = idgen.GenerateUUID4()
		}
		return uuid
	},
	"zero": func(s string) string {
		switch {
		case s == "string":
			return "\"\""
		case s == "bool":
			return "false"
		case strings.HasPrefix(s, "int"):
			return "0"
		case strings.HasPrefix(s, "float"):
			return "0.0"
		case s == "time.Duration":
			return "0"
		default:
			return "nil"
		}
	},
	"argtype": func(s string) string {
		switch s {
		case "time.Duration":
			return "Duration"
		default:
			return cases.Title(language.English).String(s)
		}
	},
	"isVariadicArg": func(s string) bool {
		return strings.HasPrefix(strings.TrimSpace(s), "...")
	},
}

func populateTemplate(ctx context.Context, data SDK, templateFile string, output string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	tmpl, err := template.
		New("sdk").
		Funcs(templateFuncMap).
		ParseFS(fs, templateFile)
	if err != nil {
		err = fmt.Errorf("%w: could not create template: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	if filesystem.Exists(output) {
		err = filesystem.Rm(output)
		if err != nil {
			return
		}
	}

	outDir := filepath.Dir(output)
	if !filesystem.Exists(outDir) {
		err = filesystem.MkDir(outDir)
		if err != nil {
			return
		}
	}

	outFile, err := filesystem.CreateFile(output)
	if err != nil {
		return
	}
	defer func() {
		if outFile != nil {
			_ = outFile.Close()
		}
	}()

	err = tmpl.ExecuteTemplate(outFile, filepath.Base(templateFile), data)
	if err != nil {
		err = fmt.Errorf("%w: could not execute template: %v", commonerrors.ErrUnexpected, err.Error())
		return
	}

	if filepath.Ext(output) == ".go" {
		cleaned, subErr := formatImports(ctx, output)
		if subErr != nil {
			err = subErr
			return
		}

		err = filesystem.WriteFile(output, cleaned, 0644)
		if err != nil {
			return
		}
	}

	return
}

func Generate(ctx context.Context, cfg *config.GenerateConfig) (err error) {
	if cfg == nil {
		err = fmt.Errorf("%w: generate configuration must be defined", commonerrors.ErrUndefined)
		return
	}

	spec, err := filesystem.ReadFile(cfg.Input)
	if err != nil {
		err = fmt.Errorf("%w: could not read file '%v'", commonerrors.ErrUnexpected, cfg.Input)
		return
	}

	var sdk SDK
	if ext := filepath.Ext(cfg.Input); ext == ".yml" || ext == ".yaml" {
		err = sdk.UnmarshalYAML(spec)
	} else {
		err = sdk.UnmarshalJSON(spec)
	}
	if err != nil {
		err = fmt.Errorf("%w: the file '%v' does not comply with the schema and thus, cannot be parsed: %v", commonerrors.ErrInvalid, cfg.Input, err.Error())
		return
	}

	templatesToPopulate := templates{
		NewTemplateToPopulate(implementationFile, fmt.Sprintf("%v/%v/%v.go", cfg.Output, sdk.Specification.MainModule, sdk.Specification.MainModule)),
		NewTemplateToPopulate(interfaceFile, fmt.Sprintf("%v/%v/interfaces.go", cfg.Output, sdk.Specification.MainModule)),
		NewTemplateToPopulate(configFile, fmt.Sprintf("%v/%v/config/%v.go", cfg.Output, sdk.Specification.MainModule, sdk.Specification.MainModule)),
		NewTemplateToPopulate(readmeFile, fmt.Sprintf("%v/README.md", cfg.Output)),
		NewTemplateToPopulate(contribFile, fmt.Sprintf("%v/CONTRIBUTING.md", cfg.Output)),
		NewTemplateToPopulate(developmentFile, fmt.Sprintf("%v/DEVELOPMENT.md", cfg.Output)),
		NewTemplateToPopulate(codeOfConductFile, fmt.Sprintf("%v/CODE_OF_CONDUCT.md", cfg.Output)),
		NewTemplateToPopulate(dockerFile, fmt.Sprintf("%v/Dockerfile.%v", cfg.Output, sdk.Specification.MainModule)),
		NewTemplateToPopulate(knownIssuesFile, fmt.Sprintf("%v/KNOWN_ISSUES.md", cfg.Output)),
		NewTemplateToPopulate(golangciFile, fmt.Sprintf("%v/.golangci.yaml", cfg.Output)),
		NewTemplateToPopulate(golangciPrecommitFile, fmt.Sprintf("%v/.golangci.pre-commit.yaml", cfg.Output)),
		NewTemplateToPopulate(codeOwnersFile, fmt.Sprintf("%v/.github/CODEOWNERS", cfg.Output)),
		NewTemplateToPopulate(dependabotFile, fmt.Sprintf("%v/.github/dependabot.yml", cfg.Output)),
		NewTemplateToPopulate(prTemplateFile, fmt.Sprintf("%v/.github/PULL_REQUEST_TEMPLATE.md", cfg.Output)),
		NewTemplateToPopulate(ciWorkflowFile, fmt.Sprintf("%v/.github/workflows/ci.yml", cfg.Output)),
		NewTemplateToPopulate(dependabotWorkflowFile, fmt.Sprintf("%v/.github/workflows/dependabot.yml", cfg.Output)),
		NewTemplateToPopulate(releaseWorkflowFile, fmt.Sprintf("%v/.github/workflows/release.yml", cfg.Output)),
	}

	if !cfg.NoRegenGoMod {
		templatesToPopulate = append(templatesToPopulate,
			NewTemplateToPopulate(goModSdkFile, fmt.Sprintf("%v/%v/go.mod", cfg.Output, sdk.Specification.MainModule)),
		)
	}

	if len(sdk.Specification.AdditionalTypes) > 0 {
		templatesToPopulate = append(templatesToPopulate, NewTemplateToPopulate(typesFile, fmt.Sprintf("%v/%v/types/%v.go", cfg.Output, sdk.Specification.MainModule, sdk.Specification.MainModule)))
	}

	if sdk.Specification.ReleaseToArtifactory != nil && *sdk.Specification.ReleaseToArtifactory {
		templatesToPopulate = append(templatesToPopulate, NewTemplateToPopulate(goreleaserFile, fmt.Sprintf("%v/.goreleaser.yml", cfg.Output)))
	} else {
		templatesToPopulate = append(templatesToPopulate, NewTemplateToPopulate(goreleaserLibraryFile, fmt.Sprintf("%v/.goreleaser.yml", cfg.Output)))
	}

	if sdk.Specification.CommandLineCommands != nil {
		templatesToPopulate = append(templatesToPopulate, []templateToPopulate{
			NewTemplateToPopulate(cliFile, fmt.Sprintf("%v/cli/cmd/%v.go", cfg.Output, sdk.Specification.MainModule)),
			NewTemplateToPopulate(rootCliFile, fmt.Sprintf("%v/cli/cmd/root.go", cfg.Output)),
			NewTemplateToPopulate(mainCliFile, fmt.Sprintf("%v/cli/main.go", cfg.Output)),
		}...)

		if !cfg.NoRegenGoMod {
			templatesToPopulate = append(templatesToPopulate,
				NewTemplateToPopulate(goModCliFile, fmt.Sprintf("%v/cli/go.mod", cfg.Output)),
			)
		}
	}

	var modulePropertiesFilePath string
	if sdk.Specification.CommandLineCommands == nil {
		modulePropertiesFilePath = fmt.Sprintf("%v/module.properties", cfg.Output)
	} else {
		modulePropertiesFilePath = fmt.Sprintf("%v/cli/cmd/module.properties", cfg.Output)
	}
	// Don't regenerate files that contain state
	if !filesystem.Exists(modulePropertiesFilePath) {
		templatesToPopulate = append(templatesToPopulate, NewTemplateToPopulate(modulePropertiesFile, modulePropertiesFilePath))
	}

	pyprojectFilePath := fmt.Sprintf("%v/pyproject.toml", cfg.Output)
	if !filesystem.Exists(pyprojectFilePath) {
		templatesToPopulate = append(templatesToPopulate, NewTemplateToPopulate(pyprojectFile, pyprojectFilePath))
	}

	err = templatesToPopulate.Populate(ctx, sdk)
	return
}
