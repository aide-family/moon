// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: prom/v1/node.proto

package v1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on CreateNodeRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *CreateNodeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CreateNodeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CreateNodeRequestMultiError, or nil if none found.
func (m *CreateNodeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *CreateNodeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return CreateNodeRequestMultiError(errors)
	}

	return nil
}

// CreateNodeRequestMultiError is an error wrapping multiple validation errors
// returned by CreateNodeRequest.ValidateAll() if the designated constraints
// aren't met.
type CreateNodeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CreateNodeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CreateNodeRequestMultiError) AllErrors() []error { return m }

// CreateNodeRequestValidationError is the validation error returned by
// CreateNodeRequest.Validate if the designated constraints aren't met.
type CreateNodeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CreateNodeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CreateNodeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CreateNodeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CreateNodeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CreateNodeRequestValidationError) ErrorName() string {
	return "CreateNodeRequestValidationError"
}

// Error satisfies the builtin error interface
func (e CreateNodeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateNodeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CreateNodeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CreateNodeRequestValidationError{}

// Validate checks the field values on CreateNodeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *CreateNodeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CreateNodeReply with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// CreateNodeReplyMultiError, or nil if none found.
func (m *CreateNodeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *CreateNodeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return CreateNodeReplyMultiError(errors)
	}

	return nil
}

// CreateNodeReplyMultiError is an error wrapping multiple validation errors
// returned by CreateNodeReply.ValidateAll() if the designated constraints
// aren't met.
type CreateNodeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CreateNodeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CreateNodeReplyMultiError) AllErrors() []error { return m }

// CreateNodeReplyValidationError is the validation error returned by
// CreateNodeReply.Validate if the designated constraints aren't met.
type CreateNodeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CreateNodeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CreateNodeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CreateNodeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CreateNodeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CreateNodeReplyValidationError) ErrorName() string { return "CreateNodeReplyValidationError" }

// Error satisfies the builtin error interface
func (e CreateNodeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateNodeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CreateNodeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CreateNodeReplyValidationError{}

// Validate checks the field values on UpdateNodeRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *UpdateNodeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpdateNodeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// UpdateNodeRequestMultiError, or nil if none found.
func (m *UpdateNodeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *UpdateNodeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return UpdateNodeRequestMultiError(errors)
	}

	return nil
}

// UpdateNodeRequestMultiError is an error wrapping multiple validation errors
// returned by UpdateNodeRequest.ValidateAll() if the designated constraints
// aren't met.
type UpdateNodeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpdateNodeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpdateNodeRequestMultiError) AllErrors() []error { return m }

// UpdateNodeRequestValidationError is the validation error returned by
// UpdateNodeRequest.Validate if the designated constraints aren't met.
type UpdateNodeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpdateNodeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpdateNodeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpdateNodeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpdateNodeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpdateNodeRequestValidationError) ErrorName() string {
	return "UpdateNodeRequestValidationError"
}

// Error satisfies the builtin error interface
func (e UpdateNodeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpdateNodeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpdateNodeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpdateNodeRequestValidationError{}

// Validate checks the field values on UpdateNodeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *UpdateNodeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpdateNodeReply with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// UpdateNodeReplyMultiError, or nil if none found.
func (m *UpdateNodeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *UpdateNodeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return UpdateNodeReplyMultiError(errors)
	}

	return nil
}

// UpdateNodeReplyMultiError is an error wrapping multiple validation errors
// returned by UpdateNodeReply.ValidateAll() if the designated constraints
// aren't met.
type UpdateNodeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpdateNodeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpdateNodeReplyMultiError) AllErrors() []error { return m }

// UpdateNodeReplyValidationError is the validation error returned by
// UpdateNodeReply.Validate if the designated constraints aren't met.
type UpdateNodeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpdateNodeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpdateNodeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpdateNodeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpdateNodeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpdateNodeReplyValidationError) ErrorName() string { return "UpdateNodeReplyValidationError" }

// Error satisfies the builtin error interface
func (e UpdateNodeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpdateNodeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpdateNodeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpdateNodeReplyValidationError{}

// Validate checks the field values on DeleteNodeRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *DeleteNodeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteNodeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DeleteNodeRequestMultiError, or nil if none found.
func (m *DeleteNodeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteNodeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return DeleteNodeRequestMultiError(errors)
	}

	return nil
}

// DeleteNodeRequestMultiError is an error wrapping multiple validation errors
// returned by DeleteNodeRequest.ValidateAll() if the designated constraints
// aren't met.
type DeleteNodeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteNodeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteNodeRequestMultiError) AllErrors() []error { return m }

// DeleteNodeRequestValidationError is the validation error returned by
// DeleteNodeRequest.Validate if the designated constraints aren't met.
type DeleteNodeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteNodeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteNodeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteNodeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteNodeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteNodeRequestValidationError) ErrorName() string {
	return "DeleteNodeRequestValidationError"
}

// Error satisfies the builtin error interface
func (e DeleteNodeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteNodeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteNodeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteNodeRequestValidationError{}

// Validate checks the field values on DeleteNodeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *DeleteNodeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteNodeReply with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// DeleteNodeReplyMultiError, or nil if none found.
func (m *DeleteNodeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteNodeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return DeleteNodeReplyMultiError(errors)
	}

	return nil
}

// DeleteNodeReplyMultiError is an error wrapping multiple validation errors
// returned by DeleteNodeReply.ValidateAll() if the designated constraints
// aren't met.
type DeleteNodeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteNodeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteNodeReplyMultiError) AllErrors() []error { return m }

// DeleteNodeReplyValidationError is the validation error returned by
// DeleteNodeReply.Validate if the designated constraints aren't met.
type DeleteNodeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteNodeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteNodeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteNodeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteNodeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteNodeReplyValidationError) ErrorName() string { return "DeleteNodeReplyValidationError" }

// Error satisfies the builtin error interface
func (e DeleteNodeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteNodeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteNodeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteNodeReplyValidationError{}

// Validate checks the field values on GetNodeRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *GetNodeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetNodeRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in GetNodeRequestMultiError,
// or nil if none found.
func (m *GetNodeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetNodeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return GetNodeRequestMultiError(errors)
	}

	return nil
}

// GetNodeRequestMultiError is an error wrapping multiple validation errors
// returned by GetNodeRequest.ValidateAll() if the designated constraints
// aren't met.
type GetNodeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetNodeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetNodeRequestMultiError) AllErrors() []error { return m }

// GetNodeRequestValidationError is the validation error returned by
// GetNodeRequest.Validate if the designated constraints aren't met.
type GetNodeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetNodeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetNodeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetNodeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetNodeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetNodeRequestValidationError) ErrorName() string { return "GetNodeRequestValidationError" }

// Error satisfies the builtin error interface
func (e GetNodeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetNodeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetNodeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetNodeRequestValidationError{}

// Validate checks the field values on GetNodeReply with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *GetNodeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetNodeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in GetNodeReplyMultiError, or
// nil if none found.
func (m *GetNodeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *GetNodeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return GetNodeReplyMultiError(errors)
	}

	return nil
}

// GetNodeReplyMultiError is an error wrapping multiple validation errors
// returned by GetNodeReply.ValidateAll() if the designated constraints aren't met.
type GetNodeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetNodeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetNodeReplyMultiError) AllErrors() []error { return m }

// GetNodeReplyValidationError is the validation error returned by
// GetNodeReply.Validate if the designated constraints aren't met.
type GetNodeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetNodeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetNodeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetNodeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetNodeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetNodeReplyValidationError) ErrorName() string { return "GetNodeReplyValidationError" }

// Error satisfies the builtin error interface
func (e GetNodeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetNodeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetNodeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetNodeReplyValidationError{}

// Validate checks the field values on ListNodeRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *ListNodeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListNodeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ListNodeRequestMultiError, or nil if none found.
func (m *ListNodeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ListNodeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return ListNodeRequestMultiError(errors)
	}

	return nil
}

// ListNodeRequestMultiError is an error wrapping multiple validation errors
// returned by ListNodeRequest.ValidateAll() if the designated constraints
// aren't met.
type ListNodeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListNodeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListNodeRequestMultiError) AllErrors() []error { return m }

// ListNodeRequestValidationError is the validation error returned by
// ListNodeRequest.Validate if the designated constraints aren't met.
type ListNodeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListNodeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListNodeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListNodeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListNodeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListNodeRequestValidationError) ErrorName() string { return "ListNodeRequestValidationError" }

// Error satisfies the builtin error interface
func (e ListNodeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListNodeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListNodeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListNodeRequestValidationError{}

// Validate checks the field values on ListNodeReply with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ListNodeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListNodeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ListNodeReplyMultiError, or
// nil if none found.
func (m *ListNodeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *ListNodeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return ListNodeReplyMultiError(errors)
	}

	return nil
}

// ListNodeReplyMultiError is an error wrapping multiple validation errors
// returned by ListNodeReply.ValidateAll() if the designated constraints
// aren't met.
type ListNodeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListNodeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListNodeReplyMultiError) AllErrors() []error { return m }

// ListNodeReplyValidationError is the validation error returned by
// ListNodeReply.Validate if the designated constraints aren't met.
type ListNodeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListNodeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListNodeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListNodeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListNodeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListNodeReplyValidationError) ErrorName() string { return "ListNodeReplyValidationError" }

// Error satisfies the builtin error interface
func (e ListNodeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListNodeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListNodeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListNodeReplyValidationError{}
