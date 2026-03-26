package generator

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/semver"
	"golang.org/x/tools/imports"

	"github.com/Arm-Debug/go-sdk-generator/generator/config"
)

//go:generate go tool go-jsonschema -p generator ../schema.yml -o ../generator/types.gen.go --struct-name-from-title --capitalization SDK

const (
	agentFile                    = "templates/agent.md.tmpl"
	cliSkillFile                 = "templates/cli.skill.md.tmpl"
	installSkillFile             = "templates/install.skill.md.tmpl"
	precommitConfigFile          = "templates/preconfig.yaml.tmpl"
	editorConfigFile             = "templates/editorconfig.tmpl"
	implementationFile           = "templates/implementation.go.tmpl"
	implementationCLIFile        = "templates/implementation.cli.go.tmpl"
	implementationWrapperCLIFile = "templates/wrapper.cli.go.tmpl"
	interfaceFile                = "templates/interface.go.tmpl"
	configCLIFile                = "templates/config.cli.go.tmpl"
	configFile                   = "templates/config.go.tmpl"
	cliFile                      = "templates/cli.go.tmpl"
	readmeFile                   = "templates/readme.md.tmpl"
	documentationFile            = "templates/documentation.md.tmpl"
	rootCliFile                  = "templates/root.go.tmpl"
	mainCliFile                  = "templates/main.go.tmpl"
	typesFile                    = "templates/types.go.tmpl"
	goModSdkFile                 = "templates/go.mod.sdk.tmpl"
	goModCliFile                 = "templates/go.mod.cli.tmpl"
	codeOfConductFile            = "templates/CODE_OF_CONDUCT.md.tmpl"
	dockerFile                   = "templates/Dockerfile.tmpl"
	pyprojectFile                = "templates/pyproject.toml.tmpl"
	modulePropertiesFile         = "templates/module.properties.tmpl"
	knownIssuesFile              = "templates/KNOWN_ISSUES.md.tmpl"
	developmentFile              = "templates/DEVELOPMENT.md.tmpl"
	contribFile                  = "templates/CONTRIBUTING.md.tmpl"
	codeOwnersFile               = "templates/CODEOWNERS.tmpl"
	dependabotFile               = "templates/dependabot.yml.tmpl"
	prTemplateFile               = "templates/PULL_REQUEST_TEMPLATE.md.tmpl"
	qltyTemplateFile             = "templates/qlty.toml.tmpl"
	ciWorkflowFile               = "templates/ci.yml.tmpl"
	dependabotWorkflowFile       = "templates/dependabot.workflow.yml.tmpl"
	releaseWorkflowFile          = "templates/release.yml.tmpl"
	goreleaserFile               = "templates/goreleaser.yaml.tmpl"
	goreleaserLibraryFile        = "templates/goreleaser-library.yaml.tmpl"
	golangciFile                 = "templates/golangci.yaml.tmpl"
	golangciPrecommitFile        = "templates/golangci.pre-commit.yaml.tmpl"
	preinstallFile               = "templates/pre.sh.tmpl"
	gitignoreFile                = "templates/gitignore.tmpl"
	gitAttributesFile            = "templates/gitattributes.tmpl"
)

//go:embed templates
var fs embed.FS

type templateToPopulate struct {
	templateFile string
	output       string
}

func (t *templateToPopulate) Populate(ctx context.Context, data *DataModel) (err error) {
	err = populateTemplate(ctx, data, t.templateFile, t.output)
	return
}

func NewTemplateToPopulate(templateFile, output string) templateToPopulate {
	return templateToPopulate{
		templateFile: templateFile,
		output:       output,
	}
}

type templates struct {
	toGenerate []templateToPopulate
	toTrack    []string
}

func (t *templates) Populate(ctx context.Context, data *DataModel) (files []string, err error) {
	for _, tp := range t.toGenerate {
		err = tp.Populate(ctx, data)
		if err != nil {
			return
		}
	}

	files = t.toTrack
	return
}

func formatImports(ctx context.Context, path string) (bytes []byte, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	src, err := filesystem.ReadFile(path)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrNotFound, err, "file '%v' doesn't exist", path)
		return
	}

	// FIXME It would be great to also simplify code as part of the process (gofmt -s). However, this is not possible in current go implementation https://github.com/golang/go/issues/21476
	cleanedSrc, err := imports.Process(path, src, &imports.Options{
		Comments:   true,
		TabIndent:  true,
		TabWidth:   4,
		Fragment:   true,
		FormatOnly: false,
	})
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not cleanup imports for file '%v'", path)
		return
	}

	bytes = cleanedSrc
	return
}

type DataModel struct {
	SchemaFileName string
	SDK
	TargetedGoVersion     string
	ToolVersion           string
	FlagRegistrations     []FlagRegistration
	CliEnvVarBindings     []ArgFlagBindingInfo
	CliArgConsts          []ArgDefinition
	RootArgConsts         []ArgDefinition
	RootFlagRegistrations []FlagRegistration
	RootEnvVarBindings    []ArgFlagBindingInfo
}

type GenerationOutput struct {
	ToolVersion    string
	GeneratedFiles []string
}

func populateTemplate[T DataModel | GenerationOutput](ctx context.Context, data *T, templateFile string, output string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	if data == nil {
		err = commonerrors.UndefinedVariable("data model")
		return
	}

	tmpl, err := template.
		New("sdk").
		Funcs(templateFuncMap).
		ParseFS(fs, templateFile)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not read template")
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
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not execute template")
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

func Generate(ctx context.Context, cfg *config.SdkGeneratorManagerConfiguration) (err error) {
	if cfg == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}
	if reflection.IsEmpty(cfg.Input) {
		err = commonerrors.UndefinedVariable("SDK specification file")
		return
	}
	ext := filepath.Ext(cfg.Input)
	ext = strings.ToLower(ext)
	_, found := collection.FindInSlice(false, []string{".yml", ".yaml", ".json"}, ext)
	if !found {
		err = commonerrors.Newf(commonerrors.ErrUnsupported, "file [%v]'s format [%v] is not supported by the generator", cfg.Input, ext)
		return
	}

	if !filesystem.Exists(cfg.Input) {
		if strings.EqualFold(ext, ".json") {
			err = commonerrors.Newf(commonerrors.ErrNotFound, "could not read file '%v'", cfg.Input)
			return
		}
		if strings.EqualFold(ext, ".yml") {
			ext = ".yaml"
		} else if strings.EqualFold(ext, ".yaml") {
			ext = ".yml"
		}
		alternativeInput := filepath.Join(filepath.Dir(cfg.Input), fmt.Sprintf("%v%v", filesystem.FilepathStem(cfg.Input), ext))
		if filesystem.Exists(alternativeInput) {
			cfg.Input = alternativeInput
		} else {
			err = commonerrors.Newf(commonerrors.ErrNotFound, "could not read file '%v'", cfg.Input)
			return
		}
	}

	spec, err := filesystem.ReadFile(cfg.Input)
	if err != nil {
		err = commonerrors.Newf(commonerrors.ErrUnexpected, "could not read file '%v'", cfg.Input)
		return
	}

	var sdk SDK
	if ext == ".yml" || ext == ".yaml" {
		err = sdk.UnmarshalYAML(spec)
	} else {
		err = sdk.UnmarshalJSON(spec)
	}
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "the file '%v' does not comply with the schema and thus, cannot be parsed", cfg.Input)
		return
	}

	err = ProcessBindings(ctx, &sdk)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed preprocessing the SDK definition for bindings")
		return
	}

	err = ProcessArguments(ctx, &sdk)
	if err != nil {
		return
	}

	requestedGoVersion, err := semver.Canonical(cfg.TargetedGoVersion)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid version for 'go'")
		return
	}
	defaultGoVersion, err := semver.Canonical(config.DefaultSdkGeneratorManagerConfiguration().TargetedGoVersion)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid default version for 'go'")
		return
	}

	var templatesToPopulate templates
	templatesToPopulate.addStateFile(implementationFile, fmt.Sprintf("%v/%v/%v.go", cfg.Output, sdk.Specification.MainModule, sdk.Specification.MainModule), true, true)
	templatesToPopulate.addStateFile(interfaceFile, fmt.Sprintf("%v/%v/interfaces.go", cfg.Output, sdk.Specification.MainModule), true, true)
	templatesToPopulate.addStateFile(configFile, fmt.Sprintf("%v/%v/config/%v.go", cfg.Output, sdk.Specification.MainModule, sdk.Specification.MainModule), true, true)
	templatesToPopulate.addStateFile(documentationFile, fmt.Sprintf("%v/DOCUMENTATION.md", cfg.Output), true, true)
	templatesToPopulate.addStateFile(contribFile, fmt.Sprintf("%v/CONTRIBUTING.md", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(developmentFile, fmt.Sprintf("%v/DEVELOPMENT.md", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(codeOfConductFile, fmt.Sprintf("%v/CODE_OF_CONDUCT.md", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(knownIssuesFile, fmt.Sprintf("%v/KNOWN_ISSUES.md", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(golangciFile, fmt.Sprintf("%v/.golangci.yaml", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(editorConfigFile, fmt.Sprintf("%v/.editorconfig", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(precommitConfigFile, fmt.Sprintf("%v/.pre-commit-config.yaml", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(golangciPrecommitFile, fmt.Sprintf("%v/.golangci.pre-commit.yaml", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(dependabotFile, fmt.Sprintf("%v/.github/dependabot.yml", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(prTemplateFile, fmt.Sprintf("%v/.github/PULL_REQUEST_TEMPLATE.md", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(gitignoreFile, fmt.Sprintf("%v/.gitignore", cfg.Output), cfg.Overwrite, false)
	templatesToPopulate.addStateFile(qltyTemplateFile, fmt.Sprintf("%v/.qlty/qlty.toml", cfg.Output), cfg.Overwrite, true) // See https://docs.qlty.sh/qlty-toml

	templatesToPopulate.addStateFile(codeOwnersFile, fmt.Sprintf("%v/.github/CODEOWNERS", cfg.Output), cfg.Overwrite, true)
	templatesToPopulate.addStateFile(readmeFile, fmt.Sprintf("%v/README.md", cfg.Output), cfg.Overwrite, false)
	templatesToPopulate.addStateFile(goModSdkFile, fmt.Sprintf("%v/%v/go.mod", cfg.Output, sdk.Specification.MainModule), cfg.Overwrite, true)

	updateGoVersion := cfg.Overwrite || requestedGoVersion != defaultGoVersion
	templatesToPopulate.addStateFile(dockerFile, fmt.Sprintf("%v/Dockerfile.%v", cfg.Output, sdk.Specification.MainModule), updateGoVersion, true)
	templatesToPopulate.addStateFile(ciWorkflowFile, fmt.Sprintf("%v/.github/workflows/ci.yml", cfg.Output), updateGoVersion, true)
	templatesToPopulate.addStateFile(dependabotWorkflowFile, fmt.Sprintf("%v/.github/workflows/dependabot.yml", cfg.Output), updateGoVersion, true)
	templatesToPopulate.addStateFile(releaseWorkflowFile, fmt.Sprintf("%v/.github/workflows/release.yml", cfg.Output), updateGoVersion, true)

	if len(sdk.Specification.AdditionalTypes) > 0 {
		templatesToPopulate.addStateFile(typesFile, fmt.Sprintf("%v/%v/types/%v.go", cfg.Output, sdk.Specification.MainModule, sdk.Specification.MainModule), true, true)
	}

	if sdk.Specification.ReleaseToArtifactory != nil && *sdk.Specification.ReleaseToArtifactory {
		templatesToPopulate.addStateFile(goreleaserFile, fmt.Sprintf("%v/.goreleaser.yml", cfg.Output), cfg.Overwrite, true)
	} else {
		templatesToPopulate.addStateFile(goreleaserLibraryFile, fmt.Sprintf("%v/.goreleaser.yml", cfg.Output), cfg.Overwrite, true)
	}

	if sdk.Specification.CommandLineCommands != nil {
		templatesToPopulate.addStateFile(agentFile, fmt.Sprintf("%v/agent.md", cfg.Output), true, true)
		templatesToPopulate.addStateFile(cliSkillFile, fmt.Sprintf("%v/.agents/skills/%v/SKILL.md", cfg.Output, strings.ToLower(sdk.Specification.MainModule)), true, true)
		templatesToPopulate.addStateFile(installSkillFile, fmt.Sprintf("%v/.agents/skills/install/SKILL.md", cfg.Output), cfg.Overwrite, true)
		templatesToPopulate.addStateFile(cliFile, fmt.Sprintf("%v/cli/cmd/%v.go", cfg.Output, sdk.Specification.MainModule), true, true)
		templatesToPopulate.addStateFile(rootCliFile, fmt.Sprintf("%v/cli/cmd/root.go", cfg.Output), true, true)
		templatesToPopulate.addStateFile(mainCliFile, fmt.Sprintf("%v/cli/main.go", cfg.Output), true, true)
		templatesToPopulate.addStateFile(configCLIFile, fmt.Sprintf("%v/cli/cmd/config/%v.go", cfg.Output, sdk.Specification.MainModule), true, true)
		templatesToPopulate.addStateFile(implementationWrapperCLIFile, fmt.Sprintf("%v/cli/cmd/%v/wrappers.go", cfg.Output, sdk.Specification.MainModule), true, true)
		templatesToPopulate.addStateFile(preinstallFile, fmt.Sprintf("%v/scripts/pre.sh", cfg.Output), cfg.Overwrite, true)

		templatesToPopulate.addStateFile(implementationCLIFile, fmt.Sprintf("%v/cli/cmd/%v/implementations.go", cfg.Output, sdk.Specification.MainModule), cfg.Overwrite, false)
		templatesToPopulate.addStateFile(goModCliFile, fmt.Sprintf("%v/cli/go.mod", cfg.Output), cfg.Overwrite, true)
	}

	var modulePropertiesFilePath string
	if sdk.Specification.CommandLineCommands == nil {
		modulePropertiesFilePath = fmt.Sprintf("%v/module.properties", cfg.Output)
	} else {
		modulePropertiesFilePath = fmt.Sprintf("%v/cli/cmd/module.properties", cfg.Output)
	}
	templatesToPopulate.addStateFile(modulePropertiesFile, modulePropertiesFilePath, cfg.Overwrite, true)
	templatesToPopulate.addStateFile(pyprojectFile, fmt.Sprintf("%v/pyproject.toml", cfg.Output), cfg.Overwrite, true)

	flagRegs, cliEnvVarBindings, err := PreComputeFlagsAndBindings(sdk.Specification.CommandLineCommands)
	if err != nil {
		return
	}

	argConsts, err := PrecomputeArgNames(sdk.Specification.CommandLineCommands)
	if err != nil {
		return
	}

	rootArgConsts, err := PrecomputeRootArgNames(sdk.Specification.CommandLineCommands)
	if err != nil {
		return
	}
	rootFlagRegs, rootEnvVarBindings, err := PreComputeRootFlagsAndBindings(sdk.Specification.CommandLineCommands)
	if err != nil {
		return
	}

	generated, err := templatesToPopulate.Populate(ctx, &DataModel{
		SchemaFileName:        filesystem.FilepathStem(cfg.Input),
		SDK:                   sdk,
		TargetedGoVersion:     cfg.TargetedGoVersion,
		ToolVersion:           field.OptionalString(cfg.ToolVersion, ""),
		CliEnvVarBindings:     cliEnvVarBindings,
		FlagRegistrations:     flagRegs,
		CliArgConsts:          argConsts,
		RootArgConsts:         rootArgConsts,
		RootFlagRegistrations: rootFlagRegs,
		RootEnvVarBindings:    rootEnvVarBindings,
	})
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not populate templates")
		return
	}

	err = generateGitAttributesFile(ctx, cfg, generated)
	return
}

func (t *templates) addStateFile(templateFile, dest string, overwrite bool, markAsGenerated bool) {
	if markAsGenerated {
		t.toTrack = append(t.toTrack, dest)
	}
	if !filesystem.Exists(dest) || overwrite {
		t.toGenerate = append(t.toGenerate,
			NewTemplateToPopulate(templateFile, dest),
		)
	}
}

func generateGitAttributesFile(ctx context.Context, cfg *config.SdkGeneratorManagerConfiguration, generated []string) (err error) {
	generated, err = filesystem.GetGlobalFileSystem().ConvertToRelativePath(cfg.Output, generated...)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not convert generated files to path relative to output directory '%v'", cfg.Output)
		return
	}
	generated = filesystem.FilePathsToSlash(filesystem.GetGlobalFileSystem(), generated...)
	slices.Sort(generated)

	err = populateTemplate(ctx, &GenerationOutput{
		ToolVersion:    field.OptionalString(cfg.ToolVersion, ""),
		GeneratedFiles: generated,
	}, gitAttributesFile, fmt.Sprintf("%v/.gitattributes", cfg.Output))
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create .gitattributes file")
		return
	}

	return
}
