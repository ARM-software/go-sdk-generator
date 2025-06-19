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
			Bindings: []Binding{expectedBinding},
		}
		sdk.Specification.GlobalConfiguration = append(sdk.Specification.GlobalConfiguration, configuration)
		require.NoError(t, ProcessBindings(context.Background(), &sdk))
		for i := range originalValidTypes {
			assert.Contains(t, sdk.Specification.ExtraValidTypes, originalValidTypes[i])
		}
		assert.Contains(t, sdk.Specification.ExtraValidTypes, bindingTypes)
	})

}
