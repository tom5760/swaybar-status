package mpris

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	metadataKeyTrackID = "mpris:trackid"
	metadataKeyLength  = "mpris:length"
	metadataKeyArtURL  = "mpris:artUrl"

	metadataKeyAlbum  = "xesam:album"
	metadataKeyArtist = "xesam:artist"
	metadataKeyTitle  = "xesam:title"
)

var (
	ErrMetadataNoEntry = errors.New("metadata entry not found")
)

// MetadataMap is a mapping from metadata attribute names to values.
//
// The mpris:trackid attribute must always be present, and must be of D-Bus
// type "o". This contains a D-Bus path that uniquely identifies the track
// within the scope of the playlist. There may or may not be an actual D-Bus
// object at that path; this specification says nothing about what interfaces
// such an object may implement.
//
// If the length of the track is known, it should be provided in the metadata
// property with the "mpris:length" key. The length must be given in
// microseconds, and be represented as a signed 64-bit integer.
//
// If there is an image associated with the track, a URL for it may be
// provided using the "mpris:artUrl" key. For other metadata, fields defined
// by the Xesam ontology should be used, prefixed by "xesam:". See the
// metadata page on the freedesktop.org wiki for a list of common fields.
//
// Lists of strings should be passed using the array-of-string ("as") D-Bus
// type. Dates should be passed as strings using the ISO 8601 extended format
// (eg: 2007-04-29T14:35:51). If the timezone is known, RFC 3339's internet
// profile should be used (eg: 2007-04-29T14:35:51+02:00).
type MetadataMap map[string]dbus.Variant

// TrackID returns a unique identity for this track within the context of an
// MPRIS object (eg: tracklist).
func (m MetadataMap) TrackID() (dbus.ObjectPath, error) {
	v, ok := m[metadataKeyTrackID]
	if !ok {
		return "", ErrMetadataNoEntry
	}

	x, ok := v.Value().(dbus.ObjectPath)
	if !ok {
		return "", fmt.Errorf("unexpected metadata type; got %T; expected %T", v, x)
	}

	return x, nil
}

// Length returns the duration of the track in microseconds.
func (m MetadataMap) Length() (int64, error) {
	v, ok := m[metadataKeyLength]
	if !ok {
		return 0, ErrMetadataNoEntry
	}

	x, ok := v.Value().(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected metadata type; got %T; expected %T", v, x)
	}

	return x, nil
}

// ArtURL returns the location of an image representing the track or album.
// Clients should not assume this will continue to exist when the media player
// stops giving out the URL.
func (m MetadataMap) ArtURL() (string, error) {
	return m.propertyString(metadataKeyArtURL)
}

// Album returns the album name.
func (m MetadataMap) Album() (string, error) {
	return m.propertyString(metadataKeyAlbum)
}

// Artist returns the track artist(s).
func (m MetadataMap) Artist() ([]string, error) {
	return m.propertyStringSlice(metadataKeyArtist)
}

// Title returns the track title.
func (m MetadataMap) Title() (string, error) {
	return m.propertyString(metadataKeyTitle)
}

func (m MetadataMap) propertyString(key string) (string, error) {
	v, ok := m[key]
	if !ok {
		return "", ErrMetadataNoEntry
	}

	x, ok := v.Value().(string)
	if !ok {
		return "", fmt.Errorf("unexpected metadata type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (m MetadataMap) propertyStringSlice(key string) ([]string, error) {
	v, ok := m[key]
	if !ok {
		return nil, ErrMetadataNoEntry
	}

	x, ok := v.Value().([]string)
	if !ok {
		return nil, fmt.Errorf("unexpected metadata type; got %T; expected %T", v, x)
	}

	return x, nil
}
