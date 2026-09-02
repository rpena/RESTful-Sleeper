package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewHandlerServesHTMLWithoutRedirect(t *testing.T) {
	server := httptest.NewServer(NewHandler(http.NotFoundHandler()))
	defer server.Close()

	for _, path := range []string{"/", "/index.html"} {
		response, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, response.StatusCode, http.StatusOK)
		}
		if !strings.Contains(response.Header.Get("Content-Type"), "text/html") {
			t.Fatalf("%s content type = %q, want HTML", path, response.Header.Get("Content-Type"))
		}
		if !strings.Contains(string(body), "Gameday board") {
			t.Fatalf("%s body does not contain dashboard HTML", path)
		}
	}
}
