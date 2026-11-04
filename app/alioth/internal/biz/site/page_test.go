package site

import (
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestNormalizeLocale(t *testing.T) {
	cases := map[string]string{
		"":     "zh",
		"zh":   "zh",
		"EN":   "en",
		" en ": "en",
		"fr":   "zh",
	}
	for input, want := range cases {
		if got := normalizeLocale(input); got != want {
			t.Fatalf("normalizeLocale(%q) = %q, want %q", input, got, want)
		}
	}
	if otherLocale("zh") != "en" || otherLocale("en") != "zh" {
		t.Fatal("otherLocale should flip between zh and en")
	}
}

// 从未发布过的页面不算「有未发布的改动」：列表上只显示状态徽标，
// 没发布过就没有可比较的基线。
func TestSitePageHasUnpublishedChanges(t *testing.T) {
	page := bo.SitePage{Draft: bo.SitePageContent{Title: "草稿"}}
	if page.HasUnpublishedChanges() {
		t.Fatal("page that was never published must not report unpublished changes")
	}
	publishedAt := time.Now()
	page.PublishedAt = &publishedAt
	page.Published = bo.SitePageContent{Title: "线上"}
	if !page.HasUnpublishedChanges() {
		t.Fatal("draft different from the published snapshot should report changes")
	}
	page.Published = page.Draft
	if page.HasUnpublishedChanges() {
		t.Fatal("identical draft and snapshot should not report changes")
	}
}

// 标题与摘要会直接显示在官网页面，前后空格必须裁剪；正文不裁剪，Markdown 空行有意义。
func TestNormalizePageContent(t *testing.T) {
	in := bo.SitePageContent{
		Title:          "  产品说明 ",
		Summary:        " 一句话摘要 ",
		BodyMD:         "\n\n# 标题\n\n",
		CoverImageURL:  " https://cdn.example.com/a.png ",
		SEOTitle:       " 标题 ",
		SEODescription: " 描述 ",
	}
	got := normalizePageContent(in)
	if got.Title != "产品说明" || got.Summary != "一句话摘要" {
		t.Fatalf("title/summary not trimmed: %+v", got)
	}
	if got.BodyMD != "\n\n# 标题\n\n" {
		t.Fatalf("body must keep whitespace: %q", got.BodyMD)
	}
	if got.CoverImageURL != "https://cdn.example.com/a.png" {
		t.Fatalf("cover url not trimmed: %q", got.CoverImageURL)
	}
}

// 封面图最终会成为公开页面上的 <img src>，只接受 http(s)。
func TestValidatePageContentRejectsUnsafeCover(t *testing.T) {
	for _, value := range []string{"javascript:alert(1)", "data:image/png;base64,AAA", "not-a-url"} {
		if err := validatePageContent(bo.SitePageContent{CoverImageURL: value}); err == nil {
			t.Fatalf("cover url %q should be rejected", value)
		}
	}
	if err := validatePageContent(bo.SitePageContent{CoverImageURL: "https://cdn.example.com/a.png"}); err != nil {
		t.Fatalf("https cover should be accepted: %v", err)
	}
	if err := validatePageContent(bo.SitePageContent{}); err != nil {
		t.Fatalf("empty cover should be accepted: %v", err)
	}
}

// 公开页面只读发布快照：草稿里的内容不能出现在公开接口的响应里。
func TestPublicSitePageOfReadsPublishedOnly(t *testing.T) {
	publishedAt := time.Now()
	page := bo.SitePage{
		Slug:   "product",
		Locale: "zh",
		Draft:  bo.SitePageContent{Title: "草稿标题"},
		Published: bo.SitePageContent{
			Title:  "线上标题",
			BodyMD: "线上正文",
		},
		PublishedAt: &publishedAt,
	}
	got := publicSitePageOf(page, "en", true)
	if got.Title != "线上标题" || got.BodyMD != "线上正文" {
		t.Fatalf("public page must read the published snapshot: %+v", got)
	}
	if got.Locale != "en" || !got.LocaleFallback {
		t.Fatalf("locale fallback not reported: %+v", got)
	}
	if got.PublishedAt.IsZero() {
		t.Fatal("published time should be copied")
	}
}
