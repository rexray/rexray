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

	"github.com/AVENTER-UG/rexray/libstorage/api/types"
	"github.com/AVENTER-UG/rexray/libstorage/api/utils"
)

var errServerFingerprint = errors.New("invalid server fingerprint")

func verifyKnownHost(
	ctx types.Context,
	host string,
	peerCerts []*x509.Certificate,
	knownHost *types.TLSKnownHost) (bool, error) {

	if knownHost == nil {
		return false, nil
	}

	ok, err := verifyPeerCerts(ctx, host, knownHost, peerCerts)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return false, newErrKnownHost(host, peerCerts)
}

func verifyKnownHostFiles(
	ctx types.Context,
	host string,
	peerCerts []*x509.Certificate,
	usrKnownHostsFilePath,
	sysKnownHostsFilePath string) (bool, error) {

	if len(usrKnownHostsFilePath) == 0 && len(sysKnownHostsFilePath) == 0 {
		return false, nil
	}

	if len(usrKnownHostsFilePath) > 0 {
		ok, err := verifyKnownHostsFile(
			ctx, host, peerCerts, usrKnownHostsFilePath)
		if ok {
			return true, nil
		}
		if _, ok := err.(*types.ErrKnownHost); !ok {
			return false, err
		}
	}

	if len(sysKnownHostsFilePath) > 0 {
		return verifyKnownHostsFile(
			ctx, host, peerCerts, sysKnownHostsFilePath)
	}

	return false, newErrKnownHost(host, peerCerts)
}

func verifyKnownHostsFile(
	ctx types.Context,
	host string,
	peerCerts []*x509.Certificate,
	knownHostsFilePath string) (bool, error) {

	r, err := os.Open(knownHostsFilePath)
	if err != nil {
		ctx.WithField("path", knownHostsFilePath).Error(
			"error opening known_hosts file")
		return false, err
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
			return false, err
		}
		if kh == nil {
			continue
		}
		ok, err := verifyPeerCerts(ctx, host, kh, peerCerts)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, newErrKnownHost(host, peerCerts)
}

func verifyPeerCerts(
	ctx types.Context,
	host string,
	knownHost *types.TLSKnownHost,
	peerCerts []*x509.Certificate) (bool, error) {

	szExpectedFP := hex.EncodeToString(knownHost.Fingerprint)

	for _, cert := range peerCerts {
		h := sha256.New()
		h.Write(cert.Raw)
		certFP := h.Sum(nil)

		szActualFP := hex.EncodeToString(certFP)

		logFields := map[string]interface{}{
			"actualFingerprint":   szActualFP,
			"expectedFingerprint": szExpectedFP,
			"actualHost":          host,
			"expectedHost":        knownHost.Host,
		}

		ctx.WithFields(logFields).Debug(
			"comparing tls known host information")

		// does the targeted host equal the saved, known host name?
		if strings.EqualFold(host, knownHost.Host) {

			// are the fingerprints equal? if so this is a validated,
			// known host
			if bytes.EqualFold(knownHost.Fingerprint, certFP) {

				ctx.WithFields(logFields).Debug(
					"matched tls known host information")

				return true, nil
			}

			// the saved fingerprint does not equal the remote, peer
			// fingerprint meaning there is a possible mitm attack
			// where a remote host has usurped another host's identity
			ctx.WithFields(logFields).Error(
				"known host conflict has occurred")

			return false, newErrKnownHostConflict(host, knownHost)
		}

	}

	return false, nil
}

func newErrKnownHost(host string, peerCerts []*x509.Certificate) error {
	err := &types.ErrKnownHost{
		HostName: host,
	}

	if len(peerCerts) == 0 {
		return err
	}

	err.PeerAlg = "sha256"

	h := sha256.New()
	h.Write(peerCerts[0].Raw)
	err.PeerFingerprint = h.Sum(nil)

	return err
}

func newErrKnownHostConflict(
	host string,
	knownHost *types.TLSKnownHost) error {

	return &types.ErrKnownHostConflict{
		HostName:        host,
		KnownHostName:   knownHost.Host,
		PeerAlg:         knownHost.Alg,
		PeerFingerprint: knownHost.Fingerprint,
	}
}
