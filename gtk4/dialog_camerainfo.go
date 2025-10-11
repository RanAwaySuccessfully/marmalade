//go:build withgtk4

package gtk4

import (
	"fmt"
	"marmalade/v4l2"
	"slices"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func create_camera_info_window(camera_id uint8) error {
	camera, err := v4l2.GetDetailsForDevice(camera_id)
	if err != nil {
		return err
	}

	window := gtk.NewWindow()
	titlebar := gtk.NewHeaderBar()
	window.SetTitlebar(titlebar)

	window.SetTitle("Marmalade - " + camera.Name)
	window.SetDefaultSize(450, 300)
	window.SetResizable(false)
	window.SetVisible(true)

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	window.SetChild(box)

	content_box := gtk.NewBox(gtk.OrientationHorizontal, 10)

	stack := gtk.NewStack()

	for _, format := range camera.Formats {

		grid := gtk.NewGrid()
		grid.SetRowSpacing(7)
		grid.SetColumnSpacing(10)
		grid.SetMarginEnd(10)
		grid.SetMarginTop(10)
		grid.SetMarginBottom(10)

		compressed := "No"
		if format.Compressed {
			compressed = "Yes"
		}

		emulated := "No"
		if format.Emulated {
			emulated = "Yes"
		}

		create_line("ID:", format.Id, grid, 0)
		create_line("Name:", format.Name, grid, 1)
		create_line("Compressed:", compressed, grid, 2)
		create_line("Emulated:", emulated, grid, 3)

		create_resolution_list(&format, grid)

		stack.AddChild(grid)
		page := stack.Page(grid)
		page.SetTitle(format.Id)
	}

	sidebar := gtk.NewStackSidebar()
	sidebar.SetStack(stack)
	sidebar.SetVExpand(true)

	content_box.Append(sidebar)
	content_box.Append(stack)

	box.Append(content_box)

	action_bar := gtk.NewActionBar()
	box.Append(action_bar)

	button := gtk.NewButton()
	button.SetLabel("Close")

	action_bar.SetCenterWidget(button)

	button.Connect("clicked", func() {
		window.Destroy()
	})

	return nil
}

func create_resolution_list(format *v4l2.VideoFormat, grid *gtk.Grid) {
	line_index := 3

	var header_text string

	switch format.ResolutionType {
	case v4l2.FrameSizeTypeDiscrete:
		header_text = "Discrete resolutions"
	case v4l2.FrameSizeTypeContinuous:
		header_text = "Continouous resolutions"
	case v4l2.FrameSizeTypeStepwise:
		header_text = "Stepwise resolutions"
	}

	line_index++
	sep_1 := gtk.NewSeparator(gtk.OrientationHorizontal)
	grid.Attach(sep_1, 0, line_index, 2, 1)

	line_index++
	header := gtk.NewLabel(header_text)
	grid.Attach(header, 0, line_index, 2, 1)

	line_index++
	sep_2 := gtk.NewSeparator(gtk.OrientationHorizontal)
	grid.Attach(sep_2, 0, line_index, 2, 1)

	if format.ResolutionType == v4l2.FrameSizeTypeDiscrete {

		line_index++
		name_value := gtk.NewLabel("Supported frame rates:")
		name_value.SetHAlign(gtk.AlignStart)
		name_value.SetHExpand(true)

		grid.Attach(name_value, 1, line_index, 1, 1)

		slices.SortFunc(format.Resolutions, func(a v4l2.VideoFormatResolution, b v4l2.VideoFormatResolution) int {
			if b.Width != a.Width {
				return int(b.Width) - int(a.Width)
			} else {
				return int(b.Height) - int(a.Height)
			}
		})

		for _, resolution := range format.Resolutions {
			line_index++
			label_text := fmt.Sprintf("%dx%d:", resolution.Width, resolution.Height)
			label := gtk.NewLabel(label_text)
			label.SetHAlign(gtk.AlignEnd)
			label.SetSelectable(true)
			grid.Attach(label, 0, line_index, 1, 1)

			create_frame_rate_line(&resolution, grid, line_index)
		}

	} else {

		resolution := format.Resolutions[0]

		line_index++
		minimum := fmt.Sprintf("%dx%d", resolution.RangeWidth[0], resolution.RangeHeight[1])
		create_line("Minimum:", minimum, grid, line_index)

		line_index++
		maximum := fmt.Sprintf("%dx%d", resolution.RangeWidth[0], resolution.RangeHeight[1])
		create_line("Maximum:", maximum, grid, line_index)

		if format.ResolutionType == v4l2.FrameSizeTypeStepwise {
			line_index++
			step_res := fmt.Sprintf("%dx%d", resolution.Width, resolution.Height)
			create_line("Step:", step_res, grid, line_index)
		}

		line_index++
		fps_label := gtk.NewLabel("Frame rates:")
		fps_label.SetHAlign(gtk.AlignEnd)
		fps_label.SetSelectable(true)
		grid.Attach(fps_label, 0, line_index, 1, 1)

		create_frame_rate_line(&resolution, grid, line_index)
	}

}

func create_line(label_text string, value_text string, grid *gtk.Grid, line_index int) {
	label := gtk.NewLabel(label_text)
	label.SetHAlign(gtk.AlignEnd)

	value := gtk.NewLabel(value_text)
	value.SetHAlign(gtk.AlignStart)
	value.SetHExpand(true)
	value.SetSelectable(true)

	grid.Attach(label, 0, line_index, 1, 1)
	grid.Attach(value, 1, line_index, 1, 1)
}

func create_frame_rate_line(resolution *v4l2.VideoFormatResolution, grid *gtk.Grid, line_index int) {

	var label_text string

	switch resolution.FrameRateType {
	case v4l2.FrameIntervalTypeDiscrete:
		frame_rates := make([]string, 0, len(resolution.FrameRates))

		slices.SortFunc(resolution.FrameRates, func(a uint32, b uint32) int {
			return int(b) - int(a)
		})

		for _, frame_rate := range resolution.FrameRates {
			frame_rates = append(frame_rates, strconv.FormatUint(uint64(frame_rate), 10))
		}

		label_text = strings.Join(frame_rates, ", ")
	case v4l2.FrameIntervalTypeContinuous:
		label_text = fmt.Sprintf("Min: %d / Max: %d", resolution.FrameRates[0], resolution.FrameRates[1])
	case v4l2.FrameIntervalTypeStepwise:
		label_text = fmt.Sprintf("Min: %d / Max: %d / Step: %d", resolution.FrameRates[0], resolution.FrameRates[1], resolution.FrameRates[2])
	}

	label := gtk.NewLabel(label_text)
	label.SetHAlign(gtk.AlignStart)
	label.SetSelectable(true)
	grid.Attach(label, 1, line_index, 1, 1)
}
