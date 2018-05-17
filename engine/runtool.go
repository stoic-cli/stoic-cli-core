package engine

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stoic-cli/stoic-cli-core/tool"
)

type uncommittedCheckout struct {
	CheckoutVersion tool.Version
	CheckoutPath    string
}

func (uc uncommittedCheckout) Version() tool.Version { return uc.CheckoutVersion }
func (uc uncommittedCheckout) Path() string          { return uc.CheckoutPath }

func (uc uncommittedCheckout) Dispose() {
	os.RemoveAll(uc.CheckoutPath)
}

func (e engine) doCheckout(
	endpoint *url.URL, version tool.Version, state State, getter tool.Getter) (tool.Checkout, error) {
	parts := []string{e.checkoutsDir, endpoint.Hostname()}
	parts = append(parts, strings.Split(endpoint.EscapedPath(), "/")...)
	baseDir := filepath.Join(parts...)

	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, err
	}

	checkoutPath, err := ioutil.TempDir(baseDir, string(version)+"-")
	if err != nil {
		return nil, err
	}

	err = getter.CheckoutTo(version, checkoutPath)
	if err != nil {
		os.RemoveAll(checkoutPath)
		return nil, err
	}

	return uncommittedCheckout{
		CheckoutVersion: version,
		CheckoutPath:    checkoutPath,
	}, nil
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

func (e engine) pinnedVersionCheckout(
	endpoint *url.URL, pinVersion tool.Version, state State, getter tool.Getter) (tool.Checkout, error) {
	checkout := state.CheckoutForVersion(pinVersion)
	if isValidCheckout(checkout) {
		return checkout, nil
	}

	err := getter.FetchVersion(pinVersion)
	if err != nil {
		jww.DEBUG.Printf(
			"unable to fetch pinned version %v from %v: %v",
			pinVersion, state.ToolId(), err)
		return nil, err
	}

	return e.doCheckout(endpoint, pinVersion, state, getter)
}

func (e engine) combinedUpdateFrequency(toolSetting tool.UpdateFrequency) tool.UpdateFrequency {
	if e.updateFrequencyOverride != tool.UpdateDefault {
		return e.updateFrequencyOverride
	}
	if toolSetting != tool.UpdateDefault {
		return toolSetting
	}
	return e.updateFrequencyFallback
}

func (e engine) fetchUpstreamAndCheckout(
	endpoint *url.URL, state State, channel tool.Channel, getter tool.Getter) (tool.Checkout, error) {
	version, err := getter.FetchLatest()
	if err != nil {
		jww.DEBUG.Printf("unable to check %v for updates: %v", state.ToolId(), err)
		return nil, err
	}
	if version == tool.NullVersion {
		return nil, fmt.Errorf("got empty version from getter for channel '%v'", channel)
	}

	state.(*toolState).setUpstreamVersion(channel, version)

	checkout := state.CheckoutForVersion(version)
	if isValidCheckout(checkout) {
		return checkout, nil
	}

	return e.doCheckout(endpoint, version, state, getter)
}

func (e engine) RunTool(toolName string, args []string) error {
	toolConfig, ok := e.tools[toolName]
	if !ok {
		return fmt.Errorf("unknown tool: %v", toolName)
	}

	if toolConfig.Getter.Type == "" {
		toolConfig.Getter.Type = DefaultToolGetterType
	}
	if toolConfig.Runner.Type == "" {
		toolConfig.Runner.Type = DefaultToolRunnerType
	}

	getter, err := e.NewGetter(toolConfig)
	if err != nil {
		return err
	}
	runner, err := e.NewRunner(toolConfig)
	if err != nil {
		return err
	}

	url, err := toolConfig.Endpoint.MarshalBinary()
	if err != nil {
		return err
	}
	state := e.LoadState(string(url))

	if toolConfig.PinVersion != tool.NullVersion {
		checkout, err := e.pinnedVersionCheckout(toolConfig.Endpoint, toolConfig.PinVersion, state, getter)
		if uc, ok := checkout.(uncommittedCheckout); ok {
			err = runner.Setup(uc)
			if err == nil {
				state.(*toolState).addCheckout(uc.CheckoutVersion, uc.CheckoutPath, false)
				return runner.Run(uc, toolName, args)
			}

			uc.Dispose()
		}

		if err != nil {
			return fmt.Errorf(
				"unable to create checkout for (pinned) version %v of %v: %v",
				toolConfig.PinVersion, toolName, err)
		}

		return runner.Run(checkout, toolName, args)
	}

	var checkout tool.Checkout

	channel := toolConfig.Channel
	if state.UpstreamVersion(channel) == tool.NullVersion ||
		e.combinedUpdateFrequency(toolConfig.UpdateFrequency).
			IsTimeToUpdate(state.LastUpstreamUpdate(channel)) {
		checkout, err = e.fetchUpstreamAndCheckout(toolConfig.Endpoint, state, channel, getter)
		if err != nil {
			jww.WARN.Printf(
				"unable to update %v to latest upstream version: %v",
				toolName, err)
		}

		if uc, ok := checkout.(uncommittedCheckout); ok {
			err = runner.Setup(uc)
			if err == nil {
				state.(*toolState).addCheckout(uc.CheckoutVersion, uc.CheckoutPath, true)
				return runner.Run(uc, toolName, args)
			}

			jww.WARN.Printf(
				"failed to setup latest upstream version %v of %v: %v",
				uc.CheckoutVersion, toolName, err)

			uc.Dispose()
			checkout = nil
		}
	}

	if checkout == nil {
		checkout = state.CurrentCheckout()
		if isValidCheckout(checkout) {
			return runner.Run(checkout, toolName, args)
		}
	}

	// Last chance, fallback!

	var version tool.Version
	if checkout == nil {
		// No current version, use upstream
		version = state.UpstreamVersion(channel)
	} else {
		// There used to be a current checkout, it is now invalid
		version = checkout.Version()
	}

	checkout, err = e.doCheckout(toolConfig.Endpoint, version, state, getter)
	if err != nil {
		return fmt.Errorf(
			"failed to checkout version %v of %v: %v", version, toolName, err)
	}

	err = runner.Setup(checkout)
	if err != nil {
		checkout.(uncommittedCheckout).Dispose()
		return fmt.Errorf(
			"failed to setup version %v of tool %v: %v", version, toolName, err)
	}

	state.(*toolState).addCheckout(checkout.Version(), checkout.Path(), true)
	return runner.Run(checkout, toolName, args)
}
