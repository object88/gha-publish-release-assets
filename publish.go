package publish

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type GithubConfig struct {
	Auth       string
	ReleaseID  string
	Repository string
	Token      string
	Workspace  string
}

type Publisher struct {
	Github GithubConfig

	req *Request

	includes []glob.Glob
	excludes []glob.Glob
}

func NewPublisher(req *Request) *Publisher {
	return &Publisher{
		Github: GithubConfig{},
		req:    req,
	}
}

func (p *Publisher) AddInclude(in string) error {
	g, err := glob.Compile(in)
	if err != nil {
		return errors.Wrapf(err, "Failed to compile '%s' for inclusion", in)
	}
	p.includes = append(p.includes, g)
	return nil
}

func (p *Publisher) AddExclude(ex string) error {
	g, err := glob.Compile(ex)
	if err != nil {
		return errors.Wrapf(err, "Failed to compile '%s' for exclusion", ex)
	}
	p.excludes = append(p.excludes, g)
	return nil
}

func (p *Publisher) Publish() error {
	pathchan := make(chan string, 10)
	done := make(chan struct{}, 1)

	go func() {
		for filepath := range pathchan {
			fmt.Printf("Acceptable path: %s\n", filepath)
		}
		done <- struct{}{}
	}()

	c := len(p.Github.Workspace)
	if !strings.HasSuffix(p.Github.Workspace, string(os.PathSeparator)) {
		c++
	}
	err := filepath.Walk(p.Github.Workspace, func(filepath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		testpath := filepath[c:]

		fmt.Printf("Have path %s", testpath)

		apply := false
		for _, g := range p.includes {
			fmt.Printf("; glob: %#v", g)
			if g.Match(testpath) {
				apply = true
				break
			}
		}

		if !apply {
			fmt.Printf("... not included\n")
			return nil
		}

		for _, g := range p.excludes {
			if g.Match(testpath) {
				// An exclusion matched; move on
				fmt.Printf("... excluded\n")
				return nil
			}
		}

		// path is included!
		fmt.Printf("... processing\n")
		err = p.sendfile(filepath)
		if err != nil {
			return err
		}

		return nil
	})

	close(pathchan)

	<-done

	return err
}

func (p *Publisher) sendfile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return errors.Wrapf(err, "Failed to open file '%s' for upload", filepath)
	}
	defer f.Close()

	filename := path.Base(filepath)

	urlpath := fmt.Sprintf("repos/%s/releases/%s/assets", p.Github.Repository, p.Github.ReleaseID)
	query := map[string]string{"name": filename}
	headers := map[string]string{
		"Authorization": fmt.Sprintf("token %s", p.Github.Auth),
	}
	rc, err := p.req.ProcessPost(urlpath, query, headers, f)
	if err != nil {
		return errors.Wrapf(err, "Failed to upload file '%s'", filepath)
	}
	defer rc.Close()

	return nil
}
