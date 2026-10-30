package domain

import (
	"regexp"
	"strings"
)

const MaximumIdentifierLength = 255

var identifierPattern = regexp.MustCompile(`^[0-9a-zA-Z._=-]+$`)

var reservedIdentifierWords = []string{"null", "undefined", "nil", "empty"}

func identifierValid(value string) bool {
	if value == "" || len(value) >= MaximumIdentifierLength || !identifierPattern.MatchString(value) {
		return false
	}
	lower := strings.ToLower(value)
	for _, word := range reservedIdentifierWords {
		if strings.Contains(lower, word) {
			return false
		}
	}
	return true
}

var (
	ErrSubjectIDInvalid = &Error{kind: KindBadRequest, code: "SUBJECT_ID_INVALID", field: "subject_id",
		msg: "subject_id must match [0-9a-zA-Z-._=]+, be shorter than 255 characters and must not contain null, undefined, nil or empty"}
	ErrEntitlementIDInvalid = &Error{kind: KindBadRequest, code: "ENTITLEMENT_ID_INVALID", field: "entitlement_id",
		msg: "entitlement_id must match [0-9a-zA-Z-._=]+, be shorter than 255 characters and must not contain null, undefined, nil or empty"}
)

func ValidateSubjectID(subjectID string) error {
	if !identifierValid(subjectID) {
		return ErrSubjectIDInvalid
	}
	return nil
}

func ValidateEntitlementID(entitlementID string) error {
	if !identifierValid(entitlementID) {
		return ErrEntitlementIDInvalid
	}
	return nil
}
