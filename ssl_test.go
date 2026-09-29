package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseCertEndTime(t *testing.T) {
	got, ok := parseCertEndTime("2026-10-01 12:00:00")
	if !ok {
		t.Fatal("expected parse ok")
	}
	if got.Year() != 2026 || got.Month() != time.October || got.Day() != 1 {
		t.Fatalf("unexpected time: %v", got)
	}
	if _, ok := parseCertEndTime(""); ok {
		t.Fatal("empty should fail")
	}
}

func TestCollectBoundTypes(t *testing.T) {
	types := collectBoundTypes([]bindResourceTypeResult{
		{
			ResourceType: "clb",
			BindResourceRegionResult: []struct {
				Region     string `json:"Region"`
				TotalCount uint64 `json:"TotalCount"`
				Error      string `json:"Error"`
			}{
				{Region: "ap-guangzhou", TotalCount: 2},
			},
		},
		{
			ResourceType: "cdn",
			BindResourceRegionResult: []struct {
				Region     string `json:"Region"`
				TotalCount uint64 `json:"TotalCount"`
				Error      string `json:"Error"`
			}{
				{Region: "ap-guangzhou", TotalCount: 0},
			},
		},
	})
	if len(types) != 1 || types[0] != "clb" {
		t.Fatalf("got %v", types)
	}
}

func TestSSLConfigDefaults(t *testing.T) {
	cfg := SSLConfig{}
	if !cfg.autoDelete() {
		t.Fatal("autoDelete default true")
	}
	if !cfg.dryRun() {
		t.Fatal("dryRun default true")
	}
	if cfg.checkInterval() != 86400 {
		t.Fatalf("interval=%d", cfg.checkInterval())
	}
	if !cfg.bindUseCache() {
		t.Fatal("bindUseCache default true")
	}
	f := false
	cfg.DryRun = &f
	if cfg.dryRun() {
		t.Fatal("dryRun should honor false")
	}
}

func TestReasonAndDisplay(t *testing.T) {
	if reasonLabel("expired") != "已过期" {
		t.Fatal(reasonLabel("expired"))
	}
	cert := SSLCertificate{CertificateID: "id1", Domain: "a.com"}
	if certDisplayName(cert) != "a.com" {
		t.Fatal(certDisplayName(cert))
	}
	cert.Domain = ""
	cert.Alias = "alias"
	if certDisplayName(cert) != "alias" {
		t.Fatal(certDisplayName(cert))
	}
}

func TestToSSLCertificate(t *testing.T) {
	c := toSSLCertificate(sslCertListItem{
		CertificateID: "cid",
		Domain:        "d.com",
		Status:        3,
		CertEndTime:   "2020-01-01 00:00:00",
		IsExpiring:    true,
	})
	if c.CertificateID != "cid" || c.Status != 3 || !c.IsExpiring {
		t.Fatalf("%+v", c)
	}
}

func TestFormatSSLAccountNotify(t *testing.T) {
	if formatSSLAccountNotify("a", nil, nil, nil) != "" {
		t.Fatal("empty should be empty")
	}
	msg := formatSSLAccountNotify("acc",
		[]string{"证书[c1/x] 即将过期，仍关联[teo]，跳过删除"},
		[]string{"证书[c2/y] 已过期，无关联，已删除"},
		[]string{"证书[c3/z] 删除失败：boom"},
	)
	for _, want := range []string{
		"[SSL] 账号[acc]",
		"有资源跳过删除：",
		"成功删除：",
		"失败删除：",
		"- 证书[c1/x]",
		"- 证书[c2/y]",
		"- 证书[c3/z]",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("missing %q in:\n%s", want, msg)
		}
	}
	// empty sections omitted
	msg2 := formatSSLAccountNotify("acc", nil, []string{"only"}, nil)
	if strings.Contains(msg2, "有资源跳过删除") || strings.Contains(msg2, "失败删除") {
		t.Fatal(msg2)
	}
	if !strings.Contains(msg2, "成功删除：") {
		t.Fatal(msg2)
	}
}

func TestBindQueryBatchSize(t *testing.T) {
	if bindQueryBatchSize != 100 {
		t.Fatalf("batch size=%d", bindQueryBatchSize)
	}
}
