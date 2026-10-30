package domain

import "fmt"

type Kind uint8

const (
	KindDB Kind = iota + 1
	KindNotFound
	KindUnauthorized
	KindBadRequest
	KindServiceUnavailable
	KindBadGateway
	KindInternal
	KindConflict
)

type Error struct {
	kind   Kind
	reason string
	code   string
	field  string
	msg    string
	cause  error
}

func (e *Error) Kind() Kind      { return e.kind }
func (e *Error) Reason() string  { return e.reason }
func (e *Error) Code() string    { return e.code }
func (e *Error) Field() string   { return e.field }
func (e *Error) Message() string { return e.msg }

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return t.kind == e.kind && (t.reason == "" || t.reason == e.reason) && (t.code == "" || t.code == e.code)
}

func (e *Error) Error() string {
	switch e.kind {
	case KindDB:
		return fmt.Sprintf("database error: %v", e.cause)
	case KindNotFound:
		return withReason("resource not found", e.reason)
	case KindUnauthorized:
		return withReason("unauthorized", e.reason)
	case KindBadRequest, KindConflict:
		return e.msg
	case KindServiceUnavailable, KindBadGateway:
		if e.cause != nil {
			return fmt.Sprintf("%s: %v", e.msg, e.cause)
		}
		return e.msg
	case KindInternal:
		return fmt.Sprintf("internal error: %v", e.cause)
	default:
		return fmt.Sprintf("internal error: unknown kind %d", e.kind)
	}
}

func withReason(base, reason string) string {
	if reason == "" {
		return base
	}
	return base + ": " + reason
}

func DB(cause error) *Error { return &Error{kind: KindDB, cause: cause} }

func Internal(cause error) *Error { return &Error{kind: KindInternal, cause: cause} }

func Internalf(format string, args ...any) *Error {
	return &Error{kind: KindInternal, cause: fmt.Errorf(format, args...)}
}

func BadRequest(message string) *Error { return &Error{kind: KindBadRequest, msg: message} }

func BadRequestCoded(code, message string) *Error {
	return &Error{kind: KindBadRequest, code: code, msg: message}
}

func Invalid(field, message string) *Error {
	return &Error{kind: KindBadRequest, code: "VALIDATION_FAILED", field: field, msg: message}
}

func ConflictCoded(code, message string) *Error {
	return &Error{kind: KindConflict, code: code, msg: message}
}

func NotFound(reason string) *Error { return &Error{kind: KindNotFound, reason: reason} }

func NotFoundCoded(code, reason string) *Error {
	return &Error{kind: KindNotFound, code: code, reason: reason}
}

func Unauthorized(reason string) *Error { return &Error{kind: KindUnauthorized, reason: reason} }

func UnauthorizedCoded(code, reason string) *Error {
	return &Error{kind: KindUnauthorized, code: code, reason: reason}
}

func ServiceUnavailableCoded(code, message string) *Error {
	return &Error{kind: KindServiceUnavailable, code: code, msg: message}
}

func BadGatewayCaused(code, message string, cause error) *Error {
	return &Error{kind: KindBadGateway, code: code, msg: message, cause: cause}
}

var (
	ErrDB                 = &Error{kind: KindDB}
	ErrNotFound           = &Error{kind: KindNotFound}
	ErrUnauthorized       = &Error{kind: KindUnauthorized}
	ErrBadRequest         = &Error{kind: KindBadRequest}
	ErrServiceUnavailable = &Error{kind: KindServiceUnavailable}
	ErrBadGateway         = &Error{kind: KindBadGateway}
	ErrConflict           = &Error{kind: KindConflict}
	ErrInternal           = &Error{kind: KindInternal}
)

var (
	ErrTokenInvalid            = UnauthorizedCoded("TOKEN_INVALID", "")
	ErrTokenIssuerUnknown      = UnauthorizedCoded("TOKEN_ISSUER_UNKNOWN", "token_issuer_unknown")
	ErrTokenSigningKeyNotFound = UnauthorizedCoded("TOKEN_SIGNING_KEY_NOT_FOUND", "token_signing_key_not_found")
	ErrTokenClaimInvalid       = UnauthorizedCoded("TOKEN_CLAIM_INVALID", "")
)

func TokenInvalid(reason string) *Error { return UnauthorizedCoded("TOKEN_INVALID", reason) }

func TokenClaimInvalid(reason string) *Error { return UnauthorizedCoded("TOKEN_CLAIM_INVALID", reason) }

var ErrSigningKeyNotFound = ServiceUnavailableCoded("SIGNING_KEY_NOT_FOUND", "no signing key is active")

var (
	ErrSigningKeyAlreadyPending     = ConflictCoded("SIGNING_KEY_ALREADY_PENDING", "a pending signing key already exists; rotate again after it becomes active")
	ErrSigningKeyRotationInProgress = ConflictCoded("SIGNING_KEY_ROTATION_IN_PROGRESS", "another signing key rotation is running; retry shortly")
)
