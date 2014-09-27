package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/goji/httpauth"
	"github.com/jmoiron/sqlx"
	"github.com/justinas/nosurf"
	flag "github.com/ogier/pflag"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"gopkg.in/unrolled/secure.v1"

	_ "github.com/mattn/go-sqlite3"
)

var _ = fmt.Println

var (
	flagDatabaseType  string
	flagDatabaseConn  string
	flagAdminPassword string
	flagListenPort    uint16
)

func init() {
	flag.StringVar(&flagDatabaseType, "dbtype", "sqlite3", "database type")
	flag.StringVar(&flagDatabaseConn, "dbconn", ":memory:", "database connection string")
	flag.StringVar(&flagAdminPassword, "password", "s33kret", "admin password")
	flag.Uint16VarP(&flagListenPort, "port", "p", 8080, "default listen port")
}

func RenderIndex(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "index", nil)
	if err != nil {
		log.Printf("Error rendering template: %s", err)
		http.Error(w, "error rendering template", 500)
	}
}

func RenderForm(c web.C, w http.ResponseWriter, r *http.Request) {
	db := c.Env["db"].(*sqlx.DB)

	r.ParseForm()

	token := r.Form.Get("token")
	if len(token) < 1 {
		http.Error(w, "no user token found", 403)
		return
	}

	resp := Respondent{}
	err := db.Get(&resp, "SELECT * FROM respondents WHERE token=?", token)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid token", 403)
		return
	}

	err = renderTemplate(w, "form", M{
		"token":      token,
		"name":       resp.Name,
		"csrf_token": nosurf.Token(r),
	})
	if err != nil {
		log.Printf("Error rendering template: %s", err)
		http.Error(w, "error rendering template", 500)
	}
}

func RenderAdmin(c web.C, w http.ResponseWriter, r *http.Request) {
	db := c.Env["db"].(*sqlx.DB)

	// Load all respondents
	respondents := []Respondent{}
	err := db.Select(&respondents, "SELECT * FROM respondents ORDER BY id ASC")
	if err != nil {
		log.Printf("Error getting respondents: %s", err)
		http.Error(w, "database error", 500)
		return
	}

	// Load all responses
	responses := []Response{}
	err = db.Select(&responses, "SELECT * FROM responses ORDER BY id ASC")
	if err != nil {
		log.Printf("Error getting responses: %s", err)
		http.Error(w, "database error", 500)
		return
	}

	// Massage the data a bit so that responses have names, not IDs
	type NiceResponse struct {
		ID         int64
		Respondent string
		Item       string
		Quantity   int
		MaxPrice   int
		Notes      string
	}

	respondentNames := make(map[int64]string)
	for _, resp := range respondents {
		respondentNames[resp.ID] = resp.Name
	}

	niceResponses := []NiceResponse{}
	for _, resp := range responses {
		niceResponses = append(niceResponses, NiceResponse{
			ID:         resp.ID,
			Respondent: respondentNames[resp.ID],
			Item:       resp.Item,
			Quantity:   resp.Quantity,
			MaxPrice:   resp.MaxPrice,
			Notes:      resp.Notes,
		})
	}

	err = renderTemplate(w, "admin", M{
		"respondents": respondents,
		"responses":   niceResponses,
		"csrf_token":  nosurf.Token(r),
	})
	if err != nil {
		log.Printf("Error rendering template: %s", err)
		http.Error(w, "error rendering template", 500)
	}
}

func HandleNewRespondent(c web.C, w http.ResponseWriter, r *http.Request) {
	db := c.Env["db"].(*sqlx.DB)

	r.ParseForm()

	_, err := db.NamedExec(`INSERT INTO respondents (name, token) VALUES (:name, :token)`,
		map[string]interface{}{
			"name":  r.Form.Get("name"),
			"token": randString(20),
		})
	if err != nil {
		log.Printf("Error getting respondents: %s", err)
		http.Error(w, "database error", 500)
		return
	}

	http.Redirect(w, r, "/admin/main", 303)
}

func HandleSubmit(c web.C, w http.ResponseWriter, r *http.Request) {
	db := c.Env["db"].(*sqlx.DB)

	r.ParseForm()

	token := r.Form.Get("token")
	if len(token) < 1 {
		http.Error(w, "no user token found", 403)
		return
	}

	// Validate input.
	quantity, err := strconv.ParseUint(r.Form.Get("quantity"), 10, 64)
	if err != nil {
		http.Error(w, "invalid quantity", 400)
		return
	}
	max_price, err := strconv.ParseUint(r.Form.Get("max_price"), 10, 64)
	if err != nil {
		http.Error(w, "invalid max price", 400)
		return
	}
	if len(r.Form.Get("item")) < 1 {
		http.Error(w, "no item provided", 400)
		return
	}

	// Ensure token exists
	resp := Respondent{}
	err = db.Get(&resp, "SELECT * FROM respondents WHERE token=?", token)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid token", 403)
		return
	}

	// Add the new response
	_, err = db.NamedExec(`INSERT INTO responses (respondent, item, quantity, max_price, notes) `+
		`VALUES (:respondent, :item, :quantity, :max_price, :notes)`,
		map[string]interface{}{
			"respondent": resp.ID,
			"item":       r.Form.Get("item"),
			"quantity":   quantity,
			"max_price":  max_price,
			"notes":      r.Form.Get("notes"),
		})
	if err != nil {
		log.Printf("Error getting respondents: %s", err)
		http.Error(w, "database error", 500)
		return
	}

	err = renderTemplate(w, "success", M{
		"token": token,
	})
	if err != nil {
		log.Printf("Error rendering template: %s", err)
		http.Error(w, "error rendering template", 500)
	}
}

func main() {
	flag.Parse()

	log.Println("Started")

	db, err := sqlx.Connect(flagDatabaseType, flagDatabaseConn)
	if err != nil {
		log.Fatalln(err)
	}

	// Create tables
	tx := db.MustBegin()
	for _, stmt := range createStatements {
		tx.MustExec(stmt)
	}
	tx.Commit()

	// Create middleware
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:          true,
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
	})

	auth := httpauth.SimpleBasicAuth("admin", flagAdminPassword)

	// Setup mux + middleware
	m := web.New()
	m.Use(middleware.RequestID)
	m.Use(middleware.Logger)
	m.Use(middleware.Recoverer)
	m.Use(middleware.AutomaticOptions)
	m.Use(nosurf.NewPure)
	m.Use(secureMiddleware.Handler)
	m.Use(func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			c.Env["db"] = db
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handler)
	})

	// Setup routes
	m.Get("/", RenderIndex)
	m.Get("/form", RenderForm)
	m.Post("/submit", HandleSubmit)

	admin := web.New()
	admin.Use(auth)
	admin.Get("/admin/main", RenderAdmin)
	admin.Post("/admin/respondent", HandleNewRespondent)
	m.Handle("/admin/*", admin)

	// Static assets
	for _, asset := range AssetDescriptors() {
		if !strings.HasSuffix(asset.Path, ".tmpl") {
			log.Printf("Serving asset: %s", asset.Path)
			m.Get("/"+asset.Path, ServeAsset(asset.Path, asset.Mime))
		}
	}

	addr := fmt.Sprintf(":%d", flagListenPort)
	log.Printf("Serving HTTP on port %d...", flagListenPort)
	log.Fatal(http.ListenAndServe(addr, m))
}
