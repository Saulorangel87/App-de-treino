package activityimport

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/tormoder/fit"
)

// buildFIT encodes a minimal, valid Activity FIT file using the same
// library this package decodes with, so these tests exercise the adapter
// code in parse_fit.go rather than re-testing the binary format itself
// (already covered by tormoder/fit's own test suite).
func buildFIT(t *testing.T, configure func(f *fit.File, activity *fit.ActivityFile)) []byte {
	t.Helper()
	f, err := fit.NewFile(fit.FileTypeActivity, fit.NewHeader(fit.V10, false))
	if err != nil {
		t.Fatalf("fit.NewFile: %v", err)
	}
	f.FileId.Type = fit.FileTypeActivity
	f.FileId.Manufacturer = fit.ManufacturerGarmin
	f.FileId.TimeCreated = time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	activity, err := f.Activity()
	if err != nil {
		t.Fatalf("f.Activity: %v", err)
	}
	configure(f, activity)
	var buf bytes.Buffer
	if err := fit.Encode(&buf, f, binary.LittleEndian); err != nil {
		t.Fatalf("fit.Encode: %v", err)
	}
	return buf.Bytes()
}

func TestParseFIT_FromSession(t *testing.T) {
	start := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	hr, maxHR, power, np, cadence := uint8(140), uint8(178), uint16(210), uint16(225), uint8(88)
	data := buildFIT(t, func(_ *fit.File, activity *fit.ActivityFile) {
		session := fit.NewSessionMsg()
		session.Sport = fit.SportCycling
		session.StartTime = start
		session.Timestamp = start.Add(90 * time.Minute)
		session.TotalTimerTime = 90 * 60 * 1000 // 90 min, in ms
		session.TotalElapsedTime = 92 * 60 * 1000
		session.TotalDistance = 4500000 // 45 km in cm
		session.TotalAscent = 620
		session.AvgHeartRate = hr
		session.MaxHeartRate = maxHR
		session.AvgPower = power
		session.NormalizedPower = np
		session.AvgCadence = cadence
		activity.Sessions = append(activity.Sessions, session)
	})

	parsed, err := ParseFIT(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseFIT: %v", err)
	}
	if !parsed.StartedAt.Equal(start) {
		t.Errorf("StartedAt = %v, want %v", parsed.StartedAt, start)
	}
	if parsed.MovingSeconds != 90*60 {
		t.Errorf("MovingSeconds = %d, want %d", parsed.MovingSeconds, 90*60)
	}
	if parsed.DistanceKM != 45 {
		t.Errorf("DistanceKM = %v, want 45", parsed.DistanceKM)
	}
	if parsed.ElevationGainM == nil || *parsed.ElevationGainM != 620 {
		t.Errorf("ElevationGainM = %v, want 620", parsed.ElevationGainM)
	}
	if parsed.AverageHeartRate == nil || *parsed.AverageHeartRate != int(hr) {
		t.Errorf("AverageHeartRate = %v, want %d", parsed.AverageHeartRate, hr)
	}
	if parsed.MaxHeartRate == nil || *parsed.MaxHeartRate != int(maxHR) {
		t.Errorf("MaxHeartRate = %v, want %d", parsed.MaxHeartRate, maxHR)
	}
	if parsed.AveragePowerW == nil || *parsed.AveragePowerW != int(power) {
		t.Errorf("AveragePowerW = %v, want %d", parsed.AveragePowerW, power)
	}
	if parsed.NormalizedPowerW == nil || *parsed.NormalizedPowerW != int(np) {
		t.Errorf("NormalizedPowerW = %v, want %d", parsed.NormalizedPowerW, np)
	}
	if parsed.AverageCadenceRPM == nil || *parsed.AverageCadenceRPM != int(cadence) {
		t.Errorf("AverageCadenceRPM = %v, want %d", parsed.AverageCadenceRPM, cadence)
	}
}

func TestParseFIT_RejectsNonCyclingSport(t *testing.T) {
	data := buildFIT(t, func(_ *fit.File, activity *fit.ActivityFile) {
		session := fit.NewSessionMsg()
		session.Sport = fit.SportRunning
		session.StartTime = time.Now().UTC()
		session.TotalTimerTime = 30 * 60 * 1000
		session.TotalElapsedTime = 30 * 60 * 1000
		activity.Sessions = append(activity.Sessions, session)
	})

	_, err := ParseFIT(bytes.NewReader(data))
	if err != ErrUnsupportedSport {
		t.Fatalf("err = %v, want ErrUnsupportedSport", err)
	}
}

func TestParseFIT_FallbackFromRecords(t *testing.T) {
	start := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	data := buildFIT(t, func(_ *fit.File, activity *fit.ActivityFile) {
		for i := 0; i < 3; i++ {
			rec := fit.NewRecordMsg()
			rec.Timestamp = start.Add(time.Duration(i) * time.Minute)
			rec.Distance = uint32(i * 100000)        // meters * 100 (cm), i.e. i * 1000 m
			rec.Altitude = uint16((100 + i*100) * 5) // scale 1/5
			rec.HeartRate = uint8(130 + i)
			rec.Power = uint16(200 + i)
			rec.Cadence = uint8(80 + i)
			activity.Records = append(activity.Records, rec)
		}
	})

	parsed, err := ParseFIT(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseFIT: %v", err)
	}
	if parsed.MovingSeconds != 120 {
		t.Errorf("MovingSeconds = %d, want 120", parsed.MovingSeconds)
	}
	if parsed.DistanceKM != 2 {
		t.Errorf("DistanceKM = %v, want 2 (last record's distance)", parsed.DistanceKM)
	}
	if parsed.ElevationGainM == nil || *parsed.ElevationGainM != 200 {
		t.Errorf("ElevationGainM = %v, want 200", parsed.ElevationGainM)
	}
	if parsed.AverageHeartRate == nil || *parsed.AverageHeartRate != 131 {
		t.Errorf("AverageHeartRate = %v, want 131", parsed.AverageHeartRate)
	}
	if parsed.MaxHeartRate == nil || *parsed.MaxHeartRate != 132 {
		t.Errorf("MaxHeartRate = %v, want 132", parsed.MaxHeartRate)
	}
}

func TestParseFIT_CorruptFile(t *testing.T) {
	_, err := ParseFIT(bytes.NewReader([]byte("not a fit file")))
	if err != ErrCorruptFile {
		t.Fatalf("err = %v, want ErrCorruptFile", err)
	}
}
