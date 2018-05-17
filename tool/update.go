package tool

import (
	"time"
)

type UpdateFrequency int

const (
	UpdateDefault UpdateFrequency = iota
	UpdateDaily
	UpdateWeekly
	UpdateMonthly

	UpdateAlways UpdateFrequency = -1
	UpdateNever  UpdateFrequency = -2
)

func (uf UpdateFrequency) Combine(otherUF ...UpdateFrequency) UpdateFrequency {
	for i := len(otherUF) - 1; i >= 0; i-- {
		if otherUF[i] != UpdateDefault {
			return otherUF[i]
		}
	}
	return uf
}

func (uf UpdateFrequency) String() string {
	switch uf {
	case UpdateDaily:
		return "daily"
	case UpdateWeekly:
		return "weekly"
	case UpdateMonthly:
		return "monthly"

	case UpdateAlways:
		return "always"
	case UpdateNever:
		return "never"

	default:
		return ""
	}
}

func (uf UpdateFrequency) IsTimeToUpdate(lastUpdate time.Time) bool {
	switch uf {
	case UpdateDaily:
		// Update after a good night's sleep
		return time.Since(lastUpdate) > 12*time.Hour

	case UpdateWeekly:
		return time.Since(lastUpdate) > 6.5*24*time.Hour

	case UpdateMonthly:
		return time.Since(lastUpdate) > 29.5*24*time.Hour

	case UpdateAlways:
		return true

	case UpdateNever:
		return false

	default:
		panic("unable to decide whether to update")
	}
}

func UpdateFrequencyFromString(value string) UpdateFrequency {
	switch value {
	case "daily":
		return UpdateDaily
	case "weekly":
		return UpdateWeekly
	case "monthly":
		return UpdateMonthly

	case "always":
		return UpdateAlways
	case "never":
		return UpdateNever

	default:
		return UpdateDefault
	}
}

func (uf *UpdateFrequency) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var frequency string
	if err := unmarshal(&frequency); err != nil {
		return err
	}

	*uf = UpdateFrequencyFromString(frequency)
	return nil
}
