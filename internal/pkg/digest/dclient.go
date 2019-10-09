package digest

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is an interface with a single Do() method, allowing functions to accept
// either an http.Client or a digest.DClient
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// DClient represents an HTTP client used for making requests authenticated
// with http digest authentication.
type DClient struct {
	client     *http.Client
	username   string
	password   string
	snonce     string
	realm      string
	qop        string
	nonceCount uint32
}

// NewDClient returns a DClient that wraps a given standard library http Client with the given username and password
func NewDClient(stdClient *http.Client, username string, password string) *DClient {
	return &DClient{
		client:   stdClient,
		username: username,
		password: password,
	}
}

// Do performs an http request, wrapping it with digest authentication
func (dc *DClient) Do(req *http.Request) (*http.Response, error) {
	return dc.doDigestAuth(req)
}

func (dc *DClient) doDigestAuth(req *http.Request) (*http.Response, error) {
	if dc.snonce != "" {
		req.Header.Set("Authorization", dc.getDigestAuth(req.Method, req.URL.String()))
	}

	// Attempt the request using the underlying client
	resp, err := dc.client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	// We will need to return the response from another request, so defer a close on this one
	defer resp.Body.Close()

	dc.getDigestParts(resp)

	authedReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		return &http.Response{}, fmt.Errorf("http.NewRequest: %v", err.Error())
	}

	authedReq.Header = req.Header
	authedReq.Header.Set("Authorization", dc.getDigestAuth(authedReq.Method, authedReq.URL.String()))

	resp, err = dc.client.Do(authedReq)
	if err != nil {
		return &http.Response{}, err
	}

	// resp.Body intentionally not closed

	return resp, err
}

func (dc *DClient) getDigestParts(resp *http.Response) {
	result := map[string]string{}
	if len(resp.Header["WWW-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop"}
		responseHeaders := strings.Split(resp.Header["WWW-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	dc.snonce = result["nonce"]
	dc.realm = result["realm"]
	dc.qop = result["qop"]
}

func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func (dc *DClient) getDigestAuth(method string, uri string) string {
	ha1 := getMD5(dc.username + ":" + dc.realm + ":" + dc.password)
	ha2 := getMD5(method + ":" + uri)
	cnonce := getCnonce()
	dc.nonceCount++
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, dc.snonce, dc.nonceCount, cnonce, dc.qop, ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		dc.username, dc.realm, dc.snonce, uri, cnonce, dc.nonceCount, dc.qop, response)
	return authorization
}
