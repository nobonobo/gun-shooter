package ui

import (
	"math"
	"time"

	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/game"
	"github.com/nobonobo/gun-shooter/schema"
)

type ActiveMember struct {
	Time        time.Time
	Info        *schema.Info
	Score       int
	Calibration [4]schema.Point
	Calibrated  int
}

type GlobalState struct {
	AudioAPI    audio.API
	Engine      *game.Engine
	ResourceSet *game.ResourceSet
	Actives     map[string]ActiveMember
}

// Calibrate はキャリブレーション4点を用いてバイリニア逆変換で座標を補正する。
// am.Calibration[0..3] は TL, TR, BR, BL のターゲットを狙った際の生マーカー座標。
// raw は補正対象の生座標。戻り値は補正後の正規化座標 (0-1)。
func (am *ActiveMember) Calibrate() schema.Point {
	raw := schema.Point{X: am.Info.X, Y: am.Info.Y}
	// キャリブレーションターゲットの既知位置（正規化座標）
	// TL(0.25,0.25), TR(0.75,0.25), BR(0.75,0.75), BL(0.25,0.75)
	dst := [4]schema.Point{
		{X: 0.25, Y: 0.25},
		{X: 0.75, Y: 0.25},
		{X: 0.75, Y: 0.75},
		{X: 0.25, Y: 0.75},
	}

	// Newton法でバイリニア逆変換: calib四角形内の (u,v) を求める
	// bilinear(u,v) = (1-u)(1-v)*P0 + u(1-v)*P1 + uv*P2 + (1-u)v*P3
	u, v := 0.5, 0.5
	for i := 0; i < 20; i++ {
		// 残差
		fx := (1-u)*(1-v)*am.Calibration[0].X + u*(1-v)*am.Calibration[1].X + u*v*am.Calibration[2].X + (1-u)*v*am.Calibration[3].X - raw.X
		fy := (1-u)*(1-v)*am.Calibration[0].Y + u*(1-v)*am.Calibration[1].Y + u*v*am.Calibration[2].Y + (1-u)*v*am.Calibration[3].Y - raw.Y

		if math.Abs(fx) < 1e-10 && math.Abs(fy) < 1e-10 {
			break
		}

		// ヤコビアン
		dxdu := -(1-v)*am.Calibration[0].X + (1-v)*am.Calibration[1].X + v*am.Calibration[2].X - v*am.Calibration[3].X
		dxdv := -(1-u)*am.Calibration[0].X - u*am.Calibration[1].X + u*am.Calibration[2].X + (1-u)*am.Calibration[3].X
		dydu := -(1-v)*am.Calibration[0].Y + (1-v)*am.Calibration[1].Y + v*am.Calibration[2].Y - v*am.Calibration[3].Y
		dydv := -(1-u)*am.Calibration[0].Y - u*am.Calibration[1].Y + u*am.Calibration[2].Y + (1-u)*am.Calibration[3].Y

		det := dxdu*dydv - dxdv*dydu
		if math.Abs(det) < 1e-12 {
			break
		}

		du := -(fx*dydv - fy*dxdv) / det
		dv := -(dxdu*fy - dydu*fx) / det
		u += du
		v += dv
	}

	// (u,v) → ターゲット空間にマッピング
	rx := (1-u)*(1-v)*dst[0].X + u*(1-v)*dst[1].X + u*v*dst[2].X + (1-u)*v*dst[3].X
	ry := (1-u)*(1-v)*dst[0].Y + u*(1-v)*dst[1].Y + u*v*dst[2].Y + (1-u)*v*dst[3].Y

	return schema.Point{X: rx, Y: ry}
}
