package activityimport

import (
	"bytes"
	"testing"
	"time"

	"github.com/muktihari/fit/encoder"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/muktihari/fit/proto"
)

// buildFIT encodes a minimal, valid Activity FIT file using the same
// library this package decodes with, so these tests exercise the adapter
// code in parse_fit.go rather than re-testing the binary format itself
// (already covered by muktihari/fit's own test suite, and by the real XOSS
// file this parser was validated against during development).
func buildFIT(t *testing.T, messages ...proto.Message) []byte {
	t.Helper()
	fileId := mesgdef.NewFileId(nil)
	fileId.Type = typedef.FileActivity
	fileId.TimeCreated = time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)

	data := proto.FIT{
		FileHeader: proto.FileHeader{ProtocolVersion: proto.V1, DataType: ".FIT"},
		Messages:   append([]proto.Message{fileId.ToMesg(nil)}, messages...),
	}
	var buf bytes.Buffer
	if err := encoder.New(&buf).Encode(&data); err != nil {
		t.Fatalf("encoder.Encode: %v", err)
	}
	return buf.Bytes()
}

func TestParseFIT_FromSession(t *testing.T) {
	start := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	hr, maxHR, power, np, cadence := uint8(140), uint8(178), uint16(210), uint16(225), uint8(88)
	session := mesgdef.NewSession(nil)
	session.Sport = typedef.SportCycling
	session.StartTime = start
	session.Timestamp = start.Add(90 * time.Minute)
	session.TotalTimerTime = 90 * 60 * 1000 // 90 min, scale 1000
	session.TotalElapsedTime = 92 * 60 * 1000
	session.TotalDistance = 4500000 // 45 km, scale 100 -> m
	session.TotalAscent = 620
	session.AvgHeartRate = hr
	session.MaxHeartRate = maxHR
	session.AvgPower = power
	session.NormalizedPower = np
	session.AvgCadence = cadence
	data := buildFIT(t, session.ToMesg(nil))

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

// TestParseFIT_TreatsZeroVitalsAsMissing documents a real quirk found in a
// ride exported from a XOSS head unit: without a paired sensor, it writes 0
// into the session's average heart rate/power/cadence fields instead of the
// FIT protocol's invalid sentinel. A whole-ride average of exactly zero is
// not physiologically plausible, so it is treated the same as "no sensor".
func TestParseFIT_TreatsZeroVitalsAsMissing(t *testing.T) {
	session := mesgdef.NewSession(nil)
	session.Sport = typedef.SportCycling
	session.StartTime = time.Now().UTC()
	session.TotalTimerTime = 3600 * 1000
	session.TotalElapsedTime = 3600 * 1000
	session.TotalDistance = 3000000
	session.AvgHeartRate, session.MaxHeartRate = 0, 0
	session.AvgPower, session.NormalizedPower = 0, 0
	session.AvgCadence = 0
	data := buildFIT(t, session.ToMesg(nil))

	parsed, err := ParseFIT(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseFIT: %v", err)
	}
	if parsed.AverageHeartRate != nil || parsed.MaxHeartRate != nil {
		t.Errorf("heart rate should be nil when the device reports 0, got avg=%v max=%v", parsed.AverageHeartRate, parsed.MaxHeartRate)
	}
	if parsed.AveragePowerW != nil || parsed.NormalizedPowerW != nil {
		t.Errorf("power should be nil when the device reports 0, got avg=%v np=%v", parsed.AveragePowerW, parsed.NormalizedPowerW)
	}
	if parsed.AverageCadenceRPM != nil {
		t.Errorf("cadence should be nil when the device reports 0, got %v", parsed.AverageCadenceRPM)
	}
	// A distância real da sessão não deve ser descartada pela mesma regra.
	if parsed.DistanceKM != 30 {
		t.Errorf("DistanceKM = %v, want 30", parsed.DistanceKM)
	}
}

func TestParseFIT_RejectsNonCyclingSport(t *testing.T) {
	session := mesgdef.NewSession(nil)
	session.Sport = typedef.SportRunning
	session.StartTime = time.Now().UTC()
	session.TotalTimerTime = 30 * 60 * 1000
	session.TotalElapsedTime = 30 * 60 * 1000
	data := buildFIT(t, session.ToMesg(nil))

	_, err := ParseFIT(bytes.NewReader(data))
	if err != ErrUnsupportedSport {
		t.Fatalf("err = %v, want ErrUnsupportedSport", err)
	}
}

func TestParseFIT_FallbackFromRecords(t *testing.T) {
	start := time.Date(2026, 9, 20, 8, 0, 0, 0, time.UTC)
	var messages []proto.Message
	for i := 0; i < 3; i++ {
		rec := mesgdef.NewRecord(nil)
		rec.Timestamp = start.Add(time.Duration(i) * time.Minute)
		rec.Distance = uint32(i * 100000)           // scale 100 -> i*1000 m
		rec.Altitude = uint16((100+i*100)*5 + 2500) // scale 5, offset 500 -> 100+i*100 meters
		rec.HeartRate = uint8(130 + i)
		rec.Power = uint16(200 + i)
		rec.Cadence = uint8(80 + i)
		messages = append(messages, rec.ToMesg(nil))
	}
	data := buildFIT(t, messages...)

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
