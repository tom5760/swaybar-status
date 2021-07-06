// Package mpris implements a D-Bus client to the MPRIS interface.
// https://specifications.freedesktop.org/mpris-spec/2.2/
package mpris

const (
	mediaPlayerPropCanQuit             = mprisIface + ".CanQuit"
	mediaPlayerPropFullscreen          = mprisIface + ".Fullscreen"
	mediaPlayerPropCanSetFullscreen    = mprisIface + ".CanSetFullscreen"
	mediaPlayerPropCanRaise            = mprisIface + ".CanRaise"
	mediaPlayerPropHasTrackList        = mprisIface + ".HasTrackList"
	mediaPlayerPropIdentity            = mprisIface + ".Identity"
	mediaPlayerPropDesktopEntry        = mprisIface + ".DesktopEntry"
	mediaPlayerPropSupportedURISchemes = mprisIface + ".SupportedUriSchemes"
	mediaPlayerPropSupportedMIMETypes  = mprisIface + ".SupportedMimeTypes"

	mediaPlayerMethodRaise = mprisIface + ".Raise"
	mediaPlayerMethodQuit  = mprisIface + ".Quit"
)

// CanQuit reports what the Quit() method will do.  If false, calling Quit will
// have no effect, and may raise a NotSupported error. If true, calling Quit
// will cause the media application to attempt to quit (although it may still
// be prevented from quitting by the user, for example).
func (p *Player) CanQuit() (bool, error) {
	return p.obj.PropertyBool(mediaPlayerPropCanQuit)
}

// Fullscreen reports whether the media player is occupying the fullscreen.
//
// This is typically used for videos. A value of true indicates that the media
// player is taking up the full screen.
//
// Media centre software may well have this value fixed to true.
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) Fullscreen() (bool, error) {
	return p.obj.PropertyBool(mediaPlayerPropFullscreen)
}

// SetFullscreen sets whether the media player is occupying the fullscreen.
//
// If CanSetFullscreen is true, clients may set this property to true to tell
// the media player to enter fullscreen mode, or to false to return to windowed
// mode.
//
// If CanSetFullscreen is false, then attempting to set this property should
// have no effect, and may raise an error. However, even if it is true, the
// media player may still be unable to fulfil the request, in which case
// attempting to set this property will have no effect (but should not raise an
// error).
//
// This allows remote control interfaces, such as LIRC or mobile devices like
// phones, to control whether a video is shown in fullscreen.
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) SetFullscreen(v bool) error {
	return p.obj.SetProperty(mediaPlayerPropFullscreen, v)
}

// CanSetFullscreen determines whether SetFullscreen has an effect.
//
// If false, attempting to set Fullscreen will have no effect, and may raise an
// error. If true, attempting to set Fullscreen will not raise an error, and
// (if it is different from the current value) will cause the media player to
// attempt to enter or exit fullscreen mode.
//
// Note that the media player may be unable to fulfil the request. In this
// case, the value will not change. If the media player knows in advance that
// it will not be able to fulfil the request, however, this property should be
// false.
//
// This allows clients to choose whether to display controls for entering or
// exiting fullscreen mode.
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) CanSetFullscreen() (bool, error) {
	return p.obj.PropertyBool(mediaPlayerPropCanSetFullscreen)
}

// CanRaise reports what the Raise() method will do.  If false, calling Raise
// will have no effect, and may raise a NotSupported error. If true, calling
// Raise will cause the media application to attempt to bring its user
// interface to the front, although it may be prevented from doing so (by the
// window manager, for example).
func (p *Player) CanRaise() (bool, error) {
	return p.obj.PropertyBool(mediaPlayerPropCanRaise)
}

// HasTrackList indicates whether the /org/mpris/MediaPlayer2 object implements
// the org.mpris.MediaPlayer2.TrackList interface.
func (p *Player) HasTrackList() (bool, error) {
	return p.obj.PropertyBool(mediaPlayerPropHasTrackList)
}

// Identity returns a friendly name to identify the media player to users.
//
// This should usually match the name found in .desktop files (eg: "VLC media
// player").
func (p *Player) Identity() (string, error) {
	return p.obj.PropertyString(mediaPlayerPropIdentity)
}

// DesktopEntry returns the basename of an installed .desktop file which
// complies with the Desktop entry specification, with the ".desktop" extension
// stripped.
//
// Example: The desktop entry file is "/usr/share/applications/vlc.desktop",
// and this property contains "vlc".
//
// This property is optional. Clients should handle its absence gracefully.
func (p *Player) DesktopEntry() (string, error) {
	return p.obj.PropertyString(mediaPlayerPropDesktopEntry)
}

// SupportedURISchemes returns the URI schemes supported by the media player.
//
// This can be viewed as protocols supported by the player in almost all cases.
// Almost every media player will include support for the "file" scheme. Other
// common schemes are "http" and "rtsp".
//
// Note that URI schemes should be lower-case.
//
// This is important for clients to know when using the editing capabilities of the Playlist interface, for example.
func (p *Player) SupportedURISchemes() ([]string, error) {
	return p.obj.PropertySliceString(mediaPlayerPropSupportedURISchemes)
}

// SupportedMIMETypes returns the mime-types supported by the media player.
//
// Mime-types should be in the standard format (eg: audio/mpeg or
// application/ogg).
//
// This is important for clients to know when using the editing capabilities of
// the Playlist interface, for example.
func (p *Player) SupportedMIMETypes() ([]string, error) {
	return p.obj.PropertySliceString(mediaPlayerPropSupportedMIMETypes)
}

// Raise brings the media player's user interface to the front using any
// appropriate mechanism available.
//
// The media player may be unable to control how its user interface is
// displayed, or it may not have a graphical user interface at all. In this
// case, the CanRaise property is false and this method does nothing.
func (p *Player) Raise() error {
	return p.obj.Call(mediaPlayerMethodRaise, 0).Err
}

// Quit causes the media player to stop running.
//
// The media player may refuse to allow clients to shut it down. In this case,
// the CanQuit property is false and this method does nothing.
//
// Note: Media players which can be D-Bus activated, or for which there is no
// sensibly easy way to terminate a running instance (via the main interface or
// a notification area icon for example) should allow clients to use this
// method. Otherwise, it should not be needed.
//
// If the media player does not have a UI, this should be implemented.
func (p *Player) Quit() error {
	return p.obj.Call(mediaPlayerMethodQuit, 0).Err
}
