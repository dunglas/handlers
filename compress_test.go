// Copyright 2013 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handlers

import (
	"compress/flate"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var defaultContentType = "text/plain; charset=utf-8"

func compressedRequest(w *httptest.ResponseRecorder, compression string, level int, contentType string) {
	CompressHandlerLevel(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(9*1024))
		w.Header().Set("Content-Type", contentType)
		for i := 0; i < 1024; i++ {
			io.WriteString(w, "Gorilla!\n")
		}
	}), level).ServeHTTP(w, &http.Request{
		Method: "GET",
		Header: http.Header{
			"Accept-Encoding": []string{compression},
		},
	})

}

func TestCompressHandlerNoCompression(t *testing.T) {
	w := httptest.NewRecorder()
	compressedRequest(w, "", flate.DefaultCompression, defaultContentType)
	if enc := w.HeaderMap.Get("Content-Encoding"); enc != "" {
		t.Errorf("wrong content encoding, got %q want %q", enc, "")
	}
	if ct := w.HeaderMap.Get("Content-Type"); ct != defaultContentType {
		t.Errorf("wrong content type, got %q want %q", ct, defaultContentType)
	}
	if w.Body.Len() != 1024*9 {
		t.Errorf("wrong len, got %d want %d", w.Body.Len(), 1024*9)
	}
	if l := w.HeaderMap.Get("Content-Length"); l != "9216" {
		t.Errorf("wrong content-length. got %q expected %d", l, 1024*9)
	}
}

func TestCompressHandlerGzip(t *testing.T) {
	w := httptest.NewRecorder()
	compressedRequest(w, "gzip", flate.DefaultCompression, defaultContentType)
	if w.HeaderMap.Get("Content-Encoding") != "gzip" {
		t.Errorf("wrong content encoding, got %q want %q", w.HeaderMap.Get("Content-Encoding"), "gzip")
	}
	if vary := w.HeaderMap.Get("Vary"); vary != "Accept-Encoding" {
		t.Errorf("wrong content type, got %s want %s", vary, "Accept-Encoding")
	}
	if w.HeaderMap.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("wrong content type, got %s want %s", w.HeaderMap.Get("Content-Type"), "text/plain; charset=utf-8")
	}
	if w.Body.Len() != 72 {
		t.Errorf("wrong len, got %d want %d", w.Body.Len(), 72)
	}
	if l := w.HeaderMap.Get("Content-Length"); l != "" {
		t.Errorf("wrong content-length. got %q expected %q", l, "")
	}
}

func TestCompressHandlerDeflate(t *testing.T) {
	w := httptest.NewRecorder()
	compressedRequest(w, "deflate", flate.DefaultCompression, defaultContentType)
	if w.HeaderMap.Get("Content-Encoding") != "deflate" {
		t.Fatalf("wrong content encoding, got %q want %q", w.HeaderMap.Get("Content-Encoding"), "deflate")
	}
	if vary := w.HeaderMap.Get("Vary"); vary != "Accept-Encoding" {
		t.Errorf("wrong content type, got %s want %s", vary, "Accept-Encoding")
	}
	if w.HeaderMap.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("wrong content type, got %s want %s", w.HeaderMap.Get("Content-Type"), "text/plain; charset=utf-8")
	}
	if w.Body.Len() != 54 {
		t.Fatalf("wrong len, got %d want %d", w.Body.Len(), 54)
	}
}

func TestCompressHandlerGzipDeflate(t *testing.T) {
	w := httptest.NewRecorder()
	compressedRequest(w, "gzip, deflate ", flate.DefaultCompression, defaultContentType)
	if w.HeaderMap.Get("Content-Encoding") != "gzip" {
		t.Fatalf("wrong content encoding, got %q want %q", w.HeaderMap.Get("Content-Encoding"), "gzip")
	}
	if w.HeaderMap.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("wrong content type, got %s want %s", w.HeaderMap.Get("Content-Type"), "text/plain; charset=utf-8")
	}
}

func TestCompressHandlerTypes(t *testing.T) {
	w := httptest.NewRecorder()
	compressedRequest(w, "gzip", flate.DefaultCompression, "image/png")

	if ce := w.HeaderMap.Get("Content-Encoding"); ce == "gzip" {
		t.Fatalf("wrong content encoding: got %q want %q", ce, "dunno?")
	}
}

func BenchmarkCompressHandlerGzipDefaultCompression(b *testing.B) {
	b.StopTimer()
	w := httptest.NewRecorder()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		compressedRequest(w, "gzip", flate.DefaultCompression, defaultContentType)
	}
}

func BenchmarkCompressHandlerGzipBestCompression(b *testing.B) {
	b.StopTimer()
	w := httptest.NewRecorder()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		compressedRequest(w, "gzip", flate.BestCompression, defaultContentType)
	}
}

func BenchmarkCompressHandlerGzipBestSpeed(b *testing.B) {
	b.StopTimer()
	w := httptest.NewRecorder()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		compressedRequest(w, "gzip", flate.BestSpeed, defaultContentType)
	}
}
