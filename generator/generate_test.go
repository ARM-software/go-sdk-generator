package generator

import (
	"bytes"
	"context"
	"fmt"
	fSys "io/fs"
	"os/exec"
	"path"
	"slices"
	"sort"
	"strings"
	"testing"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/hashing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Arm-Debug/go-sdk-generator/generator/config"
)

func generatedModuleDirectories(ctx context.Context, t *testing.T, localFs filesystem.FS, generatedDir string) []string {
	t.Helper()

	modules, err := filesystem.Glob(filesystem.FilePathJoin(localFs, generatedDir, "**", "go.mod"))
	require.NoError(t, err)
	modules = collection.Map(modules, func(path string) string {
		return filesystem.FilePathDir(localFs, path)
	})
	modules = slices.Compact(modules)
	sort.Strings(modules)
	return modules
}

func isPrivateDependencyFailure(output string) bool {
	trimmed := strings.TrimSpace(output)
	lowerTrimmed := strings.ToLower(trimmed)
	if trimmed == "" || !strings.Contains(lowerTrimmed, strings.ToLower("github.com/Arm-Debug/")) {
		return false
	}
	return collection.AnyFunc([]string{
		"could not read username",
		"authentication required",
		"repository not found",
		"terminal prompts disabled",
		"no secure protocol found",
		"does not contain package",
	}, func(marker string) bool {
		return strings.Contains(lowerTrimmed, marker)
	})
}

func TestIsPrivateDependencyFailure(t *testing.T) {
	assert.True(t, isPrivateDependencyFailure(`go: github.com/Arm-Debug/example imports
		github.com/Arm-Debug/missing/module: module github.com/Arm-Debug/missing/module@latest found (v1.0.0), but does not contain package github.com/Arm-Debug/missing/module/pkg`))
	assert.False(t, isPrivateDependencyFailure(`go: finding module for package github.com/example/public`))
}

func normaliseGeneratedGoWork(ctx context.Context, t *testing.T, localFs filesystem.FS, verificationDir string) string {
	t.Helper()

	goWorkPath := filesystem.FilePathJoin(localFs, verificationDir, "go.work")
	if filesystem.Exists(goWorkPath) {
		require.NoError(t, filesystem.Rm(goWorkPath))
	}

	moduleDirs := generatedModuleDirectories(ctx, t, localFs, verificationDir)
	if len(moduleDirs) == 0 {
		return "off"
	}

	cmd := exec.CommandContext(ctx, "go", "work", "init")
	cmd.Dir = verificationDir
	cmd.Env = append(cmd.Environ(), "GOWORK=off")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	require.NoError(t, err, "go work init failed in %s\n%s", verificationDir, output.String())

	relativeModuleDirs := collection.Map(moduleDirs, func(moduleDir string) string {
		rel, err := filesystem.FilePathRel(localFs, verificationDir, moduleDir)
		require.NoError(t, err)
		return rel
	})

	goWorkUseArgs := append([]string{"work", "use"}, relativeModuleDirs...)
	cmd = exec.CommandContext(ctx, "go", goWorkUseArgs...)
	cmd.Dir = verificationDir
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("GOWORK=%s", goWorkPath))

	output.Reset()
	cmd.Stdout = &output
	cmd.Stderr = &output

	err = cmd.Run()
	require.NoError(t, err, "go work use failed in %s\n%s", verificationDir, output.String())

	return goWorkPath
}

func runGeneratedGoCommand(ctx context.Context, t *testing.T, workDir string, goWorkValue string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = workDir
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("GOWORK=%s", goWorkValue))

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	return output.String(), err
}

func buildGeneratedModules(ctx context.Context, localFs filesystem.FS, generatedDir string, t *testing.T) {
	t.Helper()

	verificationDir := filesystem.FilePathJoin(localFs, t.TempDir(), "verification-copy")
	require.NoError(t, localFs.CopyWithContext(ctx, generatedDir, verificationDir))
	goWorkValue := normaliseGeneratedGoWork(ctx, t, localFs, verificationDir)

	for _, moduleDir := range generatedModuleDirectories(ctx, t, localFs, verificationDir) {
		moduleName, err := filesystem.FilePathRel(localFs, verificationDir, moduleDir)
		require.NoError(t, err)
		modulePattern := fmt.Sprintf("./%s/...", strings.ReplaceAll(moduleName, "\\", "/"))

		t.Run(moduleName, func(t *testing.T) {
			t.Helper()

			skipModuleVerification := false
			skipReason := ""

			t.Run("go mod tidy", func(t *testing.T) {
				output, err := runGeneratedGoCommand(ctx, t, moduleDir, "off", "mod", "tidy")
				if err != nil && isPrivateDependencyFailure(output) {
					skipModuleVerification = true
					skipReason = output
					t.Skipf("skipping generated module compilation/verification in %s because the generated project requires dependencies that are private or otherwise inaccessible in this environment while running 'go %s'\n%s", moduleDir, strings.Join([]string{"mod", "tidy"}, " "), output)
				}
				require.NoError(t, err, "go %s failed in %s\n%s", strings.Join([]string{"mod", "tidy"}, " "), moduleDir, output)
			})

			t.Run("go generate", func(t *testing.T) {
				if skipModuleVerification {
					t.Skipf("skipping generated module compilation/verification in %s because the generated project requires dependencies that are private or otherwise inaccessible in this environment\n%s", moduleDir, skipReason)
				}
				output, err := runGeneratedGoCommand(ctx, t, verificationDir, goWorkValue, "generate", modulePattern)
				require.NoError(t, err, "go %s failed in %s\n%s", strings.Join([]string{"generate", modulePattern}, " "), moduleDir, output)
			})

			for _, command := range []struct {
				name string
				args []string
			}{
				{name: "go build", args: []string{"build", "./..."}},
				{name: "go test", args: []string{"test", "./..."}},
			} {
				t.Run(command.name, func(t *testing.T) {
					if skipModuleVerification {
						t.Skipf("skipping generated module compilation/verification in %s because the generated project requires dependencies that are private or otherwise inaccessible in this environment\n%s", moduleDir, skipReason)
					}
					args := []string{command.args[0], modulePattern}
					output, err := runGeneratedGoCommand(ctx, t, verificationDir, goWorkValue, args...)
					require.NoError(t, err, "go %s failed in %s\n%s", strings.Join(args, " "), moduleDir, output)
				})
			}
		})
	}
}

func compareGeneratedOutputWithRegressionArchive(ctx context.Context, localFs filesystem.FS, generatedDir, baselineZip string, t *testing.T) {
	t.Helper()

	buildGeneratedModules(ctx, localFs, generatedDir, t)

	generatedZip := filesystem.FilePathJoin(localFs, t.TempDir(), "generated.zip")
	err := localFs.ZipWithContext(ctx, generatedDir, generatedZip)
	require.NoError(t, err)

	baselineHash, err := localFs.FileHashWithContext(ctx, hashing.HashSha256, baselineZip)
	require.NoError(t, err)

	generatedHash, err := localFs.FileHashWithContext(ctx, hashing.HashSha256, generatedZip)
	require.NoError(t, err)

	if generatedHash == baselineHash {
		t.Skipf("generated archive hash matches baseline fixture %s, so unpacked content comparison is unnecessary", filesystem.FilePathBase(localFs, baselineZip))
	}

	baselineDir := t.TempDir()
	_, err = localFs.Unzip(baselineZip, baselineDir)
	require.NoError(t, err)

	checkFilesAreEqual(t, generatedDir, baselineDir, filesystem.NewStandardFileSystem(), "pyproject.toml")
}

func TestFormatImports(t *testing.T) {
	localFs := filesystem.NewStandardFileSystem()

	t.Run("Valid import add", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filesystem.FilePathJoin(localFs, tmpDir, "test.go")
		tmpFileContent := "package main\n\nfunc main() {\nfmt.Println(\"Hello World\")\n}\n"

		err := filesystem.WriteFile(tmpFile, []byte(tmpFileContent), 0655)
		require.NoError(t, err)

		cleaned, err := formatImports(context.Background(), tmpFile)
		assert.NoError(t, err)
		assert.Equal(t, "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}\n", string(cleaned))
	})

	t.Run("Invalid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filesystem.FilePathJoin(localFs, tmpDir, "test.go")
		tmpFileContent := "main {\nfmt.Println(\"Hello World\")\n}\n"

		err := filesystem.WriteFile(tmpFile, []byte(tmpFileContent), 0655)
		require.NoError(t, err)

		cleaned, err := formatImports(context.Background(), tmpFile)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
		assert.Empty(t, cleaned)
	})

	t.Run("No file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filesystem.FilePathJoin(localFs, tmpDir, "test.go")

		cleaned, err := formatImports(context.Background(), tmpFile)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, cleaned)
	})

	t.Run("Context cancel", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filesystem.FilePathJoin(localFs, tmpDir, "test.go")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		cleaned, err := formatImports(ctx, tmpFile)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.Empty(t, cleaned)
	})
}

func TestFuncMap(t *testing.T) {
	t.Run("FuncMap", func(t *testing.T) {
		tmpl, err := template.
			New("test").
			Funcs(templateFuncMap).
			Parse(strings.Join([]string{
				"{{ .Data | title }}",
				"{{ .Data | lower }}",
			}, "\n"))

		require.NoError(t, err)

		var b strings.Builder
		err = tmpl.ExecuteTemplate(&b, "test", struct{ Data string }{"this_is_A_WORD"})
		require.NoError(t, err)

		assert.Equal(t, b.String(), "This_is_A_WORD\nthis_is_a_word")
	})
}

func TestPopulateTemplate(t *testing.T) {
	localFs := filesystem.NewStandardFileSystem()

	t.Run("Valid import add", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filesystem.FilePathJoin(localFs, tmpDir, "out1", "out2", "out.go")
		require.NoDirExists(t, filesystem.FilePathDir(localFs, out))
		testFs, err := filesystem.NewEmbedFileSystem(&fs)
		require.NoError(t, err)

		ctx := context.Background()
		err = populateTemplate(ctx, &DataModel{}, filesystem.FilePathJoin(testFs, "templates", "test.go.tmpl"), out)
		assert.NoError(t, err)

		content, err := filesystem.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}\n", string(content))
	})

	t.Run("Valid import add (not .go so no import tidy)", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filesystem.FilePathJoin(localFs, tmpDir, "out1", "out2")
		require.NoDirExists(t, filesystem.FilePathDir(localFs, out))
		testFs, err := filesystem.NewEmbedFileSystem(&fs)
		require.NoError(t, err)
		ctx := context.Background()
		err = populateTemplate(ctx, &DataModel{}, filesystem.FilePathJoin(testFs, "templates", "test.go.tmpl"), out)
		assert.NoError(t, err)

		content, err := filesystem.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "package main\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}\n", string(content))
	})

	t.Run("Missing template", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filesystem.FilePathJoin(localFs, tmpDir, "out")

		ctx := context.Background()
		err := populateTemplate(ctx, &DataModel{}, path.Join("templates", "test-not-exist.go.tmpl"), out)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
		assert.NoFileExists(t, out)
	})

	t.Run("Cancel Context", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filesystem.FilePathJoin(localFs, tmpDir, "out")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := populateTemplate(ctx, &DataModel{}, path.Join("templates", "test.go.tmpl"), out)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.NoFileExists(t, out)
	})
}

func checkFilesAreEqual(t assert.TestingT, actualDir string, expectedDir string, localFs filesystem.FS, ignoreFiles ...string) {
	err := localFs.WalkWithContext(context.Background(), expectedDir, func(path string, info fSys.FileInfo, err error) error {
		if slices.ContainsFunc(ignoreFiles, func(ignore string) bool {
			return strings.HasSuffix(path, filesystem.FilePathClean(localFs, ignore))
		}) {
			return nil
		}

		assert.NoError(t, err)

		relativePath, err := filesystem.FilePathRel(localFs, expectedDir, path)
		assert.NoError(t, err)

		expectedFilePath := filesystem.FilePathJoin(localFs, actualDir, relativePath)

		if info.IsDir() {
			folderExists, err := filesystem.IsDir(expectedFilePath)
			assert.NoError(t, err)
			assert.Equal(t, folderExists, true)
			return nil
		}
		createdFileContent, err := filesystem.ReadFile(path)
		assert.NoError(t, err)

		expectedFileContent, err := filesystem.ReadFile(expectedFilePath)
		assert.NoError(t, err)

		assert.Equal(t, strings.ReplaceAll(string(expectedFileContent), "\r", ""), strings.ReplaceAll(string(createdFileContent), "\r", ""), "file %v", relativePath)
		return nil
	})
	assert.NoError(t, err)
}

func TestGenerate(t *testing.T) {
	localFs := filesystem.NewStandardFileSystem()

	tests := []struct {
		title         string
		recipe        string
		regressionSrc string
		err           error
	}{
		{
			title:         "No CLI (workspace service sdk)",
			recipe:        "workspace.json",
			regressionSrc: "workspace.zip",
			err:           nil,
		},
		{
			title:         "Yes CLI (verify cmsis example sdk (json))",
			recipe:        "verify.json",
			regressionSrc: "verify.zip",
			err:           nil,
		},
		{
			title:         "Yes CLI (verify cmsis example sdk (yaml))",
			recipe:        "verify.yml",
			regressionSrc: "verify.zip",
			err:           nil,
		},
		{
			title:         "Yes CLI (FPGA service agent (duplicate arguments))",
			recipe:        "agent.yml",
			regressionSrc: "agent.zip",
			err:           nil,
		},
		{
			title:         "No CLI (custom lang types are normalised to generated type casing)",
			recipe:        "casing-types.yml",
			regressionSrc: "casing-types.zip",
			err:           nil,
		},
		{
			title:         "No CLI (list and iterator flags shape generated signatures and fields)",
			recipe:        "list-iterator.yml",
			regressionSrc: "list-iterator.zip",
			err:           nil,
		},
		{
			title:         "Yes CLI (string enums generate constants validation and cli help text)",
			recipe:        "enum.yml",
			regressionSrc: "enum.zip",
			err:           nil,
		},
		{
			title:         "Yes CLI (packaged additional types generate into grouped packages)",
			recipe:        "packaged-types.yml",
			regressionSrc: "packaged-types.zip",
			err:           nil,
		},
		{
			title:         "Yes CLI (custom CLI package changes generated project root)",
			recipe:        "cli-package.yml",
			regressionSrc: "cli-package.zip",
			err:           nil,
		},
		{
			title:         "Yes CLI (constraints generate explicit flag and validation rules)",
			recipe:        "constraints.yml",
			regressionSrc: "constraints.zip",
			err:           nil,
		},
		{
			title:  "Yes CLI (verify cmsis example sdk (yaml))",
			recipe: "verify.bad.yml",
			err:    commonerrors.ErrConflict,
		},
		{
			title:  "Yes CLI (verify cmsis example sdk (yaml))",
			recipe: "verify.bad2.yml",
			err:    commonerrors.ErrConflict,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.title, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			tmpDir := t.TempDir()
			cfg := config.DefaultSDKGeneratorManagerConfiguration()
			cfg.Input = filesystem.FilePathJoin(localFs, "testdata", test.recipe)
			cfg.Output = tmpDir
			cfg.ToolVersion = field.ToOptionalString("0.31.0")

			err := Generate(ctx, cfg)
			if test.err != nil {
				errortest.AssertError(t, err, test.err)
			} else {
				assert.NoError(t, err)
				zipped := filesystem.FilePathJoin(localFs, "testdata", test.regressionSrc)
				compareGeneratedOutputWithRegressionArchive(ctx, localFs, tmpDir, zipped, t)
			}
		})
	}

	t.Run("Missing config", func(t *testing.T) {
		ctx := context.Background()
		err := Generate(ctx, nil)
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
	})

	t.Run("Context cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		cfg := &config.SDKGeneratorManagerConfiguration{
			Input:  filesystem.FilePathJoin(localFs, "testdata", "verify.json"),
			Output: "aaaa",
		}
		err := Generate(ctx, cfg)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})

	t.Run("Invalid specification", func(t *testing.T) {
		ctx := context.Background()
		cfg := &config.SDKGeneratorManagerConfiguration{
			Input:  filesystem.FilePathJoin(localFs, "testdata", "i-dont-exist.json"),
			Output: "bbb",
		}

		err := Generate(ctx, cfg)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected, commonerrors.ErrNotFound)
	})

	t.Run("Multiline YAML defaults generate valid code", func(t *testing.T) {
		tmpDir := t.TempDir()
		ctx := context.Background()
		cfg := config.DefaultSDKGeneratorManagerConfiguration()
		cfg.Input = filesystem.FilePathJoin(localFs, "testdata", "multiline-defaults.yml")
		cfg.Output = tmpDir

		err := Generate(ctx, cfg)
		assert.NoError(t, err, "Code generation should succeed with multiline YAML defaults")

		configFile := filesystem.FilePathJoin(localFs, tmpDir, "multilinetest", "config", "multilinetest.go")
		typesFile := filesystem.FilePathJoin(localFs, tmpDir, "multilinetest", "types", "multilinetest.go")
		interfaceFile := filesystem.FilePathJoin(localFs, tmpDir, "multilinetest", "interfaces.go")

		configContent, err := filesystem.ReadFile(configFile)
		require.NoError(t, err, "Generated config file should exist")

		typesContent, err := filesystem.ReadFile(typesFile)
		require.NoError(t, err, "Generated types file should exist")

		interfaceContent, err := filesystem.ReadFile(interfaceFile)
		require.NoError(t, err, "Generated interface file should exist")

		assert.NotContains(t, string(configContent), "\n,", "Generated config should not have commas on new lines")
		assert.NotContains(t, string(typesContent), "\n,", "Generated types should not have commas on new lines")
		assert.Contains(t, string(configContent), "// It supports multiline descriptions.")
		assert.Contains(t, string(typesContent), "// It supports multiline descriptions.")
		assert.Contains(t, string(interfaceContent), "// This also verifies multiline descriptions.")
	})

	t.Run("No actions skips action outputs", func(t *testing.T) {
		tmpDir := t.TempDir()
		ctx := context.Background()
		cfg := config.DefaultSDKGeneratorManagerConfiguration()
		cfg.Input = filesystem.FilePathJoin(localFs, "testdata", "agent.yml")
		cfg.Output = tmpDir
		cfg.ToolVersion = field.ToOptionalString("0.31.0")

		err := Generate(ctx, cfg)
		require.NoError(t, err)

		assert.NoDirExists(t, filesystem.FilePathJoin(localFs, tmpDir, "actions"))
		assert.NoFileExists(t, filesystem.FilePathJoin(localFs, tmpDir, ".agents", "skills", "using-fpga-service-agent-in-a-github-action", "SKILL.md"))
	})

}

func TestCompareGoVersions(t *testing.T) {
	assert.Equal(t, 0, compareGoVersions("1.24.6", "1.24.6"))
	assert.Equal(t, 1, compareGoVersions("1.24.6", "1.24.5"))
	assert.Equal(t, -1, compareGoVersions("1.24.6", "1.26.5"))
	assert.Equal(t, 0, compareGoVersions("v1.24.6", "1.24.6"))
}

func TestShouldRegenerateVersionSensitiveStateFile(t *testing.T) {
	assert.False(t, shouldRegenerateVersionSensitiveStateFile(false, "1.24.6", "1.26.5"))
	assert.True(t, shouldRegenerateVersionSensitiveStateFile(false, "1.24.6", "1.24.5"))
	assert.True(t, shouldRegenerateVersionSensitiveStateFile(true, "1.24.6", "1.26.5"))
}

func TestAddStateFileFollowsOverwriteFlag(t *testing.T) {
	tmpls := &templates{}
	existingFile := filesystem.FilePathJoin(filesystem.NewStandardFileSystem(), t.TempDir(), "state.yml")
	require.NoError(t, filesystem.WriteFile(existingFile, []byte("keep"), 0644))

	tmpls.addStateFile("template", existingFile, false, true)
	assert.Empty(t, tmpls.toGenerate)
	assert.Equal(t, []string{existingFile}, tmpls.toTrack)

	tmpls = &templates{}
	tmpls.addStateFile("template", existingFile, true, true)
	require.Len(t, tmpls.toGenerate, 1)
	assert.Equal(t, existingFile, tmpls.toGenerate[0].output)
	assert.Equal(t, []string{existingFile}, tmpls.toTrack)
}

func TestValidateManifest(t *testing.T) {
	localFs := filesystem.NewStandardFileSystem()

	t.Run("valid yaml manifest", func(t *testing.T) {
		validationErr, err := Validate(context.Background(), filesystem.FilePathJoin(localFs, "testdata", "verify.yml"))
		require.NoError(t, err)
		assert.NoError(t, validationErr)
	})

	t.Run("invalid yaml manifest", func(t *testing.T) {
		validationErr, err := Validate(context.Background(), filesystem.FilePathJoin(localFs, "testdata", "verify.bad.yml"))
		require.NoError(t, err)
		require.Error(t, validationErr)
		assert.True(t, commonerrors.Any(validationErr, commonerrors.ErrInvalid))
	})
}
