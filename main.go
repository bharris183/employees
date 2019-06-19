package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"employees/db"
	"employees/employees"

	"github.com/fatih/structs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var tmpls = template.Must(template.ParseGlob("templates/*.tmpl"))
var database *sql.DB

func main() {
	var err error
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/",
		http.FileServer(http.Dir("templates/css/"))))

	r.HandleFunc("/add", AddEmployee)
	r.HandleFunc("/list", ListEmployees)
	r.HandleFunc("/employee/{id}", UserDetail)
	r.HandleFunc("/delete/{id}", DeleteEmployee)
	r.HandleFunc("/", RenderHome)

	database, err = db.CreateDatabase()
	if err != nil {
		fmt.Println("Error Dr. Smith")
	}

	fmt.Println("We're listening")
	//http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", r))

	defer database.Close()
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "Hi")
	return
}

func ListEmployees(w http.ResponseWriter, r *http.Request) {

	var employeeList []employees.Employee

	rows, err := database.Query("SELECT * from employees")
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Users not found")
		} else {
			panic(err)
		}
	}

	for rows.Next() {
		var e employees.Employee
		var id int
		var lname, fname, position, department string
		err := rows.Scan(&id, &lname, &fname, &position, &department)
		if err != nil {
			if err == sql.ErrNoRows {
				// no such user id
			} else {
				panic(err)
			}
		} else {
			e = employees.Employee{Id: id, LastName: lname, FirstName: fname, Position: position, Department: department}
		}
		employeeList = append(employeeList, e)
	}

	err = tmpls.ExecuteTemplate(w, "userlist.tmpl", employeeList)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer rows.Close()
}

func UserDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]
	var e employees.Employee

	rows, err := database.Query("SELECT * from employees where user_id = " + userId)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found")
		} else {
			panic(err)
		}
	}

	for rows.Next() {
		var id int
		var lname, fname, position, department string
		err := rows.Scan(&id, &lname, &fname, &position, &department)
		if err != nil {
			if err == sql.ErrNoRows {
				// no such user id
				e = employees.Employee{Id: -1, LastName: "", FirstName: "", Position: "", Department: ""}
			} else {
				panic(err)
			}
		} else {
			e = employees.Employee{Id: id, LastName: lname, FirstName: fname, Position: position, Department: department}
		}
	}
	s := structs.Map(e)
	fmt.Println(s)

	err = tmpls.ExecuteTemplate(w, "userdetail.tmpl", s)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		//return

	}
	defer rows.Close()
}

func RenderHome(w http.ResponseWriter, r *http.Request) {
	var i interface{}
	err := tmpls.ExecuteTemplate(w, "index.tmpl", i)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AddEmployee(w http.ResponseWriter, r *http.Request) {
	var lastName, firstName, position, department string
	r.ParseForm()
	lastName = r.FormValue("lastname")
	firstName = r.FormValue("firstname")
	position = r.FormValue("position")
	department = r.FormValue("department")
	var q string
	q = "INSERT INTO employees (Last_Name, First_Name, Position, Department) VALUES ('"
	q += lastName + "','" + firstName + "','" + position + "','" + department + "')"
	fmt.Println(q)
	rows, err := database.Query(q)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Zero rows found")
		} else {
			panic(err)
		}
	}
	ListEmployees(w, r)
	defer rows.Close()
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["id"]

	cmdDelete, err := database.Prepare("DELETE from employees where user_id = ?")
	if err != nil {
		panic(err.Error)
	}
	cmdDelete.Exec(userId)

	ListEmployees(w, r)
}
