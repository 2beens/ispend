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

	err := viewsMaker.reloadViews()
	if err != nil {
		return nil, err
	}

	return viewsMaker, nil
}

func (vm *ViewsMaker) reloadViews() error {
	layoutFiles := []string{
		vm.viewsDir + "/layouts/layout.html",
		vm.viewsDir + "/layouts/footer.html",
		vm.viewsDir + "/layouts/navbar.html",
		vm.viewsDir + "/layouts/sidebar.html",
	}

	viewFileNames, err := getViewFileNames(vm.viewsDir)
	if err != nil {
		return err
	}

	for _, v := range viewFileNames {
		viewPath := vm.viewsDir + v
		t, err := template.New("layout").ParseFiles(append(layoutFiles, viewPath)...)
		if err != nil {
			return err
		}
		log.Infof(" > read template view file: " + viewPath)
		vm.templatesMap[v] = t
	}

	return nil
}

func (vm *ViewsMaker) RenderView(w http.ResponseWriter, page string, viewData interface{}) {
	err := vm.reloadViews()
	if err != nil {
		log.Error(err.Error())
	}

	t, ok := vm.templatesMap[page+".html"]
	if !ok {
		log.Error(" >>> error rendering view, cannot find view template: " + page + ".html")
		http.Error(w, "internal server error (error rendering view)", http.StatusInternalServerError)
	}

	err = t.ExecuteTemplate(w, "layout", viewData)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (vm *ViewsMaker) RenderIndex(w http.ResponseWriter) {
	vm.RenderView(w, "index", nil)
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
