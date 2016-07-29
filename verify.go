package bahamut

import (
	"fmt"

	"github.com/aporeto-inc/elemental"
)

// ValidateAdvancedSpecification verifies advanced specifications attributes like ReadOnly and CreationOnly.
//
// For instance, it will check if the given Manipulable has field marked as
// readonly, that it has not changed according to the db.
func ValidateAdvancedSpecification(obj elemental.AttributeSpecifiable, pristine elemental.AttributeSpecifiable, op Operation) elemental.Errors {

	errors := elemental.NewErrors()

	for _, field := range extractFieldNames(obj) {

		spec := obj.SpecificationForAttribute(field)

		// Read Only
		if spec.ReadOnly &&
			!fieldValuesEquals(field, obj, pristine) {
			errors = append(
				errors,
				elemental.NewError(
					"Read Only Error",
					fmt.Sprintf("Field %s is read only. You cannot set the value.", field),
					"specification",
					3001,
				),
			)
		}

		// Create Only
		if spec.CreationOnly &&
			op != OperationCreate &&
			!fieldValuesEquals(field, obj, pristine) {
			errors = append(
				errors,
				elemental.NewError(
					"Creation Only Error",
					fmt.Sprintf("Field %s can only be set during creation. You cannot change the value.", field),
					"specification",
					3001,
				),
			)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
