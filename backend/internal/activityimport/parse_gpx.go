package activityimport

import (
	"encoding/xml"
	"io"
	"math"
	"strings"
	"time"
)

// gpxFile mirrors only the elements this app reads. encoding/xml matches a
// tag by local name when the struct tag carries no namespace prefix, so this
// works across the different namespace URIs GPX exporters use (Garmin,
// Strava, Wahoo, Komoot) without listing each one.
type gpxFile struct {
	Tracks []gpxTrack `xml:"trk"`
}

type gpxTrack struct {
	Type     string       `xml:"type"`
	Segments []gpxSegment `xml:"trkseg"`
}

type gpxSegment struct {
	Points []gpxPoint `xml:"trkpt"`
}

type gpxPoint struct {
	Lat        float64       `xml:"lat,attr"`
	Lon        float64       `xml:"lon,attr"`
	Ele        *float64      `xml:"ele"`
	Time       *time.Time    `xml:"time"`
	Extensions gpxExtensions `xml:"extensions"`
}

// gpxExtensions covers the Garmin TrackPointExtension fields, the most common
// way heart rate, cadence and power ride along a <trkpt>. Some exporters put
// power directly on the point instead of inside the extension; both are read.
type gpxExtensions struct {
	HeartRate      *int `xml:"TrackPointExtension>hr"`
	Cadence        *int `xml:"TrackPointExtension>cad"`
	PowerExtension *int `xml:"TrackPointExtension>power"`
	Power          *int `xml:"power"`
}

// nonCyclingKeywords rejects only what a <trk><type> clearly states is not
// cycling; an absent or unrecognized type is accepted, since GPX has no
// standard vocabulary for sport and the app itself only offers cycling.
var nonCyclingKeywords = []string{"run", "walk", "hik", "swim", "corrida", "caminhada", "natação", "natação"}

// ParseGPX decodes a .gpx track and extracts a cycling activity summary from
// its track points. GPX has no standard field for heart rate, power or
// cadence; this function reads the de facto Garmin extension used by most
// devices and apps, and leaves the metric unset when a file omits it.
func ParseGPX(r io.Reader) (Parsed, error) {
	var doc gpxFile
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return Parsed{}, ErrCorruptFile
	}
	var points []gpxPoint
	sportType := ""
	for _, track := range doc.Tracks {
		if track.Type != "" {
			sportType = track.Type
		}
		for _, segment := range track.Segments {
			points = append(points, segment.Points...)
		}
	}
	if sportType != "" && looksNonCycling(sportType) {
		return Parsed{}, ErrUnsupportedSport
	}
	points = pointsWithTime(points)
	if len(points) < 2 {
		return Parsed{}, ErrNoData
	}
	return summarizeGPXPoints(points), nil
}

func looksNonCycling(sportType string) bool {
	lower := strings.ToLower(sportType)
	for _, keyword := range nonCyclingKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func pointsWithTime(points []gpxPoint) []gpxPoint {
	filtered := points[:0]
	for _, p := range points {
		if p.Time != nil {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func summarizeGPXPoints(points []gpxPoint) Parsed {
	first, last := points[0], points[len(points)-1]
	parsed := Parsed{
		StartedAt:     *first.Time,
		MovingSeconds: int(last.Time.Sub(*first.Time).Seconds()),
	}
	var distanceMeters, elevationGain float64
	var hrSamples, cadenceSamples, powerSamples []int
	prevElevation, hasPrevElevation := 0.0, false
	for i, p := range points {
		if i > 0 {
			distanceMeters += haversineMeters(points[i-1].Lat, points[i-1].Lon, p.Lat, p.Lon)
		}
		if p.Ele != nil {
			if hasPrevElevation && *p.Ele > prevElevation {
				elevationGain += *p.Ele - prevElevation
			}
			prevElevation, hasPrevElevation = *p.Ele, true
		}
		if hr := firstNonNil(p.Extensions.HeartRate); hr != nil {
			hrSamples = append(hrSamples, *hr)
		}
		if cad := firstNonNil(p.Extensions.Cadence); cad != nil {
			cadenceSamples = append(cadenceSamples, *cad)
		}
		if power := firstNonNil(p.Extensions.Power, p.Extensions.PowerExtension); power != nil {
			powerSamples = append(powerSamples, *power)
		}
	}
	parsed.DistanceKM = distanceMeters / 1000
	if hasPrevElevation {
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
	parsed.AverageCadenceRPM = avgInt(cadenceSamples)
	parsed.AveragePowerW = avgInt(powerSamples)
	return parsed
}

func firstNonNil(candidates ...*int) *int {
	for _, c := range candidates {
		if c != nil {
			return c
		}
	}
	return nil
}

// haversineMeters is the standard great-circle distance between two
// coordinates, accurate enough for consecutive GPS points a few seconds
// apart (the error from ignoring elevation is negligible at that scale).
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusM = 6371000.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}
