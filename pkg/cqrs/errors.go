package cqrs

import "fmt"

type NotFoundError struct {
	ResourceType string
	Identifier   string
}

func NewNotFoundError(resourceType, id string) *NotFoundError {
	return &NotFoundError{ResourceType: resourceType, Identifier: id}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with identifier %q not found", e.ResourceType, e.Identifier)
}

type ConflictError struct {
	ResourceType string
	Identifier   string
	Detail       string
}

func NewConflictError(resourceType, id, detail string) *ConflictError {
	return &ConflictError{ResourceType: resourceType, Identifier: id, Detail: detail}
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict on %s with identifier %q: %s", e.ResourceType, e.Identifier, e.Detail)
}

type FieldError struct {
	Field  string
	Detail string
}

type ValidationError struct {
	Fields []FieldError
}

func NewValidationError(fields []FieldError) *ValidationError {
	return &ValidationError{Fields: fields}
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return "validation failed"
	}
	if len(e.Fields) == 1 {
		return fmt.Sprintf("validation failed: %s: %s", e.Fields[0].Field, e.Fields[0].Detail)
	}
	return fmt.Sprintf("validation failed on %d fields", len(e.Fields))
}

type UnauthorizedError struct {
	Detail string
}

func NewUnauthorizedError(detail string) *UnauthorizedError {
	return &UnauthorizedError{Detail: detail}
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized: %s", e.Detail)
}

type ForbiddenError struct {
	Detail string
}

func NewForbiddenError(detail string) *ForbiddenError {
	return &ForbiddenError{Detail: detail}
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Detail)
}
