package mpris

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	playerIface = mprisIface + ".Player"

	playerPropPlaybackStatus = playerIface + ".PlaybackStatus"
	playerPropLoopStatus     = playerIface + ".LoopStatus"
	playerPropRate           = playerIface + ".Rate"
	playerPropShuffle        = playerIface + ".Shuffle"
	playerPropMetadata       = playerIface + ".Metadata"
	playerPropVolume         = playerIface + ".Volume"
	playerPropPosition       = playerIface + ".Position"
	playerPropMinimumRate    = playerIface + ".MinimumRate"
	playerPropMaximumRate    = playerIface + ".MaximumRate"
	playerPropCanGoNext      = playerIface + ".CanGoNext"
	playerPropCanGoPrevious  = playerIface + ".CanGoPrevious"
	playerPropCanPlay        = playerIface + ".CanPlay"
	playerPropCanPause       = playerIface + ".CanPause"
	playerPropCanSeek        = playerIface + ".CanSeek"
	playerPropCanControl     = playerIface + ".CanControl"

	playerMethodNext        = playerIface + ".Next"
	playerMethodPrevious    = playerIface + ".Previous"
	playerMethodPause       = playerIface + ".Pause"
	playerMethodPlayPause   = playerIface + ".PlayPause"
	playerMethodStop        = playerIface + ".Stop"
	playerMethodPlay        = playerIface + ".Play"
	playerMethodSeek        = playerIface + ".Seek"
	playerMethodSetPosition = playerIface + ".SetPosition"
	playerMethodOpenURI     = playerIface + ".OpenUri"

	playerSignalSeeked = playerIface + ".Seeked"
)

type (
	// PlaybackStatus is a playback state.
	PlaybackStatus string

	// LoopStatus is a repeat / loop status.
	LoopStatus string
)

const (
	PlaybackStatusPlaying PlaybackStatus = "Playing" // A track is currently playing.
	PlaybackStatusPaused  PlaybackStatus = "Paused"  // A track is currently paused.
	PlaybackStatusStopped PlaybackStatus = "Stopped" // There is no track currently playing.
)

const (
	LoopStatusNone     LoopStatus = "None"     // The playback will stop when there are no more tracks to play.
	LoopStatusTrack    LoopStatus = "Track"    // The current track will start again from the begining once it has finished playing.
	LoopStatusPlaylist LoopStatus = "Playlist" // The playback loops through a list of tracks.
)

// PlaybackStatus reports the current playback status.
//
// May be "Playing", "Paused" or "Stopped".
func (p *Player) PlaybackStatus() (PlaybackStatus, error) {
	s, err := p.obj.PropertyString(playerPropPlaybackStatus)
	return PlaybackStatus(s), err
}

// LoopStatus reports the current loop / repeat status.
//
// May be:
// * LoopStatusNone - if the playback will stop when there are no more tracks
//                    to play.
// * LoopStatusTrack - if the current track will start again from the begining
//                     once it has finished playing.
// * LoopStatusPlaylist - if the playback loops through a list of tracks.
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) LoopStatus() (LoopStatus, error) {
	s, err := p.obj.PropertyString(playerPropLoopStatus)
	return LoopStatus(s), err
}

// SetLoopStatus sets the current loop / repeat status.
//
// May be:
// * LoopStatusNone - if the playback will stop when there are no more tracks
//                    to play.
// * LoopStatusTrack - if the current track will start again from the begining
//                     once it has finished playing.
// * LoopStatusPlaylist - if the playback loops through a list of tracks.
//
// If CanControl is false, attempting to set this property should have no
// effect and raise an error.
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) SetLoopStatus(v LoopStatus) error {
	return p.obj.SetProperty(playerPropLoopStatus, v)
}

// Rate reports the current playback rate.
//
// If the media player has no ability to play at speeds other than the normal
// playback rate, this must still be implemented, and must return 1.0. The
// MinimumRate and MaximumRate properties must also be set to 1.0.
//
// This allows clients to display (reasonably) accurate progress bars without
// having to regularly query the media player for the current position.
func (p *Player) Rate() (float64, error) {
	return p.obj.PropertyFloat64(playerPropRate)
}

// SetRate sets the current playback rate.
//
// The value must fall in the range described by MinimumRate and MaximumRate,
// and must not be 0.0. If playback is paused, the PlaybackStatus property
// should be used to indicate this. A value of 0.0 should not be set by the
// client. If it is, the media player should act as though Pause was called.
//
// Not all values may be accepted by the media player. It is left to media
// player implementations to decide how to deal with values they cannot use;
// they may either ignore them or pick a "best fit" value. Clients are
// recommended to only use sensible fractions or multiples of 1 (eg: 0.5, 0.25,
// 1.5, 2.0, etc).
func (p *Player) SetRate(v float64) error {
	return p.obj.SetProperty(playerPropRate, v)
}

// Shuffle reports the shuffle state.
//
// A value of false indicates that playback is progressing linearly through a
// playlist, while true means playback is progressing through a playlist in
// some other order.
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) Shuffle() (bool, error) {
	return p.obj.PropertyBool(playerPropShuffle)
}

// SetShuffle reports the shuffle state.
//
// A value of false indicates that playback is progressing linearly through a
// playlist, while true means playback is progressing through a playlist in
// some other order.
//
// If CanControl is false, attempting to set this property should have no
// effect and raise an error.
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) SetShuffle(v bool) error {
	return p.obj.SetProperty(playerPropShuffle, v)
}

// Metadata returns metadata of the current element.
//
// If there is a current track, this must have a "mpris:trackid" entry (of
// D-Bus type "o") at the very least, which contains a D-Bus path that uniquely
// identifies this track.
//
// See the type documentation for more details.
func (p *Player) Metadata() (MetadataMap, error) {
	v, err := p.obj.Property(playerPropMetadata)
	if err != nil {
		return nil, err
	}

	x, ok := v.(map[string]dbus.Variant)
	if !ok {
		return nil, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return MetadataMap(x), nil
}

// Volume reports the volume level.
func (p *Player) Volume() (float64, error) {
	return p.obj.PropertyFloat64(playerPropVolume)
}

// SetVolume sets the volume level.
//
// When setting, if a negative value is passed, the volume should be set to
// 0.0.
//
// If CanControl is false, attempting to set this property should have no
// effect and raise an error.
func (p *Player) SetVolume(v float64) error {
	return p.obj.SetProperty(playerPropVolume, v)
}

// Position reports the current track position in microseconds, between 0 and
// the 'mpris:length' metadata entry (see Metadata).
//
// Note: If the media player allows it, the current playback position can be
// changed either the SetPosition method or the Seek method on this interface.
// If this is not the case, the CanSeek property is false, and setting this
// property has no effect and can raise an error.
//
// If the playback progresses in a way that is inconstistant with the Rate
// property, the Seeked signal is emited.
func (p *Player) Position() (int64, error) {
	return p.obj.PropertyInt64(playerPropPosition)
}

// MinimumRate reports the maximum value which the Rate property can take.
// Clients should not attempt to set the Rate property above this value.
//
// This value should always be 1.0 or greater.
func (p *Player) MinimumRate() (float64, error) {
	return p.obj.PropertyFloat64(playerPropMinimumRate)
}

// MaximumRate reports the maximum value which the Rate property can take.
// Clients should not attempt to set the Rate property above this value.
//
// This value should always be 1.0 or greater.
func (p *Player) MaximumRate() (float64, error) {
	return p.obj.PropertyFloat64(playerPropMaximumRate)
}

// CanGoNext reports whether the client can call the Next method on this
// interface and expect the current track to change.
//
// If it is unknown whether a call to Next will be successful (for example,
// when streaming tracks), this property should be set to true.
//
// If CanControl is false, this property should also be false.
//
// Even when playback can generally be controlled, there may not always be a
// next track to move to.
func (p *Player) CanGoNext() (bool, error) {
	return p.obj.PropertyBool(playerPropCanGoNext)
}

// CanGoPrevious reports whether the client can call the Previous method on
// this interface and expect the current track to change.
//
// If it is unknown whether a call to Previous will be successful (for example,
// when streaming tracks), this property should be set to true.
//
// If CanControl is false, this property should also be false.
//
// Even when playback can generally be controlled, there may not always be a
// next previous to move to.
func (p *Player) CanGoPrevious() (bool, error) {
	return p.obj.PropertyBool(playerPropCanGoPrevious)
}

// CanPlay reports whether playback can be started using Play or PlayPause.
//
// Note that this is related to whether there is a "current track": the value
// should not depend on whether the track is currently paused or playing. In
// fact, if a track is currently playing (and CanControl is true), this should
// be true.
//
// If CanControl is false, this property should also be false.
//
// Even when playback can generally be controlled, it may not be possible to
// enter a "playing" state, for example if there is no "current track".
func (p *Player) CanPlay() (bool, error) {
	return p.obj.PropertyBool(playerPropCanPlay)
}

// CanPause reports whether playback can be paused using Pause or PlayPause.
//
// Note that this is an intrinsic property of the current track: its value
// should not depend on whether the track is currently paused or playing. In
// fact, if playback is currently paused (and CanControl is true), this should
// be true.
//
// If CanControl is false, this property should also be false.
//
// Not all media is pausable: it may not be possible to pause some streamed
// media, for example.
func (p *Player) CanPause() (bool, error) {
	return p.obj.PropertyBool(playerPropCanPause)
}

// CanSeek reports whether the client can control the playback position using
// Seek and SetPosition. This may be different for different tracks.
//
// If CanControl is false, this property should also be false.
//
// Not all media is seekable: it may not be possible to seek when playing some
// streamed media, for example.
func (p *Player) CanSeek() (bool, error) {
	return p.obj.PropertyBool(playerPropCanSeek)
}

// CanControl reports whether the media player may be controlled over this
// interface.
//
// This property is not expected to change, as it describes an intrinsic
// capability of the implementation.
//
// If this is false, clients should assume that all properties on this
// interface are read-only (and will raise errors if writing to them is
// attempted), no methods are implemented and all other properties starting
// with "Can" are also false.
//
// This allows clients to determine whether to present and enable controls to
// the user in advance of attempting to call methods and write to properties.
func (p *Player) CanControl() (bool, error) {
	return p.obj.PropertyBool(playerPropCanControl)
}

// Next skips to the next track in the tracklist.
//
// If there is no next track (and endless playback and track repeat are both
// off), stop playback.
//
// If playback is paused or stopped, it remains that way.
//
// If CanGoNext is false, attempting to call this method should have no effect.
func (p *Player) Next() error {
	return p.obj.Call(playerMethodNext, 0).Err
}

// Previous skips to the previous track in the tracklist.
//
// If there is no previous track (and endless playback and track repeat are
// both off), stop playback.
//
// If playback is paused or stopped, it remains that way.
//
// If CanGoPrevious is false, attempting to call this method should have no
// effect.
func (p *Player) Previous() error {
	return p.obj.Call(playerMethodPrevious, 0).Err
}

// Pause pauses playback.
//
// If playback is already paused, this has no effect.
//
// Calling Play after this should cause playback to start again from the same
// position.
//
// If CanPause is false, attempting to call this method should have no effect.
func (p *Player) Pause() error {
	return p.obj.Call(playerMethodPause, 0).Err
}

// PlayPause pauses playback.
//
// If playback is already paused, resumes playback.
//
// If playback is stopped, starts playback.
//
// If CanPause is false, attempting to call this method should have no effect
// and raise an error.
func (p *Player) PlayPause() error {
	return p.obj.Call(playerMethodPlayPause, 0).Err
}

// Stop stops playback.
//
// If playback is already stopped, this has no effect.
//
// Calling Play after this should cause playback to start again from the
// beginning of the track.
//
// If CanControl is false, attempting to call this method should have no effect
// and raise an error.
func (p *Player) Stop() error {
	return p.obj.Call(playerMethodStop, 0).Err
}

// Play starts or resumes playback.
//
// If already playing, this has no effect.
//
// If paused, playback resumes from the current position.
//
// If there is no track to play, this has no effect.
//
// If CanPlay is false, attempting to call this method should have no effect.
func (p *Player) Play() error {
	return p.obj.Call(playerMethodPlay, 0).Err
}

// Seek seeks forward in the current track by the specified number of
// microseconds.
//
//   offset - The number of microseconds to seek forward.
//
// A negative value seeks back. If this would mean seeking back further than
// the start of the track, the position is set to 0.
//
// If the value passed in would mean seeking beyond the end of the track, acts
// like a call to Next.
//
// If the CanSeek property is false, this has no effect.
func (p *Player) Seek(offset float64) error {
	return p.obj.Call(playerMethodSeek, 0, offset).Err
}

// SetPosition sets the current track position in microseconds.
//
//   trackID: The currently playing track's identifier.
//
//     If this does not match the id of the currently-playing track, the call
//     is ignored as "stale".
//
//     /org/mpris/MediaPlayer2/TrackList/NoTrack is not a valid value for this
//     argument.
//
//   position: Track position in microseconds.
//
//     This must be between 0 and <track_length>.
//
// If the Position argument is less than 0, do nothing.
//
// If the Position argument is greater than the track length, do nothing.
//
// If the CanSeek property is false, this has no effect.
//
// The reason for having this method, rather than making Position writable, is
// to include the TrackId argument to avoid race conditions where a client
// tries to seek to a position when the track has already changed.
func (p *Player) SetPosition(trackID dbus.ObjectPath, position float64) error {
	return p.obj.Call(playerMethodSetPosition, 0, trackID, position).Err
}

// OpenURI opens the Uri given as an argument.
//
//   uri: URI of the track to load. Its uri scheme should be an element of the
//     org.mpris.MediaPlayer2.SupportedUriSchemes property and the mime-type
//     should match one of the elements of the
//     org.mpris.MediaPlayer2.SupportedMimeTypes.
//
// If the playback is stopped, starts playing
//
// If the uri scheme or the mime-type of the uri to open is not supported, this
// method does nothing and may raise an error. In particular, if the list of
// available uri schemes is empty, this method may not be implemented.
//
// Clients should not assume that the Uri has been opened as soon as this
// method returns. They should wait until the mpris:trackid field in the
// Metadata property changes.
//
// If the media player implements the TrackList interface, then the opened
// track should be made part of the tracklist, the
// org.mpris.MediaPlayer2.TrackList.TrackAdded or
// org.mpris.MediaPlayer2.TrackList.TrackListReplaced signal should be fired,
// as well as the org.freedesktop.DBus.Properties.PropertiesChanged signal on
// the tracklist interface.
func (p *Player) OpenURI(uri string) error {
	return p.obj.Call(playerMethodOpenURI, 0, uri).Err
}
