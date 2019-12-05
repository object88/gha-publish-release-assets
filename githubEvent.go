package publish

import (
	"encoding/json"
	"os"

	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
)

type GithubEvent struct {
	releaseIDJMESPath *jmespath.JMESPath
	data              interface{}
}

func NewGithubEvent(file string) (*GithubEvent, error) {
	query := ".release.id"
	releaseIDJMESPath, err := jmespath.Compile(query)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to compile query '%s'", query)
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Faield to open file '%s'", file)
	}
	defer f.Close()

	var data interface{}
	dec := json.NewDecoder(f)
	dec.Decode(&data)

	ge := &GithubEvent{
		data:              data,
		releaseIDJMESPath: releaseIDJMESPath,
	}
	return ge, nil
}

func (ge *GithubEvent) ReleaseID() (string, error) {
	result, err := ge.releaseIDJMESPath.Search(ge.data)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to execute jmes query")
	}

	if result == nil {
		return "", errors.Wrapf(err, "Did not find release ID")
	}

	sresult, ok := result.(string)
	if !ok {
		return "", errors.Errorf("Found result ID '%s', could not cast to string", result)
	}

	return sresult, nil
}
