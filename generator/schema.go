package generator

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"sigs.k8s.io/yaml"
)

const embeddedSchemaFile = "schema/schema.yml"

var (
	//go:embed schema/schema.yml
	embeddedSchemaFS embed.FS
)

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
