package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

var tpl *template.Template

// Compile templates on start of the application
func init() {
	tpl = template.Must(template.ParseGlob("src/views/*.html"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", uploadFileHandler())

	http.ListenAndServe(":"+port, mux)

}

func Connstr() (db *sql.DB) {

	dbDriver := "mssql"
	dbUser := "DESKTOP-1VA2HU8\\srave"
	//dbPass := "your_password"
	dbName := "portfolio"
	db, err := sql.Open(dbDriver, dbUser+":@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

const maxUploadSize = 10 * 1024 * 1024 // 2 mb

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {

			tpl.ExecuteTemplate(w, "index", nil)
			return
		}
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Printf("Could not parse multipart form: %v\n", err)
			renderError(w, "CANT_PARSE_FORM", http.StatusInternalServerError)
			return
		}

		// parse and validate file and post parameters
		file, fileHeader, err := r.FormFile("myFile")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		// Get and print out file size
		fileSize := fileHeader.Size
		fmt.Printf("File size (bytes): %v\n", fileSize)
		// validate file size
		if fileSize > maxUploadSize {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		fmt.Println(fileBytes)
		db := Connstr()
		if r.Method == "POST" {
			file := fileBytes

			insForm, err := db.Prepare("INSERT INTO fileupdown(filebyte) VALUES(?)")
			if err != nil {
				panic(err.Error())
			}
			insForm.Exec(file)

		}
		defer db.Close()

	})
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}
