package engine

import (
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/stoic-cli/stoic-cli-core/tool"
	"github.com/stoic-cli/stoic-cli-core/util"
)

type UnixTimestamp int64

// State represents metadata persisted by the engine for a given tool.
type State interface {
	// ToolId identifies the tool the instance pertains too.
	ToolId() string

	// Filename returns the filesystem path where the state is persisted.
	Filename() string

	// UpstreamVersion returns the latest upstream version of the tool that is
	// available locally, for the specified channel.
	UpstreamVersion(tc tool.Channel) tool.Version

	// LastUpstreamUpdate returns the time at which the upstream version for the
	// specified channel was last updated.
	LastUpstreamUpdate(tc tool.Channel) time.Time

	// CurrentCheckout returns the most recent checkout marked as current.
	CurrentCheckout() tool.Checkout

	// CheckoutForVersion returns the most recent checkout for the requested
	// version.
	CheckoutForVersion(version tool.Version) tool.Checkout
}

func (e engine) LoadState(toolId string) State {
	filename := filepath.Join(e.stateDir, url.PathEscape(toolId))
	state := &toolState{
		toolId:   toolId,
		filename: filename,
	}

	stateFile, err := os.Open(filename)
	if os.IsNotExist(err) {
		// No previous state
		return state
	}
	if err != nil {
		jww.WARN.Printf("Unable to load state from %v: %v", filename, err)
		return state
	}
	defer stateFile.Close()

	err = state.ToolStateFormat.load(stateFile)
	if err != nil {
		jww.WARN.Printf("Unable to load state from %v, is file corrupt? %v", filename, err)
	}
	return state
}

type toolState struct {
	toolId   string
	filename string

	ToolStateFormat
}

func (ts *toolState) updateInTransaction(updateFunc func(UnixTimestamp)) (err error) {
	timestamp := UnixTimestamp(time.Now().Unix())

	err = os.MkdirAll(filepath.Dir(ts.filename), 0755)
	if err != nil {
		goto updateFailedEarly
	}

	{
		var stateFile util.AtomicFileWriter

		// Begin transaction
		stateFile, err = util.OpenToChange(ts.filename)
		if err != nil {
			goto updateFailedEarly
		}
		defer stateFile.AbortIfPending()

		// Load current state
		if curr := stateFile.Current(); curr != nil {
			err = ts.ToolStateFormat.load(curr)
			if err != nil {
				goto updateFailedEarly
			}
		}

		// Update local state
		updateFunc(timestamp)

		// Persist state
		err = ts.ToolStateFormat.save(stateFile)
		if err != nil {
			return
		}

		// Commit transaction
		stateFile.Commit()
		return
	}

updateFailedEarly:
	// Failed to update persisted state. Update local state, anyway.
	updateFunc(timestamp)
	return
}

func (ts *toolState) setUpstreamVersion(tc tool.Channel, tv tool.Version) {
	err := ts.updateInTransaction(func(timestamp UnixTimestamp) {
		upstream := ToolChannelInfoFormat{
			Version:    tv,
			LastUpdate: timestamp,
		}

		if tc == tool.DefaultChannel {
			ts.Upstream = &upstream
		} else {
			if ts.Channels == nil {
				ts.Channels = map[tool.Channel]ToolChannelInfoFormat{}
			}
			ts.Channels[tc] = upstream
		}
	})
	if err != nil {
		jww.ERROR.Printf("Unable to persist upstream version to %v: %v", ts.filename, err)
	}
}
func (ts *toolState) addCheckout(tv tool.Version, path string, setCurrent bool) {
	err := ts.updateInTransaction(func(timestamp UnixTimestamp) {
		checkout := ToolCheckoutFormat{
			CheckoutVersion: tv,
			CheckoutPath:    path,
			Created:         timestamp,
		}
		if setCurrent {
			checkout.SetCurrent = timestamp
		}

		ts.Checkouts = append(ts.Checkouts, checkout)
	})
	if err != nil {
		jww.ERROR.Printf("Unable to persist checkout information to %v: %v", ts.filename, err)
	}
}
func (ts *toolState) setCurrentCheckout(path string) {
	err := ts.updateInTransaction(func(timestamp UnixTimestamp) {
		for i := range ts.Checkouts {
			if ts.Checkouts[i].CheckoutPath == path {
				ts.Checkouts[i].SetCurrent = timestamp
				return
			}
		}

		// Er, should we be concerned?
		jww.INFO.Printf("Checkout not marked as current because it was not found in state file")
	})
	if err != nil {
		jww.ERROR.Printf("Unable to persist checkout as current in %v: %v", ts.filename, err)
	}
}

func (ts *toolState) ToolId() string   { return ts.toolId }
func (ts *toolState) Filename() string { return ts.filename }

// ToolStateFormat defines the low-level format for persisting State.
type ToolStateFormat struct {
	Upstream  *ToolChannelInfoFormat                 `json:"upstream,omitempty"`
	Channels  map[tool.Channel]ToolChannelInfoFormat `json:"channels,omitempty"`
	Checkouts []ToolCheckoutFormat                   `json:"checkouts,omitempty"`
}

func (tsf *ToolStateFormat) load(r io.Reader) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(tsf)
}

func (tsf *ToolStateFormat) save(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(tsf)
}

func (tsf *ToolStateFormat) UpstreamVersion(tc tool.Channel) tool.Version {
	if tc == tool.DefaultChannel {
		if tsf.Upstream != nil {
			return tsf.Upstream.Version
		}
	} else if info, ok := tsf.Channels[tc]; ok {
		return info.Version
	}

	return tool.NullVersion
}

func (tsf *ToolStateFormat) LastUpstreamUpdate(tc tool.Channel) time.Time {
	if tc == tool.DefaultChannel {
		if tsf.Upstream != nil {
			return time.Unix(int64(tsf.Upstream.LastUpdate), 0)
		}
	} else if info, ok := tsf.Channels[tc]; ok {
		return time.Unix(int64(info.LastUpdate), 0)
	}

	return time.Time{}
}

func (tsf *ToolStateFormat) CurrentCheckout() tool.Checkout {
	if tsf.Checkouts == nil {
		return nil
	}

	youngest := -1
	for i := range tsf.Checkouts {
		if tsf.Checkouts[i].SetCurrent != 0 {
			youngest = i
			break
		}
	}
	if youngest == -1 {
		return nil
	}

	for i := range tsf.Checkouts[youngest:] {
		if tsf.Checkouts[youngest].SetCurrent < tsf.Checkouts[i].SetCurrent {
			youngest = i
		}
	}
	return tsf.Checkouts[youngest]
}

func (tsf *ToolStateFormat) CheckoutForVersion(version tool.Version) tool.Checkout {
	if tsf.Checkouts == nil {
		return nil
	}

	youngest := -1
	for i := range tsf.Checkouts {
		if tsf.Checkouts[i].CheckoutVersion == version {
			youngest = i
			break
		}
	}
	if youngest == -1 {
		return nil
	}

	for i := range tsf.Checkouts[youngest:] {
		if tsf.Checkouts[i].CheckoutVersion == version {
			if tsf.Checkouts[youngest].Created < tsf.Checkouts[i].Created {
				youngest = i
			}
		}
	}
	return tsf.Checkouts[youngest]
}

// ToolChannelInfoFormat defines the low-level format for persisting information about
// an upstream source.
type ToolChannelInfoFormat struct {
	Version    tool.Version  `json:"version,omitempty"`
	LastUpdate UnixTimestamp `json:"last-update,omitempty"`
}

// ToolCheckoutFormat defines the low-level format for persisting information
// about a tool.Checkout.
type ToolCheckoutFormat struct {
	CheckoutVersion tool.Version  `json:"version"`
	CheckoutPath    string        `json:"path"`
	Created         UnixTimestamp `json:"created"`
	SetCurrent      UnixTimestamp `json:"set-current,omitempty"`
}

func (tcf ToolCheckoutFormat) Path() string          { return tcf.CheckoutPath }
func (tcf ToolCheckoutFormat) Version() tool.Version { return tcf.CheckoutVersion }
