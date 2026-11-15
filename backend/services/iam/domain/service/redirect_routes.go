package service

// AllowedPaths mengembalikan salinan daftar awalan rute redirect internal yang diizinkan IAM.
func (r RedirectResolver) AllowedPaths() []string {
	out := make([]string, len(r.allowedPrefixes))
	copy(out, r.allowedPrefixes)
	return out
}
