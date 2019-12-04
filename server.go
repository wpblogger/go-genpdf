package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	wk "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gorilla/mux"
)

type Cfg struct {
	PageSize      string
	Orientation   string
	MarginBottom  int
	MarginTop     int
	MarginLeft    int
	MarginRight   int
	PageShrinking bool
	PageZoom      float64
}

func getPdfFile(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var Buf bytes.Buffer
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Print(err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")
	log.Print("Input file name ", name[0], ".", name[1])
	io.Copy(&Buf, file)

	cfg := &Cfg{
		PageSize:      "A4",
		Orientation:   "Landscape",
		MarginBottom:  10,
		MarginTop:     10,
		MarginLeft:    10,
		MarginRight:   10,
		PageShrinking: true,
		PageZoom:      1,
	}

	if len(r.FormValue("page_size")) > 0 {
		cfg.PageSize = r.FormValue("page_size")
	}
	if len(r.FormValue("orientation")) == 0 {
		cfg.Orientation = r.FormValue("orientation")
	}
	if len(r.FormValue("margin_bottom")) > 0 {
		cfg.MarginBottom, err = strconv.Atoi(r.FormValue("margin_bottom"))
		if err != nil {
			log.Print(err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
	}
	if len(r.FormValue("margin_top")) > 0 {
		cfg.MarginTop, err = strconv.Atoi(r.FormValue("margin_top"))
		if err != nil {
			log.Print(err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
	}
	if len(r.FormValue("margin_left")) > 0 {
		cfg.MarginLeft, err = strconv.Atoi(r.FormValue("margin_left"))
		if err != nil {
			log.Print(err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
	}
	if len(r.FormValue("margin_right")) > 0 {
		cfg.MarginRight, err = strconv.Atoi(r.FormValue("margin_right"))
		if err != nil {
			log.Print(err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
	}
	if r.FormValue("shrink") == "1" {
		cfg.PageShrinking = false
	}
	if len(r.FormValue("zoom")) > 0 {
		cfg.PageZoom, err = strconv.ParseFloat(r.FormValue("zoom"), 64)
		if err != nil {
			log.Print(err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
	}

	pdfg, err := wk.NewPDFGenerator()
	if err != nil {
		log.Print("PDFGeneratorError: " + err.Error())
		http.Error(w, "application error", http.StatusInternalServerError)
		return
	}

	pdfg.PageSize.Set(cfg.PageSize)
	pdfg.Orientation.Set(cfg.Orientation)
	pdfg.MarginBottom.Set(uint(cfg.MarginBottom))
	pdfg.MarginTop.Set(uint(cfg.MarginTop))
	pdfg.MarginLeft.Set(uint(cfg.MarginLeft))
	pdfg.MarginRight.Set(uint(cfg.MarginRight))

	page := wk.NewPageReader(bytes.NewReader(Buf.Bytes()))
	page.DisableSmartShrinking.Set(cfg.PageShrinking)
	page.Zoom.Set(cfg.PageZoom)

	pdfg.AddPage(page)
	pdfg.Dpi.Set(600)

	err = pdfg.Create()
	if err != nil {
		log.Print("RunError: " + err.Error())
		http.Error(w, "application error", http.StatusInternalServerError)
		return
	}

	Buf.Reset()
	w.Header().Set("Content-Type", "application/pdf")
	fmt.Fprint(w, string(pdfg.Bytes()))
	log.Print("Pdf generated successfully, ", time.Since(start))
	return
}

func main() {
	listenPort := "8080"
	if len(os.Getenv("PORT")) > 0 {
		listenPort = os.Getenv("PORT")
	}
	log.Print("App start on port: ", listenPort)
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { getPdfFile(w, r) }).Methods("POST")
	log.Fatal(http.ListenAndServe(":"+listenPort, r))
}
