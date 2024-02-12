package main

import (
	"fmt"
	"html/template"
	"net/http"

	"codeberg.org/meadowingc/fido/linkchecker"
)

func main() {
	templates := template.Must(template.ParseGlob("templates/*.tmpl.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "home.tmpl.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("POST /submit-eval-request", func(w http.ResponseWriter, r *http.Request) {
		// get submitted link from request and evaluate all links on that page
		// for broken links
		submittedLink := r.FormValue("link")
		if submittedLink == "" {
			http.Error(w, "Link is required", http.StatusBadRequest)
			return
		}

		resultUUID := linkchecker.SubmitLinkForCheck(submittedLink)

		// redirect to results page
		http.Redirect(w, r, fmt.Sprintf("/result/%s", resultUUID), http.StatusSeeOther)
	})

	http.HandleFunc("/result/{operation_id}", func(w http.ResponseWriter, r *http.Request) {
		opId := r.PathValue("operation_id")

		// get result for operation ID
		result := linkchecker.GetResultForUUID(opId)
		if result == nil {
			http.Error(w, "No result found for operation ID", http.StatusNotFound)
			return
		}

		// TODO remove this
		templates = template.Must(template.ParseGlob("templates/*.tmpl.html"))
		err := templates.ExecuteTemplate(w, "result.tmpl.html", result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the contact page!")
	})

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
