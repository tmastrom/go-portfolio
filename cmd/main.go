package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

type Header struct {
	Id   int
	Name string
	Path string
}

var id int = 0

func newHeader(name, path string) Header {
	id++
	return Header{
		Id:   id,
		Name: name,
		Path: path,
	}
}

type Headers = []Header

type HeaderData struct {
	Headers Headers
}

func newHeaderData() HeaderData {
	return HeaderData{
		Headers: []Header{
			newHeader("Blog", "/blog"),
			newHeader("Contact", "/contact"),
		},
	}
}

type Page struct {
	HeaderData HeaderData
	PageName   string
}

func newPage() Page {
	return Page{
		HeaderData: newHeaderData(),
	}
}

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct{}

func (fr FileReader) Read(slug string) (string, error) {
	f, err := os.Open("posts/" + slug + ".md")
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	// cast byte[] to string
	return string(b), nil
}

func PostHandler(sl SlugReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("post")
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			c.Status(404)
			return
		}

		md := goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
			),
		)

		var buf bytes.Buffer
		context := parser.NewContext()

		if err := md.Convert([]byte(postMarkdown), &buf, parser.WithContext(context)); err != nil {
			panic(err)
		}

		metaData := meta.Get(context)
		title := metaData["title"]

		c.HTML(http.StatusOK, "blog-post", gin.H{
			"Title":    title,
			"Markdown": template.HTML(buf.String()),
		})
	}
}

func RemoveExtensionFromFilename(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

func ListFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var filenames []string
	for _, v := range entries {
		if v.IsDir() {
			continue
		}
		if filepath.Ext(v.Name()) != ".md" {
			continue
		}
		filenames = append(filenames, RemoveExtensionFromFilename(v.Name()))
	}
	return filenames
}

func main() {
	router := gin.New()
	router.Use(gin.Logger())

	router.Static("/images", "images")
	router.Static("/css", "css")

	router.LoadHTMLGlob("views/*")

	page := newPage()

	fr := FileReader{}

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", page)
	})

	router.GET("/blog", func(c *gin.Context) {
		files := ListFiles("posts")
		c.HTML(http.StatusOK, "blog", files)
	})

	router.GET("/blog/:post", PostHandler(fr))

	router.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact", page)
	})

	router.Run(":8080")
}
