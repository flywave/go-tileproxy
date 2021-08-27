package gcj02

import (
	"math"
)

const earthR = 6378245.0
const earchHalfCir = 20037508.34

func outOfChina(lat, lng float64) bool {
	if lng < 72.004 || lng > 137.8347 {
		return true
	}
	if lat < 0.8293 || lat > 55.8271 {
		return true
	}
	return false
}

func transform(x, y float64) (lat, lng float64) {
	xy := x * y
	absX := math.Sqrt(math.Abs(x))
	xPi := x * math.Pi
	yPi := y * math.Pi
	d := 20.0*math.Sin(6.0*xPi) + 20.0*math.Sin(2.0*xPi)

	lat = d
	lng = d

	lat += 20.0*math.Sin(yPi) + 40.0*math.Sin(yPi/3.0)
	lng += 20.0*math.Sin(xPi) + 40.0*math.Sin(xPi/3.0)

	lat += 160.0*math.Sin(yPi/12.0) + 320*math.Sin(yPi/30.0)
	lng += 150.0*math.Sin(xPi/12.0) + 300.0*math.Sin(xPi/30.0)

	lat *= 2.0 / 3.0
	lng *= 2.0 / 3.0

	lat += -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*xy + 0.2*absX
	lng += 300.0 + x + 2.0*y + 0.1*x*x + 0.1*xy + 0.1*absX

	return
}

func delta(lat, lng float64) (dLat, dLng float64) {
	const ee = 0.00669342162296594323
	dLat, dLng = transform(lng-105.0, lat-35.0)
	radLat := lat / 180.0 * math.Pi
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((earthR * (1 - ee)) / (magic * sqrtMagic) * math.Pi)
	dLng = (dLng * 180.0) / (earthR / sqrtMagic * math.Cos(radLat) * math.Pi)
	return
}

func WGS84toGCJ02(wgsLat, wgsLng float64) (gcjLat, gcjLng float64) {
	if outOfChina(wgsLat, wgsLng) {
		gcjLat, gcjLng = wgsLat, wgsLng
		return
	}
	dLat, dLng := delta(wgsLat, wgsLng)
	gcjLat, gcjLng = wgsLat+dLat, wgsLng+dLng
	return
}

func GCJ02toWGS84(gcjLat, gcjLng float64) (wgsLat, wgsLng float64) {
	if outOfChina(gcjLat, gcjLng) {
		wgsLat, wgsLng = gcjLat, gcjLng
		return
	}
	dLat, dLng := delta(gcjLat, gcjLng)
	wgsLat, wgsLng = gcjLat-dLat, gcjLng-dLng
	return
}

func GCJ02toWGS84Exact(gcjLat, gcjLng float64) (wgsLat, wgsLng float64) {
	const initDelta = 0.01
	const threshold = 1e-6
	dLat, dLng := initDelta, initDelta
	mLat, mLng := gcjLat-dLat, gcjLng-dLng
	pLat, pLng := gcjLat+dLat, gcjLng+dLng

	for true {
		wgsLat, wgsLng = (mLat+pLat)/2, (mLng+pLng)/2
		tmpLat, tmpLng := WGS84toGCJ02(wgsLat, wgsLng)
		dLat, dLng = tmpLat-gcjLat, tmpLng-gcjLng
		if math.Abs(dLat) < threshold && math.Abs(dLng) < threshold {
			return
		}
		if dLat > 0 {
			pLat = wgsLat
		} else {
			mLat = wgsLat
		}
		if dLng > 0 {
			pLng = wgsLng
		} else {
			mLng = wgsLng
		}
	}
	return
}

const (
	X_PI = math.Pi * 3000.0 / 180.0
)

func BD09toGCJ02(bdLat, bdLng float64) (gLat float64, gLng float64) {
	x := bdLng - 0.0065
	y := bdLat - 0.006

	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*X_PI)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*X_PI)

	gLng = z * math.Cos(theta)
	gLat = z * math.Sin(theta)

	return
}

func GCJ02toBD09(gcjLat, gcjLng float64) (bdLat float64, bdLng float64) {
	z := math.Sqrt(gcjLng*gcjLng+gcjLat*gcjLat) + 0.00002*math.Sin(gcjLat*X_PI)
	theta := math.Atan2(gcjLat, gcjLng) + 0.000003*math.Cos(gcjLng*X_PI)

	bdLng = z*math.Cos(theta) + 0.0065
	bdLat = z*math.Sin(theta) + 0.006

	return
}

func BD09toWGS84(bdLat, bdLng float64) (wgsLat float64, wgsLng float64) {
	lat, lng := BD09toGCJ02(bdLat, bdLng)
	return GCJ02toWGS84(lat, lng)
}

func BD09toWGS84Exact(bdLat, bdLng float64) (wgsLat float64, wgsLng float64) {
	lat, lng := BD09toGCJ02(bdLat, bdLng)
	return GCJ02toWGS84Exact(lat, lng)
}

func WGS84toBD09(wgsLat, wgsLng float64) (bdLat float64, bdLng float64) {
	lat, lng := WGS84toGCJ02(wgsLat, wgsLng)
	return GCJ02toBD09(lat, lng)
}

func Distance(latA, lngA, latB, lngB float64) float64 {
	pi180 := math.Pi / 180
	arcLatA := latA * pi180
	arcLatB := latB * pi180
	x := math.Cos(arcLatA) * math.Cos(arcLatB) * math.Cos((lngA-lngB)*pi180)
	y := math.Sin(arcLatA) * math.Sin(arcLatB)
	s := x + y
	if s > 1 {
		s = 1
	}
	if s < -1 {
		s = -1
	}
	alpha := math.Acos(s)
	distance := alpha * earthR
	return distance
}

func EPSG3857toWGS84(mercartorY, mercartorX float64) (wgsLat float64, wgsLng float64) {
	if !(mercartorX >= -earchHalfCir && mercartorX <= earchHalfCir) {
		return 0, 0
	}
	wgsLng = mercartorX / earchHalfCir * 180
	wgsLat = mercartorY / earchHalfCir * 180
	wgsLat = 180 / math.Pi * (2*math.Atan(math.Exp(wgsLat*math.Pi/180)) - math.Pi/2)
	return
}

func WGS84toEPSG3857(lat, lng float64) (mercartorY, mercartorX float64) {
	if !(lng >= -180 && lng <= 180 && lat >= -90 && lat <= 90) {
		return 0, 0
	}
	mercartorX = lng * earchHalfCir / 180
	mercartorY = math.Log(math.Tan((90+lat)*math.Pi/360)) / (math.Pi / 180)
	mercartorY = mercartorY * earchHalfCir / 180
	return
}

func GCJ02MCtoGCJ02(mercartorY, mercartorX float64) (wgsLat float64, wgsLng float64) {
	return BDMCtoBD09(mercartorY, mercartorX)
}

func GCJ02toGCJ02MC(lat, lng float64) (mercartorY, mercartorX float64) {
	return BD09toBDMC(lat, lng)
}

func GCJ02MCtoWGS84(mercartorY, mercartorX float64) (wgsLat float64, wgsLng float64) {
	if !(mercartorX >= -earchHalfCir && mercartorX <= earchHalfCir) {
		return 0, 0
	}
	wgsLat, wgsLng = GCJ02MCtoGCJ02(mercartorY, mercartorX)
	return GCJ02toWGS84(wgsLat, wgsLng)
}

func WGS84toGCJ02MC(lat, lng float64) (mercartorY, mercartorX float64) {
	if !(lng >= -180 && lng <= 180 && lat >= -90 && lat <= 90) {
		return 0, 0
	}
	lat, lng = WGS84toGCJ02(lat, lng)
	return GCJ02toGCJ02MC(lat, lng)
}

func GCJ02toBDMC(lat, lng float64) (mercartorY float64, mercartorX float64) {
	bdlat, bdlng := GCJ02toBD09(lat, lng)
	return BD09toBDMC(bdlat, bdlng)
}

func BDMCtoGCJ02MC(bdY, bdX float64) (mercartorY, mercartorX float64) {
	lat, lng := BDMCtoBD09(bdY, bdX)
	return BD09toGCJ02MC(lat, lng)
}

func GCJ02MCtoBDMC(gcjY, gcjX float64) (mercartorY, mercartorX float64) {
	lat, lng := GCJ02MCtoGCJ02(gcjY, gcjX)
	return GCJ02toBDMC(lat, lng)
}

func GCJ02MCtoBD09(gcjY, gcjX float64) (bdLat float64, bdLng float64) {
	lat, lng := GCJ02MCtoGCJ02(gcjY, gcjX)
	return GCJ02toBD09(lat, lng)
}

func BD09toGCJ02MC(bdY, bdX float64) (mercartorY, mercartorX float64) {
	lat, lng := BD09toGCJ02(bdY, bdX)
	return GCJ02toGCJ02MC(lat, lng)
}

var mcband = []float64{12890594.86, 8362377.87, 5591021, 3481989.83, 1678043.12, 0}
var mc2ll = [][]float64{
	{1.410526172116255e-8, 0.00000898305509648872, -1.9939833816331, 200.9824383106796, -187.2403703815547, 91.6087516669843, -23.38765649603339, 2.57121317296198, -0.03801003308653, 17337981.2},
	{-7.435856389565537e-9, 0.000008983055097726239, -0.78625201886289, 96.32687599759846, -1.85204757529826, -59.36935905485877, 47.40033549296737, -16.50741931063887, 2.28786674699375, 10260144.86},
	{-3.030883460898826e-8, 0.00000898305509983578, 0.30071316287616, 59.74293618442277, 7.357984074871, -25.38371002664745, 13.45380521110908, -3.29883767235584, 0.32710905363475, 6856817.37},
	{-1.981981304930552e-8, 0.000008983055099779535, 0.03278182852591, 40.31678527705744, 0.65659298677277, -4.44255534477492, 0.85341911805263, 0.12923347998204, -0.04625736007561, 4482777.06},
	{3.09191371068437e-9, 0.000008983055096812155, 0.00006995724062, 23.10934304144901, -0.00023663490511, -0.6321817810242, -0.00663494467273, 0.03430082397953, -0.00466043876332, 2555164.4},
	{2.890871144776878e-9, 0.000008983055095805407, -3.068298e-8, 7.47137025468032, -0.00000353937994, -0.02145144861037, -0.00001234426596, 0.00010322952773, -0.00000323890364, 826088.5},
}
var llband = []float64{75, 60, 45, 30, 15, 0}
var ll2mc = [][]float64{
	{-0.0015702102444, 111320.7020616939, 1704480524535203, -10338987376042340, 26112667856603880, -35149669176653700, 26595700718403920, -10725012454188240, 1800819912950474, 82.5},
	{0.0008277824516172526, 111320.7020463578, 647795574.6671607, -4082003173.641316, 10774905663.51142, -15171875531.51559, 12053065338.62167, -5124939663.577472, 913311935.9512032, 67.5},
	{0.00337398766765, 111320.7020202162, 4481351.045890365, -23393751.19931662, 79682215.47186455, -115964993.2797253, 97236711.15602145, -43661946.33752821, 8477230.501135234, 52.5},
	{0.00220636496208, 111320.7020209128, 51751.86112841131, 3796837.749470245, 992013.7397791013, -1221952.21711287, 1340652.697009075, -620943.6990984312, 144416.9293806241, 37.5},
	{-0.0003441963504368392, 111320.7020576856, 278.2353980772752, 2485758.690035394, 6070.750963243378, 54821.18345352118, 9540.606633304236, -2710.55326746645, 1405.483844121726, 22.5},
	{-0.0003218135878613132, 111320.7020701615, 0.00369383431289, 823725.6402795718, 0.46104986909093, 2351.343141331292, 1.58060784298199, 8.77738589078284, 0.37238884252424, 7.45},
}

func BDMCtoBD09(mercartorY, mercartorX float64) (gLat float64, gLng float64) {
	mercartorX, mercartorY = math.Abs(mercartorX), math.Abs(mercartorY)
	var f []float64
	for i := 0; i < len(mcband); i++ {
		if mercartorY >= mcband[i] {
			f = mc2ll[i]
			break
		}
	}
	if len(f) == 0 {
		for i := 0; i < len(mcband); i++ {
			if -mercartorY <= -mcband[i] {
				f = mc2ll[i]
				break
			}
		}
	}
	return convert(mercartorY, mercartorX, f)
}

func convert(lat, lng float64, f []float64) (tlat float64, tlng float64) {
	if len(f) == 0 {
		return 0, 0
	}
	tlng = f[0] + f[1]*math.Abs(lng)
	cc := math.Abs(lat) / f[9]

	for i := 0; i <= 6; i++ {
		tlat += (f[i+2] * math.Pow(cc, float64(i)))
	}

	if lng < 0 {
		tlng *= -1
	}
	if lat < 0 {
		tlat *= -1
	}
	return
}

func BD09toBDMC(lat, lng float64) (mercartorY float64, mercartorX float64) {
	lng = getLoop(lng, -180, 180)
	lat = getRange(lat, -74, 74)
	var f []float64
	for i := 0; i < len(llband); i++ {
		if lat >= llband[i] {
			f = ll2mc[i]
			break
		}
	}
	if len(f) > 0 {
		for i := len(llband) - 1; i >= 0; i-- {
			if lat <= -llband[i] {
				f = ll2mc[i]
				break
			}
		}
	}
	return convert(lat, lng, f)
}

func getLoop(lng, min, max float64) float64 {
	for lng > max {
		lng -= (max - min)
	}
	for lng < min {
		lng += (max - min)
	}
	return lng
}

func getRange(lat, min, max float64) float64 {
	if min != 0 {
		lat = math.Max(lat, min)
	}
	if max != 0 {
		lat = math.Min(lat, max)
	}
	return lat
}

func BDMCtoWGS84(mercartorY, mercartorX float64) (wgsLat float64, wgsLng float64) {
	gcjLat, gcjLng := BDMCtoGCJ02(mercartorY, mercartorX)
	return GCJ02toWGS84(gcjLat, gcjLng)
}

func WGS84toBDMC(lat, lng float64) (mercartorY float64, mercartorX float64) {
	gcjLat, gcjLng := WGS84toBD09(lat, lng)
	return BD09toBDMC(gcjLat, gcjLng)
}

func BDMCtoGCJ02(mercartorX, mercartorY float64) (float64, float64) {
	gcjLng, gcjLat := BDMCtoBD09(mercartorX, mercartorY)
	return BD09toGCJ02(gcjLng, gcjLat)
}
