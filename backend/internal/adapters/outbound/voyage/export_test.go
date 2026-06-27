package voyage

// SetBaseURLForTest points the embedder at a stub server. Test-only.
func SetBaseURLForTest(e *Embedder, url string) { e.baseURL = url }
