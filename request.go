package publish

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

type Request struct {
	baseURL url.URL
	token   string

	Transport http.RoundTripper
}

func NewRequest(baseURL string) (*Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse URL '%s'", baseURL)
	}

	r := &Request{
		baseURL: *u,
	}

	return r, nil
}

// ProcessPost will send an HTTP POST request to the baseURL and query with the
// provided body.
func (r *Request) ProcessPost(urlpath string, query map[string]string, headers map[string]string, body io.Reader) (io.ReadCloser, error) {
	rc, status, err := r.process(http.MethodPost, urlpath, query, headers, body)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to make request to %s/%s", r.baseURL.String(), urlpath)
	}

	if status != http.StatusCreated {
		defer rc.Close()

		b, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, errors.Wrapf(err, "HTTP response (%d) was not expected and failed to read response", status)
		}

		return nil, errors.Errorf("HTTP response (%d) was not expected: %s", status, string(b))
	}

	return rc, nil
}

func (r *Request) process(verb, urlpath string, query map[string]string, headers map[string]string, body io.Reader) (io.ReadCloser, int, error) {
	u := r.baseURL
	u.Path = path.Join(u.Path, urlpath)

	q := u.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	completeURL := u.String()
	req, err := http.NewRequestWithContext(context.Background(), verb, completeURL, body)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "Failed to create request for '%s %s'", verb, completeURL)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	c := http.Client{
		Transport: r.Transport,
	}

	resp, err := c.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil, 0, errors.Wrapf(err, "Failed to perform request")
	}

	return resp.Body, resp.StatusCode, nil
}
