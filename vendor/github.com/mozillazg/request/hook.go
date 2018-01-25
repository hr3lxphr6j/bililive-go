package request

import "net/http"

// Hook ...
type Hook interface {
	// call BeforeRequest before send http request,
	// if resp != nil or err != nil
	// use this resp and err, no longer send http request.
	BeforeRequest(req *http.Request) (resp *http.Response, err error)
	// call AfterRequest after got response
	// if newResp != nil or newErr != nil
	// use the new NewResp instead of origin response.
	AfterRequest(req *http.Request, resp *http.Response, err error) (newResp *http.Response, newErr error)
}

func applyBeforeReqHooks(req *http.Request, hooks []Hook) (resp *http.Response, err error) {
	for _, hook := range hooks {
		resp, err = hook.BeforeRequest(req)
		if resp != nil || err != nil {
			return
		}
	}
	return
}

func applyAfterReqHooks(req *http.Request, resp *http.Response, err error, hooks []Hook) (newResp *http.Response, newErr error) {
	for _, hook := range hooks {
		newResp, newErr = hook.AfterRequest(req, resp, err)
		if newResp != nil || newErr != nil {
			return
		}
	}
	return
}
