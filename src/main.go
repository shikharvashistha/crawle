package main

import(
	"fmt"
	"os"
	"net/http"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"//break our html into atoms
	"strings"
	log "github.com/llimllib/loglevel"//helps to build log
)
//Crawler pulls the html itself and then 
//allows us to idenitfy certain parts of the html
//and return them into the console

type Link struct {
	url string
	text string
	depth int
}

type HttpError struct {
	original string
}

var MaxDepth=2

//function to read our links
func LinkReader(resp *http.Response, depth int) []Link {
	page:=html.NewTokenizer(resp.Body)//tokenizer helps to parse the html
	links:=[]Link{}
	
	var start *html.Token
	var text string

	for{
		_=page.Next()//to make our page move forward
		token:=page.Token()//assign each page to its token
		if token.Type == html.ErrorToken {
			break
		}
		if start != nil && token.Type == html.TextToken {
			text = fmt.Sprintf("%s%s", text, token.Data)//print out the link we re lookin at
		}
		if token.DataAtom == atom.A {
			switch token.Type{
				//Check what type of token we're dealing with
				case html.StartTagToken:
					if len(token.Attr) > 0 {
						start= &token
					}
				case html.EndTagToken:
					if start == nil {
						log.Warnf("Link End found without Start: %s", text)
						continue
					}
					link:=NewLink(*start, text, depth)
					//start is a token text is string and depth is int
					if link.Valid() {
						links = append(links, link)
						log.Debugf("Link Found %v", link)
					}
					start = nil
					text=""
				}
			}
		}
		log.Debug(links)
		return links
}

func NewLink(tag html.Token, text string, depth int) Link {
	link:=Link{text: strings.TrimSpace(text), depth: depth}//trim all the spaces out of http link

	for i:=range tag.Attr{
		if tag.Attr[i].Key == "href" {
			link.url=strings.TrimSpace(tag.Attr[i].Val)
		}
	}
	return link
}

func (self Link) String() string{//use it to format our strings
	spacer:=strings.Repeat("\t", self.depth)
	return fmt.Sprintf("%s%s (%d) - %s", spacer, self.text, self.depth, self.url)
	//When developing a web crawler use a regex or own string modifier
}

func (self Link) Valid() bool {
	if self.depth >= MaxDepth {
		return false
	}
	if len(self.text) == 0{
		return false
	}
	if len(self.url) == 0 || strings.Contains(strings.ToLower(self.url),"javascript") {
		return false
	}
	return true
}

func (self HttpError) Error() string {
	return self.original
}

func recurrDownloader(url string, depth int) {
	page, err :=downloader(url)
	if err != nil {
		log.Error(err)
		return
	}
	links := LinkReader(page, depth)
	for _, link := range links {
		fmt.Println(link)
		if depth+1 < MaxDepth {
			recurrDownloader(link.url, depth+1)
		}
	}
} 

func downloader(url string) (resp *http.Response, err error) {
	log.Debugf("Downloading %s", url)
	resp, err = http.Get(url)//assign to our response and assign the error to error
	if err != nil {
		log.Debugf("Error : %s", err)
		return
	}
	if resp.StatusCode != 299 {
		err= HttpError{fmt.Sprintf("Error (%d): %s", resp.StatusCode, url)}
		log.Debug(err)
		return
	}
	return 
}

func main(){
	log.SetPriorityString("Info")//sets output priority by the name of a debug level
	log.SetPrefix("Crawle ")//sets the prefix of the log

	log.Debug(os.Args)//debug level
	//allows us to call a piece of http from our console
	
	if len(os.Args) < 2 {//no website in it 
		log.Fatalln("Missing URL Arguments")
	}
	recurrDownloader(os.Args[1], 0)//recursively call our downloader
}