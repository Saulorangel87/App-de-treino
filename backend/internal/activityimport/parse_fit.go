package activityimport

import (
	"io"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
)

// ParseFIT decodes a .fit file (the format produced by Garmin, Wahoo, Polar,
// Zwift and the XOSS app, among others) and extracts a cycling activity
// summary.
//
// This uses github.com/muktihari/fit rather than the more commonly cited
// github.com/tormoder/fit: a real ride exported from a XOSS head unit (the
// device this project's own users ride with) defines several message types
// back-to-back before any data record, which is valid FIT but not something
// tormoder/fit's file_id fast path accepts — it only handles a lone
// definition immediately followed by file_id's data. muktihari/fit decodes
// the same file without issue and is the more actively maintained of the two.
func ParseFIT(r io.Reader) (Parsed, error) {
	dec := decoder.New(r)
	fit, err := dec.Decode()
	if err != nil {
		return Parsed{}, ErrCorruptFile
	}
	var session *mesgdef.Session
	var records []*mesgdef.Record
	for i := range fit.Messages {
		mesg := &fit.Messages[i]
		switch mesg.Num {
		case typedef.MesgNumSession:
			if session == nil {
				session = mesgdef.NewSession(mesg)
			}
		case typedef.MesgNumRecord:
			records = append(records, mesgdef.NewRecord(mesg))
		}
	}
	if session != nil {
		return parseFITFromSession(session)
	}
	return parseFITFromRecords(records)
}

func parseFITFromSession(session *mesgdef.Session) (Parsed, error) {
	if session.Sport != typedef.SportCycling {
		return Parsed{}, ErrUnsupportedSport
	}
	if session.TotalTimerTime == basetype.Uint32Invalid {
		return Parsed{}, ErrNoData
	}
	parsed := Parsed{
		StartedAt:     session.StartTime,
		MovingSeconds: int(session.TotalTimerTime / 1000), // scale 1000, ms -> s
		DistanceKM:    float64(session.TotalDistance) / 100 / 1000,
	}
	if session.TotalAscent != basetype.Uint16Invalid {
		v := int(session.TotalAscent) // scale 1 (meters)
		parsed.ElevationGainM = &v
	}
	// Alguns aparelhos (confirmado num arquivo real do XOSS) gravam 0 em vez
	// do valor "inválido" do FIT quando não há sensor conectado. Uma média de
	// 0 bpm/W/rpm na sessão inteira não é fisicamente plausível, então trata
	// como ausente aqui — diferente dos registros por segundo, onde 0 W é uma
	// leitura legítima (embalo).
	if session.AvgHeartRate != basetype.Uint8Invalid && session.AvgHeartRate != 0 {
		v := int(session.AvgHeartRate)
		parsed.AverageHeartRate = &v
	}
	if session.MaxHeartRate != basetype.Uint8Invalid && session.MaxHeartRate != 0 {
		v := int(session.MaxHeartRate)
		parsed.MaxHeartRate = &v
	}
	if session.AvgPower != basetype.Uint16Invalid && session.AvgPower != 0 {
		v := int(session.AvgPower)
		parsed.AveragePowerW = &v
	}
	if session.NormalizedPower != basetype.Uint16Invalid && session.NormalizedPower != 0 {
		v := int(session.NormalizedPower)
		parsed.NormalizedPowerW = &v
	}
	if session.AvgCadence != basetype.Uint8Invalid && session.AvgCadence != 0 {
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
func parseFITFromRecords(records []*mesgdef.Record) (Parsed, error) {
	if len(records) < 2 {
		return Parsed{}, ErrNoData
	}
	first, last := records[0], records[len(records)-1]
	moving := int(last.Timestamp.Sub(first.Timestamp).Seconds())
	if moving <= 0 {
		return Parsed{}, ErrNoData
	}
	parsed := Parsed{StartedAt: first.Timestamp, MovingSeconds: moving}
	if last.Distance != basetype.Uint32Invalid {
		parsed.DistanceKM = float64(last.Distance) / 100 / 1000
	}
	var hrSamples, powerSamples, cadenceSamples []int
	var elevationGain float64
	prevAltitude, hasPrevAltitude := 0.0, false
	for _, rec := range records {
		if rec.HeartRate != basetype.Uint8Invalid {
			hrSamples = append(hrSamples, int(rec.HeartRate))
		}
		if rec.Power != basetype.Uint16Invalid {
			powerSamples = append(powerSamples, int(rec.Power))
		}
		if rec.Cadence != basetype.Uint8Invalid {
			cadenceSamples = append(cadenceSamples, int(rec.Cadence))
		}
		if rec.Altitude != basetype.Uint16Invalid {
			altitude := (float64(rec.Altitude) / 5) - 500 // scale 5, offset 500
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
