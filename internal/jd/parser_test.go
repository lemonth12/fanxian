package jd

import "testing"

func TestParseURL_PC(t *testing.T) {
	sku, err := ParseURL("https://item.jd.com/123456.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sku != "123456" {
		t.Errorf("sku = %s, want 123456", sku)
	}
}

func TestParseURL_Mobile(t *testing.T) {
	sku, err := ParseURL("https://item.m.jd.com/product/654321.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sku != "654321" {
		t.Errorf("sku = %s, want 654321", sku)
	}
}

func TestParseURL_WithParams(t *testing.T) {
	sku, err := ParseURL("https://item.jd.com/123456.html?cu=true&utm_source=kong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sku != "123456" {
		t.Errorf("sku = %s, want 123456", sku)
	}
}

func TestParseURL_Invalid(t *testing.T) {
	_, err := ParseURL("https://taobao.com/item/123.html")
	if err == nil {
		t.Error("expected error for non-jd URL")
	}
}

func TestParseURL_Empty(t *testing.T) {
	_, err := ParseURL("")
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestParseURL_ShortLink(t *testing.T) {
	_, err := ParseURL("https://3.cn/abc123")
	if err == nil {
		t.Skip("short link resolution requires network")
	}
}
