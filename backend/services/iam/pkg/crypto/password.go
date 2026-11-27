package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	iamerrors "github.com/integrione/erp-retail/backend/services/iam/domain/errors"
	"golang.org/x/crypto/argon2"
)

const (
	minArgon2MemoryKiB  uint32 = 8 * 1024
	maxArgon2MemoryKiB  uint32 = 1024 * 1024
	minArgon2Iterations uint32 = 1
	maxArgon2Iterations uint32 = 20
	minArgon2Parallel   uint8  = 1
	maxArgon2Parallel   uint8  = 16
	minArgon2SaltLen    uint32 = 16
	maxArgon2SaltLen    uint32 = 64
	minArgon2KeyLen     uint32 = 16
	maxArgon2KeyLen     uint32 = 128
)

// Argon2Params mendefinisikan parameter biaya hashing Argon2id sesuai profil keamanan.
type Argon2Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params mengembalikan parameter Argon2id rekomendasi OWASP/RFC 9106 (64 MiB, 3 iterasi, 2 thread).
func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		MemoryKiB:   64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func validateArgon2Params(p Argon2Params) error {
	if p.MemoryKiB < minArgon2MemoryKiB || p.MemoryKiB > maxArgon2MemoryKiB {
		return iamerrors.NewValidation("argon2.memory", iamerrors.Code("iam.crypto.invalid_argon2_memory"), "argon2 memory must be between 8 MiB and 1 GiB")
	}
	if p.Iterations < minArgon2Iterations || p.Iterations > maxArgon2Iterations {
		return iamerrors.NewValidation("argon2.iterations", iamerrors.Code("iam.crypto.invalid_argon2_iterations"), "argon2 iterations out of allowed range")
	}
	if p.Parallelism < minArgon2Parallel || p.Parallelism > maxArgon2Parallel {
		return iamerrors.NewValidation("argon2.parallelism", iamerrors.Code("iam.crypto.invalid_argon2_parallelism"), "argon2 parallelism out of allowed range")
	}
	if p.SaltLength < minArgon2SaltLen || p.SaltLength > maxArgon2SaltLen {
		return iamerrors.NewValidation("argon2.salt_length", iamerrors.Code("iam.crypto.invalid_argon2_salt_length"), "argon2 salt length out of allowed range")
	}
	if p.KeyLength < minArgon2KeyLen || p.KeyLength > maxArgon2KeyLen {
		return iamerrors.NewValidation("argon2.key_length", iamerrors.Code("iam.crypto.invalid_argon2_key_length"), "argon2 key length out of allowed range")
	}
	return nil
}

// Argon2idHasher mengimplementasikan domainservice.PasswordHasher dengan format PHC Argon2id.
type Argon2idHasher struct {
	params     Argon2Params
	randSource io.Reader
}

// NewDefaultArgon2idHasher membentuk Argon2idHasher dengan parameter keamanan baku.
func NewDefaultArgon2idHasher() *Argon2idHasher {
	return &Argon2idHasher{
		params:     DefaultArgon2Params(),
		randSource: rand.Reader,
	}
}

// NewArgon2idHasher membentuk Argon2idHasher setelah memvalidasi batas aman parameter Argon2id.
func NewArgon2idHasher(params Argon2Params) (*Argon2idHasher, error) {
	if err := validateArgon2Params(params); err != nil {
		return nil, err
	}
	return &Argon2idHasher{
		params:     params,
		randSource: rand.Reader,
	}, nil
}

// Hash menghasilkan string hash PHC Argon2id beserta salt CSPRNG unik per pemanggilan.
func (h *Argon2idHasher) Hash(password string) (string, error) {
	if h == nil {
		return "", iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.nil_hasher"), "password hasher is nil")
	}
	if password == "" {
		return "", iamerrors.NewValidation("password", iamerrors.Code("iam.crypto.empty_password"), "password must not be empty")
	}
	if strings.ContainsRune(password, '\x00') {
		return "", iamerrors.NewValidation("password", iamerrors.Code("iam.crypto.null_byte_password"), "password must not contain null byte")
	}

	src := h.randSource
	if src == nil {
		src = rand.Reader
	}
	salt := make([]byte, h.params.SaltLength)
	if _, err := io.ReadFull(src, salt); err != nil {
		return "", iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.salt_entropy_failed"), "failed to generate password salt")
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.MemoryKiB,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.MemoryKiB,
		h.params.Iterations,
		h.params.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// Verify mencocokkan kata sandi polos terhadap string hash PHC Argon2id secara constant-time.
func (h *Argon2idHasher) Verify(password, encodedHash string) (bool, error) {
	if password == "" {
		return false, iamerrors.NewValidation("password", iamerrors.Code("iam.crypto.empty_password"), "password must not be empty")
	}
	if strings.TrimSpace(encodedHash) == "" {
		return false, iamerrors.NewValidation("password_hash", iamerrors.Code("iam.crypto.empty_hash"), "encoded hash must not be empty")
	}

	params, salt, expectedHash, err := decodeArgon2idPHC(encodedHash)
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.MemoryKiB,
		params.Parallelism,
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}

func decodeArgon2idPHC(encoded string) (Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return Argon2Params{}, nil, nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.invalid_phc_format"), "invalid argon2id hash format")
	}

	versionPart := strings.TrimPrefix(parts[2], "v=")
	version, err := strconv.Atoi(versionPart)
	if err != nil || version != argon2.Version {
		return Argon2Params{}, nil, nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.unsupported_argon2_version"), "unsupported argon2id version")
	}

	var params Argon2Params
	for _, kv := range strings.Split(parts[3], ",") {
		pair := strings.SplitN(kv, "=", 2)
		if len(pair) != 2 {
			return Argon2Params{}, nil, nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.invalid_phc_params"), "invalid argon2id parameters")
		}
		val, parseErr := strconv.ParseUint(pair[1], 10, 32)
		if parseErr != nil {
			return Argon2Params{}, nil, nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.invalid_phc_params"), "invalid argon2id parameter value")
		}
		switch pair[0] {
		case "m":
			params.MemoryKiB = uint32(val)
		case "t":
			params.Iterations = uint32(val)
		case "p":
			if val > 255 {
				return Argon2Params{}, nil, nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.invalid_phc_params"), "argon2id parallelism overflow")
			}
			params.Parallelism = uint8(val)
		default:
			return Argon2Params{}, nil, nil, iamerrors.New(iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.invalid_phc_params"), "unknown argon2id parameter")
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2Params{}, nil, nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.invalid_phc_salt"), "failed to decode argon2id salt")
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2Params{}, nil, nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.Code("iam.crypto.invalid_phc_digest"), "failed to decode argon2id digest")
	}

	params.SaltLength = uint32(len(salt))
	params.KeyLength = uint32(len(expectedHash))
	if err := validateArgon2Params(params); err != nil {
		return Argon2Params{}, nil, nil, iamerrors.Wrap(err, iamerrors.KindUnavailable, iamerrors.CodeOf(err), "stored argon2id parameters are invalid")
	}
	return params, salt, expectedHash, nil
}
