package generator

import (
	"context"
	fSys "io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"text/template"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Arm-Debug/go-sdk-generator/generator/config"
)

func TestFormatImports(t *testing.T) {
	t.Run("Valid import add", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.go")
		tmpFileContent := "package main\n\nfunc main() {\nfmt.Println(\"Hello World\")\n}\n"

		err := filesystem.WriteFile(tmpFile, []byte(tmpFileContent), 0655)
		require.NoError(t, err)

		cleaned, err := formatImports(context.Background(), tmpFile)
		assert.NoError(t, err)
		assert.Equal(t, "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}\n", string(cleaned))
	})

	t.Run("Invalid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.go")
		tmpFileContent := "main {\nfmt.Println(\"Hello World\")\n}\n"

		err := filesystem.WriteFile(tmpFile, []byte(tmpFileContent), 0655)
		require.NoError(t, err)

		cleaned, err := formatImports(context.Background(), tmpFile)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
		assert.Empty(t, cleaned)
	})

	t.Run("No file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.go")

		cleaned, err := formatImports(context.Background(), tmpFile)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, cleaned)
	})

	t.Run("Context cancel", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.go")

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
				"{{ .Data | replaceUnderscores }}",
				"{{ .Data | lower }}",
				"{{ .Data | notSeenRootCmd }}",
			}, "\n"))

		require.NoError(t, err)

		var b strings.Builder
		err = tmpl.ExecuteTemplate(&b, "test", struct{ Data string }{"this_is_A_WORD"})
		require.NoError(t, err)

		assert.Equal(t, b.String(), "This_is_A_WORD\nthis-is-A-WORD\nthis_is_a_word\ntrue")
		assert.Contains(t, rootCmds, "this_is_A_WORD") // For notSeenRootCmd
	})

	t.Run("resetNotSeenRootCmd", func(t *testing.T) {
		tmpl, err := template.
			New("test").
			Funcs(templateFuncMap).
			Parse(strings.Join([]string{
				"{{ .Data | notSeenRootCmd }}",
				"{{ resetNotSeenRootCmd }}",
			}, "\n"))

		require.NoError(t, err)

		var b strings.Builder
		err = tmpl.ExecuteTemplate(&b, "test", struct{ Data string }{"this_is_A_WORD"})
		require.NoError(t, err)

		assert.Empty(t, rootCmds)
	})

	t.Run("Zero", func(t *testing.T) {
		tmpl, err := template.
			New("test").
			Funcs(templateFuncMap).
			Parse(strings.Join([]string{
				"{{ \"string\" | zero }}",
				"{{ \"bool\" | zero }}",
				"{{ \"int\" | zero }}",
				"{{ \"int32\" | zero }}",
				"{{ \"float\" | zero }}",
				"{{ \"float32\" | zero }}",
				"{{ \"aaaaaa\" | zero }}",
			}, "\n"))

		require.NoError(t, err)

		var b strings.Builder
		err = tmpl.ExecuteTemplate(&b, "test", nil)
		require.NoError(t, err)

		assert.Equal(t, b.String(), "\"\"\nfalse\n0\n0\n0.0\n0.0\nnil")
	})

}

func TestPopulateTemplate(t *testing.T) {
	t.Run("Valid import add", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filepath.Join(tmpDir, "out1", "out2", "out.go")
		require.NoDirExists(t, filepath.Dir(out))

		ctx := context.Background()
		err := populateTemplate(ctx, &DataModel{}, filepath.Join("templates", "test.go.tmpl"), out)
		assert.NoError(t, err)

		content, err := filesystem.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}\n", string(content))
	})

	t.Run("Valid import add (not .go so no import tidy)", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filepath.Join(tmpDir, "out1", "out2")
		require.NoDirExists(t, filepath.Dir(out))

		ctx := context.Background()
		err := populateTemplate(ctx, &DataModel{}, filepath.Join("templates", "test.go.tmpl"), out)
		assert.NoError(t, err)

		content, err := filesystem.ReadFile(out)
		require.NoError(t, err)
		assert.Equal(t, "package main\n\nfunc main() {\n\tfmt.Println(\"Hello World\")\n}\n", string(content))
	})

	t.Run("Missing template", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filepath.Join(tmpDir, "out")

		ctx := context.Background()
		err := populateTemplate(ctx, &DataModel{}, filepath.Join("templates", "test-not-exist.go.tmpl"), out)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
		assert.NoFileExists(t, out)
	})

	t.Run("Cancel Context", func(t *testing.T) {
		tmpDir := t.TempDir()
		out := filepath.Join(tmpDir, "out")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := populateTemplate(ctx, &DataModel{}, filepath.Join("templates", "test.go.tmpl"), out)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
		assert.NoFileExists(t, out)
	})
}

func checkFilesAreEqual(t assert.TestingT, actualDir string, expectedDir string, localFs filesystem.FS, ignoreFiles ...string) {
	err := localFs.WalkWithContext(context.Background(), expectedDir, func(path string, info fSys.FileInfo, err error) error {
		if slices.ContainsFunc(ignoreFiles, func(ignore string) bool {
			return strings.HasSuffix(path, filepath.Clean(ignore))
		}) {
			return nil
		}

		assert.NoError(t, err)

		relativePath, err := filepath.Rel(expectedDir, path)
		assert.NoError(t, err)

		expectedFilePath := filepath.Join(actualDir, relativePath)

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
			tmpDir := t.TempDir()
			ctx := context.Background()
			cfg := config.DefaultSdkGeneratorManagerConfiguration()
			cfg.Input = filepath.Join("testdata", test.recipe)
			cfg.Output = tmpDir

			err := Generate(ctx, cfg)
			if test.err != nil {
				errortest.AssertError(t, err, test.err)
			} else {
				assert.NoError(t, err)

				tmpDir2 := t.TempDir()
				zipped := filepath.Join("testdata", test.regressionSrc)
				_, err = filesystem.Unzip(zipped, tmpDir2)
				require.NoError(t, err)

				checkFilesAreEqual(t, tmpDir, tmpDir2, filesystem.NewStandardFileSystem(), "pyproject.toml")
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
			Input:  filepath.Join("testdata", "verify.json"),
			Output: "aaaa",
		}
		err := Generate(ctx, cfg)
		errortest.AssertError(t, err, commonerrors.ErrCancelled)
	})

	t.Run("Invalid specification", func(t *testing.T) {
		ctx := context.Background()
		cfg := &config.SdkGeneratorManagerConfiguration{
			Input:  filepath.Join("testdata", "i-dont-exist.json"),
			Output: "bbb",
		}

		err := Generate(ctx, cfg)
		errortest.AssertError(t, err, commonerrors.ErrUnexpected)
	})
}
