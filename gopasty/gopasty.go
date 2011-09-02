package gopasty
import (
	"fmt"
	"http"
	"template"
	"io"
	"os"
	"strings"
	"crypto/md5"
)

type Page struct {
		Text string
		Language string
		LanguageLower string
		UrlId string
}

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
		temp, err := template.ParseFile("html/index.html", nil);
		if(err != nil) {
				write_error(writer, err)
				return
		}
		temp.Execute(writer, &Page {})
}

func handle_paste(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm() 
		if(err != nil) {
				write_error(writer, err)
				return 
		}
	    text := request.FormValue("paste_text")	
		language := request.FormValue("language")

		// TODO: We should have just one instance of the MD5 hash
		// and reuse it
		var md5hash hash.Hash = md5.New()
		md5hash.Write([]byte(text))

		urlid := md5hash.Sum().String()[0:8]
		page := &Page {Text:text,
					Language: language,
					LanguageLower: strings.ToLower(language),
					UrlId:urlid}

		temp, err := template.ParseFile("html/paste.html", nil)
		if(err != nil) {
				write_error(writer, err)
				return
		}
		temp.Execute(writer, page)
}
