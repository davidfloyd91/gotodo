package main

import (
    "database/sql"
    "encoding/json"
    // "fmt"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "github.com/rs/cors"
)

var db *sql.DB

type Todo struct {
    Id int `json: "id", db: "id"`
    Name string `json:"name", db:"name"`
    Description string `json:"description", db:"description"`
}

func main() {
    initDB()
    // fmt.Printf("it worked!")

    initRouter()
}

func initRouter() {
    router := mux.NewRouter()

    // $ curl http://localhost:8000/ -v
    router.HandleFunc("/", Home).Methods("GET")

    // $ curl http://localhost:8000/todos/ -v
    router.HandleFunc("/todos/", GetTodos).Methods("GET")

    // $ curl http://localhost:8000/todos/1/ -v
    router.HandleFunc("/todos/{id}/", GetTodo).Methods("GET")

    // $ curl -H "Content-Type: application/json" http://localhost:8000/todos/ -d '{"name":"Do something cool","description":"Or not"}' -v
    router.HandleFunc("/todos/", CreateTodo).Methods("POST")

    // $ curl -X PUT http://localhost:8000/todos/9/ -d name=Ok%20gee -d description=Sure%20nice -v
    router.HandleFunc("/todos/{id}/", UpdateTodo).Methods("PUT")

    // $ curl -X DELETE http://localhost:8000/todos/8/delete/ -v
    router.HandleFunc("/todos/{id}/delete/", DeleteTodo).Methods("DELETE")

    c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowCredentials: true,
    })

    handler := c.Handler(router)

    // start server
    log.Fatal(http.ListenAndServe(":8000", handler))
}

/*** home ***/
var Home = func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode("أهلا بالعالم")
}

/*** index ***/
var GetTodos = func(w http.ResponseWriter, r *http.Request) {
    var todos []Todo

    SqlStatement := `
        SELECT * FROM todos
    `

    rows, err := db.Query(SqlStatement)
    if err != nil {
        panic(err)
    }

    for rows.Next() {
        var Id int
        var Name string
        var Description string
        rows.Scan(&Id, &Name, &Description)

        todos = append(todos, Todo{
            Id: Id,
            Name: Name,
            Description: Description,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todos)
}

/*** show ***/
var GetTodo = func(w http.ResponseWriter, r *http.Request) {
    var todo Todo
    params := mux.Vars(r)

    SqlStatement := `
        SELECT * FROM todos
        WHERE id = $1
    `

    err := db.QueryRow(
        SqlStatement,
        params["id"],
    ).Scan(&todo.Id, &todo.Name, &todo.Description)

    if err != nil {
        panic(err)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&todo)
}

/*** create ***/
var CreateTodo = func(w http.ResponseWriter, r *http.Request) {
    todo := &Todo{}

    err := json.NewDecoder(r.Body).Decode(todo)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    SqlStatement := `
        INSERT INTO todos (name, description)
        VALUES ($1, $2)
        RETURNING id
    `

    id := 0
    err = db.QueryRow(SqlStatement, todo.Name, todo.Description).Scan(&id)
    if err != nil {
        panic(err)
    }

    todo.Id = id

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todo)
}

/*** update ***/
var UpdateTodo = func(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
  	var todo Todo

    r.ParseForm() // must be called to access r.FormValue()

    SqlStatement := `
        UPDATE todos
        SET name = $1, description = $2
        WHERE id = $3
        RETURNING *
    `

    err := db.QueryRow(
        SqlStatement,
        r.FormValue("name"),
        r.FormValue("description"),
        params["id"],
    ).Scan(&todo.Id, &todo.Name, &todo.Description)

    if err != nil {
        panic(err)
    }

    w.Header().Set("Content-Type", "application/json")
  	json.NewEncoder(w).Encode(&todo)
}

/*** delete ***/
var DeleteTodo = func(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
  	var todo Todo

    SqlStatement := `
        DELETE FROM todos
        WHERE id = $1
        RETURNING *
    `

    err := db.QueryRow(
        SqlStatement,
        params["id"],
    ).Scan(&todo.Id, &todo.Name, &todo.Description)

    if err != nil {
        panic(err)
    }

    w.Header().Set("Content-Type", "application/json")
  	json.NewEncoder(w).Encode(&todo)
}

// hook up to postgres db
func initDB() {
  var err error
  db, err = sql.Open("postgres", "dbname=gotodo sslmode=disable")

  if err != nil {
      panic(err)
  }
}
