package poxxy

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// =============================================================================
// String Transformers
// =============================================================================

// trimSpaceTransformer trims whitespace from both ends of a string.
type trimSpaceTransformer struct{}

func (t trimSpaceTransformer) Transform(value string) (string, error) {
	return strings.TrimSpace(value), nil
}

func (t trimSpaceTransformer) Name() string {
	return "TrimSpace"
}

// TrimSpace returns a transformer that trims whitespace from both ends.
func TrimSpace() Transformer[string] {
	return trimSpaceTransformer{}
}

// toLowerTransformer converts a string to lowercase.
type toLowerTransformer struct{}

func (t toLowerTransformer) Transform(value string) (string, error) {
	return strings.ToLower(value), nil
}

func (t toLowerTransformer) Name() string {
	return "ToLower"
}

// ToLower returns a transformer that converts to lowercase.
func ToLower() Transformer[string] {
	return toLowerTransformer{}
}

// toUpperTransformer converts a string to uppercase.
type toUpperTransformer struct{}

func (t toUpperTransformer) Transform(value string) (string, error) {
	return strings.ToUpper(value), nil
}

func (t toUpperTransformer) Name() string {
	return "ToUpper"
}

// ToUpper returns a transformer that converts to uppercase.
func ToUpper() Transformer[string] {
	return toUpperTransformer{}
}

// titleCaseTransformer converts a string to title case.
type titleCaseTransformer struct {
	caser cases.Caser
}

func (t titleCaseTransformer) Transform(value string) (string, error) {
	return t.caser.String(value), nil
}

func (t titleCaseTransformer) Name() string {
	return "TitleCase"
}

// TitleCase returns a transformer that converts to title case.
func TitleCase() Transformer[string] {
	return titleCaseTransformer{
		caser: cases.Title(language.English),
	}
}

// TitleCaseWithLanguage returns a transformer that converts to title case for a specific language.
func TitleCaseWithLanguage(lang language.Tag) Transformer[string] {
	return titleCaseTransformer{
		caser: cases.Title(lang),
	}
}

// capitalizeTransformer capitalizes the first letter of a string.
type capitalizeTransformer struct{}

func (t capitalizeTransformer) Transform(value string) (string, error) {
	if value == "" {
		return value, nil
	}
	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes), nil
}

func (t capitalizeTransformer) Name() string {
	return "Capitalize"
}

// Capitalize returns a transformer that capitalizes the first letter.
func Capitalize() Transformer[string] {
	return capitalizeTransformer{}
}

// sanitizeEmailTransformer normalizes email addresses (lowercase + trim).
type sanitizeEmailTransformer struct{}

func (t sanitizeEmailTransformer) Transform(value string) (string, error) {
	return strings.ToLower(strings.TrimSpace(value)), nil
}

func (t sanitizeEmailTransformer) Name() string {
	return "SanitizeEmail"
}

// SanitizeEmail returns a transformer that normalizes email addresses.
func SanitizeEmail() Transformer[string] {
	return sanitizeEmailTransformer{}
}

// trimPrefixTransformer removes a prefix from a string.
type trimPrefixTransformer struct {
	prefix string
}

func (t trimPrefixTransformer) Transform(value string) (string, error) {
	return strings.TrimPrefix(value, t.prefix), nil
}

func (t trimPrefixTransformer) Name() string {
	return "TrimPrefix"
}

// TrimPrefix returns a transformer that removes a prefix.
func TrimPrefix(prefix string) Transformer[string] {
	return trimPrefixTransformer{prefix: prefix}
}

// trimSuffixTransformer removes a suffix from a string.
type trimSuffixTransformer struct {
	suffix string
}

func (t trimSuffixTransformer) Transform(value string) (string, error) {
	return strings.TrimSuffix(value, t.suffix), nil
}

func (t trimSuffixTransformer) Name() string {
	return "TrimSuffix"
}

// TrimSuffix returns a transformer that removes a suffix.
func TrimSuffix(suffix string) Transformer[string] {
	return trimSuffixTransformer{suffix: suffix}
}

// replaceTransformer replaces occurrences in a string.
type replaceTransformer struct {
	old string
	new string
	n   int // -1 for all occurrences
}

func (t replaceTransformer) Transform(value string) (string, error) {
	return strings.Replace(value, t.old, t.new, t.n), nil
}

func (t replaceTransformer) Name() string {
	return "Replace"
}

// Replace returns a transformer that replaces all occurrences of old with new.
func Replace(old, new string) Transformer[string] {
	return replaceTransformer{old: old, new: new, n: -1}
}

// ReplaceN returns a transformer that replaces n occurrences of old with new.
func ReplaceN(old, new string, n int) Transformer[string] {
	return replaceTransformer{old: old, new: new, n: n}
}

// =============================================================================
// Numeric Transformers
// =============================================================================

// Signed is a constraint for signed numeric types.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Float is a constraint for floating-point types.
type Float interface {
	~float32 | ~float64
}

// SignedOrFloat is a constraint for signed integers and floats.
type SignedOrFloat interface {
	Signed | Float
}

// absTransformer returns the absolute value.
type absTransformer[T SignedOrFloat] struct{}

func (t absTransformer[T]) Transform(value T) (T, error) {
	if value < 0 {
		return -value, nil
	}
	return value, nil
}

func (t absTransformer[T]) Name() string {
	return "Abs"
}

// Abs returns a transformer that returns the absolute value.
func Abs[T SignedOrFloat]() Transformer[T] {
	return absTransformer[T]{}
}

// clampTransformer clamps a value to a range.
type clampTransformer[T Ordered] struct {
	min T
	max T
}

func (t clampTransformer[T]) Transform(value T) (T, error) {
	if value < t.min {
		return t.min, nil
	}
	if value > t.max {
		return t.max, nil
	}
	return value, nil
}

func (t clampTransformer[T]) Name() string {
	return "Clamp"
}

// Clamp returns a transformer that clamps a value to [min, max].
func Clamp[T Ordered](min, max T) Transformer[T] {
	return clampTransformer[T]{min: min, max: max}
}

// defaultIfZeroTransformer sets a default if the value is zero.
type defaultIfZeroTransformer[T comparable] struct {
	defaultValue T
}

func (t defaultIfZeroTransformer[T]) Transform(value T) (T, error) {
	var zero T
	if value == zero {
		return t.defaultValue, nil
	}
	return value, nil
}

func (t defaultIfZeroTransformer[T]) Name() string {
	return "DefaultIfZero"
}

// DefaultIfZero returns a transformer that sets a default if the value is zero.
func DefaultIfZero[T comparable](defaultValue T) Transformer[T] {
	return defaultIfZeroTransformer[T]{defaultValue: defaultValue}
}

// =============================================================================
// Slice Transformers
// =============================================================================

// mapTransformer applies a function to each element of a slice.
type mapTransformer[T any] struct {
	fn func(T) T
}

func (t mapTransformer[T]) Transform(value []T) ([]T, error) {
	result := make([]T, len(value))
	for i, v := range value {
		result[i] = t.fn(v)
	}
	return result, nil
}

func (t mapTransformer[T]) Name() string {
	return "Map"
}

// MapSlice returns a transformer that applies a function to each element.
func MapSlice[T any](fn func(T) T) Transformer[[]T] {
	return mapTransformer[T]{fn: fn}
}

// filterTransformer filters elements based on a predicate.
type filterTransformer[T any] struct {
	predicate func(T) bool
}

func (t filterTransformer[T]) Transform(value []T) ([]T, error) {
	result := make([]T, 0, len(value))
	for _, v := range value {
		if t.predicate(v) {
			result = append(result, v)
		}
	}
	return result, nil
}

func (t filterTransformer[T]) Name() string {
	return "Filter"
}

// FilterSlice returns a transformer that filters elements.
func FilterSlice[T any](predicate func(T) bool) Transformer[[]T] {
	return filterTransformer[T]{predicate: predicate}
}

// compactTransformer removes zero values from a slice.
type compactTransformer[T comparable] struct{}

func (t compactTransformer[T]) Transform(value []T) ([]T, error) {
	var zero T
	result := make([]T, 0, len(value))
	for _, v := range value {
		if v != zero {
			result = append(result, v)
		}
	}
	return result, nil
}

func (t compactTransformer[T]) Name() string {
	return "Compact"
}

// Compact returns a transformer that removes zero values from a slice.
func Compact[T comparable]() Transformer[[]T] {
	return compactTransformer[T]{}
}

// =============================================================================
// Custom Transformer
// =============================================================================

// funcTransformer wraps a function as a transformer.
type funcTransformer[T any] struct {
	fn   func(T) (T, error)
	name string
}

func (t funcTransformer[T]) Transform(value T) (T, error) {
	return t.fn(value)
}

func (t funcTransformer[T]) Name() string {
	if t.name != "" {
		return t.name
	}
	return "Custom"
}

// TransformerFunc creates a transformer from a function.
func TransformerFunc[T any](fn func(T) (T, error)) Transformer[T] {
	return funcTransformer[T]{fn: fn}
}

// TransformerFuncNamed creates a named transformer from a function.
func TransformerFuncNamed[T any](name string, fn func(T) (T, error)) Transformer[T] {
	return funcTransformer[T]{fn: fn, name: name}
}

// =============================================================================
// Chaining Transformers
// =============================================================================

// chainTransformer chains multiple transformers.
type chainTransformer[T any] struct {
	transformers []Transformer[T]
}

func (t chainTransformer[T]) Transform(value T) (T, error) {
	var err error
	for _, tr := range t.transformers {
		value, err = tr.Transform(value)
		if err != nil {
			return value, err
		}
	}
	return value, nil
}

func (t chainTransformer[T]) Name() string {
	return "Chain"
}

// Chain creates a transformer that applies multiple transformers in sequence.
func Chain[T any](transformers ...Transformer[T]) Transformer[T] {
	return chainTransformer[T]{transformers: transformers}
}
