package exceptions

type AuthenticationCode int

const (
	AuthenticationBadFormat          AuthenticationCode = 20001
	AuthenticationUnauthenticated    AuthenticationCode = 20002
	AuthenticationInvalidRealm       AuthenticationCode = 20003
	AuthenticationInvalidIdentity    AuthenticationCode = 20004
	AuthenticationInvalidCredentials AuthenticationCode = 20005
	AuthenticationInvalidToken       AuthenticationCode = 20006
	AuthenticationTokenRevoked       AuthenticationCode = 20007
	AuthenticationTokenExpired       AuthenticationCode = 20008
)

// refer to for best practice: https://learn.microsoft.com/en-us/azure/architecture/best-practices/api-design
const (
	DataAlreadyExist     = 10001
	DataNotFound         = 10002
	DataCreateFailed     = 10003
	DataUpdateFailed     = 10004
	DataDeleteFailed     = 10005
	DataGetFailed        = 10006
	InvalidArgument      = 10007
	OutboundFailed       = 10008
	InboundFailed        = 10009
	DataValidationFailed = 10010
	DataForbidden        = 10011
	DataConflict         = 10012
	BadData              = 10013
	//OtherError         = 10007
)
