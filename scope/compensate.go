package main

import "math"

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

type Marker struct {
	Point
	Detected bool
}

func compensateMarkers(markers [4]Marker) [4]Point {
	// 検出されていないマーカーのインデックスを特定
	missing := [4]bool{}
	validPoints := [4]Point{}
	validCount := 0

	for i, m := range markers {
		if !m.Detected {
			missing[i] = true
		} else {
			validPoints[validCount] = m.Point
			validCount++
		}
	}

	// 欠損数に応じた処理
	switch validCount {
	case 3: // 1つ欠損：3点で長方形の第4点を計算
		return compensateOneMissing(markers, missing, validPoints)

	case 2: // 2つ欠損：対角線チェック
		return compensateDiagonalMissing(markers, missing, validPoints)

	default:
		// 条件に該当しない場合はすべて有効として返す
		var result [4]Point
		for i, m := range markers {
			result[i] = m.Point
		}
		return result
	}
}

func compensateOneMissing(markers [4]Marker, missing [4]bool, valid [4]Point) [4]Point {
	// 3点から第4点を計算（対角線和が等しい性質を利用）
	// P0 + P2 = P1 + P3 の関係を利用

	var result [4]Point
	for i, m := range markers {
		if !m.Detected {
			// 欠損位置を特定
			missingIdx := i
			// 対角線上のもう一方の点を探す
			diagIdx := (missingIdx + 2) % 4

			sum := valid[0].Add(valid[1])
			if diagIdx != 0 && diagIdx != 1 {
				sum = sum.Sub(valid[0])
			}

			// 残りの有効点から計算
			p0 := Point{0, 0}
			p1 := Point{0, 0}
			count := 0
			for j := 0; j < 4; j++ {
				if j != missingIdx && markers[j].Detected {
					if count == 0 {
						p0 = markers[j].Point
					} else {
						p1 = markers[j].Point
					}
					count++
				}
			}

			// 長方形の性質：対角線の交点が中心
			result[i] = p0.Add(p1).Sub(valid[0]) // 簡易計算
		} else {
			result[i] = m.Point
		}
	}
	return result
}

func compensateDiagonalMissing(markers [4]Marker, missing [4]bool, valid [4]Point) [4]Point {
	// 2つの有効点が対角線か隣接かをチェック
	p0, p1 := valid[0], valid[1]

	// 距離が長い方を対角線と仮定
	dist01 := p0.Sub(p1).Length()

	// 対角線として扱える場合のみ補完
	if dist01 > 100 { // 画面サイズに応じた閾値
		var result [4]Point

		// 対角線の中心から長方形を再構築（簡易版）
		center := p0.Add(p1).Scale(0.5)

		// 長方形の向きを推定し、残り2点を計算
		dx, dy := p1.X-p0.X, p1.Y-p0.Y

		halfLen := math.Sqrt(dx*dx+dy*dy) / 2

		// 簡易的に中心周囲に配置（実際は傾きを考慮）
		result[0] = center.Add(Point{-halfLen / 2, -halfLen / 2})
		result[1] = center.Add(Point{halfLen / 2, -halfLen / 2})
		result[2] = center.Add(Point{halfLen / 2, halfLen / 2})
		result[3] = center.Add(Point{-halfLen / 2, halfLen / 2})

		return result
	}

	// 対角線として扱えない場合はすべて有効として返す
	var result [4]Point
	for i, m := range markers {
		result[i] = m.Point
	}
	return result
}

func calc(points [4]Point, w, h float64) (x, y float64) {
	center := Point{w / 2, h / 2}

	// P0, P1, P3 を基底にしてバリセン座標 (u, v) を解く
	p0 := points[0]
	p1 := points[1]
	p3 := points[3]

	a := p1.Sub(p0)     // v0
	b := p3.Sub(p0)     // v1
	c := center.Sub(p0) // C - P0

	// 2x2 の連立方程式を内積で解く (a,b が一次独立なとき)
	aa := a.Dot(a)
	ab := a.Dot(b)
	bb := b.Dot(b)
	ac := a.Dot(c)
	bc := b.Dot(c)

	denom := aa*bb - ab*ab
	if denom == 0 {
		// 退化している場合はとりあえず 0,0 にしておく
		return 0, 0
	}

	u := (ac*bb - bc*ab) / denom
	v := (aa*bc - ac*ab) / denom

	// そのまま返すと、指定どおり:
	// P0一致 → (0,0), P1一致 → (1,0), P2一致 → (1,1), P3一致 → (0,1)
	return u, v
}
