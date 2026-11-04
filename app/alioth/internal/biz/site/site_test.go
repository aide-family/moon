package site

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// 备案字段留空是合法状态：官网页脚只是不展示那一行，
// 不能因为没有真实备案号就拒绝保存，也不能退化成占位假号。
func TestValidateSiteSettingsAcceptsEmptyFilings(t *testing.T) {
	if err := validateSiteSettings(bo.SiteSettings{SiteName: "Alioth"}); err != nil {
		t.Fatalf("empty filings should be accepted: %v", err)
	}
}

// 备案查询地址只接受 http(s)：`javascript:` 之类的 scheme 一旦落库，
// 就会变成官网页脚上的可点链接。
func TestValidateSiteSettingsRejectsUnsafeURL(t *testing.T) {
	for _, value := range []string{
		"javascript:alert(1)",
		"ftp://example.com",
		"not-a-url",
		"//example.com",
	} {
		in := bo.SiteSettings{FilingICPNumber: "京ICP备12345678号-1", FilingICPURL: value}
		if err := validateSiteSettings(in); err == nil {
			t.Fatalf("filing icp url %q should be rejected", value)
		}
	}
}

func TestValidateSiteSettingsAcceptsHTTPS(t *testing.T) {
	in := bo.SiteSettings{
		FilingICPNumber: "京ICP备12345678号-1",
		FilingICPURL:    "https://beian.miit.gov.cn/",
		ContactEmail:    "contact@alioth.asia",
	}
	if err := validateSiteSettings(in); err != nil {
		t.Fatalf("valid settings should be accepted: %v", err)
	}
}

func TestValidateSiteSettingsRejectsBadEmail(t *testing.T) {
	if err := validateSiteSettings(bo.SiteSettings{ContactEmail: "not-an-email"}); err == nil {
		t.Fatal("invalid contact email should be rejected")
	}
}

// 备案号前后的空白会原样带进页脚，保存前必须裁剪。
func TestNormalizeSiteSettingsTrims(t *testing.T) {
	in := normalizeSiteSettings(bo.SiteSettings{
		FilingICPNumber: "  京ICP备12345678号-1  ",
		FilingExtraText: " 备案主体 ",
		ContactEmail:    " contact@alioth.asia ",
	})
	if in.FilingICPNumber != "京ICP备12345678号-1" {
		t.Fatalf("filing icp number = %q", in.FilingICPNumber)
	}
	if in.FilingExtraText != "备案主体" || in.ContactEmail != "contact@alioth.asia" {
		t.Fatalf("settings not trimmed: %+v", in)
	}
}
