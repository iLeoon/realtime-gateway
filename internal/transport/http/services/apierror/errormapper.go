package apierror

func InvalidJSONFormat() *APIError {
	return Build(BadRequestCode,
		"invalid request JSON format",
		WithTarget("RequestBody"),
		WithInnerError("InvalidFieldDataTypeOrInvalidJSONFormat"),
	)
}

func MissingUserIDContext() *APIError {
	return Build(InternalServerErrorCode,
		"user id is missing in the request header",
		WithTarget("userId"),
		WithInnerError("MissingUserIdContext"),
	)
}

func FaildToGenerateToken(code string) *APIError {
	return Build(InternalServerErrorCode,
		"unexpected error while trying to generate token",
		WithTarget("token"),
		WithInnerError(code),
	)

}

func UnknownInternalErrors() *APIError {
	return Build(InternalServerErrorCode,
		"unknown error occurred",
		WithInnerError("UnknwonInternalError"),
	)
}

func FaildToDecodeToken() *APIError {
	return Build(InternalServerErrorCode,
		"unexpected error while trying to decode token",
		WithTarget("token"),
		WithInnerError("UnexpectedInternalError"),
	)
}

func UnexpectedDatabaseFailure(code Code, target string, details string) *APIError {
	return Build(code,
		"unexpected error on processing the request",
		WithTarget(target),
		WithInnerError("DatabaseFailure"),
		WithInnerErrorDetails(details),
	)
}

func UnAuthorizedUser(target string) *APIError {
	return Build(
		ForbiddenRequestCode,
		"user is not authorized to check the resource",
		WithTarget(target),
		WithInnerError("AuthenticatedIdIsUnauthroized"),
	)
}

func InvalidToken() *APIError {
	return Build(
		UnAuthorizedRequestCode,
		"invalid token is being used",
		WithTarget("token"),
		WithInnerError("InvalidOrExpiredToken"),
	)
}

func InvalidArgument(target string, details []ErrorDetails) *APIError {
	return Build(
		BadArgumentCode,
		"invalid argument",
		WithTarget(target),
		WithDetails(details),
	)
}

func NoDataFound(target string) *APIError {
	return Build(
		NotFoundRequestCode,
		"no data was found",
		WithTarget(target),
		WithInnerError("NoRecordsFoundWithThatId"),
	)
}

func InvalidAuthParameters(target string, code string) *APIError {
	return Build(
		BadRequestCode,
		"user is not authenticated",
		WithTarget(target),
		WithInnerError(code),
	)

}

func TimeOutError(target string) *APIError {
	return Build(GatewayTimeout,
		"request took so long to process",
		WithTarget(target),
		WithInnerError("TimeOutHasBeenExceeded"),
	)
}
