package exceptions

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Failures on the server side (storage, token generation, ...) are 500 so
// clients can tell them apart from bad input, which stays 422.
func TestBusinessLogicReason_HTTPStatus(t *testing.T) {
	cases := map[string]struct {
		code int
		want int
	}{
		"get failed":        {DataGetFailed, http.StatusInternalServerError},
		"create failed":     {DataCreateFailed, http.StatusInternalServerError},
		"update failed":     {DataUpdateFailed, http.StatusInternalServerError},
		"delete failed":     {DataDeleteFailed, http.StatusInternalServerError},
		"bad data":          {BadData, http.StatusInternalServerError},
		"conflict":          {DataConflict, http.StatusUnprocessableEntity},
		"already exist":     {DataAlreadyExist, http.StatusUnprocessableEntity},
		"invalid argument":  {InvalidArgument, http.StatusUnprocessableEntity},
		"validation failed": {DataValidationFailed, http.StatusUnprocessableEntity},
		"not found":         {DataNotFound, http.StatusNotFound},
		"forbidden":         {DataForbidden, http.StatusForbidden},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, BusinessLogicReason(tc.code).HttpCode)
		})
	}
}
