package main

import (
	"encoding/json"
	"html/template"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/mmcloughlin/globe"
	geojson "github.com/paulmach/go.geojson"
)

func main() {

	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fs)
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":8080", nil)
}

var templates = template.Must(template.ParseFiles("static/globe.html"))

func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl+".html", data)
}

func createImage(filename string, uploadUrl string, rgbaColors rgba, longitude float64, latitude float64) {

	geojson, err := LoadFeatureCollection(uploadUrl)
	if err != nil {
		log.Fatal(err)
	}

	g := globe.New()
	g.DrawGraticule(20.0)
	g.DrawLandBoundaries()
	//g.DrawCountryBoundaries()
	color := color.RGBA{rgbaColors.R, rgbaColors.G, rgbaColors.B, rgbaColors.A}

	for _, geometries := range geojson.Features {
		if geometries.Geometry.IsPoint() {
			//log.Println("Point")
			coords := geometries.Geometry.Point
			// Lat, Lng, something, color
			g.DrawDot(coords[0], coords[1], 0.02, globe.Color(color))
		}
	}

	g.CenterOn(latitude, longitude)

	err = g.SavePNG("./static/generated/"+filename+".png", 800)
	if err != nil {
		log.Fatal(err)
	}
}

type rgba struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

func getRgbaColor(rgbaStr string) rgba {

	in := []byte(rgbaStr)
	var raw rgba
	err := json.Unmarshal(in, &raw)
	if err != nil {
		log.Println(err)
	}
	return raw

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
		rgbaValue := r.FormValue("rgba")
		rgbaColors := getRgbaColor(rgbaValue)

		latitude, err := strconv.ParseFloat(r.FormValue("latitude"), 64)
		if err != nil {
			latitude = 0.0
		}
		longitude, err := strconv.ParseFloat(r.FormValue("longitude"), 64)
		if err != nil {
			longitude = 0.0
		}

		log.Println(rgbaValue)
		log.Println(rgbaColors)
		log.Println(latitude)
		log.Println(longitude)

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
			uploadUrl = "./upload/" + files[i].Filename
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

		createImage(uuid, uploadUrl, rgbaColors, longitude, latitude)
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
