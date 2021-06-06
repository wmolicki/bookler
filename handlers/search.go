package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/blevesearch/bleve/v2"

	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/models"
)

const indexFile = "example.bleve"

type SearchHandler struct {
	Index bleve.Index

	bs models.BookService
}

func NewSearchHandler(bs models.BookService) *SearchHandler {
	s := &SearchHandler{bs: bs}
	s.Init()
	return s
}

func (s *SearchHandler) Init() {
	//index, err := bleve.Open(indexFile)
	index, err := s.CreateNew()
	helpers.Must(err)
	s.Index = index
}

func (s *SearchHandler) CreateNew() (bleve.Index, error) {
	if err := os.RemoveAll(indexFile); err != nil {
		return nil, err
	}
	mapping := bleve.NewIndexMapping()
	bleve.NewDocumentMapping()
	index, err := bleve.New(indexFile, mapping)
	if err != nil {
		return nil, err
	}

	books, err := s.bs.List()
	if err != nil {
		return nil, err
	}

	batch := index.NewBatch()

	for i, b := range books {
		batch.Index(strconv.Itoa(int(b.ID)), b)
		if i > 0 && i%100 == 0 {
			index.Batch(batch)
			fmt.Printf("indexed %d books", batch.Size())
			batch.Reset()
		}
	}

	index.Batch(batch)

	fmt.Printf("indexed %d books", len(books))

	return index, nil
}

func (s *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	query := bleve.NewFuzzyQuery(q)
	search := bleve.NewSearchRequest(query)
	search.Highlight = bleve.NewHighlightWithStyle("html")
	search.Highlight.AddField("Name")
	search.Highlight.AddField("Description")
	//bleve.NewFacetRequest("")
	search.Fields = []string{"*"}

	searchResults, err := s.Index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}

	serialized, _ := json.Marshal(searchResults)

	w.Write(serialized)
	w.WriteHeader(http.StatusOK)
}
