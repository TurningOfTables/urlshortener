package main

import (
	"database/sql"
	"flag"
	"math/rand"
	"net/url"
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/template/html/v2"
	_ "github.com/mattn/go-sqlite3"
)

// Configuration
const prodDbPath = "file:./data/linkdata.db?cache=shared"
const testDbPath = "file:./data/linkdata_test.db?cache=shared"
const IP = "localhost"
const port = "8080"
const shortUrlPath = "/go/"
const shortCodeLength = 6
const shortCodeCharacters = "abcdefghijklmnopqrstuvwxyz0123456789"
const shortCodeMaxAttempts = 10 // Maximum times to try again if a shortCode is already in use, to avoid a tiny possibility of it looping forever

var validModes = []string{"test", "production"}
var validUrl = regexp.MustCompile(`(https?:\/\/)([\w\-])+\.{1}([a-zA-Z]{2,63})([\/\w-]*)*\/?\??([^#\n\r]*)?#?([^\n\r]*)`)

// Command line flags
var modeFlag = flag.String("mode", "production", "Set whether to use the 'production' or 'test' database. Defaults to production mode.")
var resetDbFlag = flag.Bool("reset", false, "Set to reset the database to default on startup")

type ShortenReq struct {
	LongUrl string `json:"longUrl"`
}

type Link struct {
	Id        int    `json:"id"`
	LongUrl   string `json:"longUrl"`
	ShortCode string `json:"shortCode"`
	ShortUrl  string `json:"shortUrl"`
}

type ErrorResponse struct {
	Error       string
	Description string
}

func main() {

	flag.Parse()
	if isValidMode(*modeFlag, validModes) {
		log.Infof("Starting app in %s mode", *modeFlag)
	} else {
		log.Fatalf("App startup failed - %s is not a valid mode", *modeFlag)
	}

	app := initApp()
	app.Listen(IP + ":" + port)
}

func initApp() *fiber.App {

	// Choose database and connect
	var dbPath string
	switch *modeFlag {
	case "test":
		dbPath = testDbPath
	case "production":
		dbPath = prodDbPath
	default:
		dbPath = prodDbPath
	}

	if *resetDbFlag {
		ResetDb(dbPath)
	}

	db := ConnectToDb(dbPath)

	engine := html.New("./views", ".html")
	app := fiber.New((fiber.Config{AppName: "Url Shortener", Views: engine}))

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return indexHandler(c, db)
	})

	app.Post("/shorten", func(c *fiber.Ctx) error {
		return shortenHandler(c, db)
	})

	app.Get("/go/:shortCode", func(c *fiber.Ctx) error {
		return followLinkHandler(c, db)
	})

	app.Static("/", "./public")
	return app
}

func indexHandler(c *fiber.Ctx, db *sql.DB) error {
	return c.Render("index", nil)
}

func shortenHandler(c *fiber.Ctx, db *sql.DB) error {
	var sr ShortenReq
	err := c.BodyParser(&sr)
	if !isValidUrl(sr.LongUrl) || err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid URL", Description: "That doesn't look like a valid URL for me to shorten :("})
	}

	sc := generateUniqueShortCode(db)
	su := formShortUrl(sc)

	var l = Link{
		ShortCode: sc,
		ShortUrl:  su,
		LongUrl:   sr.LongUrl,
	}

	_, err = db.Exec("INSERT into links (longurl, shorturl, shortcode) VALUES (?, ?, ?)", l.LongUrl, l.ShortUrl, l.ShortCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Database error", Description: "Couldn't add your link to the database :("})
	}

	return c.Status(fiber.StatusCreated).JSON(l.ShortUrl)
}

func followLinkHandler(c *fiber.Ctx, db *sql.DB) error {
	sc := c.Params("shortCode")
	if sc == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid short code", Description: "No short code provided"})
	}

	var longUrl string
	res := db.QueryRow("SELECT longurl FROM links WHERE shortCode = ?", sc)
	if err := res.Scan(&longUrl); err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Not found", Description: "A link with that short code was not found"})
	}

	return c.Status(fiber.StatusPermanentRedirect).Redirect(longUrl)
}

func isValidMode(requestedMode string, validModes []string) bool {
	for _, vm := range validModes {
		if requestedMode == vm {
			return true
		}
	}

	return false
}

func isValidUrl(testUrl string) bool {

	// First check by trying to parse the URL - looser check
	if _, err := url.ParseRequestURI(testUrl); err != nil {
		return false
	}

	// Then check against a very simple regex
	return validUrl.MatchString(testUrl)
}

func generateUniqueShortCode(db *sql.DB) string {
	var letters = []rune(shortCodeCharacters)

	c := make([]rune, shortCodeLength)

	for i := range c {
		c[i] = letters[rand.Intn(len(letters))]
	}

	code := string(c)

	if shortCodeInUse(db, code) {
		for i := 0; i < shortCodeMaxAttempts; i++ {
			code = generateUniqueShortCode(db)
		}
	}

	log.Infof("Generated short code: %s", code)
	return string(code)
}

func shortCodeInUse(db *sql.DB, shortCode string) bool {
	var l Link
	res := db.QueryRow("SELECT * FROM links WHERE shortCode = ?", shortCode)
	if err := res.Scan(&l); err == sql.ErrNoRows {
		return false
	}

	return true
}

func formShortUrl(shortCode string) string {
	var portString string
	if len(port) > 0 {
		portString = ":" + port
	}

	shortUrl := "http://" + IP + portString + shortUrlPath + shortCode
	log.Infof("Generated short URL: %s", shortUrl)
	return shortUrl
}
