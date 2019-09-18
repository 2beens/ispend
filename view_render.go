package ispend

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ViewsMaker struct {
	viewsDir     string
	templatesMap map[string]*template.Template
}

func NewViewsMaker(viewsDir string) (*ViewsMaker, error) {
	viewsMaker := &ViewsMaker{
		viewsDir:     viewsDir,
		templatesMap: make(map[string]*template.Template),
	}

	layoutFiles := []string{
		viewsDir + "/layouts/layout.html",
		viewsDir + "/layouts/footer.html",
		viewsDir + "/layouts/navbar.html",
	}

	viewFileNames, err := getViewFileNames(viewsDir)
	if err != nil {
		return nil, err
	}

	for _, v := range viewFileNames {
		viewPath := viewsDir + v
		t, err := template.New("layout").ParseFiles(append(layoutFiles, viewPath)...)
		if err != nil {
			return nil, err
		}
		log.Infof(" > read template view file: " + viewPath)
		viewsMaker.templatesMap[v] = t
	}

	return viewsMaker, nil
}

func (vm *ViewsMaker) RenderView(w http.ResponseWriter, page string, viewData interface{}) {
	t, ok := vm.templatesMap[page+".html"]
	if !ok {
		log.Error(" >>> error rendering view, cannot find view template: " + page + ".html")
		http.Error(w, "internal server error (error rendering view)", http.StatusInternalServerError)
	}

	err := t.ExecuteTemplate(w, "layout", viewData)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type ErrorViewData struct {
	Title   string      `json:"title"`
	Message string      `json:"message"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}

func (vm *ViewsMaker) RenderErrorView(w http.ResponseWriter, username string, title string, status int, message string) {
	vm.RenderView(w, "error", ErrorViewData{Title: title, Error: fmt.Sprintf("Status: [%d]: %s", status, message)})
}

func getViewFileNames(viewsDir string) ([]string, error) {
	var viewFileNames []string
	viewFiles, err := ioutil.ReadDir("./" + viewsDir)
	if err != nil {
		return viewFileNames, err
	}
	for _, f := range viewFiles {
		fileName := f.Name()
		if strings.HasSuffix(fileName, ".html") {
			viewFileNames = append(viewFileNames, fileName)
		}
	}
	return viewFileNames, nil
}
