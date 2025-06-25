package generator

import (
	"context"
	"fmt"
	"testing"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessBindings(t *testing.T) {
	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, ProcessBindings(ctx, nil), commonerrors.ErrCancelled)
	})
	t.Run("no spec", func(t *testing.T) {
		errortest.AssertError(t, ProcessBindings(context.Background(), nil), commonerrors.ErrUndefined)
	})
	t.Run("happy", func(t *testing.T) {
		var sdk SDK
		err := faker.FakeData(&sdk)
		require.NoError(t, err)
		bindingTypes := fmt.Sprintf("types.%v-%v", faker.Word(), faker.Word())
		expectedBinding := Binding{
			Arg:             field.ToOptionalString(faker.Name()),
			Description:     field.ToOptionalString(faker.Sentence()),
			LangType:        field.ToOptionalString(bindingTypes),
			MapstructureTag: field.ToOptionalString(faker.Word()),
			Name:            field.ToOptionalString(faker.FirstName()),
		}

		originalValidTypes := sdk.Specification.ExtraValidTypes
		assert.NotContains(t, originalValidTypes, bindingTypes)
		configuration := ConfigurationOption{
			LangType: faker.Word(),
			Bindings: []Binding{expectedBinding},
		}
		sdk.Specification.GlobalConfiguration = append(sdk.Specification.GlobalConfiguration, configuration)
		require.NoError(t, ProcessBindings(context.Background(), &sdk))
		for i := range originalValidTypes {
			assert.Contains(t, sdk.Specification.ExtraValidTypes, originalValidTypes[i])
		}
		assert.Contains(t, sdk.Specification.ExtraValidTypes, bindingTypes)
		assert.Contains(t, sdk.Specification.ExtraValidTypes, configuration.LangType)
	})

}

func TestProcessArguments(t *testing.T) {
	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		errortest.AssertError(t, ProcessArguments(ctx, nil), commonerrors.ErrCancelled)
	})

	t.Run("no sdk", func(t *testing.T) {
		errortest.AssertError(t, ProcessArguments(context.Background(), nil), commonerrors.ErrUndefined)
	})

	t.Run("no CommandLineCommands", func(t *testing.T) {
		var sdk SDK
		err := faker.FakeData(&sdk)
		require.NoError(t, err)
		// Remove commands
		sdk.Specification.CommandLineCommands = nil
		assert.NoError(t, ProcessArguments(context.Background(), &sdk))
	})

	t.Run("empty CommandLineCommands.Functions", func(t *testing.T) {
		var sdk SDK
		err := faker.FakeData(&sdk)
		require.NoError(t, err)
		sdk.Specification.CommandLineCommands = &CommandLineCommands{
			Functions: []CliFunction{},
		}
		assert.NoError(t, ProcessArguments(context.Background(), &sdk))
	})

	t.Run("error from processArguments in bindings", func(t *testing.T) {
		cliFunc := CliFunction{
			Configuration: []FunctionConfigFields{
				{
					Bindings:        []Binding{{}},
					Description:     faker.Sentence(),
					LangType:        faker.Word(),
					MapstructureTag: faker.Word(),
					Name:            faker.FirstName(),
				},
			},
			Description: faker.Sentence(),
			Name:        faker.Name(),
		}
		sdk := &SDK{
			Specification: SDKSpecification{
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{cliFunc},
				},
				GlobalConfiguration: []ConfigurationOption{},
				Interface:           Interface{},
				MainModule:          "main",
				PackageName:         "pkg",
			},
		}
		err := ProcessArguments(context.Background(), sdk)
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
	})

	t.Run("error from processArguments in param", func(t *testing.T) {
		cliFunc := CliFunction{
			Configuration: []FunctionConfigFields{
				{
					// No Arg, no MapstructureTag
					Description:     faker.Sentence(),
					LangType:        faker.Word(),
					MapstructureTag: "", // empty tag, so will error
					Name:            faker.FirstName(),
				},
			},
			Description: faker.Sentence(),
			Name:        faker.Name(),
		}
		sdk := &SDK{
			Specification: SDKSpecification{
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{cliFunc},
				},
				GlobalConfiguration: []ConfigurationOption{},
				Interface:           Interface{},
				MainModule:          "main",
				PackageName:         "pkg",
			},
		}
		err := ProcessArguments(context.Background(), sdk)
		errortest.AssertError(t, err, commonerrors.ErrUndefined)
	})

	t.Run("duplicate short args triggers error", func(t *testing.T) {
		short := "x"
		arg1 := FunctionConfigFields{
			Arg:             field.ToOptionalString(faker.Word()),
			ArgShort:        &short,
			Description:     faker.Sentence(),
			LangType:        faker.Word(),
			MapstructureTag: faker.Word(),
			Name:            faker.FirstName(),
		}
		arg2 := FunctionConfigFields{
			Arg:             field.ToOptionalString(faker.Word()),
			ArgShort:        &short,
			Description:     faker.Sentence(),
			LangType:        faker.Word(),
			MapstructureTag: faker.Word(),
			Name:            faker.FirstName(),
		}
		cliFunc := CliFunction{
			Configuration: []FunctionConfigFields{arg1, arg2},
			Description:   faker.Sentence(),
			Name:          faker.Name(),
		}
		sdk := &SDK{
			Specification: SDKSpecification{
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{cliFunc},
				},
				GlobalConfiguration: []ConfigurationOption{},
				Interface:           Interface{},
				MainModule:          "main",
				PackageName:         "pkg",
			},
		}
		err := ProcessArguments(context.Background(), sdk)
		errortest.AssertError(t, err, commonerrors.ErrConflict)
	})

	t.Run("h and v short args triggers error", func(t *testing.T) {
		for _, arg := range []string{"h", "v"} {
			t.Run(arg, func(t *testing.T) {
				short := arg
				arg1 := FunctionConfigFields{
					Arg:             field.ToOptionalString(faker.Word()),
					ArgShort:        &short,
					Description:     faker.Sentence(),
					LangType:        faker.Word(),
					MapstructureTag: faker.Word(),
					Name:            faker.FirstName(),
				}
				arg2 := FunctionConfigFields{
					Arg:             field.ToOptionalString(faker.Word()),
					ArgShort:        &short,
					Description:     faker.Sentence(),
					LangType:        faker.Word(),
					MapstructureTag: faker.Word(),
					Name:            faker.FirstName(),
				}
				cliFunc := CliFunction{
					Configuration: []FunctionConfigFields{arg1, arg2},
					Description:   faker.Sentence(),
					Name:          faker.Name(),
				}
				sdk := &SDK{
					Specification: SDKSpecification{
						CommandLineCommands: &CommandLineCommands{
							Functions: []CliFunction{cliFunc},
						},
						GlobalConfiguration: []ConfigurationOption{},
						Interface:           Interface{},
						MainModule:          "main",
						PackageName:         "pkg",
					},
				}
				err := ProcessArguments(context.Background(), sdk)
				errortest.AssertError(t, err, commonerrors.ErrConflict)
			})
		}
	})

	t.Run("happy path: all good", func(t *testing.T) {
		short1 := "a"
		short2 := "b"
		arg1 := FunctionConfigFields{
			Arg:             field.ToOptionalString(faker.Word()),
			ArgShort:        &short1,
			Description:     faker.Sentence(),
			LangType:        faker.Word(),
			MapstructureTag: faker.Word(),
			Name:            faker.FirstName(),
		}
		arg2 := FunctionConfigFields{
			Arg:             field.ToOptionalString(faker.Word()),
			ArgShort:        &short2,
			Description:     faker.Sentence(),
			LangType:        faker.Word(),
			MapstructureTag: faker.Word(),
			Name:            faker.FirstName(),
		}
		cliFunc := CliFunction{
			Configuration: []FunctionConfigFields{arg1, arg2},
			Description:   faker.Sentence(),
			Name:          faker.Name(),
		}
		sdk := &SDK{
			Specification: SDKSpecification{
				CommandLineCommands: &CommandLineCommands{
					Functions: []CliFunction{cliFunc},
				},
				GlobalConfiguration: []ConfigurationOption{},
				Interface:           Interface{},
				MainModule:          "main",
				PackageName:         "pkg",
			},
		}
		assert.NoError(t, ProcessArguments(context.Background(), sdk))
	})
}
