package freeipa

import "net/http"

const (
	// UnauthorizedReason string extracted from
	// https://github.com/freeipa/freeipa/blob/master/ipaserver/rpcserver.py
	passwordExpiredUnauthorizedReason        = "password-expired"
	invalidSessionPasswordUnauthorizedReason = "invalid-password"
	krbPrincipalExpiredUnauthorizedReason    = "krbprincipal-expired"
	userLockedUnauthorizedReason             = "user-locked"

	ipaRejectionReasonHTTPHeader = "X-Ipa-Rejection-Reason"
)

func unauthorizedHTTPResponseToFreeipaError(resp *http.Response) *Error {
	var errorCode int
	rejectionReason := resp.Header.Get(ipaRejectionReasonHTTPHeader)

	switch rejectionReason {
	case passwordExpiredUnauthorizedReason:
		errorCode = PasswordExpiredCode
	case invalidSessionPasswordUnauthorizedReason:
		errorCode = InvalidSessionPasswordCode
	case krbPrincipalExpiredUnauthorizedReason:
		errorCode = KrbPrincipalExpiredCode
	case userLockedUnauthorizedReason:
		errorCode = UserLockedCode

	default:
		errorCode = GenericErrorCode
	}

	return &Error{
		Message: rejectionReason,
		Name:    rejectionReason,
		Code:    errorCode,
	}
}
