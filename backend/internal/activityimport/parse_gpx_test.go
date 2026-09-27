package activityimport

import (
	"strings"
	"testing"
)

const sampleGPX = `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1">
  <trk>
    <type>cycling</type>
    <trkseg>
      <trkpt lat="-23.55050" lon="-46.63330">
        <ele>760.0</ele>
        <time>2026-09-20T08:00:00Z</time>
        <extensions><gpxtpx:TrackPointExtension><gpxtpx:hr>130</gpxtpx:hr><gpxtpx:cad>82</gpxtpx:cad></gpxtpx:TrackPointExtension></extensions>
      </trkpt>
      <trkpt lat="-23.55150" lon="-46.63430">
        <ele>780.0</ele>
        <time>2026-09-20T08:05:00Z</time>
        <extensions><gpxtpx:TrackPointExtension><gpxtpx:hr>150</gpxtpx:hr><gpxtpx:cad>90</gpxtpx:cad></gpxtpx:TrackPointExtension></extensions>
      </trkpt>
      <trkpt lat="-23.55250" lon="-46.63530">
        <ele>770.0</ele>
        <time>2026-09-20T08:10:00Z</time>
        <extensions><gpxtpx:TrackPointExtension><gpxtpx:hr>140</gpxtpx:hr><gpxtpx:cad>85</gpxtpx:cad></gpxtpx:TrackPointExtension></extensions>
      </trkpt>
    </trkseg>
  </trk>
</gpx>`

func TestParseGPX_Basic(t *testing.T) {
	parsed, err := ParseGPX(strings.NewReader(sampleGPX))
	if err != nil {
		t.Fatalf("ParseGPX: %v", err)
	}
	if parsed.MovingSeconds != 600 {
		t.Errorf("MovingSeconds = %d, want 600", parsed.MovingSeconds)
	}
	if parsed.DistanceKM <= 0 {
		t.Errorf("DistanceKM = %v, want > 0", parsed.DistanceKM)
	}
	if parsed.ElevationGainM == nil || *parsed.ElevationGainM != 20 {
		t.Errorf("ElevationGainM = %v, want 20 (760->780, then 780->770 does not add)", parsed.ElevationGainM)
	}
	if parsed.AverageHeartRate == nil || *parsed.AverageHeartRate != 140 {
		t.Errorf("AverageHeartRate = %v, want 140", parsed.AverageHeartRate)
	}
	if parsed.MaxHeartRate == nil || *parsed.MaxHeartRate != 150 {
		t.Errorf("MaxHeartRate = %v, want 150", parsed.MaxHeartRate)
	}
	if parsed.AverageCadenceRPM == nil || *parsed.AverageCadenceRPM != 85 {
		t.Errorf("AverageCadenceRPM = %v, want 85", parsed.AverageCadenceRPM)
	}
}

func TestParseGPX_RejectsNonCyclingType(t *testing.T) {
	gpxRun := strings.Replace(sampleGPX, "<type>cycling</type>", "<type>running</type>", 1)
	_, err := ParseGPX(strings.NewReader(gpxRun))
	if err != ErrUnsupportedSport {
		t.Fatalf("err = %v, want ErrUnsupportedSport", err)
	}
}

func TestParseGPX_AcceptsMissingType(t *testing.T) {
	gpxNoType := strings.Replace(sampleGPX, "<type>cycling</type>", "", 1)
	_, err := ParseGPX(strings.NewReader(gpxNoType))
	if err != nil {
		t.Fatalf("ParseGPX: %v, want no error for a file with no <type>", err)
	}
}

func TestParseGPX_MalformedXML(t *testing.T) {
	_, err := ParseGPX(strings.NewReader("<gpx><trk"))
	if err != ErrCorruptFile {
		t.Fatalf("err = %v, want ErrCorruptFile", err)
	}
}

func TestParseGPX_TooFewPoints(t *testing.T) {
	_, err := ParseGPX(strings.NewReader(`<gpx><trk><trkseg><trkpt lat="0" lon="0"><time>2026-09-20T08:00:00Z</time></trkpt></trkseg></trk></gpx>`))
	if err != ErrNoData {
		t.Fatalf("err = %v, want ErrNoData", err)
	}
}
