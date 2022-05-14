package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func ServeHTML(addr string, tp TreeProvider) error {
	handler := http.HandlerFunc(CreateHTTPHandler(tp))
	log.Printf("Listening on %v", addr)
	err := http.ListenAndServe(addr, handler)
	return err
}

func CreateHTTPHandler(tp TreeProvider) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(os.DirFS(tp.RootDir())))

	handler := func(w http.ResponseWriter, r *http.Request) {
		for _, h := range httpRouteHandlers {
			handled, err := h(w, r, fileServer, tp)
			if err != nil {
				log.Printf("ERROR during '%s %s': %v", r.Method, r.URL.Path, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if handled {
				return
			}
		}

		Render404Page(w, r)
	}

	return handler
}

type HTTPRouteHandler func(w http.ResponseWriter, r *http.Request, fileServer http.Handler, tp TreeProvider) (bool, error)

var httpRouteHandlers = []HTTPRouteHandler{
	TryServePage,
	TryServeFile,
}

func TryServeFile(w http.ResponseWriter, r *http.Request, fileServer http.Handler, tp TreeProvider) (bool, error) {
	path := filepath.Clean(r.URL.Path)
	path = filepath.Join(tp.RootDir(), path)

	if filepath.Ext(path) == ".html" {
		return false, nil
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	fileServer.ServeHTTP(w, r)
	return true, nil
}

func TryServePage(w http.ResponseWriter, r *http.Request, fileServer http.Handler, tp TreeProvider) (bool, error) {
	tree, err := tp.GetTree()
	if err != nil {
		return false, err
	}

	path := filepath.Clean(r.URL.Path)
	path = strings.ReplaceAll(path, "\\", "/")

	if path == "/" {
		RenderListPage(w, r, tree.Root)
		return true, nil
	}

	node := tree.Index[path]
	if node == nil {
		return false, nil
	}

	RenderContentPage(w, r, node)
	return true, nil
}

func RenderListPage(w http.ResponseWriter, r *http.Request, root *Node) {
	model, err := NewListPageModel(root)
	if err != nil {
		Render500Page(w, r)
		return
	}

	RenderTemplatePage(w, r, 200, "list.html", model)
}

func RenderContentPage(w http.ResponseWriter, r *http.Request, node *Node) {
	model, err := NewContentPageModel(node)
	if err != nil {
		Render500Page(w, r)
		return
	}

	RenderTemplatePage(w, r, 200, "content.html", model)
}

func Render404Page(w http.ResponseWriter, r *http.Request) {
	RenderTemplatePage(w, r, 404, "404.html", nil)
}

func Render500Page(w http.ResponseWriter, r *http.Request) {
	RenderTemplatePage(w, r, 500, "500.html", nil)
}

func RenderTemplatePage(w http.ResponseWriter, r *http.Request, status int, name string, model interface{}) {
	t, err := template.ParseFiles(filepath.Join(".", "www", name))
	if err != nil {
		log.Printf("Unable to load template '%s': %v", name, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	w.WriteHeader(status)
	w.Header().Set("content-type", "text/html")
	err = t.Execute(w, model)
	if err != nil {
		log.Printf("Unable to render template '%s': %v", name, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

type ListPageModel struct {
	Nav []*NavItem
}

type NavItem struct {
	Title string
	URL   string
	Depth int
}

func NewListPageModel(root *Node) (*ListPageModel, error) {
	model := &ListPageModel{}

	err := root.Walk(func(n *Node, depth int) error {
		item := &NavItem{
			Title: n.Title,
			URL:   n.ID,
			Depth: depth,
		}
		model.Nav = append(model.Nav, item)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return model, nil
}

type ContentPageModel struct {
	Title       string
	Content     string
	Breadcrumbs []*BreadcrumbItem
}

type BreadcrumbItem struct {
	Title    string
	URL      string
	IsActive bool
}

func NewContentPageModel(node *Node) (*ContentPageModel, error) {
	content, err := node.Read()
	if err != nil {
		log.Printf("unable to read '%v': %v", node.Path, err)
		return nil, err
	}

	model := &ContentPageModel{
		Title:   node.Title,
		Content: content,
	}

	n := node
	for n != nil {
		item := &BreadcrumbItem{
			Title:    n.Title,
			URL:      n.ID,
			IsActive: n == node,
		}
		model.Breadcrumbs = append(model.Breadcrumbs, item)
		n = n.Parent
	}

	for i := 0; i < len(model.Breadcrumbs)/2; i++ {
		j := len(model.Breadcrumbs) - i - 1
		model.Breadcrumbs[i], model.Breadcrumbs[j] = model.Breadcrumbs[j], model.Breadcrumbs[i]
	}

	return model, nil
}
