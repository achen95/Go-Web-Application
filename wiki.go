package main

import (
	"errors"
	"html/template"
	"io/ioutil" //to read from/write to files aka input/ouput
	"log"
	"net/http" //must be imported to work with web servers
	"regexp"
)

/*initializing global variable so that we only call
ParseFiles once at program initialization.
template.Must is used so that if the template cannot be
loaded, exit the program immediately.
*/
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

/*initializing global variable to validate title so arbitrary
path to server is less possible
regexp.mustCompile parses and compiles the expression and
exits immediately if there is failure in compilation*/
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

//a page has a title and a body part
type Page struct {
	Title string
	Body  []byte
}

//store data from pages into memory through a text file
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

//load a text file with content and sets a page's values
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

//the "view" page
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title) //load content from file
	//if there is an error loading the page
	if err != nil {
		//redirect to another page
		/*in this case, redirect to the "edit" page
		so content can be created.*/
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
	//fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

//the "edit" page
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title) //load content from file
	//if there is an error loading the content
	if err != nil {
		//set the page with the title passed in.
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

//saving the content submitted
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	//set a page
	/*FormValue gives a string, so it must be converted to
	//[]byte*/
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save() //write the content into a text file
	//if unsuccessful save
	if err != nil {
		//send specified http response code and error message
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//redirect to view page with the updated body
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

//set template and run it
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	//execute the specified template
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	//if error in execution
	if err != nil {
		//send specified http response code and error message
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//validate the path of the title
//NOTE: code here is made simpler in makeHandler
//so this function is not used
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	//see if the title is valid
	/*FindStringSubmatch returns slice holding text of
	leftmost match*/
	m := validPath.FindStringSubmatch(r.URL.Path)
	//if invalid
	if m == nil {
		//"404 Not Found" is written to http connection
		http.NotFound(w, r)
		//return the error
		return "", errors.New("Invalid Page Title")
	}
	/*if valid, return the title, which is in the second
	subexression*/
	return m[2], nil
}

/*wrapper function, which takes handler functions as
arguments and returns a function of type http.HandlerFunc*/
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	//function of type http.HandlerFunc to be returned
	return func(w http.ResponseWriter, r *http.Request) {
		//verify that title is valid
		m := validPath.FindStringSubmatch(r.URL.Path)
		//if not valid
		if m == nil {
			//"404 Not Found" is written to http connection
			http.NotFound(w, r)
			return
		}
		//if valid, call the handler function from parameter
		fn(w, r, m[2])
	}
}

func main() {
	/*initialize http using handlers to handle the requests
	under the paths of /view/, /save/, or /edit/*/
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	//log errors, and listen in on port 8080
	log.Fatal(http.ListenAndServe(":8080", nil))
}
