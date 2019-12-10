package main

// Header is a JSON object that must be sent first.
type Header struct {
	// The protocol version to use. Currently, this must be 1.
	Version int `json:"version"`

	// Whether to receive click event informaeion to standard input.
	ClickEvents bool `json:"click_events,omitempty"`

	// The signal that swaybar should send to continue processing.
	ContSignal int `json:"cont_signal,omitempty"`

	// The signal that swaybar should send to stop processing
	StopSignal int `json:"stop_signal,omitempty"`
}

// Block is one element of a body array.  It is a representation of a single
// block in the status line.
type Block struct {
	// The text that will be displayed.  If missing, the block will be skipped.
	FullText string `json:"full_text"`

	// If given and the text needs to be shortened due to space, this will be
	// displayed instead of full_text.
	ShortText string `json:"short_text,omitempty"`

	// The text color to use in #RRGGBBAA or #RRGGBB notation.
	Color string `json:"color,omitempty"`

	// The background color for the block in #RRGGBBAA or #RRGGBB notation.
	Background string `json:"background,omitempty"`

	// The border color for the block in #RRGGBBAA or #RRGGBB notation.
	Border string `json:"border,omitempty"`

	// The height in pixels of the top border. The default is 1.
	BorderTop int `json:"border_top,omitempty"`

	// The height in pixels of the bottom border. The default is 1.
	BorderBottom int `json:"border_bottom,omitempty"`

	// The height in pixels of the left border. The default is 1.
	BorderLeft int `json:"border_left,omitempty"`

	// The height in pixels of the right border. The default is 1.
	BorderRight int `json:"border_right,omitempty"`

	// The minimum width to use for the block. This can either be given in pixels
	// or a string can be given to allow for it to be calculated based on the
	// width of the string.
	MinWidth string `json:"min_width,omitempty"`

	// If the text does not span the full width of the block, this specifies how
	// the text should be aligned inside of the block. This can be left
	// (default), right, or center.
	Align string `json:"align,omitempty"`

	// A name for the block. This is only used to identify the block for click
	// events. If set, each block should have a unique name and instance pair.
	Name string `json:"name,omitempty"`

	// The instance of the name for the block. This is only used to identify the
	// block for click events. If set, each block should have a unique name and
	// instance pair.
	Instance string `json:"instance,omitempty"`

	// Whether the block should be displayed as urgent. Currently swaybar
	// utilizes the colors set in the sway config for urgent workspace buttons.
	// See sway-bar(5) for more information on bar color configuration.
	Urgent bool `json:"urgent,omitempty"`

	// Whether the bar separator should be drawn after the block. See sway-bar(5)
	// for more information on how to set the separator text.
	Separator bool `json:"separator,omitempty"`

	// The amount of pixels to leave blank after the block. The separator text
	// will be displayed centered in this gap. The default is 9 pixels.
	SeparatorBlockWidth int `json:"separator_block_width,omitempty"`

	// The type of markup to use when parsing the text for the block. This can
	// either be pango or none (default).
	Markup string `json:"markup,omitempty"`

	// Called when the block is clicked on.
	ClickHandler func(ClickEvent) `json:"-"`

	// If true, block is removed from the status line instead of added.
	Remove bool `json:"-"`
}

// ClickEvents are reported if requested in the header.
type ClickEvent struct {
	// The name of the block, if set.
	Name string

	// The instance of the block, if set.
	Instance string

	// The x location that the click occurred at.
	X int

	// The y location that the click occurred at.
	Y int

	// The x11 button number for the click. If the button does not have an x11
	// button mapping, this will be 0.
	Button int

	// The event code that corresponds to the button for the click
	Event int

	// The x location of the click relative to the top-left of the block.
	RelativeX int `json:"relative_x"`

	// The y location of the click relative to the top-left of the block.
	RelativeY int `json:"relative_y"`

	// The width of the block in pixels
	Width int

	// The height of the block in pixels
	Height int
}
