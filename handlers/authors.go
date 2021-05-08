package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/wmolicki/bookler/config"
)

type AuthorsHandler struct {
	Env *config.Env
}

func (a *AuthorsHandler) Index(w http.ResponseWriter, r *http.Request) {
	query := `SELECT a.id, a.name, COUNT(ba.book_id) as book_count, ba.created_on as created_on 
              FROM author a JOIN book_author ba on a.id = ba.author_id
              GROUP BY (a.id) ORDER BY a.name, a.created_on;`
	db := a.Env.DB

	var authors []struct {
		Id        int
		Name      string
		CreatedOn time.Time `db:"created_on"`
		BookCount int       `db:"book_count"`
	}

	err := db.Select(&authors, query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not query for authors: %v", err)
		w.Write([]byte("something went wrong"))
		return
	}

	b, _ := json.Marshal(authors)

	w.Write(b)
}
