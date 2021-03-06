package getter

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/tool"
	"gopkg.in/src-d/go-git.v4"
	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

func NewGetter(stoic stoic.Stoic, tool stoic.Tool) (tool.Getter, error) {
	var options Options
	err := mapstructure.Decode(tool.Config().Getter.Options, &options)
	if err != nil {
		return nil, err
	}

	if options.URL == nil {
		options.URL = tool.Endpoint()
	}
	if options.Branch == "" {
		options.Branch = Branch(tool.Channel())
	}

	if !options.Branch.IsValid() {
		return nil, errors.Errorf("invalid branch name, '%v'", options.Branch)
	}

	pathElems := []string{stoic.Root(), "git", options.URL.Hostname()}
	pathElems = append(pathElems, strings.Split(options.URL.EscapedPath(), "/")...)
	gitDir := filepath.Join(pathElems...)

	return &Getter{options, gitDir}, nil
}

type Options struct {
	URL    *url.URL
	Branch Branch
}

type Getter struct {
	Options
	gitDir string
}

func (gg Getter) runNativeGit(command string, args ...string) error {
	var environment []string
	for _, envVar := range os.Environ() {
		switch strings.Split(envVar, "=")[0] {
		case "GIT_DIR",
			"GIT_INDEX_FILE",
			"GIT_WORK_TREE",
			"GIT_OBJECT_DIRECTORY",
			"GIT_ALTERNATE_OBJECT_DIRECTORIES":
			continue

		default:
			environment = append(environment, envVar)
		}
	}
	environment = append(environment, "GIT_DIR="+gg.gitDir)

	cmd := exec.Command("git", command)
	cmd.Args = append(cmd.Args, args...)

	cmd.Env = environment

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (gg Getter) remoteReference() string {
	if gg.Branch == "" {
		return "HEAD"
	}
	return "refs/heads/" + string(gg.Branch)
}

func (gg Getter) localReference() string {
	if gg.Branch == "" {
		return "refs/remotes/origin/HEAD"
	}
	return "refs/remotes/origin/" + string(gg.Branch)
}

func (gg Getter) fetch() (gitplumbing.Revision, error) {
	_, err := git.PlainOpen(gg.gitDir)
	if err == git.ErrRepositoryNotExists {
		_, err = git.PlainInit(gg.gitDir, true)
	}
	if err != nil {
		return "", err
	}

	localRef := gg.localReference()
	refspec := fmt.Sprintf("+%v:%v", gg.remoteReference(), localRef)

	// invoke native git for the authentication
	url, _ := gg.URL.MarshalBinary()
	err = gg.runNativeGit("fetch", "--quiet", string(url), refspec)
	if err != nil {
		return "", err
	}

	return gitplumbing.Revision(localRef), nil
}

func (gg Getter) FetchLatest() (tool.Version, error) {
	ref, err := gg.fetch()
	if err != nil {
		return tool.NullVersion, err
	}
	repo, err := git.PlainOpen(gg.gitDir)
	if err != nil {
		return tool.NullVersion, err
	}
	version, err := repo.ResolveRevision(ref)
	if err != nil {
		return tool.NullVersion, err
	}
	return tool.Version(version.String()), nil
}

func (gg Getter) FetchVersion(pinVersion tool.Version) error {
	localRef, err := gg.fetch()
	if err != nil {
		return err
	}

	pinHash := gitplumbing.NewHash(string(pinVersion))
	repo, err := git.PlainOpen(gg.gitDir)
	if err != nil {
		return err
	}
	pinCommit, err := repo.CommitObject(pinHash)
	if err != nil {
		return err
	}
	tipHash, err := repo.ResolveRevision(localRef)
	if err != nil {
		return err
	}
	tipCommit, err := repo.CommitObject(*tipHash)
	if err != nil {
		return err
	}

	// Ensure commit is reachable from branch?
	cIter := gitobject.NewCommitPostorderIter(tipCommit, nil)
	defer cIter.Close()

	for {
		commit, err := cIter.Next()
		switch err {
		case nil:
			if bytes.Equal(pinCommit.Hash[:], commit.Hash[:]) {
				return nil
			}

		case io.EOF:
			return errors.Errorf(
				"requested version, %.12v, is unreachable from %v branch",
				pinCommit.Hash.String(), gg.Branch)

		default:
			return err
		}
	}
}

func (gg Getter) CheckoutTo(version tool.Version, path string) error {
	dstGitDir := filepath.Join(path, ".git")

	dstHeads := filepath.Join(dstGitDir, "refs", "heads")
	err := os.MkdirAll(dstHeads, 0777)
	if err != nil {
		return err
	}

	dstRemotes := filepath.Join(dstGitDir, "refs", "remotes")
	err = os.MkdirAll(dstRemotes, 0777)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(dstHeads, "master"), []byte(version), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dstGitDir, "HEAD"), []byte("ref: refs/heads/master"), 0644)
	if err != nil {
		return err
	}

	dstObjectsDir := filepath.Join(dstGitDir, "objects")
	srcObjectsDir := filepath.Join(gg.gitDir, "objects")
	err = os.Symlink(srcObjectsDir, dstObjectsDir)
	if err != nil {
		return err
	}

	dstOrigin := filepath.Join(dstGitDir, "refs", "remotes", "origin")
	srcOrigin := filepath.Join(gg.gitDir, "refs", "remotes", "origin")
	err = os.Symlink(srcOrigin, dstOrigin)
	if err != nil {
		return err
	}

	config, err := os.Create(filepath.Join(dstGitDir, "config"))
	if err != nil {
		return err
	}

	fmt.Fprintf(config, `[remote "origin"]
url = %v
refspec = +refs/heads/*:refs/remotes/origin/*
[branch "%v"]
remote = origin
`, gg.URL, "master")
	config.Close()

	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	versionHash := gitplumbing.NewHash(string(version))
	err = wt.Reset(&git.ResetOptions{
		Commit: versionHash,
		Mode:   git.HardReset,
	})
	if err != nil {
		return err
	}

	return nil
}
