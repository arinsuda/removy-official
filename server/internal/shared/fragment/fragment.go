package fragment

import "net/url"

// URL builds base+path#params. Mailed secrets go in the fragment so they
// are never sent to servers, proxies or Referer headers.
func URL(base, path string, params url.Values) string {
	return base + path + "#" + params.Encode()
}
