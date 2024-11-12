package main

import (
    "os"
    "fmt"
    "math"
    "net/http"
    "image"
    "image/color"
    "image/png"
    "github.com/quartercastle/vector"
)

type vec = vector.Vector

func look(cam vec, target vec, uv vec) vec {
    fwd := target.Sub(cam).Unit()
    right, _ := fwd.Cross(vec{0,1,0}.Unit())
    up, _ := fwd.Cross(right)
    return fwd.Sum(right.Scale(uv.X())).Sum(up.Scale(uv.Y())).Unit()
}

func Vabs(p vec) vec {
    return vec{math.Abs(p.X()), math.Abs(p.Y()), math.Abs(p.Z())}
}

func Rect(p vec, d vec) vec {
    return vec{max(max(math.Abs(p.X())-d.X(),math.Abs(p.Y())-d.Y()), math.Abs(p.Z())-d.Z())}
}

func Max(a, b vec) vec {
    if a.X() > b.X() {
        return a
    } else {
        return b
    }
}

func gradient(p vec) vec {
    var e float64 = 0.005
    return vec{
        sdf(p.Sum(vec{e,0,0})).X() - sdf(p.Sub(vec{e,0,0})).X(),
        sdf(p.Sum(vec{0,e,0})).X() - sdf(p.Sub(vec{0,e,0})).X(),
        sdf(p.Sum(vec{0,0,e})).X() - sdf(p.Sub(vec{0,0,e})).X(),
    }.Unit()
}

func sdf(p vec) vec {

    // for i := 0; i < 6; i++ {
    //     p = Vabs(p).Sub(vec{0.1,0.2,0.3}.Scale(1.0))
    //     p = p.Rotate(math.Pi/3.12421, vector.X)
    //     p = p.Rotate(math.Pi/5.23153, vector.Z)
    // }
    // return Rect(p, vec{1.0,2.0,3.0}.Scale(0.06))

    for i:=0;i<9;i++ {
        p = Vabs(p).Sub(vec{0.1,0.2,0.3})
        p = p.Rotate(1.1, vector.X)
        p = p.Rotate(0.7, vector.Z)
    }
    return Max(Rect(p, vec{1.0,1.0,1.0}.Scale(0.7)), vec{-(p.Magnitude()-0.8)})
}

func march(origin vec, ray vec) vec {
    rayLength := float64(0)
    hit := float64(0)
    i := 0
    for i = 0; i < 70; i++ {
        distance := sdf(origin.Sum(ray.Scale(rayLength))).X()
        rayLength += distance*0.9;
        if(distance < 0.01) {
            hit = 1
            break
        }
    }
    return vec{rayLength, hit}
}

func calcPixel(x,y,width,height int) color.RGBA {
    aspectDiv := min(width,height)
    uv := vec{float64(x-(width/2)), float64(y-(height/2))}.Scale(1.0/float64(aspectDiv))

    cam := vec{-1,-1.5,-2}.Unit().Scale(5)
    target := vec{0,0,0}
    ray := look(cam, target, uv.Scale(2))
    result := march(cam, ray)
    distance := result.X()
    brightness := 0.0
    hit := result.Y() > 0
    if hit {
        // brightness = 5.0/math.Pow(1.8, distance)
        brightness = gradient(cam.Sum(ray.Scale(distance))).Dot(vec{0.6,0.4,-0.9}.Unit())*0.6 + 0.4
    }

    shade := uint8(min(1,max(0,brightness))*255)
    alpha := uint8(0xFF)

    return color.RGBA{shade, shade, shade, alpha}
}

type PixelRow struct {
    rowNum int
    pixelData []color.RGBA
}

func calcPixelRow(rowNum int, width int, height int, c chan PixelRow) {
    data := make([]color.RGBA, width, width)
    for x := 0; x < width; x++ {
        data[x] = calcPixel(x, rowNum, width, height)
    }
    c <- PixelRow{rowNum, data}
}

func createImage() {
    width := 1024/2
    height := 768/2

    upLeft := image.Point{0, 0}
    lowRight := image.Point{width, height}

    img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

    c := make(chan PixelRow)
    for y := 0; y<height; y++ {
        go calcPixelRow(y, width, height, c)
    }

    pixelRowMap := make(map[int][]color.RGBA)

    for y := 0; y<height; y++ {
        var currentRow PixelRow = <- c
        pixelRowMap[currentRow.rowNum] = currentRow.pixelData
    }

    // Set color for each pixel.
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            //img.Set(x,y, calcPixel(x,y,width,height));
            img.Set(x,y, pixelRowMap[y][x]);
        }
    }
    
    // Encode as PNG.
    f, _ := os.Create("static/image.png")
    png.Encode(f, img)
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
    })
    http.HandleFunc("/genimage", func(w http.ResponseWriter, r *http.Request) {
        createImage()
        http.ServeFile(w, r, "static/image.png")
    })
    fs := http.FileServer(http.Dir("static/"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    http.ListenAndServe(":80", nil)
}