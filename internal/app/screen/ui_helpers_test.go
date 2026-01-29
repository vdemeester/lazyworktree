package screen

import "testing"

type statusIconProvider struct {
	clean string
	dirty string
}

func (p *statusIconProvider) GetPRIcon() string {
	return ""
}

func (p *statusIconProvider) GetIssueIcon() string {
	return ""
}

func (p *statusIconProvider) GetCIIcon(conclusion string) string {
	return ""
}

func (p *statusIconProvider) GetUIIcon(icon UIIcon) string {
	switch icon {
	case UIIconStatusClean:
		return p.clean
	case UIIconStatusDirty:
		return p.dirty
	default:
		return ""
	}
}

func TestStatusIndicatorUsesIconProvider(t *testing.T) {
	prev := currentIconProvider
	t.Cleanup(func() { SetIconProvider(prev) })

	SetIconProvider(&statusIconProvider{clean: "C", dirty: "D"})

	if got := statusIndicator(true, true); got != "C" {
		t.Fatalf("expected clean icon, got %q", got)
	}
	if got := statusIndicator(false, true); got != "D" {
		t.Fatalf("expected dirty icon, got %q", got)
	}
	if got := statusIndicator(true, false); got != " " {
		t.Fatalf("expected clean fallback, got %q", got)
	}
	if got := statusIndicator(false, false); got != "~" {
		t.Fatalf("expected dirty fallback, got %q", got)
	}
}
