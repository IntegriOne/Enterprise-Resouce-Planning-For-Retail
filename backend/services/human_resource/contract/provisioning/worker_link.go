// Package provisioning mendefinisikan bukti penautan akun employee untuk RPC privat HRM.
package provisioning

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"

	"github.com/google/uuid"
)

// ErrLinkUnconfirmed menandai hasil RPC yang belum dapat direkonsiliasi.
var ErrLinkUnconfirmed = errors.New("worker account link could not be confirmed")

// WorkerLinkEvidence mengikat bukti provisioning ke pekerja, pihak, akun baru, dan operator IAM.
func WorkerLinkEvidence(key []byte, workerID, partyID, principalID, actorID string) (string, error) {
	if len(key) < sha256.Size {
		return "", errors.New("worker link evidence requires a 32-byte key")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("integrione/hrm/worker-principal-link/v1"))
	for _, id := range []string{workerID, partyID, principalID, actorID} {
		value, err := uuid.Parse(id)
		if err != nil {
			return "", err
		}
		if value == uuid.Nil {
			return "", errors.New("worker link evidence requires a nonzero identifier")
		}
		mac.Write(value[:])
	}
	var evidence uuid.UUID
	copy(evidence[:], mac.Sum(nil))
	evidence[6] = (evidence[6] & 0x0f) | 0x80
	evidence[8] = (evidence[8] & 0x3f) | 0x80
	return evidence.String(), nil
}
