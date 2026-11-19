/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"embed"
	"fmt"
	"path"
	"slices"
	"strings"
	"text/template"

	modsemver "golang.org/x/mod/semver"

	"github.com/ARM-software/go-sdk-generator/generator/config"
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/reflection"
	"github.com/ARM-software/golang-utils/utils/semver"
)

//go:generate go tool go-jsonschema -p generator ../generator/schema/schema.yml -o ../generator/types.gen.go --struct-name-from-title --capitalization SDK

const (
	copilotFile                  = "templates/copilot-instructions.md.tmpl"
	agentFile                    = "templates/agent.md.tmpl"
	cliSkillFile                 = "templates/cli.skill.md.tmpl"
	githubActionSkillFile        = "templates/github-action.skill.md.tmpl"
	installSkillFile             = "templates/install.skill.md.tmpl"
	actionDefinitionFile         = "templates/action.yml.tmpl"
	actionSharedScriptFile       = "templates/install-and-run.sh.tmpl"
	precommitConfigFile          = "templates/preconfig.yaml.tmpl"
	editorConfigFile             = "templates/editorconfig.tmpl"
	implementationFile           = "templates/implementation.go.tmpl"
	implementationCLIFile        = "templates/implementation.cli.go.tmpl"
	implementationWrapperCLIFile = "templates/wrapper.cli.go.tmpl"
	secretUtilCLIFile            = "templates/secretutil.go.tmpl"
	interfaceFile                = "templates/interface.go.tmpl"
	interfaceEnumFile            = "templates/interface.enums.go.tmpl"
	configCLIFile                = "templates/config.cli.go.tmpl"
	configCLITestFile            = "templates/config.cli_test.go.tmpl"
	configCLIEnumFile            = "templates/config.cli.enums.go.tmpl"
	configFile                   = "templates/config.go.tmpl"
	configEnumFile               = "templates/config.enums.go.tmpl"
	cliFile                      = "templates/cli.go.tmpl"
	readmeFile                   = "templates/readme.md.tmpl"
	documentationFile            = "templates/documentation.md.tmpl"
	rootCliFile                  = "templates/root.go.tmpl"
	mainCliFile                  = "templates/main.go.tmpl"
	typesFile                    = "templates/types.go.tmpl"
	typesEnumFile                = "templates/types.enums.go.tmpl"
	optionsFile                  = "templates/options.go.tmpl"
	optionsEnumFile              = "templates/options.enums.go.tmpl"
	goWorkFile                   = "templates/go.work.tmpl"
	goModSdkFile                 = "templates/go.mod.sdk.tmpl"
	goModCliFile                 = "templates/go.mod.cli.tmpl"
	codeOfConductFile            = "templates/CODE_OF_CONDUCT.md.tmpl"
	securityFile                 = "templates/SECURITY.md.tmpl"
	dockerFile                   = "templates/Dockerfile.tmpl"
	ociDockerfileAlpineFile      = "templates/Dockerfile.oci.alpine.tmpl"
	ociDockerfileUbuntuFile      = "templates/Dockerfile.oci.ubuntu.tmpl"
	pyprojectFile                = "templates/pyproject.toml.tmpl"
	modulePropertiesFile         = "templates/module.properties.tmpl"
	knownIssuesFile              = "templates/KNOWN_ISSUES.md.tmpl"
	developmentFile              = "templates/DEVELOPMENT.md.tmpl"
	contribFile                  = "templates/CONTRIBUTING.md.tmpl"
	codeOwnersFile               = "templates/CODEOWNERS.tmpl"
	dependabotFile               = "templates/dependabot.yml.tmpl"
	dependabotPublicFile         = "templates/dependabot.public.yml.tmpl"
	prTemplateFile               = "templates/PULL_REQUEST_TEMPLATE.md.tmpl"
	qltyTemplateFile             = "templates/qlty.toml.tmpl"
	ciWorkflowFile               = "templates/ci.yml.tmpl"
	ciWorkflowPublicFile         = "templates/ci.public.yml.tmpl"
	lintActionsWorkflowFile      = "templates/lint-actions.yml.tmpl"
	upgradeGeneratorWorkflowFile = "templates/upgrade-generator.yml.tmpl"
	dependabotWorkflowFile       = "templates/dependabot.workflow.yml.tmpl"
	dependabotWorkflowPublicFile = "templates/dependabot.public.workflow.yml.tmpl"
	sbomWorkflowPublicFile       = "templates/sbom.public.yml.tmpl"
	licensingWorkflowFile        = "templates/licensing.yml.tmpl"
	releaseWorkflowFile          = "templates/release.yml.tmpl"
	releaseWorkflowPublicFile    = "templates/release.public.yml.tmpl"
	goreleaserFile               = "templates/goreleaser.yaml.tmpl"
	goreleaserLibraryFile        = "templates/goreleaser-library.yaml.tmpl"
	golangciFile                 = "templates/golangci.yaml.tmpl"
	golangciPrecommitFile        = "templates/golangci.pre-commit.yaml.tmpl"
	preinstallFile               = "templates/pre.sh.tmpl"
	gitignoreFile                = "templates/gitignore.tmpl"
	gitAttributesFile            = "templates/gitattributes.tmpl"
	checkovFile                  = "templates/checkov.yaml.tmpl"
	yamlLintFile                 = "templates/yamllint.tmpl"
	overwriteExcludeGroupCI      = "ci"
	overwriteExcludeGroupLinters = "linters"
	overwriteExcludeGroupGoMod   = "gomod"
	overwriteExcludeGroupCLI     = "cli"
	overwriteExcludeGroupReadme  = "readme"
)

var supportedOverwriteExcludeGroups = []string{
	overwriteExcludeGroupCI,
	overwriteExcludeGroupLinters,
	overwriteExcludeGroupGoMod,
	overwriteExcludeGroupCLI,
	overwriteExcludeGroupReadme,
}

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

type DataModel struct {
	SchemaFileName   string
	ManifestFileName string
	SDK
	Distribution                  DistributionInfo
	OCI                           OCIImageDistributionInfo
	AdditionalTypesPackagesByName map[string]string
	SuppressRequiredValidationMap map[string]bool
	CurrentTypePackage            string
	CurrentAdditionalTypes        Types
	CurrentOptionPackage          string
	CurrentOptionSets             OptionSets
	GeneratedActions              []GeneratedAction
	GeneratedAction               *GeneratedAction
	InstallAction                 *GeneratedAction
	CommandActions                []GeneratedAction
	TargetedGoVersion             string
	ToolVersion                   string
	FlagRegistrations             []FlagRegistration
	FlagConstraintRegistrations   []FlagConstraintRegistration
	CliEnvVarBindings             []ArgFlagBindingInfo
	CliArgConsts                  []ArgDefinition
	RootArgConsts                 []ArgDefinition
	RootFlagRegistrations         []FlagRegistration
	RootEnvVarBindings            []ArgFlagBindingInfo
}

type GenerationOutput struct {
	ToolVersion    string
	GeneratedFiles []string
}

type overwriteExclusionMatcher struct {
	outputDir             string
	explicitBasenames     map[string]struct{}
	explicitRelativePaths map[string]struct{}
	groupRelativePaths    map[string]struct{}
}

func normaliseGeneratedOutputRelativePath(outputDir string, filePath string) (relative string, err error) {
	localFs := filesystem.GetGlobalFileSystem()
	relativePaths, err := localFs.ConvertToRelativePath(outputDir, filePath)
	if err != nil {
		return "", commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not convert %q to a path relative to %q", filePath, outputDir)
	}
	if len(relativePaths) == 0 || reflection.IsEmpty(relativePaths[0]) {
		return "", commonerrors.Newf(commonerrors.ErrInvalid, "could not determine path relative to output directory for %q", filePath)
	}
	return strings.TrimPrefix(path.Clean(filesystem.FilePathsToSlash(localFs, relativePaths[0])[0]), "./"), nil
}

func looksLikeAbsolutePath(value string) bool {
	if strings.HasPrefix(value, "/") {
		return true
	}
	if len(value) >= 2 && ((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) && value[1] == ':' {
		return true
	}
	return false
}

func normaliseExplicitOverwriteFileSelector(raw string) (selector string, pathLike bool, err error) {
	trimmed := strings.TrimSpace(raw)
	if reflection.IsEmpty(trimmed) {
		return "", false, nil
	}
	normalised := strings.ReplaceAll(trimmed, "\\", "/")
	pathLike = strings.Contains(normalised, "/")
	if looksLikeAbsolutePath(normalised) {
		return "", false, commonerrors.Newf(commonerrors.ErrInvalid, "exclude_files entry %q must be a filename or output-relative path, not an absolute path", raw)
	}
	cleaned := path.Clean(normalised)
	if cleaned == "." {
		return "", false, commonerrors.Newf(commonerrors.ErrInvalid, "exclude_files entry %q must not resolve to the output directory itself", raw)
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false, commonerrors.Newf(commonerrors.ErrInvalid, "exclude_files entry %q escapes the output directory", raw)
	}
	return strings.TrimPrefix(cleaned, "./"), pathLike, nil
}

func generatedOverwriteExclusionGroupPaths(sdk *SDK, generatedActions []GeneratedAction, onGithub bool) map[string][]string {
	if sdk == nil {
		return nil
	}
	groups := map[string][]string{
		overwriteExcludeGroupCLI: {},
		overwriteExcludeGroupLinters: {
			".editorconfig",
			".golangci.yaml",
			".golangci.pre-commit.yaml",
			".pre-commit-config.yaml",
			".checkov.yaml",
			".yamllint",
			".qlty/qlty.toml",
		},
		overwriteExcludeGroupReadme: {"README.md"},
		overwriteExcludeGroupCI:     {".goreleaser.yml"},
		overwriteExcludeGroupGoMod:  {path.Join(sdk.Specification.MainModule, "go.mod")},
	}
	if onGithub {
		groups[overwriteExcludeGroupCI] = append(groups[overwriteExcludeGroupCI],
			".github/workflows/ci.yml",
			".github/workflows/upgrade-generator.yml",
			".github/workflows/dependabot.yml",
			".github/workflows/release.yml",
		)
		if sdk.Specification.UpdateCopyrightHeaders {
			groups[overwriteExcludeGroupCI] = append(groups[overwriteExcludeGroupCI], ".github/workflows/licensing.yml")
		}
		if sdk.Specification.Public {
			groups[overwriteExcludeGroupCI] = append(groups[overwriteExcludeGroupCI], ".github/workflows/sbom.yml")
		}
	}
	if len(generatedActions) > 0 {
		groups[overwriteExcludeGroupCI] = append(groups[overwriteExcludeGroupCI], ".github/workflows/lint-actions.yml", "actions/_shared/install-and-run.sh")
		for _, action := range generatedActions {
			groups[overwriteExcludeGroupCI] = append(groups[overwriteExcludeGroupCI], path.Join("actions", action.DirectoryName, "action.yml"))
		}
	}
	if sdk.Specification.CommandLineCommands != nil {
		cliDir := sdk.Specification.CommandLineCommands.Package
		groups[overwriteExcludeGroupGoMod] = append(groups[overwriteExcludeGroupGoMod], "go.work", path.Join(cliDir, "go.mod"))
		groups[overwriteExcludeGroupCLI] = []string{
			path.Join(cliDir, "main.go"),
			path.Join(cliDir, "cmd", fmt.Sprintf("%v.go", sdk.Specification.MainModule)),
			path.Join(cliDir, "cmd", "root.go"),
			path.Join(cliDir, "cmd", "config", fmt.Sprintf("%v.go", sdk.Specification.MainModule)),
			path.Join(cliDir, "cmd", sdk.Specification.MainModule, "wrappers.go"),
			path.Join(cliDir, "cmd", sdk.Specification.MainModule, "implementations.go"),
		}
		if sdk.Specification.CommandLineCommands.HasSummaryValidationTests() {
			groups[overwriteExcludeGroupCLI] = append(groups[overwriteExcludeGroupCLI], path.Join(cliDir, "cmd", "config", fmt.Sprintf("%v_test.go", sdk.Specification.MainModule)))
		}
		if hasCLIConfigEnums(sdk.Specification.CommandLineCommands) {
			groups[overwriteExcludeGroupCLI] = append(groups[overwriteExcludeGroupCLI], path.Join(cliDir, "cmd", "config", "enums.go"))
		}
		if hasSecretStringParameters(sdk.Specification.CommandLineCommands) {
			groups[overwriteExcludeGroupCLI] = append(groups[overwriteExcludeGroupCLI], path.Join(cliDir, "internal", "secretutil", "secret.go"))
		}
	}
	return groups
}

func newOverwriteExclusionMatcher(outputDir string, excludeGroups []string, excludeFiles []string, groupPaths map[string][]string) (matcher *overwriteExclusionMatcher, err error) {
	matcher = &overwriteExclusionMatcher{
		outputDir:             outputDir,
		explicitBasenames:     map[string]struct{}{},
		explicitRelativePaths: map[string]struct{}{},
		groupRelativePaths:    map[string]struct{}{},
	}
	for _, rawGroup := range excludeGroups {
		group := strings.ToLower(strings.TrimSpace(rawGroup))
		if reflection.IsEmpty(group) {
			continue
		}
		paths, found := groupPaths[group]
		if !found {
			return nil, commonerrors.Newf(commonerrors.ErrInvalid, "unsupported exclude_groups value %q (expected one of %s)", rawGroup, strings.Join(supportedOverwriteExcludeGroups, ", "))
		}
		for _, candidate := range paths {
			matcher.groupRelativePaths[strings.TrimPrefix(path.Clean(strings.ReplaceAll(candidate, "\\", "/")), "./")] = struct{}{}
		}
	}
	for _, rawFile := range excludeFiles {
		selector, pathLike, subErr := normaliseExplicitOverwriteFileSelector(rawFile)
		if subErr != nil {
			return nil, subErr
		}
		if reflection.IsEmpty(selector) {
			continue
		}
		if pathLike {
			matcher.explicitRelativePaths[selector] = struct{}{}
			continue
		}
		matcher.explicitBasenames[selector] = struct{}{}
	}
	if len(matcher.explicitBasenames) == 0 && len(matcher.explicitRelativePaths) == 0 && len(matcher.groupRelativePaths) == 0 {
		return nil, nil
	}
	return matcher, nil
}

func (matcher *overwriteExclusionMatcher) ShouldExclude(filePath string) bool {
	if matcher == nil {
		return false
	}
	relativePath, err := normaliseGeneratedOutputRelativePath(matcher.outputDir, filePath)
	if err != nil || reflection.IsEmpty(relativePath) || relativePath == ".." || strings.HasPrefix(relativePath, "../") {
		return false
	}
	if _, found := matcher.groupRelativePaths[relativePath]; found {
		return true
	}
	if _, found := matcher.explicitRelativePaths[relativePath]; found {
		return true
	}
	_, found := matcher.explicitBasenames[path.Base(relativePath)]
	return found
}

func shouldGenerateFile(dest string, overwrite bool, overwriteRequested bool, matcher *overwriteExclusionMatcher, logger logs.Loggers) bool {
	localFs := filesystem.GetGlobalFileSystem()
	if !localFs.Exists(dest) {
		return true
	}
	if !overwrite {
		return false
	}
	if overwriteRequested && matcher != nil && matcher.ShouldExclude(dest) {
		if logger != nil {
			logger.Log(fmt.Sprintf("preserving existing %s; it was excluded from overwrite regeneration", dest))
		}
		return false
	}
	return true
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

	localFs := filesystem.GetGlobalFileSystem()

	tmpl, err := template.
		New("sdk").
		Funcs(templateFuncMap).
		ParseFS(fs, templateFile)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not read template")
		return
	}

	if localFs.Exists(output) {
		err = localFs.Rm(output)
		if err != nil {
			return
		}
	}

	outDir := filesystem.FilePathDir(localFs, output)
	if !localFs.Exists(outDir) {
		err = localFs.MkDir(outDir)
		if err != nil {
			return
		}
	}

	outFile, err := localFs.CreateFile(output)
	if err != nil {
		return
	}
	defer func() {
		if outFile != nil {
			_ = outFile.Close()
		}
	}()

	err = tmpl.ExecuteTemplate(outFile, path.Base(templateFile), data)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not execute template")
		return
	}

	if filesystem.FilePathExt(localFs, output) == ".go" {
		localPrefix := ""
		if typed, ok := any(data).(*DataModel); ok {
			localPrefix = importPrefix(&typed.Specification)
		}

		cleaned, subErr := formatImports(ctx, output, localPrefix)
		if subErr != nil {
			err = subErr
			return
		}

		err = localFs.WriteFile(output, cleaned, 0644)
		if err != nil {
			return
		}
	}

	return
}

func Generate(ctx context.Context, logger logs.Loggers, cfg *config.SDKGeneratorManagerConfiguration) (err error) {
	if cfg == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}

	if reflection.IsEmpty(cfg.Input) {
		err = commonerrors.UndefinedVariable("SDK specification file")
		return
	}

	localFs := filesystem.GetGlobalFileSystem()

	ext := filesystem.FilePathExt(localFs, cfg.Input)
	ext = strings.ToLower(ext)
	_, found := collection.FindInSlice(false, []string{".yml", ".yaml", ".json"}, ext)
	if !found {
		err = commonerrors.Newf(commonerrors.ErrUnsupported, "file [%v]'s format [%v] is not supported by the generator", cfg.Input, ext)
		return
	}

	if !localFs.Exists(cfg.Input) {
		alternativeInput := resolveYAMLManifestTwin(localFs, cfg.Input)
		if reflection.IsEmpty(alternativeInput) {
			err = commonerrors.Newf(commonerrors.ErrNotFound, "could not read file '%v'", cfg.Input)
			return
		}
		cfg.Input = alternativeInput
	}

	spec, err := localFs.ReadFile(cfg.Input)
	if err != nil {
		err = commonerrors.Newf(commonerrors.ErrUnexpected, "could not read file '%v'", cfg.Input)
		return
	}

	err = validateProtectedCLIArgumentNamesInManifest(ext, spec)
	if err != nil {
		return
	}

	err = validateManifestContent(ctx, ext, spec)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "the file '%v' does not comply with the schema and thus, cannot be parsed", cfg.Input)
		return
	}

	var sdk SDK
	err = unmarshalManifestContent(ext, spec, &sdk)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrInvalid, err, "the file '%v' does not comply with the schema and thus, cannot be parsed", cfg.Input)
		return
	}

	err = ProcessExtraImports(ctx, &sdk)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed preprocessing the SDK definition for extra imports")
		return
	}

	err = ProcessRepositoryAndModuleMetadata(ctx, &sdk)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed preprocessing the SDK definition for determining repository and module metadata")
		return
	}

	err = ProcessAdditionalTypePackages(ctx, &sdk)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed preprocessing the SDK definition for additional type packages")
		return
	}

	err = ProcessOptionPackages(ctx, &sdk)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed preprocessing the SDK definition for option packages")
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
	err = ProcessConstraintMirrors(ctx, &sdk)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed preprocessing the SDK definition for constraint mirrors")
		return
	}

	requestedGoVersion, err := semver.Canonical(cfg.TargetedGoVersion)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid version for 'go'")
		return
	}
	defaultGoVersion, err := semver.Canonical(config.DefaultSDKGeneratorManagerConfiguration().TargetedGoVersion)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, err, "invalid default version for 'go'")
		return
	}

	onGithub := isHostedOnGitHub(&sdk.Specification)

	var generatedActions []GeneratedAction
	if sdk.Specification.CommandLineCommands != nil && !sdk.Specification.CommandLineCommands.NoActions {
		generatedActions, err = BuildGeneratedActions(&sdk)
		if err != nil {
			return
		}
	}
	generateAction := len(generatedActions) > 0

	matcher, err := newOverwriteExclusionMatcher(cfg.Output, cfg.ExcludeGroups, cfg.ExcludeFiles, generatedOverwriteExclusionGroupPaths(&sdk, generatedActions, onGithub))
	if err != nil {
		return
	}

	effectiveOverwrite := func(dest string, overwrite bool) bool {
		return shouldGenerateFile(dest, overwrite, cfg.Overwrite, matcher, logger)
	}

	var templatesToPopulate templates
	addStateFile := func(templateFile, dest string, overwrite bool, markAsGenerated bool) {
		templatesToPopulate.addStateFile(templateFile, dest, effectiveOverwrite(dest, overwrite), markAsGenerated)
	}
	addProtectedStateFile := func(templateFile, dest string, overwrite bool, markAsGenerated bool) {
		templatesToPopulate.addProtectedStateFile(templateFile, dest, effectiveOverwrite(dest, overwrite), markAsGenerated, logger)
	}

	addStateFile(implementationFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
	addStateFile(interfaceFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "interfaces.go"), true, true)
	if hasInterfaceEnums(&sdk.Specification) {
		addStateFile(interfaceEnumFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "enums.go"), true, true)
	}
	addStateFile(configFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "config", fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
	if hasConfigEnums(&sdk.Specification) {
		addStateFile(configEnumFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "config", "enums.go"), true, true)
	}
	addStateFile(documentationFile, filesystem.FilePathJoin(localFs, cfg.Output, "DOCUMENTATION.md"), true, true)
	addStateFile(contribFile, filesystem.FilePathJoin(localFs, cfg.Output, "CONTRIBUTING.md"), cfg.Overwrite, true)
	addStateFile(developmentFile, filesystem.FilePathJoin(localFs, cfg.Output, "DEVELOPMENT.md"), cfg.Overwrite, true)
	addStateFile(codeOfConductFile, filesystem.FilePathJoin(localFs, cfg.Output, "CODE_OF_CONDUCT.md"), cfg.Overwrite, true)
	addStateFile(securityFile, filesystem.FilePathJoin(localFs, cfg.Output, "SECURITY.md"), cfg.Overwrite, true)
	addStateFile(knownIssuesFile, filesystem.FilePathJoin(localFs, cfg.Output, "KNOWN_ISSUES.md"), cfg.Overwrite, true)
	addStateFile(golangciFile, filesystem.FilePathJoin(localFs, cfg.Output, ".golangci.yaml"), cfg.Overwrite, true)
	addStateFile(editorConfigFile, filesystem.FilePathJoin(localFs, cfg.Output, ".editorconfig"), cfg.Overwrite, true)
	addStateFile(precommitConfigFile, filesystem.FilePathJoin(localFs, cfg.Output, ".pre-commit-config.yaml"), cfg.Overwrite, true)
	addStateFile(golangciPrecommitFile, filesystem.FilePathJoin(localFs, cfg.Output, ".golangci.pre-commit.yaml"), cfg.Overwrite, true)
	addStateFile(gitignoreFile, filesystem.FilePathJoin(localFs, cfg.Output, ".gitignore"), cfg.Overwrite, false)
	addStateFile(readmeFile, filesystem.FilePathJoin(localFs, cfg.Output, "README.md"), cfg.Overwrite, false)
	addStateFile(goModSdkFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "go.mod"), cfg.Overwrite, true)
	addStateFile(checkovFile, filesystem.FilePathJoin(localFs, cfg.Output, ".checkov.yaml"), cfg.Overwrite, false)
	addStateFile(yamlLintFile, filesystem.FilePathJoin(localFs, cfg.Output, ".yamllint"), cfg.Overwrite, false)

	updateGoVersion := shouldRegenerateVersionSensitiveStateFile(cfg.Overwrite, defaultGoVersion, requestedGoVersion)
	addStateFile(dockerFile, filesystem.FilePathJoin(localFs, cfg.Output, fmt.Sprintf("Dockerfile.%v", sdk.Specification.MainModule)), updateGoVersion, true)
	if sdk.Specification.CommandLineCommands != nil && sdk.Specification.DistributeOci != nil {
		addStateFile(ociDockerfileAlpineFile, filesystem.FilePathJoin(localFs, cfg.Output, "Dockerfile.oci.alpine"), updateGoVersion, true)
		addStateFile(ociDockerfileUbuntuFile, filesystem.FilePathJoin(localFs, cfg.Output, "Dockerfile.oci.ubuntu"), updateGoVersion, true)
	}

	selectedDependabotFile := dependabotFile
	selectedCIWorkflowFile := ciWorkflowFile
	selectedDependabotWorkflowFile := dependabotWorkflowFile
	selectedReleaseWorkflowFile := releaseWorkflowFile
	if sdk.Specification.Public {
		selectedDependabotFile = dependabotPublicFile
		selectedCIWorkflowFile = ciWorkflowPublicFile
		selectedDependabotWorkflowFile = dependabotWorkflowPublicFile
		selectedReleaseWorkflowFile = releaseWorkflowPublicFile
	}

	if onGithub {
		addStateFile(selectedDependabotFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "dependabot.yml"), cfg.Overwrite, true)
		addStateFile(prTemplateFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "PULL_REQUEST_TEMPLATE.md"), cfg.Overwrite, true)
		if !sdk.Specification.Public {
			addStateFile(codeOwnersFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "CODEOWNERS"), cfg.Overwrite, true)
		}
		addStateFile(selectedCIWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "ci.yml"), cfg.Overwrite, true)
		addStateFile(upgradeGeneratorWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "upgrade-generator.yml"), cfg.Overwrite, true)
		addStateFile(selectedDependabotWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "dependabot.yml"), cfg.Overwrite, true)
		addStateFile(selectedReleaseWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "release.yml"), cfg.Overwrite, true)
		if sdk.Specification.UpdateCopyrightHeaders {
			addStateFile(licensingWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "licensing.yml"), cfg.Overwrite, true)
		}
		if sdk.Specification.Public {
			addStateFile(sbomWorkflowPublicFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "sbom.yml"), cfg.Overwrite, true)
		}
	}

	if onGithub && generateAction {
		addStateFile(lintActionsWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "lint-actions.yml"), cfg.Overwrite, true)
	}

	if sdk.Specification.CommandLineCommands != nil {
		addStateFile(goreleaserFile, filesystem.FilePathJoin(localFs, cfg.Output, ".goreleaser.yml"), cfg.Overwrite, true)
	} else {
		addStateFile(goreleaserLibraryFile, filesystem.FilePathJoin(localFs, cfg.Output, ".goreleaser.yml"), cfg.Overwrite, true)
	}

	if sdk.Specification.CommandLineCommands != nil {
		cliDir := sdk.Specification.CommandLineCommands.Package
		addStateFile(agentFile, filesystem.FilePathJoin(localFs, cfg.Output, "agent.md"), true, true)
		if onGithub {
			addStateFile(copilotFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "copilot-instructions.md"), true, true)
		}
		binary := field.OptionalString(sdk.Specification.CommandLineCommands.Binary, "")
		addStateFile(cliSkillFile, filesystem.FilePathJoin(localFs, cfg.Output, ".agents", "skills", runSkillName(binary), "SKILL.md"), true, true)
		addStateFile(installSkillFile, filesystem.FilePathJoin(localFs, cfg.Output, ".agents", "skills", installSkillName(binary), "SKILL.md"), cfg.Overwrite, true)
		if generateAction {
			addStateFile(githubActionSkillFile, filesystem.FilePathJoin(localFs, cfg.Output, ".agents", "skills", githubActionSkillName(binary), "SKILL.md"), true, true)
		}
		addStateFile(cliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
		addStateFile(rootCliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "root.go"), true, true)
		addStateFile(mainCliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "main.go"), true, true)
		addStateFile(configCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "config", fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
		if sdk.Specification.CommandLineCommands.HasSummaryValidationTests() {
			addStateFile(configCLITestFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "config", fmt.Sprintf("%v_test.go", sdk.Specification.MainModule)), true, true)
		}
		if hasCLIConfigEnums(sdk.Specification.CommandLineCommands) {
			addStateFile(configCLIEnumFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "config", "enums.go"), true, true)
		}
		addStateFile(implementationWrapperCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", sdk.Specification.MainModule, "wrappers.go"), true, true)
		addStateFile(preinstallFile, filesystem.FilePathJoin(localFs, cfg.Output, "scripts", "pre.sh"), cfg.Overwrite, true)
		if hasSecretStringParameters(sdk.Specification.CommandLineCommands) {
			addStateFile(secretUtilCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "internal", "secretutil", "secret.go"), true, true)
		}

		addStateFile(implementationCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", sdk.Specification.MainModule, "implementations.go"), cfg.Overwrite, false)
		addStateFile(goWorkFile, filesystem.FilePathJoin(localFs, cfg.Output, "go.work"), cfg.Overwrite, true)
		addStateFile(goModCliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "go.mod"), cfg.Overwrite, true)
	}

	var modulePropertiesFilePath string
	if sdk.Specification.CommandLineCommands == nil {
		modulePropertiesFilePath = filesystem.FilePathJoin(localFs, cfg.Output, "module.properties")
	} else {
		modulePropertiesFilePath = filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.CommandLineCommands.Package, "cmd", "module.properties")
	}
	addProtectedStateFile(modulePropertiesFile, modulePropertiesFilePath, cfg.Overwrite, true)
	addStateFile(pyprojectFile, filesystem.FilePathJoin(localFs, cfg.Output, "pyproject.toml"), cfg.Overwrite, true)

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
	flagConstraintRegs, err := PreComputeFlagConstraintRegistrations(sdk.Specification.CommandLineCommands)
	if err != nil {
		return
	}

	distribution := precomputeDistribution(&sdk.Specification)
	oci := precomputeOCIImageDistribution(&sdk.Specification)

	data := &DataModel{
		SchemaFileName:                filesystem.FilePathStemOnFilesystem(localFs, cfg.Input),
		ManifestFileName:              resolveManifestFileName(localFs, cfg.Input),
		SDK:                           sdk,
		Distribution:                  distribution,
		OCI:                           oci,
		AdditionalTypesPackagesByName: additionalTypePackagesByName(sdk.Specification.AdditionalTypes),
		SuppressRequiredValidationMap: sdk.Specification.SuppressedRequiredValidationMap(),
		CurrentAdditionalTypes:        sdk.Specification.AdditionalTypes,
		GeneratedActions:              generatedActions,
		InstallAction:                 nil,
		CommandActions:                nil,
		TargetedGoVersion:             cfg.TargetedGoVersion,
		ToolVersion:                   field.OptionalString(cfg.ToolVersion, ""),
		CliEnvVarBindings:             cliEnvVarBindings,
		FlagRegistrations:             flagRegs,
		FlagConstraintRegistrations:   flagConstraintRegs,
		CliArgConsts:                  argConsts,
		RootArgConsts:                 rootArgConsts,
		RootFlagRegistrations:         rootFlagRegs,
		RootEnvVarBindings:            rootEnvVarBindings,
	}
	data.InstallAction, data.CommandActions = splitGeneratedActions(generatedActions)

	generated, err := templatesToPopulate.Populate(ctx, data)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not populate templates")
		return
	}

	if reflection.IsNotEmpty(sdk.Specification.AdditionalTypes) {
		groupedTypes := groupAdditionalTypesByPackage(sdk.Specification.AdditionalTypes)
		err = collection.ForAll(sortedAdditionalTypePackages(sdk.Specification.AdditionalTypes), func(pkg string) error {
			packageData := *data
			packageData.CurrentTypePackage = pkg
			packageData.CurrentAdditionalTypes = groupedTypes[pkg]

			typeOutput := filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, pkg, fmt.Sprintf("%v.go", sdk.Specification.MainModule))
			generated = append(generated, typeOutput)
			if shouldGenerateFile(typeOutput, true, cfg.Overwrite, matcher, logger) {
				if subErr := populateTemplate(ctx, &packageData, typesFile, typeOutput); subErr != nil {
					return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate types template")
				}
			}

			if groupedTypes[pkg].HasEnum() {
				enumOutput := filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, pkg, "enums.go")
				generated = append(generated, enumOutput)
				if shouldGenerateFile(enumOutput, true, cfg.Overwrite, matcher, logger) {
					if subErr := populateTemplate(ctx, &packageData, typesEnumFile, enumOutput); subErr != nil {
						return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate type enums template")
					}
				}
			}
			return nil
		})
		if err != nil {
			return
		}
	}

	if reflection.IsNotEmpty(sdk.Specification.Options) {
		groupedOptionSets := groupOptionSetsByPackage(sdk.Specification.Options)
		err = collection.ForAll(sortedOptionPackages(sdk.Specification.Options), func(pkg string) error {
			packageData := *data
			packageData.CurrentOptionPackage = pkg
			packageData.CurrentOptionSets = groupedOptionSets[pkg]

			optionOutput := filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, pkg, fmt.Sprintf("%v.go", sdk.Specification.MainModule))
			generated = append(generated, optionOutput)
			if shouldGenerateFile(optionOutput, true, cfg.Overwrite, matcher, logger) {
				if subErr := populateTemplate(ctx, &packageData, optionsFile, optionOutput); subErr != nil {
					return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate options template")
				}
			}

			if groupedOptionSets[pkg].HasEnum() {
				enumOutput := filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, pkg, "enums.go")
				generated = append(generated, enumOutput)
				if shouldGenerateFile(enumOutput, true, cfg.Overwrite, matcher, logger) {
					if subErr := populateTemplate(ctx, &packageData, optionsEnumFile, enumOutput); subErr != nil {
						return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate option enums template")
					}
				}
			}
			return nil
		})
		if err != nil {
			return
		}
	}

	if sdk.Specification.PackageName == "go-sdk-generator" {
		generated, err = copyGoSDKGeneratorSkillsForEmbedding(cfg.Output, generated, cfg.Overwrite, matcher, logger)
		if err != nil {
			return
		}
	}

	for _, action := range generatedActions {
		actionData := *data // copy
		actionData.GeneratedAction = &action

		actionDir := filesystem.FilePathJoin(localFs, cfg.Output, "actions", action.DirectoryName)
		actionOutputs := []templateToPopulate{
			NewTemplateToPopulate(actionDefinitionFile, filesystem.FilePathJoin(localFs, actionDir, "action.yml")),
		}

		for _, actionOutput := range actionOutputs {
			generated = append(generated, actionOutput.output)
			if shouldGenerateFile(actionOutput.output, true, cfg.Overwrite, matcher, logger) {
				if err = actionOutput.Populate(ctx, &actionData); err != nil {
					err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not populate action %q", action.DirectoryName)
					return
				}
			}
		}
	}

	if generateAction {
		sharedActionOutput := NewTemplateToPopulate(actionSharedScriptFile, filesystem.FilePathJoin(localFs, cfg.Output, "actions", "_shared", "install-and-run.sh"))
		generated = append(generated, sharedActionOutput.output)
		if shouldGenerateFile(sharedActionOutput.output, true, cfg.Overwrite, matcher, logger) {
			if err = sharedActionOutput.Populate(ctx, data); err != nil {
				err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not populate shared action runtime")
				return
			}
		}
	}
	qltyPath := filesystem.FilePathJoin(localFs, cfg.Output, ".qlty", "qlty.toml")
	generated = append(generated, qltyPath)
	err = generateQltyFile(ctx, cfg, generated, matcher, logger)
	if err != nil {
		return
	}
	err = generateGitAttributesFile(ctx, cfg, generated, matcher, logger)
	return
}

func relativeGeneratedFiles(outputDir string, generated []string) (relative []string, err error) {
	localFs := filesystem.GetGlobalFileSystem()
	relative, err = localFs.ConvertToRelativePath(outputDir, generated...)
	if err != nil {
		return nil, commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not convert generated files to path relative to output directory '%v'", outputDir)
	}
	relative = filesystem.FilePathsToSlash(localFs, relative...)
	slices.Sort(relative)
	return relative, nil
}

func resolveManifestFileName(localFs filesystem.FS, inputPath string) string {
	fileName := filesystem.FilePathBase(localFs, inputPath)
	if !strings.EqualFold(filesystem.FilePathExt(localFs, inputPath), ".json") {
		return fileName
	}
	// Prefer a sibling YAML manifest for links and references when generation was
	// triggered from JSON, so human-facing docs point at the easier-to-read form.
	if alternativeInput := resolveYAMLManifestTwin(localFs, inputPath); !reflection.IsEmpty(alternativeInput) {
		return filesystem.FilePathBase(localFs, alternativeInput)
	}

	return fileName
}

func compareGoVersions(a, b string) int {
	canonicalA := modsemver.Canonical(fmt.Sprintf("v%s", semver.TrimGoPrefix(a)))
	canonicalB := modsemver.Canonical(fmt.Sprintf("v%s", semver.TrimGoPrefix(b)))
	return modsemver.Compare(canonicalA, canonicalB)
}

func shouldRegenerateVersionSensitiveStateFile(overwrite bool, defaultGoVersion, requestedGoVersion string) bool {
	return overwrite || compareGoVersions(defaultGoVersion, requestedGoVersion) > 0
}

func resolveYAMLManifestTwin(localFs filesystem.FS, inputPath string) string {
	ext := strings.ToLower(filesystem.FilePathExt(localFs, inputPath))
	if ext == ".json" {
		candidates := collection.Filter(collection.Map([]string{".yml", ".yaml"}, func(candidateExt string) string {
			return filesystem.FilePathJoin(localFs, filesystem.FilePathDir(localFs, inputPath), fmt.Sprintf("%v%v", filesystem.FilepathStem(inputPath), candidateExt))
		}), func(candidate string) bool {
			return localFs.Exists(candidate)
		})
		if candidate, found := collection.First(candidates); found {
			return candidate
		}
		return ""
	}

	if ext != ".yml" && ext != ".yaml" {
		return ""
	}

	otherExt := ".yml"
	if ext == ".yml" {
		otherExt = ".yaml"
	}

	candidate := filesystem.FilePathJoin(localFs, filesystem.FilePathDir(localFs, inputPath), fmt.Sprintf("%v%v", filesystem.FilepathStem(inputPath), otherExt))
	if localFs.Exists(candidate) {
		return candidate
	}

	return ""
}

func copyGoSDKGeneratorSkillsForEmbedding(outputDir string, generated []string, overwriteRequested bool, matcher *overwriteExclusionMatcher, logger logs.Loggers) (updatedGenerated []string, err error) {
	updatedGenerated = generated
	localFs := filesystem.GetGlobalFileSystem()

	skillsRoot := filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills")
	if !localFs.Exists(skillsRoot) {
		return
	}

	for _, skillPath := range embeddedAgentSkillPaths {
		sourcePath := filesystem.FilePathJoin(localFs, append([]string{skillsRoot}, strings.Split(skillPath, "/")...)...)
		if !localFs.Exists(sourcePath) {
			continue
		}

		skill, subErr := localFs.ReadFile(sourcePath)
		if subErr != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, subErr, "could not read tool-related skill %q", sourcePath)
			return
		}

		embeddedPath := filesystem.FilePathJoin(localFs, append([]string{outputDir, "generator", "skills"}, strings.Split(skillPath, "/")...)...)
		updatedGenerated = append(updatedGenerated, embeddedPath)
		if !shouldGenerateFile(embeddedPath, true, overwriteRequested, matcher, logger) {
			continue
		}
		embeddedDir := filesystem.FilePathDir(localFs, embeddedPath)
		if !localFs.Exists(embeddedDir) {
			err = localFs.MkDir(embeddedDir)
			if err != nil {
				err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not create %q", embeddedDir)
				return
			}
		}

		err = localFs.WriteFile(embeddedPath, skill, 0644)
		if err != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not write embedded skill %q", embeddedPath)
			return
		}

	}

	return
}

func runSkillName(binary string) string {
	return fmt.Sprintf("run-%s", binary)
}

func installSkillName(binary string) string {
	return fmt.Sprintf("install-%s", binary)
}

func githubActionSkillName(binary string) string {
	return fmt.Sprintf("using-%s-in-a-github-action", binary)
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

func (t *templates) addProtectedStateFile(templateFile, dest string, overwrite bool, markAsGenerated bool, logger logs.Loggers) {
	if markAsGenerated {
		t.toTrack = append(t.toTrack, dest)
	}
	if filesystem.Exists(dest) {
		if overwrite && logger != nil {
			logger.Log(fmt.Sprintf("preserving existing %s; it was not touched even though overwrite was requested", dest))
		}
		return
	}
	t.toGenerate = append(t.toGenerate,
		NewTemplateToPopulate(templateFile, dest),
	)
}

func generateGitAttributesFile(ctx context.Context, cfg *config.SDKGeneratorManagerConfiguration, generated []string, matcher *overwriteExclusionMatcher, logger logs.Loggers) (err error) {
	gitAttributesPath := filesystem.FilePathJoin(filesystem.GetGlobalFileSystem(), cfg.Output, ".gitattributes")
	if !shouldGenerateFile(gitAttributesPath, true, cfg.Overwrite, matcher, logger) {
		return nil
	}

	generated, err = relativeGeneratedFiles(cfg.Output, generated)
	if err != nil {
		return
	}

	err = populateTemplate(ctx, &GenerationOutput{
		ToolVersion:    field.OptionalString(cfg.ToolVersion, ""),
		GeneratedFiles: generated,
	}, gitAttributesFile, gitAttributesPath)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create .gitattributes file")
		return
	}

	return
}

func generateQltyFile(ctx context.Context, cfg *config.SDKGeneratorManagerConfiguration, generated []string, matcher *overwriteExclusionMatcher, logger logs.Loggers) (err error) {
	localFs := filesystem.GetGlobalFileSystem()
	qltyPath := filesystem.FilePathJoin(localFs, cfg.Output, ".qlty", "qlty.toml")
	if !shouldGenerateFile(qltyPath, cfg.Overwrite, cfg.Overwrite, matcher, logger) {
		return nil
	}

	generated, err = relativeGeneratedFiles(cfg.Output, generated)
	if err != nil {
		return
	}

	err = populateTemplate(ctx, &GenerationOutput{
		ToolVersion:    field.OptionalString(cfg.ToolVersion, ""),
		GeneratedFiles: generated,
	}, qltyTemplateFile, qltyPath)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create .qlty/qlty.toml file")
		return
	}

	return nil
}
