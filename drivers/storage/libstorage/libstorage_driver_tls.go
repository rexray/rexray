package libstorage

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	"github.com/codedellemc/libstorage/api/types"
	"github.com/codedellemc/libstorage/api/utils"
)

var errServerFingerprint = errors.New("invalid server fingerprint")

func verifyKnownHost(
	ctx types.Context,
	peerCerts []*x509.Certificate,
	knownHost *types.TLSKnownHost) error {

	if knownHost == nil {
		return nil
	}

	expectedFP := hex.EncodeToString(knownHost.Fingerprint)
	for _, cert := range peerCerts {
		h := sha256.New()
		h.Write(cert.Raw)
		certFP := h.Sum(nil)
		actualFP := hex.EncodeToString(certFP)
		ctx.WithFields(map[string]interface{}{
			"actualFingerprint":   actualFP,
			"expectedFingerprint": expectedFP,
			"actualHost":          cert.Subject.CommonName,
			"expectedHost":        knownHost.Host,
		}).Debug("comparing tls known host information")
		if bytes.EqualFold(knownHost.Fingerprint, certFP) &&
			strings.EqualFold(knownHost.Host, cert.Subject.CommonName) {
			ctx.WithFields(map[string]interface{}{
				"actualFingerprint":   actualFP,
				"expectedFingerprint": expectedFP,
				"actualHost":          cert.Subject.CommonName,
				"expectedHost":        knownHost.Host,
			}).Debug("matched tls known host information")
			return nil
		}
	}
	return errServerFingerprint
}

func verifyKnownHostFiles(
	ctx types.Context,
	peerCerts []*x509.Certificate,
	usrKnownHostsFilePath,
	sysKnownHostsFilePath string) error {

	if len(usrKnownHostsFilePath) == 0 && len(sysKnownHostsFilePath) == 0 {
		return nil
	}

	if len(usrKnownHostsFilePath) > 0 {
		err := verifyKnownHostsFile(ctx, peerCerts, usrKnownHostsFilePath)
		if err == nil {
			return nil
		}
		if _, ok := err.(*types.ErrKnownHost); !ok {
			return err
		}
	}

	if len(sysKnownHostsFilePath) > 0 {
		return verifyKnownHostsFile(ctx, peerCerts, sysKnownHostsFilePath)
	}

	return newErrKnownHost(peerCerts)
}

func verifyKnownHostsFile(
	ctx types.Context,
	peerCerts []*x509.Certificate,
	knownHostsFilePath string) error {

	r, err := os.Open(knownHostsFilePath)
	if err != nil {
		ctx.WithField("path", knownHostsFilePath).Error(
			"error opening known_hosts file")
		return err
	}
	defer r.Close()

	ctx.WithField("path", knownHostsFilePath).Debug("opened known_hosts file")

	scn := bufio.NewScanner(r)
	for scn.Scan() {
		l := scn.Text()
		if len(l) == 0 {
			continue
		}
		ctx.WithField("line", l).Debug("scanning known_hosts file")
		kh, err := utils.ParseKnownHost(ctx, l)
		if err != nil {
			ctx.WithField("path", knownHostsFilePath).Error(
				"error scanning known_hosts file")
			return err
		}
		if kh == nil {
			continue
		}
		expectedFP := hex.EncodeToString(kh.Fingerprint)
		for _, cert := range peerCerts {
			h := sha256.New()
			h.Write(cert.Raw)
			certFP := h.Sum(nil)
			actualFP := hex.EncodeToString(certFP)
			ctx.WithFields(map[string]interface{}{
				"actualFingerprint":   actualFP,
				"expectedFingerprint": expectedFP,
				"actualHost":          cert.Subject.CommonName,
				"expectedHost":        kh.Host,
			}).Debug("comparing tls known host information")
			if bytes.EqualFold(kh.Fingerprint, certFP) &&
				strings.EqualFold(kh.Host, cert.Subject.CommonName) {
				ctx.WithFields(map[string]interface{}{
					"actualFingerprint":   actualFP,
					"expectedFingerprint": expectedFP,
					"actualHost":          cert.Subject.CommonName,
					"expectedHost":        kh.Host,
				}).Debug("matched tls known host information")
				return nil
			}
		}
	}

	return newErrKnownHost(peerCerts)
}

func newErrKnownHost(peerCerts []*x509.Certificate) error {
	err := &types.ErrKnownHost{}

	if len(peerCerts) == 0 {
		return err
	}

	err.PeerHost = peerCerts[0].Subject.CommonName
	err.PeerAlg = "sha256"

	h := sha256.New()
	h.Write(peerCerts[0].Raw)
	err.PeerFingerprint = h.Sum(nil)

	return err
}
