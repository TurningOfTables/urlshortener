package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitApp(t *testing.T) {

	testApp := initApp(Config{Testing: true})
	assert.IsType(t, &fiber.App{}, testApp)

	prodApp := initApp(Config{Testing: false})
	assert.IsType(t, &fiber.App{}, prodApp)

	testAppResetDb := initApp(Config{Testing: true, Reset: true})
	assert.IsType(t, &fiber.App{}, testAppResetDb)
}

func TestGetIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app := initApp(Config{Testing: true})
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestShortenHandler(t *testing.T) {
	var sr = ShortenReq{LongUrl: "https://www.example.com"}
	sRes, resp, err := shortenUrl(sr)
	if err != nil {
		t.Errorf("TestShortenHandler() failed to shorten URL")
	}

	requiredSubstrings := []string{
		"http", port, shortUrlPath,
	}

	for _, string := range requiredSubstrings {
		assert.Contains(t, sRes.ShortUrl, string)
	}

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode, "TestShortenHandler() responded with unexpected status code")
}

func TestShortenHandlerInvalidUrl(t *testing.T) {
	var sr = ShortenReq{LongUrl: "foo"}
	_, resp, err := shortenUrl(sr)
	if err != nil {
		t.Errorf("TestShortenHandlerInvalidUrl() failed to shorten URL")
	}

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestFollowLinkHandler(t *testing.T) {
	var sr = ShortenReq{LongUrl: "https://www.example.com"}
	sRes, _, err := shortenUrl(sr)
	if err != nil {
		t.Errorf("TestShortenHandler() failed to shorten URL")
	}

	req := httptest.NewRequest(http.MethodGet, sRes.ShortUrl, nil)
	app := initApp(Config{Testing: true})
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, fiber.StatusFound, resp.StatusCode)
	locationHeader := resp.Header.Get("Location")
	assert.Equal(t, sr.LongUrl, locationHeader)

}

func TestFollowLinkHandlerNotFound(t *testing.T) {
	app := initApp(Config{Testing: true})

	req := httptest.NewRequest(http.MethodGet, "http://"+IP+":"+port+shortUrlPath+"doesnotexist", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
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

func TestFormShortUrl(t *testing.T) {
	testShortCode := "12345"

	requiredSubstrings := []string{
		"12345", "http", port, shortUrlPath,
	}

	res := formShortUrl(testShortCode)

	for _, substring := range requiredSubstrings {
		if !strings.Contains(res, substring) {
			t.Errorf("TestFormShortUrl() = %s, expected to contain %s", res, substring)
		}
	}
}

func TestIsValidMode(t *testing.T) {
	validRes := isValidMode(validModes[0], validModes)
	assert.Equal(t, true, validRes)

	invalidRes := isValidMode("foo", validModes)
	assert.Equal(t, false, invalidRes)
}

func shortenUrl(sr ShortenReq) (ShortenRes, *http.Response, error) {
	var sRes ShortenRes
	postBody, err := json.Marshal(sr)
	if err != nil {
		return sRes, nil, err
	}

	reader := bytes.NewReader(postBody)
	req := httptest.NewRequest(http.MethodPost, "/shorten", reader)
	req.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}
	app := initApp(Config{Testing: true})
	ResetDb(testDbPath)
	resp, err := app.Test(req)
	if err != nil {
		return sRes, nil, err
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return sRes, nil, err
	}

	err = json.Unmarshal(rb, &sRes)
	if err != nil {
		return sRes, nil, err
	}
	return sRes, resp, nil
}
