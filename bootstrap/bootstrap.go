// Package bootstrap holds the startup sequence every consuming game's
// main() repeats identically: connect to the database with retry, apply an
// ordered SQL schema manifest, mount static assets, and serve. Unlike the
// rest of this framework, these functions log.Fatalln and exit the process
// on failure rather than returning an error — every current call site did
// exactly that already, so this only removes the repetition, not the
// behavior. Game-specific setup (Register, SetBrandName, SetPagePolicy,
// route wiring) still lives in each game's own main().
package bootstrap

import (
	"database/sql"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gerp93/gameshell-framework/database"
	"github.com/gerp93/gameshell-framework/static"
)

// ConnectWithRetry connects to the database, retrying up to maxAttempts
// times with delay between attempts. Exits the process on final failure.
func ConnectWithRetry(maxAttempts int, delay time.Duration) *sql.DB {
	db, err := database.CreateDatabaseConnection()
	attempt := 0
	for err != nil && attempt < maxAttempts {
		time.Sleep(delay)
		attempt++
		db, err = database.CreateDatabaseConnection()
	}
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

// ApplySchema executes each file in fsys, in the given order, against the
// database. Exits the process on failure. Games call it twice at startup —
// once for the framework's own static.SQLFiles, once for their own — since
// the framework schema must be applied first (game tables FK to it).
func ApplySchema(fsys fs.FS, files []string) {
	for _, file := range files {
		bytes, err := fs.ReadFile(fsys, file)
		if err != nil {
			log.Fatalln(err)
		}
		if err := database.Execute(string(bytes)); err != nil {
			log.Fatalln(err)
		}
	}
}

// MountStaticAssets mounts the game's own static files at /static/ (from
// gameFS) and this framework's shared assets at /gs/.
func MountStaticAssets(gameFS fs.FS) {
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(gameFS))))
	http.Handle("GET /gs/", http.StripPrefix("/gs/", http.FileServer(http.FS(static.StaticFiles))))
}

// Serve blocks running the HTTP server on the default ServeMux (every route
// a game registers via http.Handle is already there), honoring
// <envPrefix>_PORT (default ":2016"), <envPrefix>_CERT_FILE/_KEY_FILE for
// TLS, and <envPrefix>_LOG_FILE to redirect log output first. Exits the
// process if the server returns an error.
func Serve(envPrefix string) {
	if logFile := os.Getenv(envPrefix + "_LOG_FILE"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	port := ":2016"
	if p := os.Getenv(envPrefix + "_PORT"); p != "" {
		port = ":" + p
	}

	log.Println("server is running...")

	var err error
	certFile := os.Getenv(envPrefix + "_CERT_FILE")
	keyFile := os.Getenv(envPrefix + "_KEY_FILE")
	if certFile != "" && keyFile != "" {
		err = http.ListenAndServeTLS(port, certFile, keyFile, nil)
	} else {
		err = http.ListenAndServe(port, nil)
	}
	if err != nil {
		log.Fatalln(err)
	}
}
