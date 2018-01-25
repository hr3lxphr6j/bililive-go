# Changelog


## 0.8.0 (2016-12-20)

* Add `Request.Hooks` for supports before and after request hooks (via #26)
* Add `var ErrMaxRedirect = errors.New("Exceeded max redirects")`


## 0.7.0 (2016-11-07)

* Add auto decompress deflated body (via [#23][#23]. thanks [@DavidCai1993][@DavidCai1993])


## 0.6.0 (2016-09-04)

* Remove vendor/, use glide instead

### API Changes

* Add `var DefaultClient = new(http.Client)`
* Support `NewArgs(nil)`
* Support `NewRequest(nil)`
* Support `GET(url, nil), POST(url, nil), ...`


## 0.5.1 (2016-06-05)

* Add vendor/


## 0.5.0 (2015-11-15)

### API Changes

* Add `func (req *Request) PostForm(url interface{}, data interface{}) (resp *Response, err error)`
* Add `func (req *Request) Reset()`


## 0.4.0 (2015-10-31)

### API Changes

* `type Request struct` add `Body` field
* Add new variables:
  * `DefaultHeaders`
  * `DefaultContentType`
  * `DefaultJsonType`
  * `DefaultRedirectLimit`
  * `DefaultUserAgent`

## 0.3.1 (2015-09-24)

* Updated docs
* Improve code
* No API change, No bugfix

## 0.3.0 (2015-02-11)

### API Changes

* Add `type Request struct`, recommend use it instead of `Args`
* Add `func NewRequest(c *http.Client) *Request`
* Add `func (req *Request) Get(url interface{}) (resp *Response, err error)`
* Add `func (req *Request) Head(url interface{}) (resp *Response, err error)`
* Add `func (req *Request) Post(url interface{}) (resp *Response, err error)`
* Add `func (req *Request) Put(url interface{}) (resp *Response, err error)`
* Add `func (req *Request) Patch(url interface{}) (resp *Response, err error)`
* Add `func (req *Request) Delete(url interface{}) (resp *Response, err error)`
* Add `func (req *Request) Options(url interface{}) (resp *Response, err error)`


## 0.2.0 (2015-01-15)

### API Changes

* Add `Args.Proxy` for support HTTP/HTTPS/SOCKS5 proxy
* Add `Args.BasicAuth` for Support HTTP Basic Authentication
* Add `func (resp *Response) URL() (*url.URL, error)`
* Add `var DefaultRedirectLimit = 10`

### Bugfixes

* Fix "http.Client don't use original Header when it do redirect" [#6](https://github.com/mozillazg/request/issues/6)


## 0.1.0 (2015-01-08)

* Initial Release

[#23]: https://github.com/mozillazg/request/pull/23
[@DavidCai1993]: https://github.com/DavidCai1993
[#26]: https://github.com/mozillazg/request/pull/26
