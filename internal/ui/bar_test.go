package ui

import (
	"testing"

	"go.xchunk.org/beautify"

	"go.xchunk.org/fetchify/internal/config"
)

func TestBarCells(t *testing.T) {
	tests := []struct{ percent, width, want int }{
		{0, 16, 0}, {1, 16, 1}, {50, 16, 8}, {99, 16, 15}, {100, 16, 16},
		{3, 10, 1}, {96, 10, 9}, {50, 0, 0},
	}
	for _, tt := range tests {
		if got := barCells(tt.percent, tt.width); got != tt.want {
			t.Errorf("barCells(%d, %d) = %d, want %d", tt.percent, tt.width, got, tt.want)
		}
	}
}

func TestBarShrinks(t *testing.T) {
	b := bar{percent: 50, width: 16, style: config.BarBlock}
	for _, w := range []int{16, 10, 1, 0} {
		m, _ := b.Render(beautify.NewContext(nil).WithWidth(w))
		if m.Width() != w {
			t.Errorf("bound %d: width %d", w, m.Width())
		}
	}
	m, _ := bar{percent: 30, width: 4, style: config.BarThin}.Render(beautify.NewContext(nil))
	if m.String() != "━━━━" {
		t.Errorf("thin bar = %q", m.String())
	}
}

func TestLevel(t *testing.T) {
	th := [2]int{60, 85}
	tests := []struct {
		m    meter
		want role
	}{
		{meter{percent: 59}, roleGood}, {meter{percent: 60}, roleWarning}, {meter{percent: 85}, roleDanger},
		{meter{percent: 19, inverted: true}, roleDanger}, {meter{percent: 39, inverted: true}, roleWarning},
		{meter{percent: 40, inverted: true}, roleGood},
	}
	for _, tt := range tests {
		if got := level(tt.m, th); got != tt.want {
			t.Errorf("level(%+v) = %d, want %d", tt.m, got, tt.want)
		}
	}
}

func TestSizes(t *testing.T) {
	tests := []struct {
		used, total uint64
		want        meterText
	}{
		{512 * bytesMiB, 900 * bytesMiB, meterText{"512", "900", "MiB", "0%"}},
		{8 * bytesGiB, 32 * bytesGiB, meterText{"8.0", "32.0", "GiB", "0%"}},
		{250 * bytesGiB, 1000 * bytesGiB, meterText{"250", "1000", "GiB", "0%"}},
		{bytesTiB / 2, 2 * bytesTiB, meterText{"0.5", "2.0", "TiB", "0%"}},
	}
	for _, tt := range tests {
		got := formatMeter(meter{used: tt.used, total: tt.total, sizes: true})
		if got != tt.want {
			t.Errorf("formatMeter(%d, %d) = %+v, want %+v", tt.used, tt.total, got, tt.want)
		}
	}
	if got := formatAmount(99.96); got != "100" {
		t.Errorf("formatAmount(99.96) = %q", got)
	}
}

func TestMeterColumnsAlign(t *testing.T) {
	ctx := beautify.NewContext(nil)
	lay := layout{details: true, bars: true, barWidth: 4, barStyle: config.BarBlock, thresholds: [2]int{60, 85}}
	rows := []row{
		{label: "RAM", meter: &meter{percent: 25, used: 8 * bytesGiB, total: 32 * bytesGiB, sizes: true}},
		{label: "/home", meter: &meter{percent: 25, used: 250 * bytesGiB, total: 1000 * bytesGiB, sizes: true},
			value: []segment{{text: "btrfs", role: roleMuted, detail: true}}},
	}
	got := plainLines(renderRows(ctx, testTheme(), lay, rows, 5))
	want := "RAM    █░░░  8.0 / 32.0 GiB  25%\n/home  █░░░  250 / 1000 GiB  25%  btrfs"
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	lay.bars = false
	if got := plainLines(renderRows(ctx, testTheme(), lay, rows[:1], 3)); got != "RAM  8.0 / 32.0 GiB  25%" {
		t.Errorf("without bars: %q", got)
	}
}
