package box

import (
	"math"

	"github.com/gdamore/tcell"
	"retort.dev/r"
)

func calculateBlockLayout(
	props Properties,
) r.CalculateLayout {
	return func(
		s tcell.Screen,
		stage r.CalculateLayoutStage,
		parentBlockLayout r.BlockLayout,
		children []r.BlockLayoutWithProperties,
	) (
		outerBlockLayout r.BlockLayout,
		innerBlockLayout r.BlockLayout,
		childrenBlockLayouts []r.BlockLayoutWithProperties,
	) {
		childrenBlockLayouts = children

		switch stage {
		case r.CalculateLayoutStageInitial:
			// if any widths or heights are explicitly set, set them here
			// otherwise inherit from the parentBlockLayout
			rows := parentBlockLayout.Rows
			columns := parentBlockLayout.Columns

			if props.Rows == 0 && props.Height != 0 {
				rows = int(
					math.Round(
						float64(parentBlockLayout.Rows) * float64(props.Height) / 100,
					),
				)
				outerBlockLayout.FixedRows = true
			} else if props.Rows != 0 {
				rows = props.Rows
				outerBlockLayout.FixedRows = true
			}

			if props.Columns == 0 && props.Width != 0 {
				columns = int(
					math.Round(
						float64(parentBlockLayout.Columns) * float64(props.Width) / 100,
					),
				)
				outerBlockLayout.FixedColumns = true
			} else if props.Columns != 0 {
				columns = props.Columns
				outerBlockLayout.FixedColumns = true
			}

			outerBlockLayout = r.BlockLayout{
				ZIndex:       props.ZIndex,
				Rows:         rows,
				Columns:      columns,
				Grow:         props.Grow,
				X:            parentBlockLayout.X,
				Y:            parentBlockLayout.Y,
				FixedColumns: outerBlockLayout.FixedColumns,
				FixedRows:    outerBlockLayout.FixedRows,
				Valid:        true,
			}

			// Calculate margin
			outerBlockLayout.X = parentBlockLayout.X + props.Margin.Left
			outerBlockLayout.Columns = outerBlockLayout.Columns - props.Margin.Right
			outerBlockLayout.Y = parentBlockLayout.Y + props.Margin.Top
			outerBlockLayout.Rows = outerBlockLayout.Rows - props.Margin.Bottom

			innerBlockLayout = r.BlockLayout{
				ZIndex:       props.ZIndex,
				Rows:         outerBlockLayout.Rows,
				Columns:      outerBlockLayout.Columns,
				X:            outerBlockLayout.X,
				Y:            outerBlockLayout.Y,
				FixedColumns: outerBlockLayout.FixedColumns,
				FixedRows:    outerBlockLayout.FixedRows,
				Valid:        true,
			}

			innerBlockLayout.Columns = innerBlockLayout.Columns -
				props.Padding.Left - props.Padding.Right

			innerBlockLayout.Rows = innerBlockLayout.Rows -
				props.Padding.Top - props.Padding.Bottom

			// // Calculate padding box
			innerBlockLayout.Y = innerBlockLayout.Y + props.Padding.Top
			innerBlockLayout.Columns = innerBlockLayout.Columns - props.Padding.Right
			innerBlockLayout.Rows = innerBlockLayout.Rows - props.Padding.Bottom
			innerBlockLayout.X = innerBlockLayout.X + props.Padding.Left

			// Border Sizing
			if props.Border.Style != BorderStyleNone {
				outerBlockLayout.Columns = outerBlockLayout.Columns - 2 // 1 for each side
				outerBlockLayout.Rows = outerBlockLayout.Rows - 2       // 1 for each side

				// only one border type at the moment
				outerBlockLayout.Border.Top = 1
				outerBlockLayout.Border.Right = 1
				outerBlockLayout.Border.Bottom = 1
				outerBlockLayout.Border.Left = 1

				innerBlockLayout.X = innerBlockLayout.X + 1
				innerBlockLayout.Y = innerBlockLayout.Y + 1
				innerBlockLayout.Rows = innerBlockLayout.Rows - 2
				innerBlockLayout.Columns = innerBlockLayout.Columns - 2
			}

			// Ensure the rows and cols are not below 0
			if outerBlockLayout.Rows < 0 {
				outerBlockLayout.Rows = 0
			}
			if outerBlockLayout.Columns < 0 {
				outerBlockLayout.Columns = 0
			}
			if innerBlockLayout.Rows < 0 {
				innerBlockLayout.Rows = 0
			}
			if innerBlockLayout.Columns < 0 {
				innerBlockLayout.Columns = 0
			}
			return
		case r.CalculateLayoutStageWithChildren:
			if len(children) == 0 {
				return
			}

			// Look at all the children who have widths, and add them up
			// then split the remainder between those without widths
			innerBlockLayout = r.BlockLayout{
				ZIndex:  props.ZIndex,
				Rows:    parentBlockLayout.Rows,
				Columns: parentBlockLayout.Columns,
				X:       parentBlockLayout.X,
				Y:       parentBlockLayout.Y,
				Valid:   true,
			}

			colsRemaining := innerBlockLayout.Columns
			rowsRemaining := innerBlockLayout.Rows
			growCount := 0
			// growDivision is the number of cols/rows each grow is worth
			growDivision := 0

			// Find all children with fixed row,col sizing, and count all grow's
			for i, c := range children {
				// debug.Spew("c", c)
				if c.BlockLayout.FixedColumns {
					colsRemaining = colsRemaining - c.BlockLayout.Columns
				}
				if c.BlockLayout.FixedRows {
					rowsRemaining = rowsRemaining - c.BlockLayout.Rows
				}

				cProps := c.Properties.GetOptionalProperty(
					Properties{},
				).(Properties)

				grow := cProps.Grow
				c.BlockLayout.Grow = grow

				growCount = growCount + c.BlockLayout.Grow

				if c.BlockLayout.Grow <= 0 {
					growCount = growCount + 1 // we force grow to be at least 1
				}

				childrenBlockLayouts[i].BlockLayout.Grow = grow

			}

			switch props.Direction {
			case DirectionRow:
				growDivision = colsRemaining / growCount
			case DirectionRowReverse:
				growDivision = colsRemaining / growCount
			case DirectionColumn:
				growDivision = rowsRemaining / growCount
			case DirectionColumnReverse:
				growDivision = rowsRemaining / growCount
			}

			// Reverse the slices if needed
			if props.Direction == DirectionRowReverse ||
				props.Direction == DirectionColumnReverse {
				for i := len(children)/2 - 1; i >= 0; i-- {
					opp := len(children) - 1 - i
					children[i], children[opp] = children[opp], children[i]
				}
			}

			// Get our starting position
			x := innerBlockLayout.X
			y := innerBlockLayout.Y

			// Calculate initial blockLayout for children
			for i, c := range children {

				grow := c.BlockLayout.Grow

				rows := 0
				columns := 0

				if !c.BlockLayout.FixedColumns || !c.BlockLayout.FixedRows {
					// Calculate the size of this block based on the direction of the parent
					switch props.Direction {
					case DirectionRow:
						columns = growDivision * grow
						rows = innerBlockLayout.Rows
					case DirectionRowReverse:
						columns = growDivision * grow
						rows = innerBlockLayout.Rows
					case DirectionColumn:
						columns = innerBlockLayout.Columns
						rows = growDivision * grow
					case DirectionColumnReverse:
						columns = innerBlockLayout.Columns
						rows = growDivision * grow
					}
				}

				// Ensure rows and columns aren't negative
				// if rows < 0 {
				// 	rows = 0
				// }
				// if columns < 0 {
				// 	columns = 0
				// }

				// if props.MinHeight != 0 {
				// 	rows = intmath.Min(rows, props.MinHeight)
				// }
				// if props.MinWidth != 0 {
				// 	columns = intmath.Min(columns, props.MinWidth)
				// }

				blockLayout := r.BlockLayout{
					X:       x,
					Y:       y,
					Rows:    rows,
					Columns: columns,
					ZIndex:  c.BlockLayout.ZIndex,
					Order:   i,
					Valid:   true,
				}

				switch props.Direction {
				case DirectionRow:
					fallthrough
				case DirectionRowReverse:
					colsRemaining = colsRemaining - columns
					x = x + columns - 1
				case DirectionColumn:
					fallthrough
				case DirectionColumnReverse:
					rowsRemaining = rowsRemaining - rows
					y = y + rows - 1
				}

				childrenBlockLayouts[i].BlockLayout = blockLayout

			}

			// expand any possible boxes to fill the remaining space
			xOffset := 0
			yOffset := 0

			xIncrease := 0
			xRemainder := 0
			if len(children) != 0 && colsRemaining != 0 {
				xIncrease = len(children) / colsRemaining
				xRemainder = colsRemaining - (xIncrease * len(children)) + 2
			}

			yIncrease := 0
			yRemainder := 0
			if len(children) != 0 && rowsRemaining != 0 {
				yIncrease = rowsRemaining / len(children)
				yRemainder = rowsRemaining - (yIncrease * len(children)) + 2
			}

			for i := range childrenBlockLayouts {

				if i == len(childrenBlockLayouts)-1 {
					xIncrease = xIncrease + xRemainder
					yIncrease = yIncrease + yRemainder
				}

				switch props.Direction {
				case DirectionRow:
					fallthrough
				case DirectionRowReverse:
					childrenBlockLayouts[i].BlockLayout.X =
						childrenBlockLayouts[i].BlockLayout.X + xOffset
					childrenBlockLayouts[i].BlockLayout.Columns =
						childrenBlockLayouts[i].BlockLayout.Columns + xIncrease
					xOffset = xOffset + xIncrease
				case DirectionColumn:
					fallthrough
				case DirectionColumnReverse:
					childrenBlockLayouts[i].BlockLayout.Y =
						childrenBlockLayouts[i].BlockLayout.Y + yOffset
					childrenBlockLayouts[i].BlockLayout.Rows =
						childrenBlockLayouts[i].BlockLayout.Rows + yIncrease
					yOffset = yOffset + yIncrease
				}
			}

			// If we reversed them, reverse them back
			if props.Direction == DirectionRowReverse ||
				props.Direction == DirectionColumnReverse {
				for i := len(children)/2 - 1; i >= 0; i-- {
					opp := len(children) - 1 - i
					children[i], children[opp] = children[opp], children[i]
				}
			}

		case r.CalculateLayoutStageFinal:
		}

		return
	}
}
