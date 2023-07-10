package altsvc

// Jar is a container of AltSvc.
type Jar interface {
	// SetAltSvc store the AltSvc.
	SetAltSvc(addr string, as *AltSvc)
	// GetAltSvc get the AltSvc.
	GetAltSvc(addr string) *AltSvc
}
