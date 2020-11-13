package main

import (
	"database/sql"
	"fmt"
	"text/template"
	"log"
	"net/http"
	"os"
	_ "github.com/go-sql-driver/mysql"
)

type Employee struct{
	Id uint
	Name string
	City string
}


// app struct contains global state.
type app struct {
	// db is the global database connection pool.
	db *sql.DB
	// tmpl is the parsed HTML template.
	// tmpl *template.Template
}

var tmpl = template.Must(template.ParseGlob("form/*"))
// indexHandler handles requests to the / route.
func (app *app) Index(w http.ResponseWriter, r *http.Request) {
	selDB, err := app.db.Query("select * from employee order by id desc")
	if err != nil{
		panic(err.Error())
	}
	emp := Employee{}
	res := []Employee{}
	for selDB.Next(){
		var id uint
		var name, city string
		err = selDB.Scan(&id, &name, &city)
		if err != nil {
			panic(err.Error())
		}
		emp.Id=id
		emp.Name=name
		emp.City=city
		res = append(res,emp)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	// defer app.db.Close()
}

func (app *app)Show(w http.ResponseWriter, r *http.Request){
	nId := r.URL.Query().Get("id")
	selDB, err := app.db.Query("select * from employee where id=?",nId) 
	if err != nil {
		panic(err.Error())
	}
	emp := Employee{}
	for selDB.Next(){
		var id uint
		var name, city string
		err = selDB.Scan(&id,&name,&city)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name=name
		emp.City=city
	}
	tmpl.ExecuteTemplate(w,"Show",emp)
	// defer app.db.Close()
}

func (app *app)New(w http.ResponseWriter, r *http.Request){
	tmpl.ExecuteTemplate(w, "New", nil)
}

func (app *app)Edit(w http.ResponseWriter, r *http.Request){
	nId := r.URL.Query().Get("id")
	selDB, err := app.db.Query("select * from employee where id=?",nId) 
	if err != nil {
		panic(err.Error())
	}
	emp := Employee{}
	for selDB.Next(){
		var id uint
		var name, city string
		err = selDB.Scan(&id,&name,&city)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name=name
		emp.City=city
	}
	tmpl.ExecuteTemplate(w,"Edit",emp)
	// defer app.db.Close()
}

func (app *app)Insert(w http.ResponseWriter, r *http.Request){
	if r.Method == "POST" {
		name := r.FormValue("name")
		city := r.FormValue("city")
		insForm, err := app.db.Prepare("insert into employee(name,city) values(?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name,city)
		log.Println("INSERT: Name: " + name + " | City: " + city)
	}
	// defer app.db.Close()
	http.Redirect(w,r,"/",301)
}

func (app *app)Update(w http.ResponseWriter, r *http.Request){
	if r.Method == "POST"{
		name := r.FormValue("name")
		city := r.FormValue("city")
		id := r.FormValue("uid")
		insForm, err := app.db.Prepare("update employee set name=?, city=? where id=?")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name,city,id)
		log.Println("UPDATE: Name: " + name + " | City: " + city)
	}
	// defer app.db.Close()
	http.Redirect(w,r,"/",301)
}


func (app *app)Delete(w http.ResponseWriter, r *http.Request){
	emp := r.URL.Query().Get("id")
	delForm, err := app.db.Prepare("delete from employee where id=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(emp)
	log.Println("DELETE")
	// defer app.db.Close()
	http.Redirect(w,r,"/",301)
}


func main() {

	 app := &app{}

	// If the optional DB_TCP_HOST environment variable is set, it contains
	// the IP address and port number of a TCP connection pool to be created,
	// such as "127.0.0.1:3306". If DB_TCP_HOST is not set, a Unix socket
	// connection pool will be created instead.
	if os.Getenv("DB_TCP_HOST") != "" {
		app.db, _ = initTCPConnectionPool()
		// app.db, err = initTCPConnectionPool()
		// if err != nil {
		// 	log.Fatalf("initTCPConnectionPool: unable to connect: %v", err)
		// }
	} else {
		app.db, _= initSocketConnectionPool()
		// app.db, err = initSocketConnectionPool()
		// if err != nil {
		// 	log.Fatalf("initSocketConnectionPool: unable to connect: %v", err)
		// }
	}

	// Create the employee table if it does not already exist.
	if _, err := app.db.Exec(`CREATE TABLE IF NOT EXISTS employee
	( id int(6) unsigned NOT NULL AUTO_INCREMENT,name varchar(30) NOT NULL,
	city varchar(30) NOT NULL, PRIMARY KEY (id) ) ENGINE=InnoDB AUTO_INCREMENT=1;`); err != nil {
		log.Fatalf("DB.Exec: unable to create table: %s", err)
	}

	http.HandleFunc("/", app.Index)
    http.HandleFunc("/show",app.Show)
    http.HandleFunc("/new", app.New)
    http.HandleFunc("/edit", app.Edit)
    http.HandleFunc("/insert", app.Insert)
    http.HandleFunc("/update", app.Update)
    http.HandleFunc("/delete", app.Delete)
    // http.ListenAndServe(":8080", nil)


	// http.HandleFunc("/", app.indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// mustGetEnv is a helper function for getting environment variables.
// Displays a warning if the environment variable is not set.
func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Warning: %s environment variable not set.\n", k)
	}
	return v
}

// initSocketConnectionPool initializes a Unix socket connection pool for
// a Cloud SQL instance of SQL Server.
func initSocketConnectionPool() (*sql.DB, error) {
	// [START cloud_sql_mysql_databasesql_create_socket]
	var (
		dbUser                 = mustGetenv("DB_USER")                  // e.g. 'my-db-user'
		dbPwd                  = mustGetenv("DB_PASS")                  // e.g. 'my-db-password'
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME") // e.g. 'project:region:instance'
		dbName                 = mustGetenv("DB_NAME")                  // e.g. 'my-database'
	)

	socketDir, isSet := os.LookupEnv("DB_SOCKET_DIR")
	if !isSet {
		socketDir = "/cloudsql"
	}

	var dbURI string
	dbURI = fmt.Sprintf("%s:%s@unix(/%s/%s)/%s?parseTime=true", dbUser, dbPwd, socketDir, instanceConnectionName, dbName)

	// dbPool is the pool of database connections.
	dbPool, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	// [START_EXCLUDE]
	configureConnectionPool(dbPool)
	// [END_EXCLUDE]

	return dbPool, nil
	// [END cloud_sql_mysql_databasesql_create_socket]
}

// initTCPConnectionPool initializes a TCP connection pool for a Cloud SQL
// instance of SQL Server.
func initTCPConnectionPool() (*sql.DB, error) {
	// [START cloud_sql_mysql_databasesql_create_tcp]
	var (
		dbUser    = mustGetenv("DB_USER")     // e.g. 'my-db-user'
		dbPwd     = mustGetenv("DB_PASS")     // e.g. 'my-db-password'
		dbTcpHost = mustGetenv("DB_TCP_HOST") // e.g. '127.0.0.1' ('172.17.0.1' if deployed to GAE Flex)
		dbPort    = mustGetenv("DB_PORT")     // e.g. '3306'
		dbName    = mustGetenv("DB_NAME")     // e.g. 'my-database'
	)

	var dbURI string
	dbURI = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPwd, dbTcpHost, dbPort, dbName)

	// dbPool is the pool of database connections.
	dbPool, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	// [START_EXCLUDE]
	configureConnectionPool(dbPool)
	// [END_EXCLUDE]

	return dbPool, nil
	// [END cloud_sql_mysql_databasesql_create_tcp]
}

// configureConnectionPool sets database connection pool properties.
// For more information, see https://golang.org/pkg/database/sql
func configureConnectionPool(dbPool *sql.DB) {
	// [START cloud_sql_mysql_databasesql_limit]

	// Set maximum number of connections in idle connection pool.
	dbPool.SetMaxIdleConns(5)

	// Set maximum number of open connections to the database.
	dbPool.SetMaxOpenConns(7)

	// [END cloud_sql_mysql_databasesql_limit]

	// [START cloud_sql_mysql_databasesql_lifetime]

	// Set Maximum time (in seconds) that a connection can remain open.
	dbPool.SetConnMaxLifetime(1800)

	// [END cloud_sql_mysql_databasesql_lifetime]
}
