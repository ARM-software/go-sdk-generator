package generator

import (
	"context"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	jsonutils "github.com/ARM-software/golang-utils/utils/serialization/json" //nolint:misspell
	yamlutils "github.com/ARM-software/golang-utils/utils/serialization/yaml" //nolint:misspell
	jsonschemavalidation "github.com/ARM-software/golang-utils/utils/validation/jsonschema"
)

func (j *SDK) UnmarshalYAML(b []byte) error {
	data, err := yamlutils.ToJSON(b)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not convert YAML to JSON")
	}

	return j.UnmarshalJSON(data)
}

func validateManifestContent(ctx context.Context, extension string, content []byte) error {
	schemaFS, err := filesystem.NewEmbedFileSystem(&embeddedSchemaFS)
	if err != nil {
		return commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not prepare embedded schema filesystem for manifest validation")
	}

	options := []jsonschemavalidation.SchemaOption{
		jsonschemavalidation.WithFilesystem(schemaFS),
		jsonschemavalidation.WithLocalPath(embeddedSchemaFile),
		jsonschemavalidation.WithTitle("SDK"),
		jsonschemavalidation.IgnoreYAMLAliases(),
	}

	if yamlutils.IsYAMLFile(extension) {
		return jsonschemavalidation.ValidateRawYAMLAgainstSchemaOptions(ctx, content, options...)
	}
	return jsonschemavalidation.ValidateRawJSONAgainstSchemaOptions(ctx, content, options...)
}

func unmarshalManifestContent(extension string, content []byte, sdk *SDK) error {
	if sdk == nil {
		return commonerrors.UndefinedVariable("sdk")
	}
	if yamlutils.IsYAMLFile(extension) {
		return yamlutils.Unmarshal(content, sdk)
	}
	return jsonutils.Unmarshal(content, sdk)
}

func Validate(ctx context.Context, path string) (validationErr, err error) {
	localFs := filesystem.GetGlobalFileSystem()

	data, err := filesystem.ReadFile(path)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not read file '%v'", path)
		return nil, err
	}

	ext := filesystem.FilePathExt(localFs, path)
	if err = validateManifestContent(ctx, ext, data); err != nil {
		return err, nil
	}

	var sdk SDK
	err = unmarshalManifestContent(ext, data, &sdk)
	if err != nil {
		return err, nil
	}

	if err = ProcessExtraImports(ctx, &sdk); err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate manifest preprocessing for extra imports"), nil
	}
	if err = ProcessRepositoryAndModuleMetadata(ctx, &sdk); err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate manifest preprocessing for repository and module metadata"), nil
	}
	if err = ProcessAdditionalTypePackages(ctx, &sdk); err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate manifest preprocessing for additional type packages"), nil
	}
	if err = ProcessBindings(ctx, &sdk); err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate manifest preprocessing for bindings"), nil
	}
	if err = ProcessArguments(ctx, &sdk); err != nil {
		return commonerrors.WrapError(commonerrors.ErrInvalid, err, "failed to validate manifest preprocessing for arguments"), nil
	}

	return nil, nil
}
