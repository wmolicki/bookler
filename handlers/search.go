package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve/v2"

	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/models"
)

const indexFile = "example.bleve"

type SearchHandler struct {
	Index bleve.Index

	ba *models.BookAuthorService
}

func NewSearchHandler(ba *models.BookAuthorService) *SearchHandler {
	s := &SearchHandler{ba: ba}
	s.Init()
	return s
}

func (s *SearchHandler) Init() {
	index, err := bleve.Open(indexFile)
	//index, err := s.CreateNew()
	helpers.Must(err)
	s.Index = index
}

func (s *SearchHandler) CreateNew() (bleve.Index, error) {
	if err := os.RemoveAll(indexFile); err != nil {
		return nil, err
	}
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexFile, mapping)
	if err != nil {
		return nil, err
	}

	books, err := s.ba.BookWithAuthorList()
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
	q = strings.TrimSpace(q)
	q = strings.ToLower(q)

	finalQuery := bleve.NewDisjunctionQuery()

	fuzzyQuery := bleve.NewFuzzyQuery(q)
	fuzzyQuery.SetField("Name")
	fuzzyQuery.SetBoost(0.75)
	finalQuery.AddQuery(fuzzyQuery)

	fuzzyDescriptionQuery := bleve.NewFuzzyQuery(q)
	fuzzyDescriptionQuery.SetField("Description")
	fuzzyDescriptionQuery.SetBoost(0.35)
	finalQuery.AddQuery(fuzzyDescriptionQuery)

	// match with some analysis
	matchQuery := bleve.NewMatchQuery(q)
	matchQuery.SetBoost(2.0)
	matchQuery.SetField("Name")
	matchQuery.Analyzer = "en"
	finalQuery.AddQuery(matchQuery)

	// exact match
	termQ := bleve.NewTermQuery(q)
	termQ.SetField("Name")
	termQ.SetBoost(3.0)
	finalQuery.AddQuery(termQ)

	terms := strings.Split(q, " ")
	for _, term := range terms {
		prefixQuery := bleve.NewPrefixQuery(term)
		prefixQuery.SetField("Name")
		prefixQuery.SetBoost(1)
		finalQuery.AddQuery(prefixQuery)

		fuzzyTermQuery := bleve.NewFuzzyQuery(term)
		fuzzyTermQuery.SetBoost(0.85)
		fuzzyTermQuery.SetField("Name")
		fuzzyQuery.SetPrefix(2)
		fuzzyTermQuery.SetFuzziness(2)
		finalQuery.AddQuery(fuzzyTermQuery)
	}

	if len(terms) > 1 {
		matchPhraseQuery := bleve.NewMatchPhraseQuery(q)
		matchPhraseQuery.Analyzer = "en"
		matchPhraseQuery.SetBoost(3.0)
		finalQuery.AddQuery(matchPhraseQuery)
	}

	//. query := bleve.NewQueryStringQuery(fmt.Sprintf("Description:%s~1 Name:%s~1 Name:\"%s\"^10", q, q))

	search := bleve.NewSearchRequest(finalQuery)
	search.Highlight = bleve.NewHighlightWithStyle("html")
	search.Highlight.AddField("Name")
	search.Highlight.AddField("Description")
	search.Size = 10
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
