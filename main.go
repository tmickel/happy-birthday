package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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
	db, err := sql.Open("postgres", os.Getenv("PG_DSN"))
	if err != nil {
		return err
	}

	http.HandleFunc("/daily", func(w http.ResponseWriter, r *http.Request) {
		dailyCheck(w, db)
	})

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

		row := db.QueryRow(`SELECT name, month, day FROM birthdays WHERE id = $1 LIMIT 1`, id)

		var birthday Birthday
		if err := row.Scan(&birthday.Name, &birthday.Month, &birthday.Day); err != nil {
			log.Printf("could not scan for delete: %v", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if err := sendDeleteNotice(&birthday); err != nil {
			log.Printf("could not delete: %v", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		db.Exec(`delete from birthdays where id = $1`, id)

		http.Redirect(w, r, "/", http.StatusFound)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		tpl, err := template.ParseFiles("./index.gohtml")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, fmt.Sprintf("template parse error: %v", err))
			return
		}

		birthdays, err := birthdaysQuery(db)
		if err != nil {
			http.Error(w, "failed to get birthdays", http.StatusInternalServerError)
		}

		templateData := TemplateData{
			Birthdays: *birthdays,
		}
		if err := tpl.ExecuteTemplate(w, "index", templateData); err != nil {
			fmt.Fprint(w, fmt.Sprintf("template error: %v", err))
		}

	})

	http.ListenAndServe(":80", nil)
	return nil
}

func birthdaysQuery(db *sql.DB) (*[]Birthday, error) {
	rows, err := db.Query(`SELECT id, name, month, day FROM birthdays order by to_date(month || ' ' || day,'Month DD');`)
	if err != nil {
		return nil, err
	}
	birthdays := []Birthday{}
	for rows.Next() {
		birthday := Birthday{}
		if err := rows.Scan(&birthday.Id, &birthday.Name, &birthday.Month, &birthday.Day); err != nil {
			return nil, errors.New("failed to scan")
		}
		birthdays = append(birthdays, birthday)
	}
	return &birthdays, nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func dailyCheck(w http.ResponseWriter, db *sql.DB) {
	sentCount := 0
	birthdays, err := birthdaysQuery(db)
	if err != nil {
		http.Error(w, "failed to get birthdays", http.StatusInternalServerError)
	}

	for _, birthday := range *birthdays {
		if birthday.Month == time.Now().Month().String() && birthday.Day == time.Now().Day() {
			sendBirthdayNotice(birthday.Name, "today")
			sentCount++
		}
		future := time.Now().Add(3 * 24 * time.Hour)
		if birthday.Month == future.Month().String() && birthday.Day == future.Day() {
			sendBirthdayNotice(birthday.Name, "on "+future.Weekday().String())
			sentCount++
		}
	}

	fmt.Fprintf(w, "success; sent %d", sentCount)
}

func sendBirthdayNotice(name string, day string) error {
	from := mail.NewEmail("Tim Mickel", "tim@tmickel.com")
	subject := name + "'s birthday " + day
	to := mail.NewEmail("Tim Mickel", "tim@tmickel.com")
	plainTextContent := "🥳"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, plainTextContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func sendDeleteNotice(birthday *Birthday) error {
	from := mail.NewEmail("Tim Mickel", "tim@tmickel.com")
	subject := birthday.Name + " has been defriended"
	to := mail.NewEmail("Tim Mickel", "tim@tmickel.com")
	plainTextContent := fmt.Sprintf("Old birthday: %s %d", birthday.Month, birthday.Day)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, plainTextContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	if err != nil {
		return err
	}
	return nil
}
