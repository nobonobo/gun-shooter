package schema

import "math"

type Info struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Fire bool    `json:"fire"`
}

type Point struct {
	X, Y float64
}

// Pointのヘルパー関数
func (p Point) Add(q Point) Point {
	return Point{p.X + q.X, p.Y + q.Y}
}

func (p Point) Sub(q Point) Point {
	return Point{p.X - q.X, p.Y - q.Y}
}

func (p Point) Scale(f float64) Point {
	return Point{p.X * f, p.Y * f}
}

func (p Point) Length() float64 {
	return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

func (p Point) Dot(q Point) float64 {
	return p.X*q.X + p.Y*q.Y
}

func (p Point) Normalize() Point {
	l := p.Length()
	if l == 0 {
		return Point{0, 0}
	}
	return p.Scale(1 / l)
}

func (p Point) Dist(q Point) float64 {
	return math.Sqrt(math.Pow(p.X-q.X, 2) + math.Pow(p.Y-q.Y, 2))
}
