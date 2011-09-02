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
		Text string
		Language string
		LanguageLower string
		UrlId string
}

const LENGTH = 8

func init() {
		http.HandleFunc("/", new_pasty);
		http.HandleFunc("/paste", handle_paste);
}

func write_error(writer io.Writer, err os.Error) {
		if(err == nil) {
				return;
		}
		fmt.Fprint(writer, "err: " + err.String())
}

func new_pasty(writer http.ResponseWriter, request *http.Request) {
		if(request.URL.Path == "/") {
			temp, err := template.ParseFile("html/index.html", nil);
			if(err != nil) {
					write_error(writer, err)
					return
			}
			temp.Execute(writer, &Page {})
		} else {
				var err os.Error
				path := request.URL.Path[1:]
				context := appengine.NewContext(request)
				query := datastore.NewQuery("page").Filter("UrlId =",path) 
				for t := query.Run(context);; {
						var page Page;
						_, err = t.Next(&page)
						if err == datastore.Done {
								break
						}
						if err != nil {
								write_error(writer, err)
								return
						}
						RenderPage(&page, writer)
				}
		}
}

func handle_paste(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm() 
		if(err != nil) {
				write_error(writer, err)
				return 
		}
	    text := request.FormValue("paste_text")	
		language := request.FormValue("language")

		var page *Page
		page, err = NewPage(text, language)

		if(err != nil) {
				write_error(writer, err)
				return
		}
		err = StorePage(page, request)
		if(err != nil) {
				write_error(writer, err)
				return
		}
		//RenderPage(page, writer)
		//fmt.Fprint(writer, page.UrlId)
		http.Redirect(writer, request, "/"+page.UrlId[0:LENGTH], http.StatusFound)
}

func StorePage(page *Page, request *http.Request) os.Error { 
		context := appengine.NewContext(request)
		_, err := datastore.Put(context, datastore.NewIncompleteKey("page"),
		page)
		if(err != nil) {
				return err
		}
		return nil
}

func RenderPage(page *Page, writer http.ResponseWriter) os.Error {
		temp, err := template.ParseFile("html/paste.html", nil)
		if(err != nil) {
				return err
		}
		temp.Execute(writer, page)
		return nil
}


func NewPage(text string, language string) (*Page, os.Error) {

		// TODO: We should have just one instance of the MD5 hash
		// and reuse it
		var md5hash hash.Hash = md5.New()
		md5hash.Write([]byte(text))
		var urlid string
		urlid, err := ByteToString(md5hash.Sum()) 
		if(err != nil) {
				return nil, err
		}
		page := &Page {Text:text,
					Language: language,
					LanguageLower: strings.ToLower(language),
					UrlId:urlid[0:LENGTH]}
		return page, nil
}

func ByteToString(bytes []byte) (string, os.Error) {
		var s = ""
		for _, v := range bytes {
				s = s + strconv.Uitob(uint(v), 16)
		}
		return s, nil
}

