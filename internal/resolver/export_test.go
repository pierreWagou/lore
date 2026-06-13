// export_test.go exposes internal functions for use by external test packages.
// This file is compiled only during testing.
package resolver

// FetchArchiveForTest calls fetchArchiveWithBase with a custom baseURL,
// allowing tests to point archive requests at a mock HTTP server.
func FetchArchiveForTest(h Handle, token, cacheDir, baseURL string) (FetchResult, error) {
	return fetchArchiveWithBase(h, token, cacheDir, baseURL)
}
