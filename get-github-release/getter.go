package getter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/mitchellh/mapstructure"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

func NewGetter(stoic stoic.Stoic, tool stoic.Tool) (tool.Getter, error) {
	endpoint := tool.Endpoint()

	var options ghrGetterOptions
	err := mapstructure.Decode(tool.Config().Getter.Options, &options)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("").Parse(options.Asset)
	if err != nil {
		return nil, err
	}

	return &ghrGetter{
		Stoic:      stoic,
		Endpoint:   endpoint,
		AssetTempl: tmpl,
	}, nil
}

type ghrGetterOptions struct {
	Asset string
}

type ghrGetter struct {
	Stoic      stoic.Stoic
	Endpoint   *url.URL
	AssetTempl *template.Template
}

func (gg ghrGetter) getRepositoriesServices() (*github.RepositoriesService, error) {
	var client *github.Client
	var err error

	if gg.Endpoint.Hostname() == "github.com" {
		client = github.NewClient(nil)
	} else {
		apiBase, _ := url.Parse("/api/v3/")
		baseURL := gg.Endpoint.ResolveReference(apiBase).String()
		client, err = github.NewEnterpriseClient(baseURL, baseURL, nil)
	}

	if err != nil {
		return nil, err
	}
	return client.Repositories, nil
}

func (gg ghrGetter) getOwnerRepo() (string, string) {
	path := gg.Endpoint.EscapedPath()
	parts := strings.SplitN(path, "/", 4)

	if len(parts) < 3 {
		return "", ""
	}
	return parts[1], parts[2]
}

func (gg ghrGetter) getAssetName(version tool.Version) (string, error) {
	var builder strings.Builder

	parameters := gg.Stoic.Parameters()
	parameters["Version"] = string(version)

	err := gg.AssetTempl.Execute(&builder, parameters)
	if err != nil {
		return "", err
	}
	return builder.String(), nil
}

func (gg ghrGetter) getCacheKey(version tool.Version, assetName string) string {
	host := gg.Endpoint.Hostname()
	owner, repo := gg.getOwnerRepo()
	return strings.Join(
		[]string{"ghr", host, owner, repo, string(version), assetName}, "/")
}

func (gg ghrGetter) getRelease(version tool.Version, wantLatest bool) (tool.Version, error) {
	repos, err := gg.getRepositoriesServices()
	if err != nil {
		return tool.NullVersion, err
	}

	owner, repo := gg.getOwnerRepo()
	if owner == "" || repo == "" {
		return tool.NullVersion, fmt.Errorf(
			"unable to infer owner and repository from %v", gg.Endpoint)
	}

	var release *github.RepositoryRelease

	if wantLatest {
		release, _, err = repos.GetLatestRelease(context.Background(), owner, repo)
	} else {
		release, _, err = repos.GetReleaseByTag(context.Background(), owner, repo, string(version))
	}

	if err != nil {
		return tool.NullVersion, err
	}

	version = tool.Version(*release.TagName)

	assetName, err := gg.getAssetName(version)
	if err != nil {
		return tool.NullVersion, err
	}

	for _, asset := range release.Assets {
		if assetName == *asset.Name {
			req, err := http.NewRequest("GET", *asset.URL, nil)
			if err != nil {
				return tool.NullVersion, err
			}
			req.Header.Add("Accept", "application/octet-stream")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return tool.NullVersion, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return tool.NullVersion, fmt.Errorf(
					"unexpected HTTP status while fetching %v for version %v of %v",
					assetName, version, gg.Endpoint)
			}

			err = gg.Stoic.Cache().Put(gg.getCacheKey(version, assetName), resp.Body)
			if err != nil {
				return tool.NullVersion, err
			}

			return version, nil
		}
	}

	return tool.NullVersion, fmt.Errorf(
		"release %v has no asset matching %v", version, assetName)
}

func (gg ghrGetter) FetchLatest() (tool.Version, error) {
	return gg.getRelease(tool.NullVersion, true)
}

func (gg ghrGetter) FetchVersion(version tool.Version) error {
	_, err := gg.getRelease(version, false)
	return err
}

func (gg ghrGetter) CheckoutTo(version tool.Version, path string) error {
	assetName, err := gg.getAssetName(version)
	if err != nil {
		return nil
	}

	assetPath := filepath.Join(path, assetName)
	asset, err := os.Create(assetPath)
	if err != nil {
		return err
	}
	defer asset.Close()

	cacheReader := gg.Stoic.Cache().Get(gg.getCacheKey(version, assetName))
	if cacheReader == nil {
		return fmt.Errorf(
			"unable to retrieve %v for version %v of %v from cache",
			assetName, version, gg.Endpoint)
	}
	defer cacheReader.Close()

	_, err = io.Copy(asset, cacheReader)
	if err != nil {
		return err
	}

	return nil
}
