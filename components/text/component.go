package text

import (
	"time"

	"github.com/gdamore/tcell"
	"retort.dev/components/box"
	"retort.dev/r"
	"retort.dev/r/debug"
	"retort.dev/r/intmath"
)

type boxState struct {
	OffsetX, OffsetY int
	lastUpdated      time.Time
}

// Text is the basic building block for a retort app.
// Text implements the Text Model, see Properties
func Text(props r.Properties) r.Element {
	// Get our Properties
	textProps := props.GetProperty(
		Properties{},
		"Text requires Properties",
	).(Properties)

	// Get our Properties
	boxProps := props.GetOptionalProperty(
		box.Properties{},
	).(box.Properties)

	// Get our BlockLayout
	parentBlockLayout := props.GetProperty(
		r.BlockLayout{},
		"Text requires a parent BlockLayout.",
	).(r.BlockLayout)

	// Get any children
	children := props.GetOptionalProperty(
		r.Children{},
	).(r.Children)
	if len(children) != 0 {
		panic("Text cannot have children")
	}

	s, setState := r.UseState(r.State{
		boxState{lastUpdated: time.Now()},
	})
	state := s.GetState(
		boxState{},
	).(boxState)

	mouseEventHandler := func(up, down, left, right bool) {
		now := time.Now()

		if now.Sub(state.lastUpdated) < 2*time.Millisecond {
			// throttle to one update a second
			return
		}

		setState(func(s r.State) r.State {
			state := s.GetState(
				boxState{},
			).(boxState)

			offsetXDelta := 0
			offsetYDelta := 0

			switch {
			case up:
				offsetXDelta = -1
				if state.OffsetX == 0 {
					return r.State{state}
				}
			case down:
				offsetXDelta = 1
			case left:
				offsetYDelta = -1
				if state.OffsetY == 0 {
					return r.State{state}
				}

			case right:
				offsetYDelta = 1
			}

			if offsetXDelta == 0 && offsetYDelta == 0 {
				return r.State{state}
			}

			offsetX := state.OffsetX
			offsetY := state.OffsetY

			if boxProps.Overflow == box.OverflowScroll ||
				boxProps.Overflow == box.OverflowScrollX {
				// When the offset is near the top, we just set the value
				// this prevents issues with the float64 conversion below
				// that was casuing jankiness
				if state.OffsetX < 2 {
					offsetX = state.OffsetX + offsetXDelta
				} else {
					offsetX = intmath.Min(
						intmath.Abs(state.OffsetX+offsetXDelta),
						int(float64(parentBlockLayout.Columns)/0.2),
					)
				}
			}

			if boxProps.Overflow == box.OverflowScroll ||
				boxProps.Overflow == box.OverflowScrollY {
				if offsetY < 2 {
					offsetY = state.OffsetY + offsetYDelta
				} else {
					offsetY = intmath.Min(
						intmath.Abs(state.OffsetY+offsetYDelta),
						int(float64(parentBlockLayout.Rows)/0.2),
					)
				}
			}

			return r.State{boxState{
				OffsetX:     offsetX,
				OffsetY:     offsetY,
				lastUpdated: time.Now(),
			},
			}
		})
	}

	p := r.Properties{
		boxProps,
	}
	debug.Spew(boxProps)
	if boxProps.Overflow != box.OverflowNone {
		p = append(p, mouseEventHandler)
	}

	return r.CreateElement(
		box.Box,
		p,
		r.Children{
			r.CreateScreenElement(
				calculateBlockLayout(textProps, boxProps),
				func(s tcell.Screen, blockLayout r.BlockLayout) {
					if s == nil {
						panic("Text can't render no screen")
					}

					w, h := s.Size()

					if w == 0 || h == 0 {
						panic("Text can't render on a zero size screen")
					}

					render(
						s,
						textProps,
						boxProps,
						blockLayout,
						state.OffsetX, state.OffsetY,
					)
				},
				r.Properties{},
				nil,
			),
		},
	)
}
