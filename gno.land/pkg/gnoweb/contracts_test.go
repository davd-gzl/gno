package gnoweb_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPHandler_LatestPackagesView(t *testing.T) {
	t.Parallel()

	// Create mock packages
	mockPkgs := []*gnoweb.MockPackage{
		{
			Path:   "/r/demo/users",
			Domain: "gno.land",
			Files:  map[string]string{"users.gno": "package users"},
		},
		{
			Path:   "/r/demo/boards",
			Domain: "gno.land",
			Files:  map[string]string{"boards.gno": "package boards"},
		},
		{
			Path:   "/p/demo/avl",
			Domain: "gno.land",
			Files:  map[string]string{"avl.gno": "package avl"},
		},
	}

	mockClient := gnoweb.NewMockClient(mockPkgs...)

	cfg := &gnoweb.HTTPHandlerConfig{
		ClientAdapter: mockClient,
		Renderer:      &rawRenderer{},
		Aliases: map[string]gnoweb.AliasTarget{
			"/latest-packages": {Value: "", Kind: gnoweb.LatestPackagesList},
		},
		Meta: gnoweb.StaticMetadata{
			Domain: "gno.land",
		},
	}

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/latest-packages", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()

	// Check that the page contains expected content
	assert.True(t, strings.Contains(body, "Latest Packages"), "Page should have Latest Packages title")
}

func TestHTTPHandler_LatestPackagesViewPagination(t *testing.T) {
	t.Parallel()

	// Create mock packages (more than one page)
	mockPkgs := []*gnoweb.MockPackage{}
	for i := 0; i < 60; i++ {
		mockPkgs = append(mockPkgs, &gnoweb.MockPackage{
			Path:   "/r/demo/realm" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
			Domain: "gno.land",
			Files:  map[string]string{"main.gno": "package main"},
		})
	}

	mockClient := gnoweb.NewMockClient(mockPkgs...)

	cfg := &gnoweb.HTTPHandlerConfig{
		ClientAdapter: mockClient,
		Renderer:      &rawRenderer{},
		Aliases: map[string]gnoweb.AliasTarget{
			"/latest-packages": {Value: "", Kind: gnoweb.LatestPackagesList},
		},
		Meta: gnoweb.StaticMetadata{
			Domain: "gno.land",
		},
	}

	handler, err := gnoweb.NewHTTPHandler(
		slog.New(slog.NewTextHandler(&testingLogger{t}, nil)),
		cfg,
	)
	require.NoError(t, err)

	// Test page 1
	req := httptest.NewRequest(http.MethodGet, "/latest-packages", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.True(t, strings.Contains(body, "Page 1"), "Should show page 1")
	assert.True(t, strings.Contains(body, "?page=2"), "Should have next page link")

	// Test page 2
	req = httptest.NewRequest(http.MethodGet, "/latest-packages?page=2", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body = rr.Body.String()
	assert.True(t, strings.Contains(body, "Page 2"), "Should show page 2")
	assert.True(t, strings.Contains(body, "?page=1"), "Should have previous page link")
}
