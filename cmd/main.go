package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
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

		md := goldmark.New()

		var buf bytes.Buffer
		if err := md.Convert([]byte(postMarkdown), &buf); err != nil {
			panic(err)
		}

		// c.HTML(http.StatusOK, "post", gin.H{
		// 	"Title":    slug,
		// 	"Markdown": buf.String(),
		// })
		res := map[string]any{
			"ParsedMarkdown": buf.String(),
		}

		c.HTML(http.StatusOK, "post", res)
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
