package board

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func validPanel() bo.BoardPanel {
	return bo.BoardPanel{
		Title:      "CPU 使用率",
		GrafanaURL: "https://grafana.example.com/d/abc123/cpu?panelId=2&from=now-6h",
		Width:      6,
		Height:     320,
	}
}

func TestNormalizeBoardTrimsAndSortsPanels(t *testing.T) {
	in := bo.Board{
		Name:        "  值班大盘  ",
		Description: "  核心服务  ",
		Panels: []bo.BoardPanel{
			validPanel(),
			{Title: " 延迟 ", GrafanaURL: "http://grafana.local/d-solo/x/y?panelId=1", Width: 12, Height: 200},
		},
	}
	if err := normalizeBoard(&in); err != nil {
		t.Fatalf("valid board rejected: %v", err)
	}
	if in.Name != "值班大盘" || in.Description != "核心服务" {
		t.Fatalf("name/description not trimmed: %q %q", in.Name, in.Description)
	}
	if in.Panels[1].Title != "延迟" {
		t.Fatalf("panel title not trimmed: %q", in.Panels[1].Title)
	}
	for index, panel := range in.Panels {
		if panel.Sort != index {
			t.Fatalf("panel %d sort = %d, want %d", index, panel.Sort, index)
		}
	}
}

func TestNormalizeBoardRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		in   bo.Board
	}{
		{name: "empty name", in: bo.Board{Name: "  "}},
		{name: "long name", in: bo.Board{Name: strings.Repeat("a", 65)}},
		{name: "too many panels", in: bo.Board{Name: "board", Panels: manyPanels(51)}},
		{name: "empty panel title", in: bo.Board{Name: "board", Panels: []bo.BoardPanel{{GrafanaURL: "https://g/d/x/y", Width: 6, Height: 320}}}},
		{name: "width zero", in: bo.Board{Name: "board", Panels: []bo.BoardPanel{{Title: "p", GrafanaURL: "https://g/d/x/y", Width: 0, Height: 320}}}},
		{name: "width too large", in: bo.Board{Name: "board", Panels: []bo.BoardPanel{{Title: "p", GrafanaURL: "https://g/d/x/y", Width: 13, Height: 320}}}},
		{name: "height too small", in: bo.Board{Name: "board", Panels: []bo.BoardPanel{{Title: "p", GrafanaURL: "https://g/d/x/y", Width: 6, Height: 119}}}},
		{name: "height too large", in: bo.Board{Name: "board", Panels: []bo.BoardPanel{{Title: "p", GrafanaURL: "https://g/d/x/y", Width: 6, Height: 1601}}}},
	}
	for _, item := range cases {
		if err := normalizeBoard(&item.in); err == nil {
			t.Fatalf("%s: expected error", item.name)
		}
	}
}

func TestValidateGrafanaURL(t *testing.T) {
	valid := []string{
		"https://grafana.example.com/d/abc/dashboard",
		"https://grafana.example.com/grafana/d-solo/abc/panel?panelId=1",
		"http://localhost:3000/public-dashboards/abc",
	}
	for _, raw := range valid {
		if err := validateGrafanaURL(raw); err != nil {
			t.Fatalf("valid url %q rejected: %v", raw, err)
		}
	}
	invalid := []string{
		"ftp://grafana.example.com/d/abc",
		"javascript:alert(1)",
		"https://user:pass@grafana.example.com/d/abc",
		"https://grafana.example.com/d/abc#frag",
		"https://grafana.example.com/not-a-dashboard",
		"https://grafana.example.com/d/abc?api_key=secret",
		"https://grafana.example.com/d/abc?token=abc",
	}
	for _, raw := range invalid {
		if err := validateGrafanaURL(raw); err == nil {
			t.Fatalf("invalid url %q accepted", raw)
		}
	}
}

func manyPanels(count int) []bo.BoardPanel {
	panels := make([]bo.BoardPanel, 0, count)
	for i := 0; i < count; i++ {
		panel := validPanel()
		panels = append(panels, panel)
	}
	return panels
}
