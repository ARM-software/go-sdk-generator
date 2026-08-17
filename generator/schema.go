package generator

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"sync"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"sigs.k8s.io/yaml"
)

const embeddedSchemaFile = "schema/schema.yml"

var (
	//go:embed schema/schema.yml
	embeddedSchemaFS              embed.FS
	protectedCLIArgumentNamesOnce sync.Once
	protectedCLIArgumentNamesErr  error
	alwaysProtectedCLIArgNames    []string
	resultProtectedCLIArgNames    []string
)

const (
	alwaysProtectedCLIArgumentNamesKey = "x-always-protected-cli-argument-names"
	resultProtectedCLIArgumentNamesKey = "x-result-protected-cli-argument-names"
)

func embeddedProtectedCLIArgumentNames(key string) ([]string, error) {
	// embeddedProtectedCLIArgumentNames reads the protected CLI argument-name lists
	// from the embedded schema so schema metadata remains the source of truth for
	// runtime validation as well as schema validation.
	protectedCLIArgumentNamesOnce.Do(func() {
		schema, err := EmbeddedSchemaYAML()
		if err != nil {
			protectedCLIArgumentNamesErr = err
			return
		}

		var values map[string]any
		if err := yaml.Unmarshal(schema, &values); err != nil {
			protectedCLIArgumentNamesErr = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not parse embedded schema YAML")
			return
		}

		readStringList := func(name string) ([]string, error) {
			raw, ok := values[name]
			if !ok {
				return nil, commonerrors.UndefinedVariable(name)
			}
			entries, ok := raw.([]any)
			if !ok {
				return nil, commonerrors.Newf(commonerrors.ErrInvalid, "%s must be a list of strings", name)
			}
			if err := collection.ForAll(entries, func(entry any) error {
				if _, ok := entry.(string); !ok {
					return commonerrors.Newf(commonerrors.ErrInvalid, "%s must contain only strings", name)
				}
				return nil
			}); err != nil {
				return nil, err
			}
			return collection.Map(entries, func(entry any) string {
				return entry.(string)
			}), nil
		}

		alwaysProtectedCLIArgNames, err = readStringList(alwaysProtectedCLIArgumentNamesKey)
		if err != nil {
			protectedCLIArgumentNamesErr = err
			return
		}
		resultProtectedCLIArgNames, err = readStringList(resultProtectedCLIArgumentNamesKey)
		if err != nil {
			protectedCLIArgumentNamesErr = err
			return
		}
	})
	if protectedCLIArgumentNamesErr != nil {
		return nil, protectedCLIArgumentNamesErr
	}
	switch key {
	case alwaysProtectedCLIArgumentNamesKey:
		return append([]string{}, alwaysProtectedCLIArgNames...), nil
	case resultProtectedCLIArgumentNamesKey:
		return append([]string{}, resultProtectedCLIArgNames...), nil
	default:
		return nil, commonerrors.UndefinedVariable(key)
	}
}

func EmbeddedSchemaYAML() (schema []byte, err error) {
	schema, err = embeddedSchemaFS.ReadFile(embeddedSchemaFile)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not read embedded schema %q", embeddedSchemaFile)
		return
	}

	return
}

func EmbeddedSchemaJSON() (schema []byte, err error) {
	yamlSchema, err := EmbeddedSchemaYAML()
	if err != nil {
		return
	}

	jsonSchema, subErr := yaml.YAMLToJSON(yamlSchema)
	if subErr != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not convert embedded schema to JSON")
		return
	}

	var formatted bytes.Buffer
	subErr = json.Indent(&formatted, jsonSchema, "", "  ")
	if subErr != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, subErr, "could not format embedded schema JSON")
		return
	}

	schema = append(formatted.Bytes(), '\n')
	return
}

func WriteEmbeddedSchemaJSON(output io.Writer) (err error) {
	if output == nil {
		err = commonerrors.UndefinedVariable("output writer")
		return
	}

	schema, err := EmbeddedSchemaJSON()
	if err != nil {
		return
	}

	_, err = output.Write(schema)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not write embedded schema")
		return
	}

	return
}
