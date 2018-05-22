package engine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stoic-cli/stoic-cli-core"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

var (
	ErrUpstreamVersionIsEmpty = errors.New("upstream version is empty")
)

type plainCheckout struct {
	version tool.Version
	path    string
}

func (uc plainCheckout) Version() tool.Version { return uc.version }
func (uc plainCheckout) Path() string          { return uc.path }

func (e engine) makeCheckout(t stoic.Tool, version tool.Version, getter tool.Getter, runner tool.Runner) (tool.Checkout, error) {
	endpoint := t.Endpoint()

	parts := []string{e.checkoutsDir, endpoint.Hostname()}
	parts = append(parts, strings.Split(endpoint.EscapedPath(), "/")...)
	baseDir := filepath.Join(parts...)

	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to create checkout directory for version %v of %v: %v",
			version, t.Name(), err)
	}

	checkoutPath, err := ioutil.TempDir(baseDir, string(version)+"-")
	if err != nil {
		return nil, fmt.Errorf(
			"unable to create checkout directory for version %v of %v: %v",
			version, t.Name(), err)
	}

	err = getter.CheckoutTo(version, checkoutPath)
	if err != nil {
		os.RemoveAll(checkoutPath)
		return nil, fmt.Errorf(
			"unable to checkout version %v of %v: %v", version, t.Name(), err)
	}

	checkout := plainCheckout{version, checkoutPath}

	if err := runner.Setup(checkout); err != nil {
		defer os.RemoveAll(checkoutPath)
		return nil, fmt.Errorf(
			"failed to setup checkout for version %v of %v: %v",
			version, t.Name(), err)
	}

	t.(engineTool).state.(*toolState).addCheckout(version, checkoutPath, true)
	return checkout, nil
}

func isValidCheckout(checkout tool.Checkout) bool {
	if checkout == nil {
		return false
	}

	fi, err := os.Lstat(checkout.Path())
	if err != nil || !fi.IsDir() {
		jww.DEBUG.Printf("discarding invalid checkout: %v", checkout.Path())
		return false
	}
	return true
}

func (e engine) shouldFetchUpstream(t stoic.Tool) bool {
	if t.IsVersionPinned() {
		return false
	}
	if t.UpstreamVersion() == tool.NullVersion {
		return true
	}

	updateFrequency := e.updateFrequencyFallback.Combine(
		t.UpdateFrequency(), e.updateFrequencyOverride)
	return updateFrequency.IsTimeToUpdate(t.LastUpdate())
}

func (e engine) getVersionForCheckout(t stoic.Tool, getter tool.Getter) (tool.Version, error) {
	if !e.shouldFetchUpstream(t) {
		checkout := t.CurrentCheckout()
		if checkout != nil {
			return checkout.Version(), nil
		}
	}

	if t.IsVersionPinned() {
		pinVersion := t.CurrentVersion()
		err := getter.FetchVersion(pinVersion)
		if err != nil {
			return tool.NullVersion, fmt.Errorf(
				"unable to get pinned version %v of %v from upstream: %v",
				pinVersion, t.Name(), err)
		}

		return pinVersion, nil
	}

	version, err := getter.FetchLatest()
	if err == nil {
		if version == tool.NullVersion {
			err = ErrUpstreamVersionIsEmpty
		} else {
			t.(engineTool).state.(*toolState).setUpstreamVersion(t.Channel(), version)
		}
	}
	if err != nil {
		jww.WARN.Printf(
			"unable to get upstream version of %v: %v", t.Name(), err)

		// Fallback to current, if any
		version = t.CurrentVersion()
		if version == tool.NullVersion {
			return tool.NullVersion, fmt.Errorf(
				"unable to get upstream version of %v and no fallback is available",
				t.Name())
		}
	}

	return version, nil
}

func (e engine) RunTool(toolName string, args []string) error {
	t, err := e.getTool(toolName)
	if err != nil {
		return err
	}

	getter, err := e.getterFor(t)
	if err != nil {
		return err
	}
	runner, err := e.runnerFor(t)
	if err != nil {
		return err
	}

	version, err := e.getVersionForCheckout(t, getter)
	if err != nil {
		return err
	}

	checkout := t.CheckoutForVersion(version)
	if !isValidCheckout(checkout) {
		checkout, err = e.makeCheckout(t, version, getter, runner)
		// FIXME: When cached artifacts are evicted, new checkouts fail
		if err != nil {
			return err
		}
	}

	return runner.Run(checkout, toolName, args)
}
