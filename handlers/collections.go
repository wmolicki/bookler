package handlers

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/wmolicki/bookler/context"
	"github.com/wmolicki/bookler/helpers"
	"github.com/wmolicki/bookler/models"
	"github.com/wmolicki/bookler/views"
)

func NewCollectionsHandler(us *models.UserService, bs *models.BookService, cs *models.CollectionsService) *CollectionsHandler {
	listView := views.NewView("bulma", "templates/collections.gohtml")
	editView := views.NewView("bulma", "templates/collection_edit.gohtml")
	addView := views.NewView("bulma", "templates/collection_add.gohtml")

	bs.DestructiveReset()
	return &CollectionsHandler{bs, us, cs, listView, addView, editView}
}

type CollectionsHandler struct {
	bs *models.BookService
	us *models.UserService
	cs *models.CollectionsService

	listView *views.View
	addView  *views.View
	editView *views.View
}

type CollectionsListViewModel struct {
	Collections []*models.Collection
}

func (ch *CollectionsHandler) Add(w http.ResponseWriter, r *http.Request) {
	ch.addView.Render(w, r, nil)
	return
}

func (ch *CollectionsHandler) HandleAdd(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err.Error())
	}
	var data editCollectionFormData
	err = decoder.Decode(&data, r.PostForm)
	if err != nil {
		panic(err)
	}

	// TODO: need to check if same collection exists

	user := context.User(r.Context())
	_, err = ch.cs.Create(user, data.Name)
	if err != nil {
		log.Errorf("could not create collection: %v", err)
		http.Error(w, "could not create collection", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/collections", http.StatusFound)
	return
}

func (ch *CollectionsHandler) HandleAddBook(w http.ResponseWriter, r *http.Request) {
	collectionId, err := helpers.ParseUintParam(r, "collectionId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	c, err := ch.cs.GetByID(collectionId)
	if err != nil {
		log.Errorf("could not load collection: %v", err)
		http.Error(w, "could not load collection", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		badRequest(w, err.Error())
	}

	var addBookFormData struct {
		Name string
	}

	err = decoder.Decode(&addBookFormData, r.PostForm)
	if err != nil {
		panic(err)
	}

	book, err := ch.bs.GetBookByName(addBookFormData.Name)
	if err != nil {
		log.Errorf("could not find book by name: %v", err)
		http.Error(w, "could not find book by name", http.StatusInternalServerError)
		return
	}

	err = ch.cs.AddBook(c, book)
	if err != nil {
		log.Errorf("could not add book to collection: %v", err)
		http.Error(w, "could not add book to collection", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/collections/%d", c.ID), http.StatusFound)
	return
}

func (ch *CollectionsHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	collectionId, err := helpers.ParseUintParam(r, "collectionId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	c, err := ch.cs.GetByID(collectionId)
	if err != nil {
		log.Errorf("could not load collection: %v", err)
		http.Error(w, "could not load collection", http.StatusInternalServerError)
		return
	}

	err = ch.cs.Delete(c)
	if err != nil {
		log.Errorf("could not delete collection: %v", err)
		http.Error(w, "could not delete collection", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/collections", http.StatusFound)
	return
}

func (ch *CollectionsHandler) HandleBookDelete(w http.ResponseWriter, r *http.Request) {
	collectionId, err := helpers.ParseUintParam(r, "collectionId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	bookId, err := helpers.ParseUintParam(r, "bookId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	c, err := ch.cs.GetByID(collectionId)
	if err != nil {
		log.Errorf("could not load collection: %v", err)
		http.Error(w, "could not load collection", http.StatusInternalServerError)
		return
	}

	b, err := ch.bs.GetBookByID(bookId)
	if err != nil {
		log.Errorf("error getting book: %v", err)
		http.Error(w, "could not get book", http.StatusInternalServerError)
		return
	}

	err = ch.cs.DeleteBook(c, b)
	if err != nil {
		log.Errorf("error removing book from collection: %v", err)
		http.Error(w, "could not remove book from collection", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/collections/%d", collectionId), http.StatusFound)
	return
}

func (ch *CollectionsHandler) Edit(w http.ResponseWriter, r *http.Request) {
	collectionId, err := helpers.ParseUintParam(r, "collectionId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	c, err := ch.cs.GetWithBooks(collectionId)
	if err != nil {
		log.Errorf("could not load collection: %v", err)
		http.Error(w, "could not load collection", http.StatusInternalServerError)
		return
	}

	inCollection := map[string]bool{}
	for _, b := range c.Books {
		inCollection[b.Name] = true
	}

	books, err := ch.bs.GetList()
	if err != nil {
		log.Errorf("could not get books: %v", err)
		http.Error(w, "could not get books", http.StatusInternalServerError)
		return
	}

	// filter out books that are already mapped to this collection
	var dropDownBooks []*models.Book

	for _, b := range books {
		_, ok := inCollection[b.Name]
		if ok {
			continue
		}
		dropDownBooks = append(dropDownBooks, b)
	}

	viewModel := struct {
		ID              uint
		Name            string
		CollectionBooks []*models.BookInCollection
		Books           []*models.Book
	}{c.ID, c.Name, c.Books, dropDownBooks}

	ch.editView.Render(w, r, &viewModel)
	return
}

type editCollectionFormData struct {
	Name string
}

func (ch *CollectionsHandler) HandleEdit(w http.ResponseWriter, r *http.Request) {
	collectionId, err := helpers.ParseUintParam(r, "collectionId")
	if err != nil {
		badRequest(w, fmt.Sprintf("could not convert param: %v", err))
		return
	}

	c, err := ch.cs.GetByID(collectionId)
	if err != nil {
		log.Errorf("could not load collection: %v", err)
		http.Error(w, "could not load collection", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		badRequest(w, err.Error())
	}
	var data editCollectionFormData
	err = decoder.Decode(&data, r.PostForm)
	if err != nil {
		panic(err)
	}

	c.Name = data.Name

	_, err = ch.cs.Update(c)
	if err != nil {
		log.Errorf("could not edit collection: %v", err)
		http.Error(w, "could not edit collection", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/collections/%d", collectionId), http.StatusFound)
	return
}

func (ch *CollectionsHandler) List(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	collections, err := ch.cs.List(user)
	if err != nil {
		log.Errorf("could not load collections: %v", err)
		http.Error(w, "could not load collections", http.StatusInternalServerError)
		return
	}
	ch.listView.Render(w, r, &CollectionsListViewModel{collections})
}
