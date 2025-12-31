package poxxy

// =============================================================================
// Field Templates - Reusable Field Configurations
// =============================================================================

// FieldTemplate is a reusable field configuration that can be bound to different targets.
type FieldTemplate[T any] struct {
	validators   []Validator[T]
	transformers []Transformer[T]
	defaultValue *T
	defaultEmpty bool
	strictType   bool
	desc         string
}

// Template creates a new field template.
func Template[T any]() *FieldTemplate[T] {
	return &FieldTemplate[T]{}
}

// Validate adds validators to the template.
func (t *FieldTemplate[T]) Validate(validators ...Validator[T]) *FieldTemplate[T] {
	t.validators = append(t.validators, validators...)
	return t
}

// Transform adds transformers to the template.
func (t *FieldTemplate[T]) Transform(transformers ...Transformer[T]) *FieldTemplate[T] {
	t.transformers = append(t.transformers, transformers...)
	return t
}

// Default sets a default value.
func (t *FieldTemplate[T]) Default(value T) *FieldTemplate[T] {
	t.defaultValue = &value
	return t
}

// DefaultOnEmpty sets a default value to use if the field is empty.
func (t *FieldTemplate[T]) DefaultOnEmpty(value T) *FieldTemplate[T] {
	t.defaultValue = &value
	t.defaultEmpty = true
	return t
}

// StrictType enables strict type checking.
func (t *FieldTemplate[T]) StrictType() *FieldTemplate[T] {
	t.strictType = true
	return t
}

// Describe sets the description.
func (t *FieldTemplate[T]) Describe(desc string) *FieldTemplate[T] {
	t.desc = desc
	return t
}

// Bind creates a field from this template bound to the given name and destination.
func (t *FieldTemplate[T]) Bind(name string, dest *T) *FieldBuilder[T] {
	fb := Field(name, dest)
	fb.validators = append(fb.validators, t.validators...)
	fb.transformers = append(fb.transformers, t.transformers...)
	if t.defaultValue != nil {
		if t.defaultEmpty {
			fb.DefaultOnEmpty(*t.defaultValue)
		} else {
			fb.Default(*t.defaultValue)
		}
	}
	if t.strictType {
		fb.StrictType()
	}
	if t.desc != "" {
		fb.Describe(t.desc)
	}
	return fb
}

// =============================================================================
// Pointer Field Templates
// =============================================================================

// PointerTemplate is a reusable template for optional fields.
type PointerTemplate[T any] struct {
	validators   []Validator[T]
	transformers []Transformer[T]
	defaultValue *T
	defaultEmpty bool
	strictType   bool
	desc         string
}

// PtrTemplate creates a new pointer field template.
func PtrTemplate[T any]() *PointerTemplate[T] {
	return &PointerTemplate[T]{}
}

// Validate adds validators to the template.
func (t *PointerTemplate[T]) Validate(validators ...Validator[T]) *PointerTemplate[T] {
	t.validators = append(t.validators, validators...)
	return t
}

// Transform adds transformers to the template.
func (t *PointerTemplate[T]) Transform(transformers ...Transformer[T]) *PointerTemplate[T] {
	t.transformers = append(t.transformers, transformers...)
	return t
}

// Default sets a default value.
func (t *PointerTemplate[T]) Default(value T) *PointerTemplate[T] {
	t.defaultValue = &value
	return t
}

// DefaultOnEmpty sets a default value to use if the field is empty.
func (t *PointerTemplate[T]) DefaultOnEmpty(value T) *PointerTemplate[T] {
	t.defaultValue = &value
	t.defaultEmpty = true
	return t
}

// StrictType enables strict type checking.
func (t *PointerTemplate[T]) StrictType() *PointerTemplate[T] {
	t.strictType = true
	return t
}

// Describe sets the description.
func (t *PointerTemplate[T]) Describe(desc string) *PointerTemplate[T] {
	t.desc = desc
	return t
}

// Bind creates a pointer field from this template bound to the given name and destination.
func (t *PointerTemplate[T]) Bind(name string, dest **T) *PointerBuilder[T] {
	pb := Pointer(name, dest)
	pb.validators = append(pb.validators, t.validators...)
	pb.transformers = append(pb.transformers, t.transformers...)
	if t.defaultValue != nil {
		if t.defaultEmpty {
			pb.DefaultOnEmpty(*t.defaultValue)
		} else {
			pb.Default(*t.defaultValue)
		}
	}
	if t.strictType {
		pb.StrictType()
	}
	if t.desc != "" {
		pb.Describe(t.desc)
	}
	return pb
}

// =============================================================================
// Slice Field Templates
// =============================================================================

// SliceTemplate is a reusable template for slice fields.
type SliceTemplate[T any] struct {
	validators   []Validator[[]T]
	itemCallback func(*T, *Schema)
	desc         string
}

// SliceTempl creates a new slice field template.
func SliceTempl[T any]() *SliceTemplate[T] {
	return &SliceTemplate[T]{}
}

// Validate adds validators to the template.
func (t *SliceTemplate[T]) Validate(validators ...Validator[[]T]) *SliceTemplate[T] {
	t.validators = append(t.validators, validators...)
	return t
}

// WithItemSchema sets a callback for defining schema per item.
func (t *SliceTemplate[T]) WithItemSchema(callback func(*T, *Schema)) *SliceTemplate[T] {
	t.itemCallback = callback
	return t
}

// Describe sets the description.
func (t *SliceTemplate[T]) Describe(desc string) *SliceTemplate[T] {
	t.desc = desc
	return t
}

// Bind creates a slice field from this template bound to the given name and destination.
func (t *SliceTemplate[T]) Bind(name string, dest *[]T) *SliceBuilder[T] {
	var sb *SliceBuilder[T]
	if t.itemCallback != nil {
		sb = Slice(name, dest, t.itemCallback)
	} else {
		sb = Slice[T](name, dest, nil)
	}
	sb.validators = append(sb.validators, t.validators...)
	if t.desc != "" {
		sb.Describe(t.desc)
	}
	return sb
}

// =============================================================================
// Struct Field Templates
// =============================================================================

// StructTemplate is a reusable template for nested struct fields.
type StructTemplate[T any] struct {
	callback func(*T, *Schema)
	desc     string
}

// StructTempl creates a new struct field template.
// The callback receives the target pointer and schema to define nested fields.
func StructTempl[T any](callback func(*T, *Schema)) *StructTemplate[T] {
	return &StructTemplate[T]{
		callback: callback,
	}
}

// Describe sets the description.
func (t *StructTemplate[T]) Describe(desc string) *StructTemplate[T] {
	t.desc = desc
	return t
}

// Bind creates a struct field from this template bound to the given name and destination.
func (t *StructTemplate[T]) Bind(name string, dest *T) field {
	// Wrap the callback to adapt (target, schema) -> (schema) signature
	wrappedCallback := func(s *Schema) {
		t.callback(dest, s)
	}
	return Struct(name, dest, wrappedCallback)
}

// =============================================================================
// Common Template Presets
// =============================================================================

// EmailTemplate is a preset for email fields with common validation.
var EmailTemplate = Template[string]().
	Transform(TrimSpace(), ToLower()).
	Validate(Email()).
	Describe("Email address")

// RequiredEmailTemplate is a preset for required email fields.
var RequiredEmailTemplate = Template[string]().
	Transform(TrimSpace(), ToLower()).
	Validate(Required[string](), Email()).
	Describe("Required email address")

// URLTemplate is a preset for URL fields with common validation.
var URLTemplate = Template[string]().
	Transform(TrimSpace()).
	Validate(URL()).
	Describe("URL")

// UUIDTemplate is a preset for UUID fields with common validation.
var UUIDTemplate = Template[string]().
	Transform(TrimSpace(), ToLower()).
	Validate(UUID()).
	Describe("UUID")

// NonEmptyStringTemplate is a preset for non-empty string fields.
var NonEmptyStringTemplate = Template[string]().
	Transform(TrimSpace()).
	Validate(Required[string]()).
	Describe("Required non-empty string")

// PositiveIntTemplate is a preset for positive integer fields.
var PositiveIntTemplate = Template[int]().
	Validate(Min(1)).
	Describe("Positive integer")

// NonNegativeIntTemplate is a preset for non-negative integer fields.
var NonNegativeIntTemplate = Template[int]().
	Validate(Min(0)).
	Describe("Non-negative integer")
