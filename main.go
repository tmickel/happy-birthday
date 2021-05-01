package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Birthday struct {
	Id    int
	Name  string
	Month string
	Day   int
}

type TemplateData struct {
	Birthdays []Birthday
}

func run() error {
	db, err := sql.Open("postgres", "postgres://gouser:gopassword@postgres.local:5432/godb?sslmode=disable")
	if err != nil {
		return err
	}

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "form parse error")
			return
		}

		name := r.PostForm.Get("name")
		month := r.PostForm.Get("month")
		day := r.PostForm.Get("day")
		db.Exec(`insert into birthdays (name, month, day) values ($1, $2, $3)`, name, month, day)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "form parse error")
			return
		}

		id := r.PostForm.Get("id")
		db.Exec(`delete from birthdays where id = $1`, id)
		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, name, month, day FROM birthdays order by to_date(month || ' ' || day,'Month DD');`)
		if err != nil {
			http.Error(w, "failed to read", http.StatusInternalServerError)
			return
		}
		birthdays := []Birthday{}
		for rows.Next() {
			birthday := Birthday{}
			if err := rows.Scan(&birthday.Id, &birthday.Name, &birthday.Month, &birthday.Day); err != nil {
				http.Error(w, "failed to scan", http.StatusInternalServerError)
				return
			}
			birthdays = append(birthdays, birthday)
		}

		tpl, err := template.ParseFiles("./index.gohtml")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, fmt.Sprintf("template parse error: %v", err))
			return
		}

		templateData := TemplateData{
			Birthdays: birthdays,
		}
		if err := tpl.ExecuteTemplate(w, "index", templateData); err != nil {
			fmt.Fprint(w, fmt.Sprintf("template error: %v", err))
		}

	})

	http.ListenAndServe(":80", nil)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
