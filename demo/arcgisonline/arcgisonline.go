package main

var (
	url           = "https://services.arcgisonline.com/arcgis/rest/services/{map}/{server}/tile/0/0/0"
	service_maps  = []string{"NatGeo_World_Map", "USA_Topo_Maps", "World_Imagery", "World_Physical_Map", "World_Shaded_Relief", "World_Street_Map", "World_Terrain_Base", "World_Topo_Map", "Elevation/World_Hillshade_Dark", "Elevation/World_Hillshade", "Ocean/World_Ocean_Base", "Ocean/World_Ocean_Reference"}
	maps_server   = "MapServer"
	matrixSet     = "EPSG:3857"
	center        = []float64{-11158582, 4813697}
	centerSrs     = "EPSG:3857"
	lerc_maps     = []string{"WorldElevation3D/Terrain3D", "WorldElevation3D/TopoBathy3D"}
	lerc_server   = "ImageServer"
	lerc_maxerror = 0.1
)
