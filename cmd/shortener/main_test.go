package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncodeUsesBase62Alphabet(t *testing.T) {
	tests := []struct {
		name string
		num  uint64
		want string
	}{
		{
			name: "zero",
			num:  0,
			want: "0",
		},
		{
			name: "last digit",
			num:  61,
			want: "Z",
		},
		{
			name: "first two-digit value",
			num:  62,
			want: "10",
		},
		{
			name: "next two-digit value",
			num:  63,
			want: "11",
		},
		{
			name: "first three-digit value",
			num:  3844,
			want: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.num)
			if got != tt.want {
				t.Fatalf("Encode(%d) = %q, want %q", tt.num, got, tt.want)
			}
		})
	}
}

func TestCreateShortURL(t *testing.T) {
	resetTestStore()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
	rec := httptest.NewRecorder()

	newRouter().ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "text/plain" {
		t.Fatalf("expected Content-Type text/plain, got %q", contentType)
	}

	body := rec.Body.String()
	if !strings.HasPrefix(body, "http://localhost:8080/") {
		t.Fatalf("expected short URL with localhost prefix, got %q", body)
	}

	id := strings.TrimPrefix(body, "http://localhost:8080/")
	if id == "" {
		t.Fatal("expected non-empty short URL id")
	}
}

func TestRedirectToOriginalURL(t *testing.T) {
	resetTestStore()

	originalURL := "https://practicum.yandex.ru/"
	createReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	createRec := httptest.NewRecorder()
	router := newRouter()

	router.ServeHTTP(createRec, createReq)

	id := strings.TrimPrefix(createRec.Body.String(), shortURLPrefix)
	redirectReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	redirectRec := httptest.NewRecorder()

	router.ServeHTTP(redirectRec, redirectReq)

	res := redirectRec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, res.StatusCode)
	}

	location := res.Header.Get("Location")
	if location != originalURL {
		t.Fatalf("expected Location %q, got %q", originalURL, location)
	}
}

func TestBadRequests(t *testing.T) {
	resetTestStore()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "GET root",
			method: http.MethodGet,
			path:   "/",
		},
		{
			name:   "POST non-root path",
			method: http.MethodPost,
			path:   "/unknown",
			body:   "https://practicum.yandex.ru/",
		},
		{
			name:   "PUT root",
			method: http.MethodPut,
			path:   "/",
			body:   "https://practicum.yandex.ru/",
		},
		{
			name:   "GET unknown id",
			method: http.MethodGet,
			path:   "/unknown",
		},
		{
			name:   "POST empty body",
			method: http.MethodPost,
			path:   "/",
		},
	}

	router := newRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
			}
		})
	}
}

func resetTestStore() {
	mu.Lock()
	defer mu.Unlock()

	nextID = 0
	urlStore = map[string]string{}
}
