// Package leaky is a fixture for internal/archtest: a "service" that leaks
// persistence types. It lives under testdata/, so `go build` never compiles it.
package leaky

import (
	"database/sql"
	"fmt"

	"github.com/letenk/golang-authentication/db/sqlcgen"
)

var _ = fmt.Sprint
var _ sql.NullString
var _ sqlcgen.User
