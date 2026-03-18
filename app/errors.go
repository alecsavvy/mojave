package app

import (
	"errors"

	v1 "github.com/alecsavvy/mojave/gen/mojave/v1"
)

var (
	// validation errors
	ErrValidationFileUploadTxInvalid = errors.New("file upload tx is invalid")
)

func ToTransactionResultError(err error) *v1.TransactionResultError {
	switch err {
	case ErrValidationFileUploadTxInvalid:
		return &v1.TransactionResultError{
			Code: v1.TransactionResultErrorCode_TRANSACTION_RESULT_ERROR_CODE_INVALID_REQUEST,
			Log:  err.Error(),
		}
	}
	return &v1.TransactionResultError{
		Code: v1.TransactionResultErrorCode_TRANSACTION_RESULT_ERROR_CODE_INTERNAL,
		Log:  err.Error(),
	}
}
