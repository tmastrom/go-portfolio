package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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
			newHeader("Contact", "/contact"),
			newHeader("Blog", "/blog"),
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

		c.HTML(http.StatusOK, "post", gin.H{
			"Title":    slug,
			"Markdown": postMarkdown,
		})
	}
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
		c.HTML(http.StatusOK, "blog", page)
	})

	router.GET("/blog/:post", PostHandler(fr))

	router.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact", page)
	})

	router.Run(":8080")
}
