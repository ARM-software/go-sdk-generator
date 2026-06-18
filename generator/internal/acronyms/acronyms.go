package acronyms

import (
	"context"
	"embed"
	"sync"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	yamlserialization "github.com/ARM-software/golang-utils/utils/serialization/yaml" //nolint:misspell
	"github.com/ARM-software/golang-utils/utils/strings/casing"
	validationjsonschema "github.com/ARM-software/golang-utils/utils/validation/jsonschema"
)

// This package applies known initialism rules while honouring embedded
// exception lists for words that should not be rewritten.
//
// see https://google.github.io/styleguide/go/decisions.html#initialisms
//
//go:generate go tool go-jsonschema -p acronyms ./special-words.schema.json -o ./types.gen.go --struct-name-from-title --capitalization AI,API,DB,DDOS,FPGA,GRPC,HTTP,HTTPS,ID,IOS,JSON,TTL,URL,XML,YAML,YML,AFAAS

//go:embed *
var acronymFS embed.FS

const (
	wordsFileName  = "special-words.yaml"
	schemaFileName = "special-words.schema.json"
	schemaURL      = "embedded://special-words.schema.json"
)

var (
	loadRulesOnce     sync.Once
	replacerSingleton *casing.Replacer
	loadRulesErr      error
)

// Replacer returns the lazily loaded acronym-aware casing replacer used by this
// package.
func Replacer() (*casing.Replacer, error) {
	loadRulesOnce.Do(func() {
		replacerSingleton, loadRulesErr = loadRules(context.Background())
	})
	return replacerSingleton, loadRulesErr
}

// ApplyKnownAcronyms rewrites identifier-like strings using the built-in Go
// initialism rules plus the additional acronym rules loaded from the embedded
// data files in this package.
func ApplyKnownAcronyms(value string) (applied string, err error) {
	replacer, err := Replacer()
	if err != nil {
		return
	}
	applied = replacer.Replace(value)
	return
}

// loadRules loads the embedded YAML rule file, validates it against the
// embedded JSON Schema, converts it to casing rules, and returns a configured
// casing.Replacer.
func loadRules(ctx context.Context) (*casing.Replacer, error) {
	fs, err := filesystem.NewEmbedFileSystem(&acronymFS)
	if err != nil {
		return nil, commonerrors.WrapError(commonerrors.ErrUnexpected, err, "failed to create embedded filesystem for acronym rules")
	}

	raw, err := fs.ReadFileWithContext(ctx, wordsFileName)
	if err != nil {
		return nil, commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to read embedded acronym list file")
	}

	err = validationjsonschema.ValidateRawYAMLAgainstSchemaOptions(
		ctx,
		raw,
		validationjsonschema.WithTitle("special acronym words schema"),
		validationjsonschema.WithLocalPath(schemaFileName),
		validationjsonschema.WithID(schemaURL),
		validationjsonschema.WithFilesystem(fs),
	)
	if err != nil {
		return nil, commonerrors.DescribeCircumstance(err, "failed to validate acronym list against schema")
	}

	var specialWords SpecialWordsSchemaJSON
	if err = yamlserialization.UnmarshallWithContext(ctx, raw, &specialWords); err != nil {
		return nil, commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed to decode embedded acronym list")
	}
	rules := collection.MapRef[SpecialWordsSchemaJSONAcronymsElem, casing.Rule](specialWords.Acronyms, func(entry *SpecialWordsSchemaJSONAcronymsElem) *casing.Rule {
		return &casing.Rule{
			Token:       entry.Token,
			Replacement: entry.Replacement,
			Exceptions:  entry.Exceptions,
		}
	})

	return casing.NewReplacer(append(casing.InitialismRules, rules...)...)
}
