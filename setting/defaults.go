package setting

var (
	default_server        = []string{"wms", "tms", "wmts", "mapbox"}
	default_image_formats = []string{"image/png", "image/jpeg", "image/gif", "image/GeoTIFF", "image/tiff"}
	default_srs           = []string{"EPSG:4326", "EPSG:4258", "CRS:84", "EPSG:900913", "EPSG:3857"}
)

const (
	default_strict            = false
	default_max_output_pixels = 4000 * 4000
)

var (
	default_axis_order_ne = []string{"EPSG:4326", "EPSG:4258", "EPSG:31466", "EPSG:31467", "EPSG:31468"}
	default_axis_order_en = []string{"CRS:84", "EPSG:900913", "EPSG:25831", "EPSG:25832", "EPSG:25833"}
)

const (
	default_resampling_method           = "bicubic"
	default_default_jpeg_quality        = 90
	stretch_factor                      = 1.15
	default_max_shrink_factor           = 4.0
	default_paletted                    = true
	default_transparent_color_tolerance = 5
)

const (
	default_max_tile_limit           = 500
	default_concurrent_tile_creators = 2
	default_meta_buffer              = 80
	default_minimize_meta_requests   = false
)

var (
	default_meta_size = [2]int{4, 4}
	default_tile_size = [2]uint32{256, 256}
)

var (
	default_grids = []GridOpts{
		{Srs: "EPSG:4326", Origin: "sw", Name: "GLOBAL_GEODETIC"},
		{Srs: "EPSG:900913", Origin: "sw", Name: "GLOBAL_MERCATOR"},
		{Srs: "EPSG:3857", Origin: "nw", Name: "GLOBAL_WEBMERCATOR"},
	}
)

var (
	default_expires_hours               = 72
	default_request_timeout             = 60
	default_access_control_allow_origin = "*"
)
