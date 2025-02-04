package data

import (
	"github.com/pkg/errors"
	"github.com/viant/datly/v0/shared"
)

//ValidateCardinality checks if cardinality is valid
func ValidateCardinality(cardinality string) error {
	switch cardinality {
	case shared.CardinalityMany, shared.CardinalityOne:
	default:
		return errors.Errorf("unsupported cardinality: '%s', supported: %v, %v", cardinality, shared.CardinalityMany, shared.CardinalityOne)
	}
	return nil
}
