package error_helpers

import (
	"errors"
	"fmt"

	"github.com/turbot/flowpipe/pipeparser/constants"
)

var MissingCloudTokenError = fmt.Errorf("Not authenticated for Turbot Pipes.\nPlease run %s or setup a token.", constants.Bold("steampipe login"))
var InvalidCloudTokenError = fmt.Errorf("Invalid token.\nPlease run %s or setup a token.", constants.Bold("steampipe login"))
var InvalidStateError = errors.New("invalid state")