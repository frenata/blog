package blog

import (
	"bytes"
	"errors"
	"html/template"
	"strings"
	"time"

	"github.com/russross/blackfriday"
)

/* TODO: tweak Post struct to work as the implementation of an interface?
Something like: */

// Parse takes a slice of bytes (most likely from a file) and parses them into the data structure.
// Format takes a filename of a template and uses that template to format it's data structure to a new slice of bytes.
// String simply pretty prints the data held by the formatter.
type Formatter interface {
	Parse(b []byte, date time.Time) error
	Format(templateFile string) ([]byte, error)
	String() string
}

// TODO: These should either be part of the Post struct or live in a config file.
const (
	titlePre    string = "Title: "
	authorPre   string = "Author: "
	tagPre      string = "Tag: "
	datePre     string = "Date: "
	templatePre string = "Template: "
	dateFormat  string = "2 January 2006 @ 3:04pm"
)

// Post holds the content of a post, parsed from a file, metadata and body content.
type PostFormatter struct {
	Title      string
	Author     string
	Tag        string
	Body       template.HTML //string //[]byte
	date       time.Time
	Template   string
	DateFormat string
	//comments Comments // TODO: bool for disqus comments?
}

func NewPostFormatter() *PostFormatter {
	// initialize the constants
	return new(PostFormatter)
}

// PostFormatter takes a byte slice (from a markdowned text file), a date, and creates a new Post object.
// The date should typically be the last modification time of the file, and will be overwritten
// if a date is manually set in the Post metadata.
func (p *PostFormatter) Parse(b []byte, date time.Time) error {
	content := string(b)

	c := strings.SplitN(content, "Body:", 2)
	if len(c) < 2 {
		return errors.New("invalid post file - no content detected: ")
	}

	// TODO: does this need validation / error checking?
	bf := string(blackfriday.MarkdownCommon([]byte(c[1])))
	p.Body = template.HTML(strings.TrimSpace(bf))
	meta := strings.Split(c[0], "\n")

	p.DateFormat = dateFormat

	// new version of meta yanking: just grab each line, check if it validates any Meta tags, and assign it properly
	for _, m := range meta {
		if ok, out := p.validateMeta(m, titlePre); ok {
			p.Title = out
		} else if ok, out := p.validateMeta(m, authorPre); ok {
			p.Author = out
		} else if ok, out := p.validateMeta(m, tagPre); ok {
			p.Tag = out
		} else if ok, out := p.validateMeta(m, templatePre); ok {
			p.Template = out
		} else if ok, out := p.validateMeta(m, datePre); ok {
			p.setDate(out)
		} else if p.date.IsZero() {
			p.date = date
		}

	}

	return nil
}

// parses the string against the specified dateformat, if it validates, manually set the post date
func (p *PostFormatter) setDate(s string) {
	d, err := time.Parse(dateFormat, s)
	if err != nil {
		return
	}
	p.date = d
}

// checks the string for valid metadata, as defined in the constant prefixes, and returns the data.
func (p *PostFormatter) validateMeta(m, pre string) (ok bool, output string) {
	if strings.HasPrefix(m, pre) {
		ok, output = true, strings.TrimSpace(strings.TrimPrefix(m, pre))
	} else {
		ok, output = false, ""
	}

	return ok, output
}

// String prints the Post meta content and body.
func (p *PostFormatter) String() string {
	t := "Title: " + p.Title + "\n"
	a := "Author: " + p.Author + "\n"
	g := "Tag: " + p.Tag + "\n"
	d := "Date: " + p.Date() + "\n"

	return t + a + g + d + "\n" + string(p.Body)
}

// Format takes a template file and creates a []byte representing an html document populated with the Post content,
// ready for writing to a file.
func (p *PostFormatter) Format(templateFile string) ([]byte, error) {
	buf := &bytes.Buffer{} // byte buffer to use for template execution

	t, err := template.ParseFiles(templateFile) // load the template file
	if err != nil {
		return nil, err
	}

	err = t.Execute(buf, p) // add the post content into the template file
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil // write the buffer to a []byte prepped to write to a file.
}

// Format the date into configured readable format.
func (p *PostFormatter) Date() string {
	return p.date.Format(dateFormat)
}
