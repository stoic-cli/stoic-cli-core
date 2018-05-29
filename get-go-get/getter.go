package getter

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/stoic-cli/stoic-cli-core"
	git "github.com/stoic-cli/stoic-cli-core/get-git"
	"github.com/stoic-cli/stoic-cli-core/tool"
	"golang.org/x/tools/go/vcs"
)

func NewGetter(stoic stoic.Stoic, tool stoic.Tool) (tool.Getter, error) {
	importPath := tool.Config().Endpoint
	repo, err := vcs.RepoRootForImportPath(importPath, false)
	if err != nil {
		return nil, err
	}

	if repo.VCS.Name != "Git" {
		return nil, fmt.Errorf("unsupported VCS: %v", repo.VCS.Name)
	}

	repoURL, err := url.Parse(repo.Repo)
	if err != nil {
		return nil, err
	}

	tool.Config().Getter.Options["url"] = repoURL

	vcs, err := git.NewGetter(stoic, tool)
	if err != nil {
		return nil, err
	}

	checkoutPath := strings.Replace(
		"src/"+repo.Root, "/", string(filepath.Separator), -1)

	return &Getter{vcs, checkoutPath}, nil
}

type Getter struct {
	VCS          tool.Getter
	CheckoutPath string
}

func (g Getter) FetchLatest() (tool.Version, error) {
	return g.VCS.FetchLatest()
}

func (g Getter) FetchVersion(pinVersion tool.Version) error {
	return g.VCS.FetchVersion(pinVersion)
}

func (g Getter) CheckoutTo(version tool.Version, path string) error {
	return g.VCS.CheckoutTo(version, filepath.Join(path, g.CheckoutPath))
}
