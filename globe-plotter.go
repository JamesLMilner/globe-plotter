package main

import (
	"html/template"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/mmcloughlin/globe"
	geojson "github.com/paulmach/go.geojson"
)

func main() {

	http.Handle("/", http.FileServer(http.Dir("./static/")))
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":8080", nil)

}

var templates = template.Must(template.ParseFiles("static/globe.html"))

func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl+".html", data)
}

func createImage(filename string, uploadUrl string) {

	// url := "./static/example/earthquake.geojson"
	// data, err := LoadFeatureCollection(url)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	geojson, err := LoadFeatureCollection(uploadUrl)
	if err != nil {
		log.Fatal(err)
	}

	g := globe.New()
	g.DrawGraticule(20.0)
	g.DrawLandBoundaries()
	//g.DrawCountryBoundaries()
	green := color.RGBA{100, 254, 0, 100}
	for _, geometries := range geojson.Features {
		if geometries.Geometry.IsPoint() {
			log.Println("Point")
			coords := geometries.Geometry.Point
			// Lat, Lng, something, color
			g.DrawDot(coords[0], coords[1], 0.01, globe.Color(green))
		}
	}
	g.CenterOn(40.645423, -73.903879)

	err = g.SavePNG("./static/generated/"+filename+".png", 800)
	if err != nil {
		log.Fatal(err)
	}
}

//This is where the action happens.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		display(w, "upload", nil)

	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		//parse the multipart form in the request
		err := r.ParseMultipartForm(100000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		uuid := r.FormValue("uuid")
		var uploadUrl string
		//get a ref to the parsed multipart form
		m := r.MultipartForm

		//get the *fileheaders
		files := m.File["geojson"]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			uploadUrl = "./uploads/" + files[i].Filename
			dst, err := os.Create(uploadUrl)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}

		createImage(uuid, uploadUrl)
		//display success message.
		//display(w, "upload", "Upload successful.")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func LoadFeatureCollection(inputFilepath string) (*geojson.FeatureCollection, error) {
	b, err := ioutil.ReadFile(inputFilepath)
	if err != nil {
		return nil, err
	}

	return geojson.UnmarshalFeatureCollection(b)
}
