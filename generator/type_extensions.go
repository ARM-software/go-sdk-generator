package generator

import (
	"path/filepath"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"sigs.k8s.io/yaml"
)

func (j *SDK) UnmarshalYAML(b []byte) error {
	data, err := yaml.YAMLToJSON(b)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could convert YAML to JSON")
	}

	return j.UnmarshalJSON(data)
}

func Validate(path string) (validationErr, err error) {
	data, err := filesystem.ReadFile(path)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not read file '%v'", path)
		return nil, err
	}

	var schema SDK
	if ext := filepath.Ext(path); ext == ".yml" || ext == ".yaml" {
		err = schema.UnmarshalYAML(data)
	} else {
		err = schema.UnmarshalJSON(data)
	}
	if err != nil {
		return err, nil
	}

	return nil, nil
}
