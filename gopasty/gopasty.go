package gopasty

import (
	"fmt"
	"http"
	"template"
	"io"
	"os"
	"strings"
	"crypto/md5"
	"hash"
	"appengine"
	"appengine/datastore"
	"strconv"
)

type Page struct {
	Text          string
	Language      string
	LanguageLower string
	UrlId         string
}

const LENGTH = 8

func init() {
	http.HandleFunc("/", HandleHome)
	http.HandleFunc("/paste", HandleNewPaste)
}

func write_error(writer io.Writer, err os.Error) {
	if err == nil {
		return
	}
	fmt.Fprint(writer, "err: "+err.String())
}

func HandleHome(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/" {
		temp, err := template.ParseFile("html/index.html")
		if err != nil {
			write_error(writer, err)
			return
		}
		temp.Execute(writer, &Page{})
	} else {
		var err os.Error
		var is_raw = false
		var raw_prefix = "/raw/"
		var urlid string
		if strings.Contains(request.URL.Path, raw_prefix) {
			is_raw = true
			urlid = request.URL.Path[len(raw_prefix):]
		} else {
			urlid = request.URL.Path[1:]
		}
		page, err := GetPageFromDataStore(urlid, request)
		if err != nil {
			write_error(writer, err)
			return
		}
		if is_raw {
			RenderPageRaw(page, writer)
		} else {
			err = RenderPage(page, writer)
			if err != nil {
				write_error(writer, err)
				return
			}
		}
	}
}

func HandleNewPaste(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		write_error(writer, err)
		return
	}
	text := request.FormValue("paste_text")
	language := request.FormValue("language")

	var page *Page
	page, err = NewPage(text, language)

	if err != nil {
		write_error(writer, err)
		return
	}
	// check if we already have a page
	// and not store it again
	oldpage, err := GetPageFromDataStore(page.UrlId, request)
	if oldpage == nil {
		err = StorePage(page, request)
		if err != nil {
			write_error(writer, err)
			return
		}
	}
	http.Redirect(writer, request, "/"+page.UrlId[0:LENGTH], http.StatusFound)
}

func GetPageFromDataStore(urlid string, request *http.Request) (*Page, os.Error) {
	context := appengine.NewContext(request)
	query := datastore.NewQuery("page").Filter("UrlId =", urlid)
	var page Page
	for t := query.Run(context); ; {
		_, err := t.Next(&page)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	if page.UrlId == urlid {
		return &page, nil
	}

	return nil, nil
}

func StorePage(page *Page, request *http.Request) os.Error {
	context := appengine.NewContext(request)
	_, err := datastore.Put(context, datastore.NewIncompleteKey(context, "page", nil),
		page)
	if err != nil {
		return err
	}
	return nil
}

func RenderPage(page *Page, writer http.ResponseWriter) os.Error {
	temp, err := template.ParseFile("html/paste.html")
	if err != nil {
		return err
	}
	temp.Execute(writer, page)
	return nil
}

func RenderPageRaw(page *Page, writer http.ResponseWriter) os.Error {
	header := writer.Header()
	header.Set("Content-Type", "text")
	fmt.Fprint(writer, page.Text)
	return nil
}

func NewPage(text string, language string) (*Page, os.Error) {

	// TODO: We should have just one instance of the MD5 hash
	// and reuse it
	var md5hash hash.Hash = md5.New()
	md5hash.Write([]byte(text + language))
	var url_hash string
	url_hash, err := ByteToString(md5hash.Sum())
	if err != nil {
		return nil, err
	}
	page := &Page{Text: text,
		Language:      language,
		LanguageLower: strings.ToLower(language),
		UrlId:         url_hash[0:LENGTH]}
	return page, nil
}

func ByteToString(bytes []byte) (string, os.Error) {
	var s = ""
	for _, v := range bytes {
		s = s + strconv.Uitob(uint(v), 16)
	}
	return s, nil
}
