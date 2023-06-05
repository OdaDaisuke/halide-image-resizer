//go:build arm64
// +build arm64

package main

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR} -ldl -lm
#cgo LDFLAGS: -lruntime
#cgo LDFLAGS: -lscale
#include "buffer.h"
#include "libscale.h"
int call_scale(unsigned char *in, int width, int height, int scale_width, int scale_height, unsigned char *out) {
  halide_buffer_t *in_buf = create_halide_buffer_rgba(in, width, height);
  halide_buffer_t *out_buf = create_halide_buffer_rgba(out, scale_width, scale_height);
  int ret = scale(in_buf, width, height, scale_width, scale_height, out_buf);
  free_halide_buffer(in_buf);
  free_halide_buffer(out_buf);
  return ret;
}
*/
import "C"

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"unsafe"
)

func Scale(in *image.RGBA, scaleWidth, scaleHeight int) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, scaleWidth, scaleHeight))

	ret := C.call_scale(
		(*C.uchar)(unsafe.Pointer(&in.Pix[0])),
		C.int(in.Rect.Dx()),
		C.int(in.Rect.Dy()),
		C.int(scaleWidth),
		C.int(scaleHeight),
		(*C.uchar)(unsafe.Pointer(&out.Pix[0])),
	)
	if int(ret) != 0 {
		panic("failed")
	}
	return out
}

func main() {
	log.Println("starting server")

	http.HandleFunc("/image", getImage)

	log.Println("app started on port 8080")
	http.ListenAndServe(":8080", nil)
}

func getImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()
	queryFile := queryParams.Get("f")
	if queryFile == "" {
		http.Error(w, "no `?f` specified", http.StatusInternalServerError)
		return
	}

	queryWidth := queryParams.Get("w")
	queryHeight := queryParams.Get("h")
	width, _ := strconv.Atoi(queryWidth)
	height, _ := strconv.Atoi(queryHeight)

	data, err := os.ReadFile(path.Join("assets", queryFile))
	if err != nil {
		panic(err)
	}

	src, err := pngToRGBA(data)
	if err != nil {
		panic(err)
	}

	out := Scale(src, width, height)

	if err := png.Encode(w, out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func pngToRGBA(data []byte) (*image.RGBA, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if i, ok := img.(*image.RGBA); ok {
		return i, nil
	}

	b := img.Bounds()
	rgba := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y += 1 {
		for x := b.Min.X; x < b.Max.X; x += 1 {
			c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			rgba.Set(x, y, c)
		}
	}
	return rgba, nil
}
