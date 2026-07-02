package generator

import (
	"context"
	fSys "io/fs"
	"path"
	"slices"
	"strings"
	"testing"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/hashing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Arm-Debug/go-sdk-generator/generator/config"
)

func compareGeneratedOutputWithRegressionArchive(ctx context.Context, localFs filesystem.FS, generatedDir, baselineZip string, t *testing.T) {
	t.Helper()

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
	_, err = filesystem.Unzip(baselineZip, baselineDir)
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
			ctx := context.Background()
			tmpDir := t.TempDir()
			cfg := config.DefaultSdkGeneratorManagerConfiguration()
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
		cfg := &config.SdkGeneratorManagerConfiguration{
			Input:  filesystem.FilePathJoin(localFs, "testdata", "verify.json"),
			Output: "aaaa",
		}
		err := Generate(ctx, cfg)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})

	t.Run("Invalid specification", func(t *testing.T) {
		ctx := context.Background()
		cfg := &config.SdkGeneratorManagerConfiguration{
			Input:  filesystem.FilePathJoin(localFs, "testdata", "i-dont-exist.json"),
			Output: "bbb",
		}

		err := Generate(ctx, cfg)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected, commonerrors.ErrNotFound)
	})

	t.Run("Multiline YAML defaults generate valid code", func(t *testing.T) {
		tmpDir := t.TempDir()
		ctx := context.Background()
		cfg := config.DefaultSdkGeneratorManagerConfiguration()
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
		cfg := config.DefaultSdkGeneratorManagerConfiguration()
		cfg.Input = filesystem.FilePathJoin(localFs, "testdata", "agent.yml")
		cfg.Output = tmpDir
		cfg.ToolVersion = field.ToOptionalString("0.31.0")

		err := Generate(ctx, cfg)
		require.NoError(t, err)

		assert.NoDirExists(t, filesystem.FilePathJoin(localFs, tmpDir, "actions"))
		assert.NoFileExists(t, filesystem.FilePathJoin(localFs, tmpDir, ".agents", "skills", "using-fpga-service-agent-in-a-github-action", "SKILL.md"))
	})
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
