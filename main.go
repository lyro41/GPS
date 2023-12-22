package main

import (
	"bufio"
	"fmt"
	"github.com/aclements/go-moremath/stats"
	"image/color"
	"math"
	"os"

	"github.com/adrianmo/go-nmea"
	dgStats "github.com/dgryski/go-onlinestats"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func Parse() []nmea.RMC {
	file, err := os.Open("input/20191121_ATGM336H_GNSS_Test.txt")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	result := make([]nmea.RMC, 0)
	for i := 0; scanner.Scan(); i++ {
		row := scanner.Text()
		s, err := nmea.Parse(row)
		if err != nil {
			fmt.Printf("row %v: %s\n", i+1, err)
			continue
		}
		if s.DataType() != nmea.TypeRMC {
			continue
		}
		result = append(result, s.(nmea.RMC))
	}
	return result
}

func test(xs, ys []float64, confidence float64) (test []bool) {
	Ex, Dx, Ey, Dy := dgStats.Mean(xs), dgStats.SampleStddev(xs), dgStats.Mean(ys), dgStats.SampleStddev(ys)
	r := dgStats.Pearson(xs, ys)
	fmt.Println("r: ", r)
	alpha := math.Atan(r * Dy / Dx)
	fmt.Println("alpha: ", alpha)

	newxs := make([]float64, 0, len(xs))
	newys := make([]float64, 0, len(ys))
	for i := range xs {
		newxs = append(newxs, (xs[i]-Ex)*math.Cos(alpha)+(ys[i]-Ey)*math.Sin(alpha))
		newys = append(newys, (ys[i]-Ey)*math.Cos(alpha)-(xs[i]-Ex)*math.Sin(alpha))
	}
	fmt.Println(newxs, newys)
	_, _, a := stats.MeanCI(newxs, confidence)
	_, _, b := stats.MeanCI(newys, confidence)
	fmt.Printf("a: %v\nb: %v\n", a, b)
	fmt.Println(newxs, newys)
	test = make([]bool, 0, len(xs))
	for i := range newxs {
		fmt.Println((newxs[i]/a)*(newxs[i]/a) + (newys[i]/b)*(newys[i]/b))
		test = append(test, (newxs[i]/a)*(newxs[i]/a)+(newys[i]/b)*(newys[i]/b) <= 1)
	}
	fmt.Println(test)
	return
}

func CITest() {
	confs := []float64{0.9999999, 0.99999999, 0.999999999, 0.999999999, 0.9999999999, 0.99999999999, 0.999999999999, 0.9999999999999}
	for _, conf := range confs {
		p := plot.New()
		p.X.Label.Text = "Latitude - 55.65342, deg"
		p.Y.Label.Text = "Longitude - 37.55196, deg"
		rmcs := Parse()
		xs := make([]float64, 0, len(rmcs))
		ys := make([]float64, 0, len(rmcs))
		for _, rmc := range rmcs {
			fmt.Println(rmc)
			xs = append(xs, rmc.Latitude-55.65342)
			ys = append(ys, rmc.Longitude-37.55196)
		}
		count := make(map[plotter.XY]int)
		reliability := make(map[plotter.XY]bool)
		confidence := test(xs, ys, conf)
		graphxys := make(plotter.XYs, 0, len(rmcs))
		for i := range rmcs {
			xy := plotter.XY{X: xs[i], Y: ys[i]}
			count[xy] = count[xy] + 1
			reliability[xy] = confidence[i]
			graphxys = append(graphxys, xy)
		}
		meanxy := make(plotter.XYs, 1)
		meanxy[0].X = dgStats.Mean(xs)
		meanxy[0].Y = dgStats.Mean(ys)
		xys := make(plotter.XYs, 0)
		for xy := range count {
			fmt.Println(xy)
			xys = append(xys, xy)
		}
		scatter := &plotter.Scatter{
			XYs: xys,
			GlyphStyleFunc: func(i int) draw.GlyphStyle {
				style := draw.GlyphStyle{
					Color:  color.RGBA{R: 255, A: 255},
					Radius: vg.Points(3 * float64(count[xys[i]])),
					Shape:  draw.CircleGlyph{},
				}
				if reliability[xys[i]] {
					style.Color = color.RGBA{G: 255, A: 255}
				}
				return style
			},
		}
		mscatter := &plotter.Scatter{
			XYs: meanxy,
			GlyphStyle: draw.GlyphStyle{
				Color:  color.RGBA{B: 255, A: 255},
				Radius: vg.Points(3),
				Shape:  draw.CircleGlyph{},
			},
		}
		graph, err := plotter.NewLine(&graphxys)
		if err != nil {
			panic(err)
		}
		p.Add(graph, scatter, mscatter)
		err = p.Save(1000, 1000, fmt.Sprintf("CI/%v.png", conf))
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	CITest()
}
