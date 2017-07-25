package main

import (
	"encoding/csv"
	"encoding/json"
	"html/template"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	globe "github.com/mmcloughlin/globe"
	geojson "github.com/paulmach/go.geojson"
)

func main() {

	port := getPort()
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fs)
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(port, nil)
}

func getPort() string {
	var port string
	if os.Getenv("PORT") != "" {
		port = ":" + os.Getenv("PORT")
	} else {
		port = ":8080"
	}
	return port
}

var templates = template.Must(template.ParseFiles("static/index.html"))

func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl+".html", data)
}

func createImage(filename string, uploadPath string, rgbaColors rgba, longitude float64, latitude float64, fileType string) string {

	g := globe.New()
	color := color.RGBA{rgbaColors.R, rgbaColors.G, rgbaColors.B, rgbaColors.A}

	g.DrawGraticule(20.0)
	g.DrawLandBoundaries()
	//g.DrawCountryBoundaries()
	g.CenterOn(latitude, longitude)

	log.Println(fileType)
	if fileType == "geojson" {
		drawFromGeoJSON(uploadPath, g, color)
	} else if fileType == "csv" {
		drawFromCSV(uploadPath, g, color)
	}

	pngPath := "./static/generated/" + filename + ".png"
	err := g.SavePNG(pngPath, 800)
	if err != nil {
		log.Fatal(err)
	}

	return pngPath

}

// Delete a file some period of time in the future
func deleteFile(path string, seconds int) {

	wait := time.Second * 20
	timeout := make(chan error, 1)
	go func() {
		time.Sleep(wait)
		var err = os.Remove(path)
		timeout <- err
	}()

	select {
	case err := <-timeout:
		if err != nil {
			log.Println("Error deleting file", err)
		} else {
			log.Println("File deleted!")
		}

	}
}

// For packing values into
type rgba struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"`
}

// Return the RGBA color as a nice struct
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

		log.Println("Colors:", rgbaColors)
		log.Println("Latitude:", latitude)
		log.Println("Longitude:", longitude)

		var uploadPath string
		//get a ref to the parsed multipart form
		m := r.MultipartForm

		//get the *fileheaders
		files := m.File["geojson"]
		fileUploaded := files[0]

		isGeoJson := strings.Contains(fileUploaded.Filename, "geojson")
		isCsv := strings.Contains(fileUploaded.Filename, "csv")
		fileType := ""

		if isGeoJson {
			fileType = "geojson"
		} else if isCsv {
			fileType = "csv"
		}

		if fileType != "" {

			log.Println("Upload file type is ", fileType)

			//for each fileheader, get a handle to the actual file
			file, err := fileUploaded.Open()
			log.Println(file)

			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			//create destination file making sure the path is writeable.
			uploadPath = "./upload/" + uuid + "_" + fileUploaded.Filename

			dst, err := os.Create(uploadPath)
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

			log.Println("Upload successful: " + uploadPath)

			pngPath := createImage(uuid, uploadPath, rgbaColors, longitude, latitude, fileType)

			// Tidy up files
			go deleteFile(uploadPath, 1)
			go deleteFile(pngPath, 30)

		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func loadFeatureCollection(inputFilepath string) (*geojson.FeatureCollection, error) {
	b, err := ioutil.ReadFile(inputFilepath)
	if err != nil {
		return nil, err
	}

	return geojson.UnmarshalFeatureCollection(b)
}

func drawFromGeoJSON(filePath string, g *globe.Globe, globeColor color.RGBA) {
	geojson, err := loadFeatureCollection(filePath)
	if err != nil {
		log.Fatal(err)
	}

	for _, geometries := range geojson.Features {
		if geometries.Geometry.IsPoint() {
			coords := geometries.Geometry.Point
			// Lat, Lng, size, color
			drawDot(g, globeColor, coords[0], coords[1])

		}
	}
}

func drawDot(g *globe.Globe, globeColor color.RGBA, latitude float64, longitude float64) {
	size := 0.05
	g.DrawDot(latitude, longitude, size, globe.Color(globeColor))
}

func drawFromCSV(filePath string, g *globe.Globe, globeColor color.RGBA) {

	log.Println("drawFromCSV called")
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println("csv file opened called")

	// automatically call Close() at the end of current method
	defer file.Close()
	reader := csv.NewReader(file)

	reader.Comma = ';'
	lineCount := 0
	latitude := -1
	longitude := -1

	for {
		// read just one record, but we could ReadAll() as well
		record, err := reader.Read()
		//log.Println("Read CSV")
		// end-of-file is fitted into err
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error:", err)
			return
		}

		// record is an array of string so is directly printable
		//fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")

		for i := 0; i < len(record); i++ {

			if lineCount == 0 {

				fieldNames := strings.ToLower(record[i])
				fieldNamesSlice := strings.Split(fieldNames, ",")

				for index, fieldName := range fieldNamesSlice {

					if fieldName == "latitude" || fieldName == "lat" {
						latitude = index
					}
					if fieldName == "longitude" || fieldName == "lon" || fieldName == "lng" {
						longitude = index
					}

				}

			}

			if lineCount > 0 && latitude != -1 && longitude != -1 {

				fieldNames := strings.ToLower(record[i])
				fieldSlice := strings.Split(fieldNames, ",")

				latitudeVal, latErr := strconv.ParseFloat(fieldSlice[latitude], 64)
				longitudeVal, lonErr := strconv.ParseFloat(fieldSlice[longitude], 64)

				if latErr == nil && lonErr == nil {
					log.Println("Drawing CSV Dot", latitudeVal, longitudeVal)
					drawDot(g, globeColor, latitudeVal, longitudeVal)
				} else {
					log.Println("Error with latitude or longitude", lonErr, latErr)
				}

			}
		}

		lineCount++
	}

}
