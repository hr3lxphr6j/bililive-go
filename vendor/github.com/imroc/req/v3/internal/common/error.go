package common

import "errors"

// ErrRequestCanceled is a copy of net/http's common.ErrRequestCanceled because it's not
// exported. At least they'll be DeepEqual for h1-vs-h2 comparisons tests.
var ErrRequestCanceled = errors.New("net/http: request canceled")
