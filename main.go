package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	portHttp   = ":8080"
	//dbUsername = "postgres"
	//dbPassword = "777888"
	//dbHost = "127.0.0.1"
	//dbPort = "5432"
	//dateBase = "postgres"
	dbUsername = "boris"
	dbPassword = "qwerty"
	dbHost = "10.7.27.34"
	dbPort = "5432"
	dateBase = "books"
)

type BookModel struct {
	Title  string
	Author string
	Cost   int
}

type Service struct {
	Db *pgx.Conn
}

type Rest struct {
	Router  *mux.Router
	Service *Service
}

func (s *Service) dbInit(username, password string) error {
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second * 3)
	defer cancelFunc()

	databaseUrl :=fmt.Sprintf("postgres://%s:%s@%s:%s/%s",username,password,dbHost,dbPort,dateBase)
	conn, err := pgx.Connect(timeout, databaseUrl)
	if err != nil {
		return errors.New("failed to connect db url: " + databaseUrl)
	}
	s.Db = conn
	return nil
}

func (s *Service) getBooksByAuthor(author string) ([]BookModel,error) {
	result := make([]BookModel,0)
	rows, err := s.Db.Query(context.Background(), "SELECT title, cost FROM books WHERE author=$1", author)
	if err != nil {
		return nil,errors.New("не удалось получить книги по автору")
	}

	for rows.Next() {
		var title string
		var cost int
		err = rows.Scan(&title, &cost)
		if err != nil {
			return nil,err
		}
		result = append(result, BookModel{title, author, cost})
	}

	log.Println("Успешно выполнен запрос, заполнено записей: " + strconv.Itoa(len(result)))
	return result,nil
}

func main() {
	log.Println("Server starting in " + portHttp + " port")

	service := new(Service)
	err :=service.dbInit(dbUsername, dbPassword)
    if err !=nil {
		panic(err)
	}

	rest := Rest{
		Router:  mux.NewRouter(),
		Service: service,
	}

	rest.Router.HandleFunc("/GetBookByAuthor/{author}", rest.GetBookByAuthor)

	log.Fatal(http.ListenAndServe(portHttp, rest.Router))
}

func (rs *Rest) GetBookByAuthor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author,ok := vars["author"]
	if !ok {
		http.Error(w,"author param does not exist", http.StatusBadRequest)
	}
	if author == "" {
		http.Error(w, "author param is empty", http.StatusBadRequest)
	}

	result,err := rs.Service.getBooksByAuthor(author)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	js ,err:= json.Marshal(result)
	if err !=nil {
		http.Error(w, "author param is empty", http.StatusBadRequest)
	}

	w.Write(js)
}
