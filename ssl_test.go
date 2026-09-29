package main

import (
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
	types := collectBoundTypes([]struct {
		ResourceType             string `json:"ResourceType"`
		BindResourceRegionResult []struct {
			Region     string `json:"Region"`
			TotalCount uint64 `json:"TotalCount"`
			Error      string `json:"Error"`
		} `json:"BindResourceRegionResult"`
	}{
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
