package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app := initApp()
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 200, resp.StatusCode)
}

func TestIsValidUrl(t *testing.T) {
	validUrls := []string{
		"http://www.example.com",
		"http://example.com",
		"https://example.com",
		"https://example.co.uk",
		"https://example.de",
		"https://subdomain.example.com",
		"http://www.example.com/withpath",
		"http://www.example.com?withQueryParam=true",
	}

	invalidUrls := []string{
		"htp://www.example.com",
		"http//www.example.com",
		"http:/www.example.com",
		"http://www..example.com",
		"djofisjdf",
		"35235",
		"!",
	}

	for _, url := range validUrls {
		if res := isValidUrl(url); res != true {
			t.Errorf("TestIsValidUrl(%s) = %t, expected %t", url, res, true)
		}
	}

	for _, url := range invalidUrls {
		if res := isValidUrl(url); res != false {
			t.Errorf("TestIsValidUrl(%s) = %t, expected %t", url, res, false)
		}
	}
}

func TestGenerateShortCode(t *testing.T) {
	db := ConnectToDb(testDbPath)
	code := generateUniqueShortCode(db)
	codeLength := len(code)

	if codeLength != shortCodeLength {
		t.Errorf("TestGenerateShortCode() = length %d, expected length %d", codeLength, shortCodeLength)
	}

	for _, ch := range code {
		if !strings.Contains(shortCodeCharacters, string(ch)) {
			t.Errorf("TestGenerateShortCode() = character %s found, is not in allowed shortCodeCharacters (%s)", string(ch), code)
		}
	}
}
