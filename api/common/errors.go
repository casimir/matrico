package common

type Error struct {
	Code    string `json:"errcode"`
	Message string `json:"error"`
	Status  int    `json:"-"`
}

func (e Error) Error() string { return e.Code + ": " + e.Message }

var (
	ErrBadJSON      = Error{"M_BAD_JSON", "Request contained valid JSON, but it was malformed in some way, e.g. missing required keys, invalid values for keys.", 400}
	ErrForbidden    = Error{"M_FORBIDDEN", "Forbidden access, e.g. joining a room without permission, failed login.", 403}
	ErrMissingToken = Error{"M_MISSING_TOKEN", "No access token was specified for the request.", 401}
	ErrNotFound     = Error{"M_NOT_FOUND", "No resource was found for this request.", 404}
	ErrNotJSON      = Error{"M_NOT_JSON", "Request did not contain valid JSON.", 400}
	ErrUnknown      = Error{"M_UNKNOWN", "An unknown error has occurred.", 400}
	ErrUnknownToken = Error{"M_UNKNOWN_TOKEN", "The access token specified was not recognised.", 403}
)

func New(msg string) Error {
	return Error{"M_UNKNOWN", msg, 400}
}
