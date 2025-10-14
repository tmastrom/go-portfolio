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
			newHeader("Projects", "/blog"),
			// newHeader("Chess", "/chess"),
			newHeader("Contact", "/contact"),
		},
	}
}

type BlogData struct {
	FileNames []string
}

func newBlogData() BlogData {
	return BlogData{}
}

type BlogPostData struct {
	Markdown template.HTML
}

func newBlogPostData() BlogPostData {
	return BlogPostData{}
}

type Page struct {
	HeaderData   HeaderData
	BlogData     BlogData
	BlogPostData BlogPostData
}

func newPage() Page {
	return Page{
		HeaderData:   newHeaderData(),
		BlogData:     newBlogData(),
		BlogPostData: newBlogPostData(),
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
	return string(b), nil
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
	router.Static("/public", "public")
	router.Static("/css", "css")
	router.LoadHTMLGlob("templates/*.html")

	layouts := template.Must(template.ParseGlob("templates/layout/*.html"))
	profile := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/profile.html"))
	blog := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/blog.html"))
	blogPost := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/blog-post.html"))
	contact := template.Must(template.Must(layouts.Clone()).ParseFiles("templates/contact.html"))

	page := newPage()
	router.GET("/", func(c *gin.Context) {
		profile.ExecuteTemplate(c.Writer, "index", page)
	})

	router.GET("/blog", func(c *gin.Context) {
		files := ListFiles("posts")
		page.BlogData.FileNames = files
		if c.Request.Header["Hx-Request"] != nil {
			c.HTML(http.StatusOK, "blog-partial", files)
			return
		}
		blog.ExecuteTemplate(c.Writer, "index", page)
	})

	router.GET("/contact", func(c *gin.Context) {
		if c.Request.Header["Hx-Request"] != nil {
			c.HTML(http.StatusOK, "contact-partial", page)
			return
		}
		contact.ExecuteTemplate(c.Writer, "index", page)
	})

	fr := FileReader{}
	router.GET("/blog/:post", func(c *gin.Context) {
		slug := c.Param("post")
		postMarkdown, err := fr.Read(slug)
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

		// Get Metadata
		// TODO: get
		// metaData := meta.Get(context)
		// title := metaData["title"]
		if c.Request.Header["Hx-Request"] != nil {
			c.HTML(http.StatusOK, "blog-post", gin.H{"Markdown": template.HTML(buf.String())})
			return
		}
		page.BlogPostData = BlogPostData{Markdown: template.HTML(buf.String())}
		blogPost.ExecuteTemplate(c.Writer, "index", page)
	})

	router.Run(":8080")
}
