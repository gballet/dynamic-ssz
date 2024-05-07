package dynssz

import (
	"reflect"

	fastssz "github.com/ferranbt/fastssz"
)

var sszMarshalerType = reflect.TypeOf((*fastssz.Marshaler)(nil)).Elem()
var sszUnmarshalerType = reflect.TypeOf((*fastssz.Unmarshaler)(nil)).Elem()
var sszHashRootType = reflect.TypeOf((*fastssz.HashRoot)(nil)).Elem()

// fastsszCompatibility holds information about a type's compatibility with fastssz's static encoding and decoding methods.
// It is used to determine whether a type can leverage fastssz's efficient, static code paths or if it must be handled dynamically
// due to the presence of non-default specification values or the lack of necessary interface implementations.
//
// Fields:
//   - isMarshaler: Indicates whether the type implements the fastssz Marshaler interface.
//   - isUnmarshaler: Indicates whether the type implements the fastssz Unmarshaler interface.
//   - isHashRoot: Indicates whether the type implements the fastssz HashRoot interface.
//   - hasDynamicSpecValues: Indicates the presence of dynamically applied specification values that deviate from the default
//     specifications. A true value here suggests that, despite potentially implementing the required interfaces for static processing,
//     the type may still need to be handled dynamically due to these spec values affecting its size or structure.
type fastsszCompatibility struct {
	isMarshaler          bool
	isUnmarshaler        bool
	isHashRoot           bool
	hasDynamicSpecValues bool
}

// getFastsszCompatibility evaluates the compatibility of a given type with fastssz, determining whether the type and its nested
// structures can be efficiently encoded/decoded using fastssz's static code generation approach. This assessment includes checking
// for the implementation of the Marshaler/Unmarshaler interfaces and the absence of non-default dynamically applied specification
// values within the type hierarchy. For performance optimization, the results of this compatibility check are cached, allowing
// `dynssz` to quickly reference these results in future operations without the need to re-evaluate the same types repeatedly.
//
// Parameters:
// - targetType: The reflect.Type of the value being assessed for fastssz compatibility. This type, along with its nested
//   or referenced types, is evaluated to ensure it aligns with fastssz's requirements for static encoding and decoding.
//
// Returns:
// - A pointer to a fastsszCompatibility struct, which contains detailed information about the compatibility status, including
//   whether the type implements required interfaces and lacks dynamically applied non-default spec values.
// - An error if the compatibility check encounters issues, such as reflection errors or the presence of unsupported type configurations
//   that would prevent the use of fastssz for encoding or decoding.

func (d *DynSsz) getFastsszCompatibility(targetType reflect.Type) (*fastsszCompatibility, error) {
	if cachedCompatibility := d.fastsszCompatCache[targetType]; cachedCompatibility != nil {
		return cachedCompatibility, nil
	}

	_, hasSpecVals, err := d.getSszSize(targetType, []sszSizeHint{})
	if err != nil {
		return nil, err
	}

	targetPtrType := reflect.New(targetType).Type()
	compatibility := &fastsszCompatibility{
		isMarshaler:          targetPtrType.Implements(sszMarshalerType),
		isUnmarshaler:        targetPtrType.Implements(sszUnmarshalerType),
		isHashRoot:           targetPtrType.Implements(sszHashRootType),
		hasDynamicSpecValues: hasSpecVals,
	}
	d.fastsszCompatCache[targetType] = compatibility
	return compatibility, nil
}
