// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: alarm/realtime/realtime.proto

package realtime

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

// Validate checks the field values on GetRealtimeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetRealtimeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetRealtimeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetRealtimeRequestMultiError, or nil if none found.
func (m *GetRealtimeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetRealtimeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetId() <= 0 {
		err := GetRealtimeRequestValidationError{
			field:  "Id",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return GetRealtimeRequestMultiError(errors)
	}

	return nil
}

// GetRealtimeRequestMultiError is an error wrapping multiple validation errors
// returned by GetRealtimeRequest.ValidateAll() if the designated constraints
// aren't met.
type GetRealtimeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetRealtimeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetRealtimeRequestMultiError) AllErrors() []error { return m }

// GetRealtimeRequestValidationError is the validation error returned by
// GetRealtimeRequest.Validate if the designated constraints aren't met.
type GetRealtimeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetRealtimeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetRealtimeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetRealtimeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetRealtimeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetRealtimeRequestValidationError) ErrorName() string {
	return "GetRealtimeRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetRealtimeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetRealtimeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetRealtimeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetRealtimeRequestValidationError{}

// Validate checks the field values on GetRealtimeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *GetRealtimeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetRealtimeReply with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetRealtimeReplyMultiError, or nil if none found.
func (m *GetRealtimeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *GetRealtimeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetDetail()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, GetRealtimeReplyValidationError{
					field:  "Detail",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, GetRealtimeReplyValidationError{
					field:  "Detail",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetDetail()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return GetRealtimeReplyValidationError{
				field:  "Detail",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return GetRealtimeReplyMultiError(errors)
	}

	return nil
}

// GetRealtimeReplyMultiError is an error wrapping multiple validation errors
// returned by GetRealtimeReply.ValidateAll() if the designated constraints
// aren't met.
type GetRealtimeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetRealtimeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetRealtimeReplyMultiError) AllErrors() []error { return m }

// GetRealtimeReplyValidationError is the validation error returned by
// GetRealtimeReply.Validate if the designated constraints aren't met.
type GetRealtimeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetRealtimeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetRealtimeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetRealtimeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetRealtimeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetRealtimeReplyValidationError) ErrorName() string { return "GetRealtimeReplyValidationError" }

// Error satisfies the builtin error interface
func (e GetRealtimeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetRealtimeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetRealtimeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetRealtimeReplyValidationError{}

// Validate checks the field values on ListRealtimeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ListRealtimeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListRealtimeRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ListRealtimeRequestMultiError, or nil if none found.
func (m *ListRealtimeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ListRealtimeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetPage() == nil {
		err := ListRealtimeRequestValidationError{
			field:  "Page",
			reason: "value is required",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if all {
		switch v := interface{}(m.GetPage()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ListRealtimeRequestValidationError{
					field:  "Page",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ListRealtimeRequestValidationError{
					field:  "Page",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetPage()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ListRealtimeRequestValidationError{
				field:  "Page",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for Keyword

	// no validation rules for StartAt

	// no validation rules for EndAt

	if len(errors) > 0 {
		return ListRealtimeRequestMultiError(errors)
	}

	return nil
}

// ListRealtimeRequestMultiError is an error wrapping multiple validation
// errors returned by ListRealtimeRequest.ValidateAll() if the designated
// constraints aren't met.
type ListRealtimeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListRealtimeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListRealtimeRequestMultiError) AllErrors() []error { return m }

// ListRealtimeRequestValidationError is the validation error returned by
// ListRealtimeRequest.Validate if the designated constraints aren't met.
type ListRealtimeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListRealtimeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListRealtimeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListRealtimeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListRealtimeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListRealtimeRequestValidationError) ErrorName() string {
	return "ListRealtimeRequestValidationError"
}

// Error satisfies the builtin error interface
func (e ListRealtimeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListRealtimeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListRealtimeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListRealtimeRequestValidationError{}

// Validate checks the field values on ListRealtimeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *ListRealtimeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ListRealtimeReply with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ListRealtimeReplyMultiError, or nil if none found.
func (m *ListRealtimeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *ListRealtimeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetPage()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ListRealtimeReplyValidationError{
					field:  "Page",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ListRealtimeReplyValidationError{
					field:  "Page",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetPage()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ListRealtimeReplyValidationError{
				field:  "Page",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	for idx, item := range m.GetList() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ListRealtimeReplyValidationError{
						field:  fmt.Sprintf("List[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ListRealtimeReplyValidationError{
						field:  fmt.Sprintf("List[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ListRealtimeReplyValidationError{
					field:  fmt.Sprintf("List[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return ListRealtimeReplyMultiError(errors)
	}

	return nil
}

// ListRealtimeReplyMultiError is an error wrapping multiple validation errors
// returned by ListRealtimeReply.ValidateAll() if the designated constraints
// aren't met.
type ListRealtimeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ListRealtimeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ListRealtimeReplyMultiError) AllErrors() []error { return m }

// ListRealtimeReplyValidationError is the validation error returned by
// ListRealtimeReply.Validate if the designated constraints aren't met.
type ListRealtimeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ListRealtimeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ListRealtimeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ListRealtimeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ListRealtimeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ListRealtimeReplyValidationError) ErrorName() string {
	return "ListRealtimeReplyValidationError"
}

// Error satisfies the builtin error interface
func (e ListRealtimeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sListRealtimeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ListRealtimeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ListRealtimeReplyValidationError{}

// Validate checks the field values on InterveneRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *InterveneRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on InterveneRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// InterveneRequestMultiError, or nil if none found.
func (m *InterveneRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *InterveneRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetId() <= 0 {
		err := InterveneRequestValidationError{
			field:  "Id",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetRemark()); l < 1 || l > 255 {
		err := InterveneRequestValidationError{
			field:  "Remark",
			reason: "value length must be between 1 and 255 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return InterveneRequestMultiError(errors)
	}

	return nil
}

// InterveneRequestMultiError is an error wrapping multiple validation errors
// returned by InterveneRequest.ValidateAll() if the designated constraints
// aren't met.
type InterveneRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m InterveneRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m InterveneRequestMultiError) AllErrors() []error { return m }

// InterveneRequestValidationError is the validation error returned by
// InterveneRequest.Validate if the designated constraints aren't met.
type InterveneRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e InterveneRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e InterveneRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e InterveneRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e InterveneRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e InterveneRequestValidationError) ErrorName() string { return "InterveneRequestValidationError" }

// Error satisfies the builtin error interface
func (e InterveneRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sInterveneRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = InterveneRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = InterveneRequestValidationError{}

// Validate checks the field values on InterveneReply with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *InterveneReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on InterveneReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in InterveneReplyMultiError,
// or nil if none found.
func (m *InterveneReply) ValidateAll() error {
	return m.validate(true)
}

func (m *InterveneReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return InterveneReplyMultiError(errors)
	}

	return nil
}

// InterveneReplyMultiError is an error wrapping multiple validation errors
// returned by InterveneReply.ValidateAll() if the designated constraints
// aren't met.
type InterveneReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m InterveneReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m InterveneReplyMultiError) AllErrors() []error { return m }

// InterveneReplyValidationError is the validation error returned by
// InterveneReply.Validate if the designated constraints aren't met.
type InterveneReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e InterveneReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e InterveneReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e InterveneReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e InterveneReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e InterveneReplyValidationError) ErrorName() string { return "InterveneReplyValidationError" }

// Error satisfies the builtin error interface
func (e InterveneReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sInterveneReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = InterveneReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = InterveneReplyValidationError{}

// Validate checks the field values on UpgradeRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *UpgradeRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpgradeRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in UpgradeRequestMultiError,
// or nil if none found.
func (m *UpgradeRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *UpgradeRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetId() <= 0 {
		err := UpgradeRequestValidationError{
			field:  "Id",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetRemark()); l < 1 || l > 255 {
		err := UpgradeRequestValidationError{
			field:  "Remark",
			reason: "value length must be between 1 and 255 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return UpgradeRequestMultiError(errors)
	}

	return nil
}

// UpgradeRequestMultiError is an error wrapping multiple validation errors
// returned by UpgradeRequest.ValidateAll() if the designated constraints
// aren't met.
type UpgradeRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpgradeRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpgradeRequestMultiError) AllErrors() []error { return m }

// UpgradeRequestValidationError is the validation error returned by
// UpgradeRequest.Validate if the designated constraints aren't met.
type UpgradeRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpgradeRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpgradeRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpgradeRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpgradeRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpgradeRequestValidationError) ErrorName() string { return "UpgradeRequestValidationError" }

// Error satisfies the builtin error interface
func (e UpgradeRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpgradeRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpgradeRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpgradeRequestValidationError{}

// Validate checks the field values on UpgradeReply with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *UpgradeReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpgradeReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in UpgradeReplyMultiError, or
// nil if none found.
func (m *UpgradeReply) ValidateAll() error {
	return m.validate(true)
}

func (m *UpgradeReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return UpgradeReplyMultiError(errors)
	}

	return nil
}

// UpgradeReplyMultiError is an error wrapping multiple validation errors
// returned by UpgradeReply.ValidateAll() if the designated constraints aren't met.
type UpgradeReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpgradeReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpgradeReplyMultiError) AllErrors() []error { return m }

// UpgradeReplyValidationError is the validation error returned by
// UpgradeReply.Validate if the designated constraints aren't met.
type UpgradeReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpgradeReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpgradeReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpgradeReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpgradeReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpgradeReplyValidationError) ErrorName() string { return "UpgradeReplyValidationError" }

// Error satisfies the builtin error interface
func (e UpgradeReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpgradeReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpgradeReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpgradeReplyValidationError{}

// Validate checks the field values on SuppressRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *SuppressRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SuppressRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// SuppressRequestMultiError, or nil if none found.
func (m *SuppressRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *SuppressRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if m.GetId() <= 0 {
		err := SuppressRequestValidationError{
			field:  "Id",
			reason: "value must be greater than 0",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetRemark()); l < 1 || l > 255 {
		err := SuppressRequestValidationError{
			field:  "Remark",
			reason: "value length must be between 1 and 255 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if val := m.GetDuration(); val < 60 || val > 86400 {
		err := SuppressRequestValidationError{
			field:  "Duration",
			reason: "value must be inside range [60, 86400]",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return SuppressRequestMultiError(errors)
	}

	return nil
}

// SuppressRequestMultiError is an error wrapping multiple validation errors
// returned by SuppressRequest.ValidateAll() if the designated constraints
// aren't met.
type SuppressRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SuppressRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SuppressRequestMultiError) AllErrors() []error { return m }

// SuppressRequestValidationError is the validation error returned by
// SuppressRequest.Validate if the designated constraints aren't met.
type SuppressRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SuppressRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SuppressRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SuppressRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SuppressRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SuppressRequestValidationError) ErrorName() string { return "SuppressRequestValidationError" }

// Error satisfies the builtin error interface
func (e SuppressRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSuppressRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SuppressRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SuppressRequestValidationError{}

// Validate checks the field values on SuppressReply with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *SuppressReply) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on SuppressReply with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in SuppressReplyMultiError, or
// nil if none found.
func (m *SuppressReply) ValidateAll() error {
	return m.validate(true)
}

func (m *SuppressReply) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return SuppressReplyMultiError(errors)
	}

	return nil
}

// SuppressReplyMultiError is an error wrapping multiple validation errors
// returned by SuppressReply.ValidateAll() if the designated constraints
// aren't met.
type SuppressReplyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m SuppressReplyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m SuppressReplyMultiError) AllErrors() []error { return m }

// SuppressReplyValidationError is the validation error returned by
// SuppressReply.Validate if the designated constraints aren't met.
type SuppressReplyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SuppressReplyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SuppressReplyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SuppressReplyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SuppressReplyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SuppressReplyValidationError) ErrorName() string { return "SuppressReplyValidationError" }

// Error satisfies the builtin error interface
func (e SuppressReplyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSuppressReply.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SuppressReplyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SuppressReplyValidationError{}
