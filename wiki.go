package main

import (
  "html/template"
  "io/ioutil"
  "net/http"
  "regexp"
  "strings"
)

type Page struct {
    Title string
    Body []byte
    Html template.HTML
}

var templates = template.Must(template.ParseFiles("views/edit.html", "views/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9\\_]+)$")

func (p *Page) save() error {
  filename := "pages/" + p.Title + ".txt"
  return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
  filename := "pages/" + title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  html := createHtmlTemplate(body)
  return &Page{Title: title, Body: body, Html: html}, nil
}

func createHtmlTemplate(body []byte) (template.HTML) {
  rxp := regexp.MustCompile("\\[(.*?)\\]")
  new_body := rxp.ReplaceAllFunc(body, linkHandler)
  return template.HTML(new_body)
}

func linkHandler(link []byte) []byte {
  str := string(link)
  str = strings.Trim(str, "[")
  str = strings.Trim(str, "]")

  a := "<a href=\"/view/" + str +"\">" + str + "</a>"
  return []byte(a)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }

    fn(w, r, m[2])
  }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
 http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/" + title, http.StatusFound)
    return
  }

  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func main() {
  http.HandleFunc("/", indexHandler)
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  http.ListenAndServe(":3000", nil)
}
