package tianditu

var (
	tk               = "e1b7133224d795928165a0148b33a079"
	url              = "http://t0.tianditu.com/{layer}/wmts?service=wmts&request=GetTile&version=1.0.0&LAYER=vec&tileMatrixSet=w&TileMatrix={TileMatrix}&TileRow={TileRow}&TileCol={TileCol}&style=default&format=tiles"
	layers_EPSG_3857 = []string{"cva_w", "cia_w", "img_w", "vec_w", "ter_w", "cta_w"}
	layers_EPSG_4490 = []string{"cva_c", "cia_c", "img_c", "vec_c", "ter_c", "cta_c"}
)
