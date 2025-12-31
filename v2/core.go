package poxxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"reflect"
)

// Validator validates a value of type T.
type Validator[T any] interface {
	Validate(value T, field string) error
	Rule() string
	WithMessage(msg string) Validator[T]
}

// Transformer transforms a value of type T.
type Transformer[T any] interface {
	Transform(value T) (T, error)
	Name() string
}

// Decoder decodes data into a target.
type Decoder interface {
	Decode(v any) error
}

// field is the internal interface for schema fields.
type field interface {
	name() string
	assign(data map[string]any, path string, schema *Schema) error
	validate(path string, schema *Schema) *ValidationErrors
	description() string
	isAbsent() bool
	isNull() bool
}

// FieldInfo contains metadata about a field for documentation generation.
type FieldInfo struct {
	Name        string            // Field name
	Type        string            // Go type name (string, int, []string, etc.)
	Description string            // Field description
	Required    bool              // Whether field is required
	Nullable    bool              // Whether field can be null (pointer types)
	IsSlice     bool              // Whether field is a slice
	IsMap       bool              // Whether field is a map
	IsStruct    bool              // Whether field is a nested struct
	Validators  []ValidatorInfo   // Validator metadata
	Children    []FieldInfo       // Nested fields for struct types
	Default     any               // Default value if set
}

// ValidatorInfo contains metadata about a validator.
type ValidatorInfo struct {
	Rule   string         // Validator rule name (required, min, max, etc.)
	Params map[string]any // Validator parameters (min value, max value, pattern, etc.)
}

// documentable is implemented by fields that can provide documentation metadata.
type documentable interface {
	fieldInfo() FieldInfo
}

// Schema orchestrates field assignment and validation.
type Schema struct {
	fields      []field
	logger      *slog.Logger
	strictTypes bool
	fieldMap    map[string]field
	catalog     *MessageCatalog
}

// NewSchema creates a new schema with the given fields.
func NewSchema(fields ...field) *Schema {
	s := &Schema{
		fields:   fields,
		fieldMap: make(map[string]field),
	}
	for _, f := range fields {
		s.fieldMap[f.name()] = f
	}
	return s
}

// Add adds a field to the schema. Used in callbacks for nested structures.
func (s *Schema) Add(f field) {
	s.fields = append(s.fields, f)
	s.fieldMap[f.name()] = f
}

// WithLogger sets a logger for debug tracing.
func (s *Schema) WithLogger(logger *slog.Logger) *Schema {
	s.logger = logger
	return s
}

// StrictTypes enables strict type checking (no automatic conversion).
func (s *Schema) StrictTypes(strict bool) *Schema {
	s.strictTypes = strict
	return s
}

// WithCatalog sets a message catalog for i18n support.
func (s *Schema) WithCatalog(catalog *MessageCatalog) *Schema {
	s.catalog = catalog
	return s
}

// Catalog returns the current message catalog (default if not set).
func (s *Schema) Catalog() *MessageCatalog {
	if s.catalog == nil {
		return defaultCatalog
	}
	return s.catalog
}

// Parse decodes from an io.Reader and validates.
func (s *Schema) Parse(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return s.ParseDecoder(decoder)
}

// ParseDecoder decodes using a custom decoder and validates.
func (s *Schema) ParseDecoder(d Decoder) error {
	var data map[string]any
	if err := d.Decode(&data); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return s.ParseMap(data)
}

// ParseMap validates from a map directly.
func (s *Schema) ParseMap(data map[string]any) error {
	if err := s.detectCircularReferences(); err != nil {
		return err
	}

	// Phase 1: Assignment
	for _, f := range s.fields {
		s.log("assigning", "field", f.name())
		if err := f.assign(data, "", s); err != nil {
			return err
		}
	}

	// Phase 2: Validation
	errs := &ValidationErrors{}
	for _, f := range s.fields {
		s.log("validating", "field", f.name())
		if fieldErrs := f.validate("", s); fieldErrs != nil {
			errs.Merge(fieldErrs)
		}
	}

	if errs.HasErrors() {
		// Translate errors if a catalog is set
		if s.catalog != nil {
			return errs.Translate(s.catalog)
		}
		return errs
	}
	return nil
}

// GetField returns a field by name for checking IsAbsent/IsNull.
func (s *Schema) GetField(name string) field {
	return s.fieldMap[name]
}

// Get returns the value of a field by name.
func (s *Schema) Get(name string) any {
	f := s.fieldMap[name]
	if f == nil {
		return nil
	}
	// Use reflection to get the value from the field
	rv := reflect.ValueOf(f)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	destField := rv.FieldByName("dest")
	if !destField.IsValid() {
		return nil
	}
	if destField.Kind() == reflect.Ptr && !destField.IsNil() {
		return destField.Elem().Interface()
	}
	return nil
}

func (s *Schema) log(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Debug("poxxy: "+msg, args...)
	}
}

func (s *Schema) detectCircularReferences() error {
	// Track visited types to detect cycles
	visited := make(map[reflect.Type]bool)
	for _, f := range s.fields {
		if err := s.checkFieldForCycles(f, visited, nil); err != nil {
			return err
		}
	}
	return nil
}

func (s *Schema) checkFieldForCycles(f field, visited map[reflect.Type]bool, path []string) error {
	rv := reflect.ValueOf(f)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Check if this is a struct field with nested schema
	if sf, ok := f.(*structField); ok {
		t := reflect.TypeOf(sf.dest).Elem()
		newPath := append(path, sf.fieldName)

		if visited[t] {
			return fmt.Errorf("circular reference detected: %s", joinPath(newPath))
		}
		visited[t] = true

		// Check nested fields
		for _, nested := range sf.schema.fields {
			if err := s.checkFieldForCycles(nested, visited, newPath); err != nil {
				return err
			}
		}
		delete(visited, t)
	}

	return nil
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " -> " + parts[i]
	}
	return result
}

// fieldPath builds the full field path for error reporting.
func fieldPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

// indexPath builds the path for slice/array elements.
func indexPath(prefix string, index int) string {
	return fmt.Sprintf("%s[%d]", prefix, index)
}

// FieldInfos returns metadata about all fields for documentation generation.
func (s *Schema) FieldInfos() []FieldInfo {
	infos := make([]FieldInfo, 0, len(s.fields))
	for _, f := range s.fields {
		if doc, ok := f.(documentable); ok {
			infos = append(infos, doc.fieldInfo())
		}
	}
	return infos
}
