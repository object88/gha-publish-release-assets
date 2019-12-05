package publish

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/object88/gha-publish-release-assets/mocks"
)

func Test_Publish(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", uuid.New().String())
	os.MkdirAll(path.Join(tempdir, "data"), 0777)
	ioutil.WriteFile(path.Join(tempdir, "data", "a.txt"), []byte(`aa`), 0666)
	ioutil.WriteFile(path.Join(tempdir, "data", "a.json"), []byte(`{"aa":"bb"}`), 0666)
	ioutil.WriteFile(path.Join(tempdir, "data", "b.yaml"), []byte(`aa: bb`), 0666)
	os.MkdirAll(path.Join(tempdir, "data", "subdata"), 0777)
	ioutil.WriteFile(path.Join(tempdir, "data", "subdata", "a.ini"), []byte(`aa: bb`), 0666)
	os.MkdirAll(path.Join(tempdir, "moardata"), 0777)
	ioutil.WriteFile(path.Join(tempdir, "moardata", "a.properties"), []byte(`aa: bb`), 0666)

	tcs := []struct {
		name     string
		includes []string
		excludes []string
		expected []string
	}{
		{
			name: "none",
		},
		{
			name:     "exact",
			includes: []string{"data/a.txt"},
			expected: []string{"a.txt"},
		},
		{
			name:     "exact",
			includes: []string{"data/a.txt"},
			expected: []string{"a.txt"},
		},
		{
			name:     "all a",
			includes: []string{"**/a.*"},
			expected: []string{"a.ini", "a.json", "a.properties", "a.txt"},
		},
		{
			name:     "all data",
			includes: []string{"data/*"},
			expected: []string{"a.ini", "a.json", "a.txt", "b.yaml"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mrt := mocks.NewMockRoundTripper(ctrl)

			actual := map[string]bool{}

			doFunc := func(args interface{}) {
				req, ok := args.(*http.Request)
				if !ok {
					t.Errorf("Expected a request, got %#v", args)
				}

				// Check the headers
				authheader := req.Header.Get("Authorization")
				if authheader != "token foo" {
					t.Errorf("Incorrect auth header '%s'", authheader)
				}

				// Check the path
				urlpath := req.URL.Path
				if !strings.HasPrefix(urlpath, "/repos/myrepo/releases/1234") {
					t.Errorf("Incorrect path '%s'", urlpath)
				}

				// Check the query
				name := req.URL.Query().Get("name")
				if _, ok := actual[name]; ok {
					t.Errorf("Have repeated upload for asset '%s'", name)
				}
				actual[name] = true
			}

			for range tc.expected {
				httpResp := &http.Response{
					Body:       ioutil.NopCloser(strings.NewReader("{}")),
					Status:     "Created",
					StatusCode: http.StatusCreated,
				}
				mrt.EXPECT().RoundTrip(gomock.Any()).Do(doFunc).Return(httpResp, nil)
			}

			req, err := NewRequest("http://example.com")
			if err != nil {
				t.Fatalf("Internal error: failed to set up test:\n%s\n", err.Error())
			}
			req.Transport = mrt

			p := &Publisher{
				Github: GithubConfig{
					Auth:       "foo",
					ReleaseID:  "1234",
					Repository: "myrepo",
					Workspace:  tempdir,
				},
				req: req,
			}

			for _, v := range tc.includes {
				p.AddInclude(v)
			}
			for _, v := range tc.excludes {
				p.AddExclude(v)
			}

			err = p.Publish()
			if err != nil {
				t.Errorf("Got unexpected error:\n%s\n", err.Error())
			}

			if len(tc.expected) != len(actual) {
				t.Errorf("Mismatched number of calls: expected %d, actual %d", len(tc.expected), len(actual))
			}
			for _, v := range tc.expected {
				if _, ok := actual[v]; !ok {
					t.Errorf("Did not get asset upload for '%s'", v)
				}
			}
			if t.Failed() {
				t.Logf("Expected: %#v\nActual:   %#v", tc.expected, actual)
			}
		})
	}
}
