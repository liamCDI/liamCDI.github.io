package main

//go:generate gopherjs build ./... -o js/ascartgjs.js --minify

import (
	"bytes"
	"html/template"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	ascart "github.com/liamCDI/ascartok"
)

//ArtState hold the state of the app
type ArtState struct {
	canvasjq  jquery.JQuery
	formjq    jquery.JQuery
	infojq    jquery.JQuery
	Pic       []byte
	Pallete   string
	Tansforms string
}

//NewArtState returns a default art state
func NewArtState() ArtState {
	out := ArtState{}
	out.canvasjq = jquery.NewJQuery("#canvas")
	out.formjq = jquery.NewJQuery("#artform")
	out.infojq = jquery.NewJQuery("#info")

	//remove loading message
	out.infojq.SetHtml("")

	return out

}

func (a *ArtState) PrintErr(err error) {
	a.infojq.SetHtml(`<div class="alert alert-danger" role="alert">` + err.Error() + `</div>`)
}

//ReadImage reads a local image file
func (a *ArtState) ReadImage(o *js.Object) {
	go func() { a.Pic = readImage(o); a.Render() }()
}

func readImage(blob *js.Object) []byte {
	var b = make(chan []byte)
	fileReader := js.Global.Get("FileReader").New()
	fileReader.Set("onload", func() {
		b <- js.Global.Get("Uint8Array").New(fileReader.Get("result")).Interface().([]byte)
	})
	fileReader.Call("readAsArrayBuffer", blob)
	return <-b
}

//GetImage gets an image hostet on a CORS webpage like Imgur
func (a *ArtState) GetImage(img string) {
	go a.getImage(img)
}

func (a *ArtState) getImage(img string) {
	res, err := http.Get(img)
	if err != nil {
		a.PrintErr(err)
		log.Println(err)
		return
	}
	defer res.Body.Close()

	a.Pic, err = ioutil.ReadAll(res.Body)
	if err != nil {
		a.PrintErr(err)
		log.Println(err)
		return
	}

	a.Render()
}

//Render the ascii image
func (a *ArtState) Render() {
	tpl := `
	<pre>
	{{.}}
	</pre>
	`
	out, err := ascart.Bytes2asc(a.Pic, a.Pallete, a.Tansforms)
	if err != nil {
		a.PrintErr(err)
		log.Println(err)
		return
	}

	tmpl, err := template.New("Tmpl").Parse(tpl)
	if err != nil {
		a.PrintErr(err)
		log.Println(err)
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, "\n"+out.String())
	if err != nil {
		a.PrintErr(err)
		log.Println(err)
	}
	a.canvasjq.SetHtml(b.String())
}

func (a *ArtState) DoForm() {
	a.Pallete = a.formjq.Find("#asciiPal").Val()
	a.Tansforms = a.formjq.Find("#transf").Val()

	log.Println("Pallet: " + a.Pallete)
	log.Println("TRansform String:" + a.Tansforms)

	url := a.formjq.Find("#imgUrl").Val()

	file := a.formjq.Find("#imgFile").Call("prop", "files").Underlying().Index(0)

	if url == "" && file == js.Undefined {
		a.infojq.SetHtml(` <div class="alert alert-info" role="alert">
		Need either a file or a URL
	</div>`)
	}

	if file != js.Undefined {
		a.ReadImage(file)
	} else {
		a.GetImage(url)
	}
}

func (a *ArtState) RenderForm() {
	tmpl := `
	<form>
	<div class="form-group">
	  <label for="imgUrl">Imgur URL</label>
	  <input type="text" class="form-control" id="imgUrl" aria-describedby="imgUrlHelp" placeholder="Enter Url">
	  <small id="imgUrlHelp" class="form-text text-muted">URL of image hosted on Imgur</small>
	</div>
	<div class="form-group">
		<label for="imgFile">Local Image File</label>
		<input type="file" class="form-control-file" id="imgFile" aria-describedby="imgFileHelp">
		<small id="imgFileHelp" class="form-text text-muted">Local Image File</small>
	<div>
	<div class="form-group">
	  <label for="asciiPal">ASCII Pallete</label>
	  <input type="text" class="form-control" id="asciiPal" aria-describedby="asciiPalHelp" value=" .:OI M">
	  <small id="asciiPalHelp" class="form-text text-muted">ASCII Character used to 'paint' the image</small>
	</div>
	<div class="form-group">
	  <label for="transf">ASCII Pallete</label>
	  <input type="text" class="form-control" id="transf" aria-describedby="transfHelp" value="resize=50,0;">
	  <small id="transfHelp" class="form-text text-muted">Image Transform String</small>
	</div>
	<button type="button" id="renderBut" class="btn btn-primary">Render</button>
  </form>
	`
	a.formjq.SetHtml(tmpl)

	a.formjq.Off(jquery.CLICK, "#renderBut", a.DoForm).On(jquery.CLICK, "#renderBut", a.DoForm)

}

var state ArtState

func main() {
	state = NewArtState()

	state.RenderForm()

}
