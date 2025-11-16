//go:build withgtk4

package gtk4

import (
	"fmt"
	"marmalade/camera"
	"slices"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func create_camera_info_window(camera_id uint8) error {
	camera, err := camera.GetDetailsForDevice(camera_id)
	if err != nil {
		return err
	}

	window := gtk.NewWindow()
	titlebar := gtk.NewHeaderBar()
	window.SetTitlebar(titlebar)

	window.SetTitle("Marmalade - " + camera.Name)
	window.SetDefaultSize(400, 550)
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
		create_line("Name:", format.Data.Description, grid, 1)
		create_line("Compressed:", compressed, grid, 2)
		create_line("Emulated:", emulated, grid, 3)

		create_resolution_list(&format, grid)

		scrollable_content := gtk.NewViewport(nil, nil)
		scrollable_content.SetChild(grid)

		scrollable_container := gtk.NewScrolledWindow()
		scrollable_container.SetChild(scrollable_content)

		stack.AddChild(scrollable_container)
		page := stack.Page(scrollable_container)
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

func create_resolution_list(format *camera.VideoFormat, grid *gtk.Grid) {
	line_index := 3

	var header_text string

	frameSizeType := format.Resolutions[0].Data.Type

	switch frameSizeType {
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

	if frameSizeType == v4l2.FrameSizeTypeDiscrete {

		line_index++
		name_value := gtk.NewLabel("Supported frame rates:")
		name_value.SetHAlign(gtk.AlignStart)
		name_value.SetHExpand(true)

		grid.Attach(name_value, 1, line_index, 1, 1)

		slices.SortFunc(format.Resolutions, func(a camera.VideoFormatResolution, b camera.VideoFormatResolution) int {
			if b.Data.Size.MaxWidth != a.Data.Size.MaxWidth {
				return int(b.Data.Size.MaxWidth) - int(a.Data.Size.MaxWidth)
			} else {
				return int(b.Data.Size.MaxHeight) - int(a.Data.Size.MaxHeight)
			}
		})

		for _, resolution := range format.Resolutions {
			line_index++
			label_text := fmt.Sprintf("%dx%d:", resolution.Data.Size.MaxWidth, resolution.Data.Size.MaxHeight)
			label := gtk.NewLabel(label_text)
			label.SetHAlign(gtk.AlignEnd)
			label.SetSelectable(true)
			grid.Attach(label, 0, line_index, 1, 1)

			create_frame_rate_line(&resolution, grid, line_index)
		}

	} else {

		resolution := format.Resolutions[0]

		line_index++
		minimum := fmt.Sprintf("%dx%d", resolution.Data.Size.MinWidth, resolution.Data.Size.MinHeight)
		create_line("Minimum:", minimum, grid, line_index)

		line_index++
		maximum := fmt.Sprintf("%dx%d", resolution.Data.Size.MaxWidth, resolution.Data.Size.MaxHeight)
		create_line("Maximum:", maximum, grid, line_index)

		if frameSizeType == v4l2.FrameSizeTypeStepwise {
			line_index++
			step_res := fmt.Sprintf("%dx%d", resolution.Data.Size.StepWidth, resolution.Data.Size.StepHeight)
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

func create_frame_rate_line(resolution *camera.VideoFormatResolution, grid *gtk.Grid, line_index int) {

	var label_text string

	frameRateType := resolution.FrameRates[0].Type

	switch frameRateType {
	case v4l2.FrameIntervalTypeDiscrete:
		frame_rates := make([]uint32, 0, len(resolution.FrameRates))

		for _, frame_fraction := range resolution.FrameRates {
			frame_rate := frac_to_int(frame_fraction.Interval.Max)
			frame_rates = append(frame_rates, frame_rate)
		}

		slices.SortFunc(frame_rates, func(a uint32, b uint32) int {
			return int(b) - int(a)
		})

		label_slice := make([]string, 0, len(resolution.FrameRates))

		for _, frame_rate := range frame_rates {
			label_slice = append(label_slice, strconv.FormatUint(uint64(frame_rate), 10))
		}

		label_text = strings.Join(label_slice, ", ")
	case v4l2.FrameIntervalTypeContinuous:
		frame_rate := resolution.FrameRates[0]
		label_text = fmt.Sprintf("Min: %d / Max: %d", frac_to_int(frame_rate.Interval.Min), frac_to_int(frame_rate.Interval.Max))
	case v4l2.FrameIntervalTypeStepwise:
		frame_rate := resolution.FrameRates[0]
		label_text = fmt.Sprintf("Min: %d / Max: %d / Step: %d", frac_to_int(frame_rate.Interval.Min), frac_to_int(frame_rate.Interval.Max), frac_to_int(frame_rate.Interval.Step))
	}

	label := gtk.NewLabel(label_text)
	label.SetHAlign(gtk.AlignStart)
	label.SetSelectable(true)
	grid.Attach(label, 1, line_index, 1, 1)
}

func frac_to_int(frac v4l2.Fract) uint32 {
	result := uint32(frac.Denominator) / uint32(frac.Numerator)
	return result
}
