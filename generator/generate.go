/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package generator

import (
	"context"
	"embed"
	"fmt"
	"maps"
	"os"
	"path"
	"slices"
	"strings"
	"text/template"

	modsemver "golang.org/x/mod/semver"

	"github.com/ARM-software/go-sdk-generator/generator/config"
	"github.com/ARM-software/go-sdk-generator/generator/options"
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
	skillsGuideFile              = "templates/agent-skills.md.tmpl"
	skillsCheckFile              = "templates/validate-skills.py.tmpl"
	skillsRequirementsFile       = "templates/skills-requirements.txt.tmpl"
	skillsWorkflowFile           = "templates/skills.yml.tmpl"
	skillsActionFile             = "templates/validate-skills.action.yml.tmpl"
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
	qltyTrivyFile                = "templates/qlty.trivy.yaml.tmpl"
	qltyOSVScannerFile           = "templates/qlty.osv-scanner.sh.tmpl"
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
)

type generatedFileGroupDefinition struct {
	paths           []string
	additionalPaths func(*SDKSpecification, []GeneratedAction) []string
}

// generatedFileGroups lists exact output-relative paths alongside the template
// definitions. Paths that depend on the manifest are resolved for each generation.
var generatedFileGroups = map[string]generatedFileGroupDefinition{
	options.OverwriteOptionsExcludeGroupsCI.String(): {
		paths: []string{
			".github/workflows/ci.yml", ".github/workflows/upgrade-generator.yml",
			".github/workflows/dependabot.yml", ".github/workflows/release.yml",
			".github/workflows/licensing.yml", ".github/workflows/sbom.yml", ".github/workflows/lint-actions.yml",
			".goreleaser.yml",
		},
		additionalPaths: func(spec *SDKSpecification, actions []GeneratedAction) []string {
			paths := []string{"Dockerfile." + spec.MainModule}
			if spec.CommandLineCommands != nil {
				paths = append(paths, "scripts/pre.sh", ".github/workflows/skills.yml", "actions/validate-skills/action.yml")
				if spec.DistributeOci != nil {
					paths = append(paths, "Dockerfile.oci.alpine", "Dockerfile.oci.ubuntu")
				}
			}
			for _, action := range actions {
				paths = append(paths, path.Join("actions", action.DirectoryName, "action.yml"))
			}
			if len(actions) > 0 {
				paths = append(paths, "actions/_shared/install-and-run.sh")
			}
			return paths
		},
	},
	options.OverwriteOptionsExcludeGroupsLinters.String(): {
		paths: []string{".golangci.yaml", ".golangci.pre-commit.yaml", ".editorconfig", ".pre-commit-config.yaml", ".checkov.yaml", ".yamllint", ".qlty/qlty.toml", ".qlty/configs/trivy.yaml", ".qlty/scripts/osv-scanner.sh"},
	},
	options.OverwriteOptionsExcludeGroupsGomod.String(): {
		additionalPaths: func(spec *SDKSpecification, _ []GeneratedAction) []string {
			paths := []string{path.Join(spec.MainModule, "go.mod")}
			if cli := spec.CommandLineCommands; cli != nil {
				paths = append(paths, "go.work", path.Join(cli.Package, "go.mod"))
			}
			return paths
		},
	},
	options.OverwriteOptionsExcludeGroupsCLI.String(): {
		additionalPaths: func(spec *SDKSpecification, _ []GeneratedAction) []string {
			cli := spec.CommandLineCommands
			if cli == nil {
				return nil
			}
			return collection.Map([]string{
				"main.go", "cmd/" + spec.MainModule + ".go", "cmd/root.go",
				"cmd/config/" + spec.MainModule + ".go", "cmd/config/" + spec.MainModule + "_test.go", "cmd/config/enums.go",
				"cmd/" + spec.MainModule + "/wrappers.go",
				"internal/secretutil/secret.go", "cmd/module.properties",
			}, func(file string) string { return path.Join(cli.Package, file) })
		},
	},
	options.OverwriteOptionsExcludeGroupsCLIImplementation.String(): {
		additionalPaths: func(spec *SDKSpecification, _ []GeneratedAction) []string {
			if cli := spec.CommandLineCommands; cli != nil {
				localFs := filesystem.GetGlobalFileSystem()
				return []string{filesystem.FilePathToSlash(localFs, filesystem.FilePathJoin(localFs, cli.Package, "cmd", spec.MainModule, "implementations.go"))}
			}
			return nil
		},
	},
	options.OverwriteOptionsExcludeGroupsSDK.String(): {
		additionalPaths: func(spec *SDKSpecification, _ []GeneratedAction) []string {
			localFs := filesystem.GetGlobalFileSystem()
			sdkPath := func(elements ...string) string {
				return filesystem.FilePathToSlash(localFs, filesystem.FilePathJoin(localFs, append([]string{spec.MainModule}, elements...)...))
			}
			filename := spec.MainModule + ".go"
			paths := []string{sdkPath(filename), sdkPath("interfaces.go"), sdkPath("config", filename)}
			if hasInterfaceEnums(spec) {
				paths = append(paths, sdkPath("enums.go"))
			}
			if hasConfigEnums(spec) {
				paths = append(paths, sdkPath("config", "enums.go"))
			}
			types := groupAdditionalTypesByPackage(spec.AdditionalTypes)
			paths = append(paths, collection.FlatMap(sortedAdditionalTypePackages(spec.AdditionalTypes), func(pkg string) []string {
				outputs := []string{sdkPath(pkg, filename)}
				if types[pkg].HasEnum() {
					outputs = append(outputs, sdkPath(pkg, "enums.go"))
				}
				return outputs
			})...)
			optionSets := groupOptionSetsByPackage(spec.Options)
			return append(paths, collection.FlatMap(sortedOptionPackages(spec.Options), func(pkg string) []string {
				outputs := []string{sdkPath(pkg, filename)}
				if optionSets[pkg].HasEnum() {
					outputs = append(outputs, sdkPath(pkg, "enums.go"))
				}
				return outputs
			})...)
		},
	},
	options.OverwriteOptionsExcludeGroupsReadme.String(): {
		paths: []string{"README.md"},
	},
	options.OverwriteOptionsExcludeGroupsProject.String(): {
		paths: []string{
			"pyproject.toml", ".github/CODEOWNERS", ".github/dependabot.yml", ".github/workflows/dependabot.yml",
			".github/PULL_REQUEST_TEMPLATE.md", ".gitignore", ".gitattributes", "DOCUMENTATION.md",
			"CONTRIBUTING.md", "DEVELOPMENT.md", "CODE_OF_CONDUCT.md", "SECURITY.md", "KNOWN_ISSUES.md",
		},
		additionalPaths: func(spec *SDKSpecification, _ []GeneratedAction) []string {
			if cli := spec.CommandLineCommands; cli != nil {
				return []string{path.Join(cli.Package, "cmd/module.properties")}
			}
			return []string{"module.properties"}
		},
	},
	options.OverwriteOptionsExcludeGroupsAgent.String(): {
		additionalPaths: func(spec *SDKSpecification, actions []GeneratedAction) []string {
			var paths []string
			if cli := spec.CommandLineCommands; cli != nil {
				binary := field.OptionalString(cli.Binary, "")
				paths = append(paths, "agent.md", ".github/copilot-instructions.md", "guides/agent-skills.md", "actions/validate-skills/validate_skills.py", "actions/validate-skills/requirements.txt",
					path.Join(".agents/skills", runSkillName(binary), "SKILL.md"),
					path.Join(".agents/skills", installSkillName(binary), "SKILL.md"),
				)
				if len(actions) > 0 {
					paths = append(paths, path.Join(".agents/skills", githubActionSkillName(binary), "SKILL.md"))
				}
			}
			if spec.PackageName == "go-sdk-generator" {
				for _, skill := range embeddedAgentSkillPaths {
					paths = append(paths, path.Join("generator/skills", skill))
				}
			}
			return paths
		},
	},
}

// resolveGeneratedFileGroups combines fixed paths with paths derived from the
// manifest and generated actions. Each result owns its path slices so callers
// cannot modify the shared group definitions.
func resolveGeneratedFileGroups(spec *SDKSpecification, actions []GeneratedAction) map[string][]string {
	return collection.ReducesSequence(maps.Keys(generatedFileGroups), make(map[string][]string, len(generatedFileGroups)), func(groups map[string][]string, name string) map[string][]string {
		definition := generatedFileGroups[name]
		paths := slices.Clone(definition.paths)
		if definition.additionalPaths != nil {
			paths = append(paths, definition.additionalPaths(spec, actions)...)
		}
		groups[name] = paths
		return groups
	})
}

//go:embed templates
var fs embed.FS

type templateToPopulate struct {
	templateFile string
	output       string
}

func (t *templateToPopulate) Populate(ctx context.Context, data *DataModel, overwriteRules *overwrite) (err error) {
	err = populateTemplate(ctx, data, t.templateFile, t.output, overwriteRules)
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

func (t *templates) Populate(ctx context.Context, data *DataModel, overwriteRules *overwrite) (files []string, err error) {
	for _, tp := range t.toGenerate {
		err = tp.Populate(ctx, data, overwriteRules)
		if err != nil {
			return
		}
	}

	files = t.toTrack
	return
}

type DataModel struct {
	OutputDirectory  string
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
	Dockerfile     string
}

func populateTemplate[T DataModel | GenerationOutput](ctx context.Context, data *T, templateFile string, output string, overwriteRules *overwrite) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	if data == nil {
		err = commonerrors.UndefinedVariable("data model")
		return
	}
	if !overwriteRules.shouldInclude(output) {
		return nil
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

func Generate(ctx context.Context, logger logs.Loggers, cfg *config.SDKGeneratorManagerConfiguration, overwriteOptions ...options.OverwriteOption) (err error) {
	if cfg == nil {
		err = commonerrors.UndefinedVariable("generator configuration")
		return
	}
	overwriteRules, err := newOverwrite(cfg.Output, overwriteOptions...)
	if err != nil {
		return err
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

	var templatesToPopulate templates
	templatesToPopulate.addStateFile(implementationFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
	templatesToPopulate.addStateFile(interfaceFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "interfaces.go"), true, true)
	if hasInterfaceEnums(&sdk.Specification) {
		templatesToPopulate.addStateFile(interfaceEnumFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "enums.go"), true, true)
	}
	templatesToPopulate.addStateFile(configFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "config", fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
	if hasConfigEnums(&sdk.Specification) {
		templatesToPopulate.addStateFile(configEnumFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "config", "enums.go"), true, true)
	}
	templatesToPopulate.addStateFile(documentationFile, filesystem.FilePathJoin(localFs, cfg.Output, "DOCUMENTATION.md"), true, true)
	templatesToPopulate.addStateFile(contribFile, filesystem.FilePathJoin(localFs, cfg.Output, "CONTRIBUTING.md"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(developmentFile, filesystem.FilePathJoin(localFs, cfg.Output, "DEVELOPMENT.md"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(codeOfConductFile, filesystem.FilePathJoin(localFs, cfg.Output, "CODE_OF_CONDUCT.md"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(securityFile, filesystem.FilePathJoin(localFs, cfg.Output, "SECURITY.md"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(knownIssuesFile, filesystem.FilePathJoin(localFs, cfg.Output, "KNOWN_ISSUES.md"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(golangciFile, filesystem.FilePathJoin(localFs, cfg.Output, ".golangci.yaml"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(editorConfigFile, filesystem.FilePathJoin(localFs, cfg.Output, ".editorconfig"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(precommitConfigFile, filesystem.FilePathJoin(localFs, cfg.Output, ".pre-commit-config.yaml"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(golangciPrecommitFile, filesystem.FilePathJoin(localFs, cfg.Output, ".golangci.pre-commit.yaml"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(gitignoreFile, filesystem.FilePathJoin(localFs, cfg.Output, ".gitignore"), overwriteRules.enabled, false)
	templatesToPopulate.addStateFile(readmeFile, filesystem.FilePathJoin(localFs, cfg.Output, "README.md"), overwriteRules.enabled, false)
	templatesToPopulate.addStateFile(goModSdkFile, filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, "go.mod"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(checkovFile, filesystem.FilePathJoin(localFs, cfg.Output, ".checkov.yaml"), overwriteRules.enabled, false)
	templatesToPopulate.addStateFile(yamlLintFile, filesystem.FilePathJoin(localFs, cfg.Output, ".yamllint"), overwriteRules.enabled, false)
	templatesToPopulate.addStateFile(qltyTrivyFile, filesystem.FilePathJoin(localFs, cfg.Output, ".qlty", "configs", "trivy.yaml"), overwriteRules.enabled, true)
	templatesToPopulate.addStateFile(qltyOSVScannerFile, filesystem.FilePathJoin(localFs, cfg.Output, ".qlty", "scripts", "osv-scanner.sh"), overwriteRules.enabled, true)

	updateGoVersion := shouldRegenerateVersionSensitiveStateFile(overwriteRules.enabled, defaultGoVersion, requestedGoVersion)
	templatesToPopulate.addStateFile(dockerFile, filesystem.FilePathJoin(localFs, cfg.Output, fmt.Sprintf("Dockerfile.%v", sdk.Specification.MainModule)), updateGoVersion, true)
	if sdk.Specification.CommandLineCommands != nil && sdk.Specification.DistributeOci != nil {
		templatesToPopulate.addStateFile(ociDockerfileAlpineFile, filesystem.FilePathJoin(localFs, cfg.Output, "Dockerfile.oci.alpine"), updateGoVersion, true)
		templatesToPopulate.addStateFile(ociDockerfileUbuntuFile, filesystem.FilePathJoin(localFs, cfg.Output, "Dockerfile.oci.ubuntu"), updateGoVersion, true)
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
		templatesToPopulate.addStateFile(selectedDependabotFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "dependabot.yml"), overwriteRules.enabled, true)
		templatesToPopulate.addStateFile(prTemplateFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "PULL_REQUEST_TEMPLATE.md"), overwriteRules.enabled, true)
		if !sdk.Specification.Public {
			templatesToPopulate.addStateFile(codeOwnersFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "CODEOWNERS"), overwriteRules.enabled, true)
		}
		templatesToPopulate.addStateFile(selectedCIWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "ci.yml"), overwriteRules.enabled, true)
		templatesToPopulate.addStateFile(upgradeGeneratorWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "upgrade-generator.yml"), overwriteRules.enabled, true)
		templatesToPopulate.addStateFile(selectedDependabotWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "dependabot.yml"), overwriteRules.enabled, true)
		templatesToPopulate.addStateFile(selectedReleaseWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "release.yml"), overwriteRules.enabled, true)
		if sdk.Specification.UpdateCopyrightHeaders {
			templatesToPopulate.addStateFile(licensingWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "licensing.yml"), overwriteRules.enabled, true)
		}
		if sdk.Specification.Public {
			templatesToPopulate.addStateFile(sbomWorkflowPublicFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "sbom.yml"), overwriteRules.enabled, true)
		}
	}

	var generatedActions []GeneratedAction
	if sdk.Specification.CommandLineCommands != nil && !sdk.Specification.CommandLineCommands.NoActions {
		generatedActions, err = BuildGeneratedActions(&sdk)
		if err != nil {
			return
		}
	}

	generateAction := len(generatedActions) > 0
	overwriteRules.addGeneratedGroups(resolveGeneratedFileGroups(&sdk.Specification, generatedActions))
	if onGithub && generateAction {
		templatesToPopulate.addStateFile(lintActionsWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "lint-actions.yml"), overwriteRules.enabled, true)
	}

	if sdk.Specification.CommandLineCommands != nil {
		templatesToPopulate.addStateFile(goreleaserFile, filesystem.FilePathJoin(localFs, cfg.Output, ".goreleaser.yml"), overwriteRules.enabled, true)
	} else {
		templatesToPopulate.addStateFile(goreleaserLibraryFile, filesystem.FilePathJoin(localFs, cfg.Output, ".goreleaser.yml"), overwriteRules.enabled, true)
	}

	if sdk.Specification.CommandLineCommands != nil {
		cliDir := sdk.Specification.CommandLineCommands.Package
		templatesToPopulate.addStateFile(skillsGuideFile, filesystem.FilePathJoin(localFs, cfg.Output, "guides", "agent-skills.md"), overwriteRules.enabled, true)
		templatesToPopulate.addStateFile(skillsCheckFile, filesystem.FilePathJoin(localFs, cfg.Output, "actions", "validate-skills", "validate_skills.py"), overwriteRules.enabled, true)
		templatesToPopulate.addStateFile(skillsRequirementsFile, filesystem.FilePathJoin(localFs, cfg.Output, "actions", "validate-skills", "requirements.txt"), overwriteRules.enabled, true)
		if onGithub {
			templatesToPopulate.addStateFile(skillsWorkflowFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "workflows", "skills.yml"), overwriteRules.enabled, true)
			for _, action := range generatedActions {
				if action.DirectoryName == "validate-skills" {
					return commonerrors.Newf(commonerrors.ErrInvalid, "action directory validate-skills is reserved for skill validation")
				}
			}
			templatesToPopulate.addStateFile(skillsActionFile, filesystem.FilePathJoin(localFs, cfg.Output, "actions", "validate-skills", "action.yml"), overwriteRules.enabled, true)
		}
		templatesToPopulate.addStateFile(agentFile, filesystem.FilePathJoin(localFs, cfg.Output, "agent.md"), true, true)
		if onGithub {
			templatesToPopulate.addStateFile(copilotFile, filesystem.FilePathJoin(localFs, cfg.Output, ".github", "copilot-instructions.md"), true, true)
		}
		binary := field.OptionalString(sdk.Specification.CommandLineCommands.Binary, "")
		templatesToPopulate.addStateFile(cliSkillFile, filesystem.FilePathJoin(localFs, cfg.Output, ".agents", "skills", runSkillName(binary), "SKILL.md"), true, true)
		templatesToPopulate.addStateFile(installSkillFile, filesystem.FilePathJoin(localFs, cfg.Output, ".agents", "skills", installSkillName(binary), "SKILL.md"), overwriteRules.enabled, true)
		if generateAction {
			templatesToPopulate.addStateFile(githubActionSkillFile, filesystem.FilePathJoin(localFs, cfg.Output, ".agents", "skills", githubActionSkillName(binary), "SKILL.md"), true, true)
		}
		templatesToPopulate.addStateFile(cliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
		templatesToPopulate.addStateFile(rootCliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "root.go"), true, true)
		templatesToPopulate.addStateFile(mainCliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "main.go"), true, true)
		templatesToPopulate.addStateFile(configCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "config", fmt.Sprintf("%v.go", sdk.Specification.MainModule)), true, true)
		if sdk.Specification.CommandLineCommands.HasSummaryValidationTests() {
			templatesToPopulate.addStateFile(configCLITestFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "config", fmt.Sprintf("%v_test.go", sdk.Specification.MainModule)), true, true)
		}
		if hasCLIConfigEnums(sdk.Specification.CommandLineCommands) {
			templatesToPopulate.addStateFile(configCLIEnumFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", "config", "enums.go"), true, true)
		}
		templatesToPopulate.addStateFile(implementationWrapperCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", sdk.Specification.MainModule, "wrappers.go"), true, true)
		templatesToPopulate.addStateFile(preinstallFile, filesystem.FilePathJoin(localFs, cfg.Output, "scripts", "pre.sh"), overwriteRules.enabled, true)
		if hasSecretStringParameters(sdk.Specification.CommandLineCommands) {
			templatesToPopulate.addStateFile(secretUtilCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "internal", "secretutil", "secret.go"), true, true)
		}

		templatesToPopulate.addStateFile(implementationCLIFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "cmd", sdk.Specification.MainModule, "implementations.go"), overwriteRules.enabled, false)
		templatesToPopulate.addStateFile(goWorkFile, filesystem.FilePathJoin(localFs, cfg.Output, "go.work"), overwriteRules.enabled, true)
		templatesToPopulate.addStateFile(goModCliFile, filesystem.FilePathJoin(localFs, cfg.Output, cliDir, "go.mod"), overwriteRules.enabled, true)
	}

	var modulePropertiesFilePath string
	if sdk.Specification.CommandLineCommands == nil {
		modulePropertiesFilePath = filesystem.FilePathJoin(localFs, cfg.Output, "module.properties")
	} else {
		modulePropertiesFilePath = filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.CommandLineCommands.Package, "cmd", "module.properties")
	}
	templatesToPopulate.addProtectedStateFile(modulePropertiesFile, modulePropertiesFilePath, overwriteRules.enabled, true, logger)
	templatesToPopulate.addStateFile(pyprojectFile, filesystem.FilePathJoin(localFs, cfg.Output, "pyproject.toml"), overwriteRules.enabled, true)

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
		OutputDirectory:               cfg.Output,
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

	generated, err := templatesToPopulate.Populate(ctx, data, overwriteRules)
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
			if subErr := populateTemplate(ctx, &packageData, typesFile, typeOutput, overwriteRules); subErr != nil {
				return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate types template")
			}
			generated = append(generated, typeOutput)

			if groupedTypes[pkg].HasEnum() {
				enumOutput := filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, pkg, "enums.go")
				if subErr := populateTemplate(ctx, &packageData, typesEnumFile, enumOutput, overwriteRules); subErr != nil {
					return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate type enums template")
				}
				generated = append(generated, enumOutput)
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
			if subErr := populateTemplate(ctx, &packageData, optionsFile, optionOutput, overwriteRules); subErr != nil {
				return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate options template")
			}
			generated = append(generated, optionOutput)

			if groupedOptionSets[pkg].HasEnum() {
				enumOutput := filesystem.FilePathJoin(localFs, cfg.Output, sdk.Specification.MainModule, pkg, "enums.go")
				if subErr := populateTemplate(ctx, &packageData, optionsEnumFile, enumOutput, overwriteRules); subErr != nil {
					return commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not populate option enums template")
				}
				generated = append(generated, enumOutput)
			}
			return nil
		})
		if err != nil {
			return
		}
	}

	if sdk.Specification.PackageName == "go-sdk-generator" {
		generated, err = copyGoSDKGeneratorSkillsForEmbedding(cfg.Output, generated, overwriteRules)
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
			if err = actionOutput.Populate(ctx, &actionData, overwriteRules); err != nil {
				err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not populate action %q", action.DirectoryName)
				return
			}

			generated = append(generated, actionOutput.output)
		}
	}

	if generateAction {
		sharedActionOutput := NewTemplateToPopulate(actionSharedScriptFile, filesystem.FilePathJoin(localFs, cfg.Output, "actions", "_shared", "install-and-run.sh"))
		if err = sharedActionOutput.Populate(ctx, data, overwriteRules); err != nil {
			err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not populate shared action runtime")
			return
		}

		generated = append(generated, sharedActionOutput.output)
	}
	qltyPath := filesystem.FilePathJoin(localFs, cfg.Output, ".qlty", "qlty.toml")
	generated = append(generated, qltyPath)
	err = generateQltyFile(ctx, cfg, generated, sdk.Specification.MainModule, overwriteRules)
	if err != nil {
		return
	}
	err = generateGitAttributesFile(ctx, cfg, generated, overwriteRules)
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

func copyGoSDKGeneratorSkillsForEmbedding(outputDir string, generated []string, overwriteRules *overwrite) (updatedGenerated []string, err error) {
	updatedGenerated = generated
	localFs := filesystem.GetGlobalFileSystem()

	skillsRoot := filesystem.FilePathJoin(localFs, outputDir, ".agents", "skills")
	if !localFs.Exists(skillsRoot) {
		return
	}

	var skillPaths []string
	for _, entry := range embeddedAgentSkillPaths {
		root := filesystem.FilePathJoin(localFs, skillsRoot, path.Dir(entry))
		if !localFs.Exists(root) {
			continue
		}
		err = localFs.Walk(root, func(name string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.Mode().IsRegular() {
				relative, relErr := localFs.ConvertToRelativePath(skillsRoot, name)
				if relErr != nil {
					return relErr
				}
				skillPaths = append(skillPaths, filesystem.FilePathToSlash(localFs, relative[0]))
			}
			return nil
		})
		if err != nil {
			return
		}
	}
	if overwriteRules != nil {
		embeddedPaths := collection.Map(skillPaths, func(skillPath string) string {
			return path.Join("generator/skills", skillPath)
		})
		overwriteRules.addGeneratedGroups(map[string][]string{options.OverwriteOptionsExcludeGroupsAgent.String(): embeddedPaths})
	}
	for _, skillPath := range skillPaths {
		embeddedPath := filesystem.FilePathJoin(localFs, append([]string{outputDir, "generator", "skills"}, strings.Split(skillPath, "/")...)...)
		if overwriteRules.shouldExclude(embeddedPath) {
			updatedGenerated = append(updatedGenerated, embeddedPath)
			continue
		}
		sourcePath := filesystem.FilePathJoin(localFs, append([]string{skillsRoot}, strings.Split(skillPath, "/")...)...)
		if !localFs.Exists(sourcePath) {
			continue
		}

		skill, subErr := localFs.ReadFile(sourcePath)
		if subErr != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, subErr, "could not read tool-related skill %q", sourcePath)
			return
		}

		embeddedDir := filesystem.FilePathDir(localFs, embeddedPath)
		if !localFs.Exists(embeddedDir) {
			err = localFs.MkDirAll(embeddedDir, 0755)
			if err != nil {
				err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not create %q", embeddedDir)
				return
			}
		}

		info, statErr := localFs.Stat(sourcePath)
		if statErr != nil {
			return updatedGenerated, statErr
		}
		err = localFs.WriteFile(embeddedPath, skill, info.Mode().Perm())
		if err != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not write embedded skill %q", embeddedPath)
			return
		}

		updatedGenerated = append(updatedGenerated, embeddedPath)
	}

	return
}

func runSkillName(binary string) string {
	return portableSkillName(fmt.Sprintf("run-%s", binary))
}

func installSkillName(binary string) string {
	return portableSkillName(fmt.Sprintf("install-%s", binary))
}

func githubActionSkillName(binary string) string {
	return portableSkillName(fmt.Sprintf("using-%s-in-a-github-action", binary))
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

func generateGitAttributesFile(ctx context.Context, cfg *config.SDKGeneratorManagerConfiguration, generated []string, overwriteRules *overwrite) (err error) {
	generated, err = relativeGeneratedFiles(cfg.Output, generated)
	if err != nil {
		return
	}

	err = populateTemplate(ctx, &GenerationOutput{
		ToolVersion:    field.OptionalString(cfg.ToolVersion, ""),
		GeneratedFiles: generated,
	}, gitAttributesFile, filesystem.FilePathJoin(filesystem.GetGlobalFileSystem(), cfg.Output, ".gitattributes"), overwriteRules)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create .gitattributes file")
		return
	}

	return
}

func generateQltyFile(ctx context.Context, cfg *config.SDKGeneratorManagerConfiguration, generated []string, mainModule string, overwriteRules *overwrite) (err error) {
	localFs := filesystem.GetGlobalFileSystem()
	qltyPath := filesystem.FilePathJoin(localFs, cfg.Output, ".qlty", "qlty.toml")
	if localFs.Exists(qltyPath) && !overwriteRules.enabled {
		return nil
	}

	generated, err = relativeGeneratedFiles(cfg.Output, generated)
	if err != nil {
		return
	}

	// OSV Scanner targets go.mod files, including those created by the generator.
	excluded := collection.Filter(generated, func(file string) bool {
		return path.Base(file) != "go.mod"
	})
	err = populateTemplate(ctx, &GenerationOutput{
		ToolVersion:    field.OptionalString(cfg.ToolVersion, ""),
		GeneratedFiles: excluded,
		Dockerfile:     fmt.Sprintf("Dockerfile.%s", mainModule),
	}, qltyTemplateFile, qltyPath, overwriteRules)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not create .qlty/qlty.toml file")
		return
	}

	return nil
}
