package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type PaletteColor struct {
	Name string `json:"name"`
	Hex  string `json:"hex"`
	RGB  [3]uint8 `json:"rgb"`
}

type Grid struct {
	Schema     string                 `json:"schema"`
	Name       string                 `json:"name"`
	Source     string                 `json:"source,omitempty"`
	Width      int                    `json:"width"`
	Height     int                    `json:"height"`
	Coordinate map[string]string      `json:"coordinate"`
	Palette    map[string]PaletteColor `json:"palette"`
	Cells      [][]int                `json:"cells"`
}

type ConvertRequest struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

var palette = map[string]PaletteColor{
	"1": {"black", "#000000", [3]uint8{0, 0, 0}},
	"2": {"white", "#ffffff", [3]uint8{255, 255, 255}},
	"3": {"light-yellow", "#f8f76e", [3]uint8{248, 247, 110}},
	"4": {"yellow", "#f9e601", [3]uint8{249, 230, 1}},
	"5": {"red", "#ee3d52", [3]uint8{238, 61, 82}},
	"6": {"brown", "#9e7216", [3]uint8{158, 114, 22}},
	"7": {"gold", "#d8a800", [3]uint8{216, 168, 0}},
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/v1/palette", paletteHandler)
	mux.HandleFunc("/v1/convert", convertHandler)
	mux.HandleFunc("/v1/pattern", patternHandler)

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	addr := ":" + port
	log.Printf("amimono API listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, cors(logging(mux))))
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "amimono-api"})
}

func paletteHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, palette)
}

func patternHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { methodNotAllowed(w); return }
	path := "patterns/35x53-stitch-chart.json"
	f, err := os.Open(path)
	if err != nil { http.Error(w, "pattern not found", http.StatusNotFound); return }
	defer f.Close()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	io.Copy(w, f)
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { methodNotAllowed(w); return }

	width, height := 35, 53
	if v := r.URL.Query().Get("width"); v != "" { width, _ = strconv.Atoi(v) }
	if v := r.URL.Query().Get("height"); v != "" { height, _ = strconv.Atoi(v) }
	if width < 1 || width > 1000 || height < 1 || height > 1000 {
		http.Error(w, "width/height must be between 1 and 1000", http.StatusBadRequest); return
	}

	var src io.Reader = r.Body
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(20 << 20); err != nil { http.Error(w, err.Error(), 400); return }
		file, _, err := r.FormFile("image")
		if err != nil { http.Error(w, "multipart field 'image' is required", 400); return }
		defer file.Close(); src = file
	}

	img, _, err := image.Decode(src)
	if err != nil { http.Error(w, "unsupported or invalid image: "+err.Error(), http.StatusBadRequest); return }

	cells := segment(img, width, height)
	name := r.URL.Query().Get("name")
	if name == "" { name = fmt.Sprintf("%dx%d stitch chart", width, height) }
	grid := Grid{
		Schema: "amimono/pixel-grid/v1", Name: name, Width: width, Height: height,
		Coordinate: map[string]string{"origin":"top-left", "rows":"1-based top-to-bottom", "columns":"1-based left-to-right"},
		Palette: palette, Cells: cells,
	}
	writeJSON(w, http.StatusOK, grid)
}

func segment(img image.Image, width, height int) [][]int {
	b := img.Bounds()
	out := make([][]int, height)
	for row := 0; row < height; row++ {
		out[row] = make([]int, width)
		for col := 0; col < width; col++ {
			// Sample the central 60% of each cell to avoid grid lines and labels.
			x0 := b.Min.X + int(float64(col)*float64(b.Dx())/float64(width))
			x1 := b.Min.X + int(float64(col+1)*float64(b.Dx())/float64(width))
			y0 := b.Min.Y + int(float64(row)*float64(b.Dy())/float64(height))
			y1 := b.Min.Y + int(float64(row+1)*float64(b.Dy())/float64(height))
			padX, padY := max(1, (x1-x0)/5), max(1, (y1-y0)/5)
			var sr, sg, sb float64; var n float64
			for y := y0+padY; y < y1-padY; y++ { for x := x0+padX; x < x1-padX; x++ {
				r, g, bb, _ := img.At(x,y).RGBA()
				sr += float64(r >> 8); sg += float64(g >> 8); sb += float64(bb >> 8); n++
			} }
			if n == 0 { out[row][col] = 2; continue }
			out[row][col] = nearestPalette(sr/n, sg/n, sb/n)
		}
	}
	return out
}

func nearestPalette(r,g,b float64) int {
	best, bestD := 1, math.MaxFloat64
	for k, c := range palette {
		id, _ := strconv.Atoi(k)
		d := sq(r-float64(c.RGB[0])) + sq(g-float64(c.RGB[1])) + sq(b-float64(c.RGB[2]))
		if d < bestD { bestD, best = d, id }
	}
	return best
}
func sq(x float64) float64 { return x*x }
func max(a,b int) int { if a>b{return a}; return b }

func writeJSON(w http.ResponseWriter, status int, v any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); json.NewEncoder(w).Encode(v) }
func methodNotAllowed(w http.ResponseWriter) { w.Header().Set("Allow", "GET, POST"); http.Error(w, "method not allowed", http.StatusMethodNotAllowed) }
func logging(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ log.Printf("%s %s",r.Method,r.URL.Path); next.ServeHTTP(w,r) }) }
func cors(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){ w.Header().Set("Access-Control-Allow-Origin","*"); w.Header().Set("Access-Control-Allow-Headers","Content-Type"); if r.Method==http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }; next.ServeHTTP(w,r) }) }
