package poker

import "errors"

var ErrNoRecord = errors.New("poker: no matching record found")
var ErrNoUserId = errors.New("poker: userId is required")
