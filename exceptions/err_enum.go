package exceptions

import "net/http"

type ErrorLogic struct {
	//ErrName  string
	ErrCode  int
	HttpCode int
	Message  string
}

var businessLogicReason = map[int]ErrorLogic{
	DataAlreadyExist:     {ErrCode: DataAlreadyExist, HttpCode: http.StatusUnprocessableEntity, Message: "data is already exist"},
	DataNotFound:         {ErrCode: DataNotFound, HttpCode: http.StatusNotFound, Message: "data not found"},
	DataCreateFailed:     {ErrCode: DataCreateFailed, HttpCode: http.StatusUnprocessableEntity, Message: "create data failed"},
	DataUpdateFailed:     {ErrCode: DataUpdateFailed, HttpCode: http.StatusUnprocessableEntity, Message: "update data failed"},
	DataDeleteFailed:     {ErrCode: DataDeleteFailed, HttpCode: http.StatusUnprocessableEntity, Message: "delete data failed"},
	DataGetFailed:        {ErrCode: DataGetFailed, HttpCode: http.StatusUnprocessableEntity, Message: "get data failed"},
	DataValidationFailed: {ErrCode: DataValidationFailed, HttpCode: http.StatusUnprocessableEntity, Message: "validation failed"},
	DataForbidden:        {ErrCode: DataForbidden, HttpCode: http.StatusForbidden, Message: "forbidden"},
	DataConflict:         {ErrCode: DataConflict, HttpCode: http.StatusUnprocessableEntity, Message: "conflicted data"},
	BadData:              {ErrCode: BadData, HttpCode: http.StatusInternalServerError, Message: "bad data"},
	InvalidArgument:      {ErrCode: InvalidArgument, HttpCode: http.StatusUnprocessableEntity, Message: "invalid argument"},
	//OtherError:         {ErrCode: ABC, HttpCode: http.StatusInternalServerError, Message: "your explanation of error EBL = error business logic"},
}

func BusinessLogicReason(code int) ErrorLogic {
	return businessLogicReason[code]
}

var authenticationReason = map[int]ErrorLogic{
	int(AuthenticationBadFormat):          {ErrCode: int(AuthenticationBadFormat), HttpCode: http.StatusBadRequest, Message: "bad request format"},
	int(AuthenticationUnauthenticated):    {ErrCode: int(AuthenticationUnauthenticated), HttpCode: http.StatusUnauthorized, Message: "unauthenticated"},
	int(AuthenticationInvalidRealm):       {ErrCode: int(AuthenticationInvalidRealm), HttpCode: http.StatusUnauthorized, Message: "invalid realm"},
	int(AuthenticationInvalidIdentity):    {ErrCode: int(AuthenticationInvalidIdentity), HttpCode: http.StatusUnauthorized, Message: "invalid identity"},
	int(AuthenticationInvalidCredentials): {ErrCode: int(AuthenticationInvalidCredentials), HttpCode: http.StatusUnauthorized, Message: "invalid credentials"},
	int(AuthenticationInvalidToken):       {ErrCode: int(AuthenticationInvalidToken), HttpCode: http.StatusUnauthorized, Message: "invalid token"},
	int(AuthenticationTokenRevoked):       {ErrCode: int(AuthenticationTokenRevoked), HttpCode: http.StatusUnauthorized, Message: "token revoked"},
	int(AuthenticationTokenExpired):       {ErrCode: int(AuthenticationTokenExpired), HttpCode: http.StatusUnauthorized, Message: "token expired"},
}

func AuthenticationReason(code AuthenticationCode) ErrorLogic {
	return authenticationReason[int(code)]
}
