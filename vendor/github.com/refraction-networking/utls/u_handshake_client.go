// Copyright 2022 uTLS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

// This function is called by (*clientHandshakeStateTLS13).readServerCertificate()
// to retrieve the certificate out of a message read by (*Conn).readHandshake()
func (hs *clientHandshakeStateTLS13) utlsReadServerCertificate(msg any) (processedMsg any, err error) {
	for _, ext := range hs.uconn.Extensions {
		switch ext.(type) {
		case *UtlsCompressCertExtension:
			// Included Compressed Certificate extension
			if len(hs.uconn.certCompressionAlgs) > 0 {
				compressedCertMsg, ok := msg.(*utlsCompressedCertificateMsg)
				if ok {
					if err = transcriptMsg(compressedCertMsg, hs.transcript); err != nil {
						return nil, err
					}
					msg, err = hs.decompressCert(*compressedCertMsg)
					if err != nil {
						return nil, fmt.Errorf("tls: failed to decompress certificate message: %w", err)
					} else {
						return msg, nil
					}
				}
			}
		default:
			continue
		}
	}
	return nil, nil
}

// called by (*clientHandshakeStateTLS13).utlsReadServerCertificate() when UtlsCompressCertExtension is used
func (hs *clientHandshakeStateTLS13) decompressCert(m utlsCompressedCertificateMsg) (*certificateMsgTLS13, error) {
	var (
		decompressed io.Reader
		compressed   = bytes.NewReader(m.compressedCertificateMessage)
		c            = hs.c
	)

	// Check to see if the peer responded with an algorithm we advertised.
	supportedAlg := false
	for _, alg := range hs.uconn.certCompressionAlgs {
		if m.algorithm == uint16(alg) {
			supportedAlg = true
		}
	}
	if !supportedAlg {
		c.sendAlert(alertBadCertificate)
		return nil, fmt.Errorf("unadvertised algorithm (%d)", m.algorithm)
	}

	switch CertCompressionAlgo(m.algorithm) {
	case CertCompressionBrotli:
		decompressed = brotli.NewReader(compressed)

	case CertCompressionZlib:
		rc, err := zlib.NewReader(compressed)
		if err != nil {
			c.sendAlert(alertBadCertificate)
			return nil, fmt.Errorf("failed to open zlib reader: %w", err)
		}
		defer rc.Close()
		decompressed = rc

	case CertCompressionZstd:
		rc, err := zstd.NewReader(compressed)
		if err != nil {
			c.sendAlert(alertBadCertificate)
			return nil, fmt.Errorf("failed to open zstd reader: %w", err)
		}
		defer rc.Close()
		decompressed = rc

	default:
		c.sendAlert(alertBadCertificate)
		return nil, fmt.Errorf("unsupported algorithm (%d)", m.algorithm)
	}

	rawMsg := make([]byte, m.uncompressedLength+4) // +4 for message type and uint24 length field
	rawMsg[0] = typeCertificate
	rawMsg[1] = uint8(m.uncompressedLength >> 16)
	rawMsg[2] = uint8(m.uncompressedLength >> 8)
	rawMsg[3] = uint8(m.uncompressedLength)

	n, err := decompressed.Read(rawMsg[4:])
	if err != nil {
		c.sendAlert(alertBadCertificate)
		return nil, err
	}
	if n < len(rawMsg)-4 {
		// If, after decompression, the specified length does not match the actual length, the party
		// receiving the invalid message MUST abort the connection with the "bad_certificate" alert.
		// https://datatracker.ietf.org/doc/html/rfc8879#section-4
		c.sendAlert(alertBadCertificate)
		return nil, fmt.Errorf("decompressed len (%d) does not match specified len (%d)", n, m.uncompressedLength)
	}
	certMsg := new(certificateMsgTLS13)
	if !certMsg.unmarshal(rawMsg) {
		return nil, c.sendAlert(alertUnexpectedMessage)
	}
	return certMsg, nil
}

// to be called in (*clientHandshakeStateTLS13).handshake(),
// after hs.readServerFinished() and before hs.sendClientCertificate()
func (hs *clientHandshakeStateTLS13) serverFinishedReceived() error {
	if err := hs.sendClientEncryptedExtensions(); err != nil {
		return err
	}
	return nil
}

func (hs *clientHandshakeStateTLS13) sendClientEncryptedExtensions() error {
	c := hs.c
	clientEncryptedExtensions := new(utlsClientEncryptedExtensionsMsg)
	if c.utls.hasApplicationSettings {
		clientEncryptedExtensions.hasApplicationSettings = true
		clientEncryptedExtensions.applicationSettings = c.utls.localApplicationSettings
		if _, err := c.writeHandshakeRecord(clientEncryptedExtensions, hs.transcript); err != nil {
			return err
		}
	}

	return nil
}

func (hs *clientHandshakeStateTLS13) utlsReadServerParameters(encryptedExtensions *encryptedExtensionsMsg) error {
	hs.c.utls.hasApplicationSettings = encryptedExtensions.utls.hasApplicationSettings
	hs.c.utls.peerApplicationSettings = encryptedExtensions.utls.applicationSettings

	if hs.c.utls.hasApplicationSettings {
		if hs.uconn.vers < VersionTLS13 {
			return errors.New("tls: server sent application settings at invalid version")
		}
		if len(hs.uconn.clientProtocol) == 0 {
			return errors.New("tls: server sent application settings without ALPN")
		}

		// Check if the ALPN selected by the server exists in the client's list.
		if alps, ok := hs.uconn.config.ApplicationSettings[hs.serverHello.alpnProtocol]; ok {
			hs.c.utls.localApplicationSettings = alps
		} else {
			// return errors.New("tls: server selected ALPN doesn't match a client ALPS")
			return nil // ignore if client doesn't have ALPS in use.
			// TODO: is this a issue or not?
		}
	}

	return nil
}
