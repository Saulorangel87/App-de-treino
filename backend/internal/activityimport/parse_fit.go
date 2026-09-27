package activityimport

import (
	"io"

	"github.com/tormoder/fit"
)

// ParseFIT decodes a .fit file (the format produced by Garmin, Wahoo, Polar,
// Zwift and the XOSS app, among others) and extracts a cycling activity
// summary. The FIT protocol's invalid-value sentinels (all bits set) are
// treated as "not present", matching how the tormoder/fit library reports
// them: it decodes every field to physical units already, so this function
// only selects and validates, it never re-implements the binary format.
func ParseFIT(r io.Reader) (Parsed, error) {
	decoded, err := fit.Decode(r)
	if err != nil {
		return Parsed{}, ErrCorruptFile
	}
	activity, err := decoded.Activity()
	if err != nil {
		return Parsed{}, ErrNoData
	}
	if len(activity.Sessions) > 0 {
		return parseFITFromSession(activity.Sessions[0])
	}
	return parseFITFromRecords(activity.Records)
}

func parseFITFromSession(session *fit.SessionMsg) (Parsed, error) {
	if session.Sport != fit.SportCycling {
		return Parsed{}, ErrUnsupportedSport
	}
	if session.TotalElapsedTime == 0xFFFFFFFF {
		return Parsed{}, ErrNoData
	}
	parsed := Parsed{
		StartedAt:     session.StartTime,
		MovingSeconds: int(session.GetTotalTimerTimeScaled()),
		DistanceKM:    session.GetTotalDistanceScaled() / 1000,
	}
	if session.TotalAscent != 0xFFFF {
		v := int(session.TotalAscent)
		parsed.ElevationGainM = &v
	}
	if session.AvgHeartRate != 0xFF {
		v := int(session.AvgHeartRate)
		parsed.AverageHeartRate = &v
	}
	if session.MaxHeartRate != 0xFF {
		v := int(session.MaxHeartRate)
		parsed.MaxHeartRate = &v
	}
	if session.AvgPower != 0xFFFF {
		v := int(session.AvgPower)
		parsed.AveragePowerW = &v
	}
	if session.NormalizedPower != 0xFFFF {
		v := int(session.NormalizedPower)
		parsed.NormalizedPowerW = &v
	}
	if session.AvgCadence != 0xFF {
		v := int(session.AvgCadence)
		parsed.AverageCadenceRPM = &v
	}
	return parsed, nil
}

// parseFITFromRecords is the fallback used when a file has no session
// summary message (some head units omit it in trimmed exports): it rebuilds
// the same fields from the raw per-second records. Records carry no sport
// field, so this path cannot reject a non-cycling activity the way
// parseFITFromSession does; it is accepted because the app itself is
// cycling-only, and the athlete chose to upload the file.
func parseFITFromRecords(records []*fit.RecordMsg) (Parsed, error) {
	if len(records) < 2 {
		return Parsed{}, ErrNoData
	}
	first, last := records[0], records[len(records)-1]
	moving := int(last.Timestamp.Sub(first.Timestamp).Seconds())
	if moving <= 0 {
		return Parsed{}, ErrNoData
	}
	parsed := Parsed{StartedAt: first.Timestamp, MovingSeconds: moving}
	if last.Distance != 0xFFFFFFFF {
		parsed.DistanceKM = last.GetDistanceScaled() / 1000
	}
	var hrSamples, powerSamples, cadenceSamples []int
	var elevationGain float64
	prevAltitude, hasPrevAltitude := 0.0, false
	for _, rec := range records {
		if rec.HeartRate != 0xFF {
			hrSamples = append(hrSamples, int(rec.HeartRate))
		}
		if rec.Power != 0xFFFF {
			powerSamples = append(powerSamples, int(rec.Power))
		}
		if rec.Cadence != 0xFF {
			cadenceSamples = append(cadenceSamples, int(rec.Cadence))
		}
		if rec.Altitude != 0xFFFF {
			altitude := rec.GetAltitudeScaled()
			if hasPrevAltitude && altitude > prevAltitude {
				elevationGain += altitude - prevAltitude
			}
			prevAltitude, hasPrevAltitude = altitude, true
		}
	}
	if hasPrevAltitude {
		v := int(elevationGain)
		parsed.ElevationGainM = &v
	}
	parsed.AverageHeartRate = avgInt(hrSamples)
	if len(hrSamples) > 0 {
		max := hrSamples[0]
		for _, v := range hrSamples {
			if v > max {
				max = v
			}
		}
		parsed.MaxHeartRate = &max
	}
	parsed.AveragePowerW = avgInt(powerSamples)
	parsed.AverageCadenceRPM = avgInt(cadenceSamples)
	return parsed, nil
}
