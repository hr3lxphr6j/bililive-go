package upload

type Uploader interface {
	Upload() error
}
