package resource

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"testing"
)

func TestTileJSON(t *testing.T) {
	tileJSON := &TileJSON{
		Type:            RASTER,
		Attribution:     "Test Attribution",
		Description:     "Test Description",
		Bounds:          [4]float32{-180, -85, 180, 85},
		Center:          [3]float32{0, 0, 5},
		Created:         1234567890,
		FileSize:        1024,
		FillZoom:        8,
		Format:          "png",
		ID:              "test-tileset",
		MaxZoom:         18,
		MinZoom:         0,
		Modified:        1234567890,
		Name:            "Test TileJSON",
		Scheme:          "xyz",
		TilejsonVersion: "2.2.0",
		Version:         "1.0.0",
		Tiles:           []string{"https://example.com/{z}/{x}/{y}.png"},
		Data:            []string{"https://example.com/data.json"},
		VectorLayers: []*VectorLayer{
			{
				Id:          "test-layer",
				Description: "Test vector layer",
				Maxzoom:     18,
				Minzoom:     0,
				Fileds:      map[string]string{"name": "string", "value": "number"},
				Source:      "test-source",
				SourceName:  "Test Source",
			},
		},
		Template: strPtr("<div>{{name}}</div>"),
		Legend:   strPtr("<div>Legend</div>"),
		Grids:    []string{"https://example.com/{z}/{x}/{y}.grid.json"},
		Webpage:  "https://example.com/docs",
		Location: "/tmp/test",
		Stored:   false,
		StoreID:  "test-store-id",
	}

	t.Run("GetExtension", func(t *testing.T) {
		if got := tileJSON.GetExtension(); got != "json" {
			t.Errorf("GetExtension() = %v, want %v", got, "json")
		}
	})

	t.Run("IsStored", func(t *testing.T) {
		if got := tileJSON.IsStored(); got != false {
			t.Errorf("IsStored() = %v, want %v", got, false)
		}
	})

	t.Run("SetStored", func(t *testing.T) {
		tileJSON.SetStored()
		if got := tileJSON.IsStored(); got != true {
			t.Errorf("IsStored() after SetStored() = %v, want %v", got, true)
		}
	})

	t.Run("GetFileName", func(t *testing.T) {
		if got := tileJSON.GetFileName(); got != "source" {
			t.Errorf("GetFileName() = %v, want %v", got, "source")
		}
	})

	t.Run("GetLocation", func(t *testing.T) {
		if got := tileJSON.GetLocation(); got != "/tmp/test" {
			t.Errorf("GetLocation() = %v, want %v", got, "/tmp/test")
		}
	})

	t.Run("SetLocation", func(t *testing.T) {
		tileJSON.SetLocation("/new/location")
		if got := tileJSON.GetLocation(); got != "/new/location" {
			t.Errorf("GetLocation() after SetLocation() = %v, want %v", got, "/new/location")
		}
	})

	t.Run("GetID", func(t *testing.T) {
		if got := tileJSON.GetID(); got != "test-store-id" {
			t.Errorf("GetID() = %v, want %v", got, "test-store-id")
		}
	})

	t.Run("SetID", func(t *testing.T) {
		tileJSON.SetID("new-id")
		if got := tileJSON.GetID(); got != "new-id" {
			t.Errorf("GetID() after SetID() = %v, want %v", got, "new-id")
		}
	})

	t.Run("Hash", func(t *testing.T) {
		expected := md5.Sum([]byte("new-id"))
		if got := tileJSON.Hash(); !bytes.Equal(got, expected[:]) {
			t.Errorf("Hash() = %v, want %v", got, expected[:])
		}
	})
}

func TestTileJSONSerialization(t *testing.T) {
	tileJSON := &TileJSON{
		Type:            VECTOR,
		Name:            "Test TileJSON",
		TilejsonVersion: "2.2.0",
		Version:         "1.0.0",
		Bounds:          [4]float32{-180, -85, 180, 85},
		Center:          [3]float32{0, 0, 5},
		Scheme:          "xyz",
		MinZoom:         0,
		MaxZoom:         18,
		Format:          "pbf",
	}

	t.Run("ToJson", func(t *testing.T) {
		data := tileJSON.ToJson()
		if len(data) == 0 {
			t.Error("ToJson() returned empty data")
		}

		// Verify JSON is valid
		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("ToJson() produced invalid JSON: %v", err)
		}

		if decoded["name"] != "Test TileJSON" {
			t.Errorf("Decoded name = %v, want %v", decoded["name"], "Test TileJSON")
		}
		if decoded["tilejson"] != "2.2.0" {
			t.Errorf("Decoded tilejson = %v, want %v", decoded["tilejson"], "2.2.0")
		}
	})

	t.Run("SetData", func(t *testing.T) {
		original := &TileJSON{}
		jsonData := tileJSON.ToJson()
		original.SetData(jsonData)

		if original.Name != tileJSON.Name {
			t.Errorf("SetData() failed to set Name = %v, want %v", original.Name, tileJSON.Name)
		}
		if original.Version != tileJSON.Version {
			t.Errorf("SetData() failed to set Version = %v, want %v", original.Version, tileJSON.Version)
		}
		if original.TilejsonVersion != tileJSON.TilejsonVersion {
			t.Errorf("SetData() failed to set TilejsonVersion = %v, want %v", original.TilejsonVersion, tileJSON.TilejsonVersion)
		}
	})

	t.Run("GetData", func(t *testing.T) {
		data := tileJSON.GetData()
		if len(data) == 0 {
			t.Error("GetData() returned empty data")
		}

		// Verify the data matches ToJson output
		expected := tileJSON.ToJson()
		if !bytes.Equal(data, expected) {
			t.Error("GetData() does not match ToJson() output")
		}
	})
}

func TestNewTileJSON(t *testing.T) {
	t.Run("CreateNewTileJSON", func(t *testing.T) {
		tileJSON := NewTileJSON("test-id", "Test Name")

		if tileJSON.StoreID != "test-id" {
			t.Errorf("StoreID = %v, want %v", tileJSON.StoreID, "test-id")
		}
		if tileJSON.Name != "Test Name" {
			t.Errorf("Name = %v, want %v", tileJSON.Name, "Test Name")
		}
		if tileJSON.Bounds != [4]float32{-180, -85, 180, 85} {
			t.Errorf("Bounds = %v, want %v", tileJSON.Bounds, [4]float32{-180, -85, 180, 85})
		}
		if tileJSON.Center != [3]float32{0, 0, 0} {
			t.Errorf("Center = %v, want %v", tileJSON.Center, [3]float32{0, 0, 0})
		}
		if tileJSON.Scheme != "xyz" {
			t.Errorf("Scheme = %v, want %v", tileJSON.Scheme, "xyz")
		}
	})
}

func TestCreateTileJSON(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		jsonStr := `{
			"tilejson": "2.2.0",
			"name": "Test TileJSON",
			"version": "1.0.0",
			"type": "vector",
			"format": "pbf",
			"bounds": [-180, -85, 180, 85],
			"center": [0, 0, 5],
			"minzoom": 0,
			"maxzoom": 18,
			"scheme": "xyz",
			"tiles": ["https://example.com/{z}/{x}/{y}.pbf"],
			"vector_layers": [
				{
					"id": "test-layer",
					"description": "Test layer",
					"minzoom": 0,
					"maxzoom": 18,
					"fields": {"name": "string"}
				}
			]
		}`

		tileJSON := CreateTileJSON([]byte(jsonStr))
		if tileJSON == nil {
			t.Fatal("CreateTileJSON() returned nil for valid JSON")
		}

		if tileJSON.Name != "Test TileJSON" {
			t.Errorf("Name = %v, want %v", tileJSON.Name, "Test TileJSON")
		}
		if tileJSON.TilejsonVersion != "2.2.0" {
			t.Errorf("TilejsonVersion = %v, want %v", tileJSON.TilejsonVersion, "2.2.0")
		}
		if tileJSON.Type != VECTOR {
			t.Errorf("Type = %v, want %v", tileJSON.Type, VECTOR)
		}
		if len(tileJSON.VectorLayers) != 1 {
			t.Errorf("VectorLayers count = %v, want %v", len(tileJSON.VectorLayers), 1)
		}
		if tileJSON.VectorLayers[0].Id != "test-layer" {
			t.Errorf("VectorLayer[0].Id = %v, want %v", tileJSON.VectorLayers[0].Id, "test-layer")
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := `{ invalid json }`
		tileJSON := CreateTileJSON([]byte(invalidJSON))
		if tileJSON != nil {
			t.Error("CreateTileJSON() should return nil for invalid JSON")
		}
	})

	t.Run("EmptyJSON", func(t *testing.T) {
		tileJSON := CreateTileJSON([]byte(""))
		if tileJSON != nil {
			t.Error("CreateTileJSON() should return nil for empty JSON")
		}
	})
}

func TestVectorLayer(t *testing.T) {
	t.Run("NewVectorLayer", func(t *testing.T) {
		layer := NewVectorLayer()
		if layer == nil {
			t.Fatal("NewVectorLayer() returned nil")
		}
		if layer.Fileds == nil {
			t.Error("NewVectorLayer() should initialize Fileds map")
		}
		if len(layer.Fileds) != 0 {
			t.Errorf("NewVectorLayer() Fileds length = %v, want 0", len(layer.Fileds))
		}
	})

	t.Run("VectorLayerFields", func(t *testing.T) {
		layer := &VectorLayer{
			Id:          "test-layer",
			Description: "Test layer",
			Maxzoom:     18,
			Minzoom:     0,
			Fileds: map[string]string{
				"name":  "string",
				"count": "number",
			},
			Source:     "test-source",
			SourceName: "Test Source",
		}

		if layer.Id != "test-layer" {
			t.Errorf("Id = %v, want %v", layer.Id, "test-layer")
		}
		if layer.Description != "Test layer" {
			t.Errorf("Description = %v, want %v", layer.Description, "Test layer")
		}
		if layer.Maxzoom != 18 {
			t.Errorf("Maxzoom = %v, want %v", layer.Maxzoom, 18)
		}
		if layer.Minzoom != 0 {
			t.Errorf("Minzoom = %v, want %v", layer.Minzoom, 0)
		}
		if layer.Source != "test-source" {
			t.Errorf("Source = %v, want %v", layer.Source, "test-source")
		}
		if layer.SourceName != "Test Source" {
			t.Errorf("SourceName = %v, want %v", layer.SourceName, "Test Source")
		}
	})

	t.Run("VectorLayerJSON", func(t *testing.T) {
		layer := &VectorLayer{
			Id:          "test-layer",
			Description: "Test layer",
			Maxzoom:     18,
			Minzoom:     0,
			Fileds:      map[string]string{"name": "string"},
			Source:      "test-source",
		}

		data, err := json.Marshal(layer)
		if err != nil {
			t.Fatalf("Failed to marshal VectorLayer: %v", err)
		}

		var decoded VectorLayer
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal VectorLayer: %v", err)
		}

		if decoded.Id != layer.Id {
			t.Errorf("Decoded Id = %v, want %v", decoded.Id, layer.Id)
		}
		if decoded.Description != layer.Description {
			t.Errorf("Decoded Description = %v, want %v", decoded.Description, layer.Description)
		}
	})
}

func TestTileJSONCache(t *testing.T) {
	mockStore := &mockTileStore{}
	cache := NewTileJSONCache(mockStore)

	t.Run("Save", func(t *testing.T) {
		tileJSON := &TileJSON{
			Name:    "test-tilejson",
			StoreID: "test-id",
		}

		err := cache.Save(tileJSON)
		if err != nil {
			t.Errorf("Save() returned error: %v", err)
		}

		if !mockStore.saved {
			t.Error("Save() did not call store.Save()")
		}
	})

	t.Run("Load", func(t *testing.T) {
		tileJSON := &TileJSON{
			StoreID: "test-id",
		}

		err := cache.Load(tileJSON)
		if err != nil {
			t.Errorf("Load() returned error: %v", err)
		}

		if !mockStore.loaded {
			t.Error("Load() did not call store.Load()")
		}
	})
}

func TestTileStats(t *testing.T) {
	t.Run("NewTileStats", func(t *testing.T) {
		tileStats := NewTileStats("test-tileset")
		if tileStats == nil {
			t.Fatal("NewTileStats() returned nil")
		}
		if tileStats.TilesetId != "test-tileset" {
			t.Errorf("TilesetId = %v, want %v", tileStats.TilesetId, "test-tileset")
		}
		if tileStats.Layers == nil {
			t.Error("NewTileStats() should initialize Layers slice")
		}
		if len(tileStats.Layers) != 0 {
			t.Errorf("NewTileStats() Layers length = %v, want 0", len(tileStats.Layers))
		}
	})

	t.Run("TileStatsMethods", func(t *testing.T) {
		tileStats := &TileStats{
			Account:    "test-account",
			TilesetId:  "test-tileset",
			LayerCount: 1,
			Location:   "/tmp/test",
			Stored:     false,
			StoreID:    "test-store-id",
			Layers: []*LayerAtrribute{
				{
					Account:   "test-account",
					TilesetId: "test-tileset",
					Layer:     "test-layer",
					Geometry:  "polygon",
					Count:     100,
					Attributes: []*Attribute{
						{
							Attr:   "name",
							Type:   "string",
							Values: []interface{}{"test1", "test2"},
						},
					},
				},
			},
		}

		if tileStats.GetExtension() != "json" {
			t.Errorf("GetExtension() = %v, want %v", tileStats.GetExtension(), "json")
		}

		if tileStats.IsStored() != false {
			t.Errorf("IsStored() = %v, want %v", tileStats.IsStored(), false)
		}

		tileStats.SetStored()
		if tileStats.IsStored() != true {
			t.Errorf("IsStored() after SetStored() = %v, want %v", tileStats.IsStored(), true)
		}

		if tileStats.GetFileName() != "tilestats" {
			t.Errorf("GetFileName() = %v, want %v", tileStats.GetFileName(), "tilestats")
		}

		if tileStats.GetLocation() != "/tmp/test" {
			t.Errorf("GetLocation() = %v, want %v", tileStats.GetLocation(), "/tmp/test")
		}

		tileStats.SetLocation("/new/location")
		if tileStats.GetLocation() != "/new/location" {
			t.Errorf("GetLocation() after SetLocation() = %v, want %v", tileStats.GetLocation(), "/new/location")
		}

		if tileStats.GetID() != "test-store-id" {
			t.Errorf("GetID() = %v, want %v", tileStats.GetID(), "test-store-id")
		}

		tileStats.SetID("new-id")
		if tileStats.GetID() != "new-id" {
			t.Errorf("GetID() after SetID() = %v, want %v", tileStats.GetID(), "new-id")
		}

		expected := md5.Sum([]byte("new-id"))
		if got := tileStats.Hash(); !bytes.Equal(got, expected[:]) {
			t.Errorf("Hash() = %v, want %v", got, expected[:])
		}
	})

	t.Run("TileStatsSerialization", func(t *testing.T) {
		tileStats := &TileStats{
			Account:    "test-account",
			TilesetId:  "test-tileset",
			LayerCount: 1,
			Layers: []*LayerAtrribute{
				NewLayerAtrribute("test-tileset", "test-layer"),
			},
		}

		t.Run("ToJson", func(t *testing.T) {
			data := tileStats.ToJson()
			if len(data) == 0 {
				t.Error("ToJson() returned empty data")
			}

			var decoded map[string]interface{}
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("ToJson() produced invalid JSON: %v", err)
			}

			if decoded["account"] != "test-account" {
				t.Errorf("Decoded account = %v, want %v", decoded["account"], "test-account")
			}
		})

		t.Run("SetData", func(t *testing.T) {
			original := &TileStats{}
			jsonData := tileStats.ToJson()
			original.SetData(jsonData)

			if original.Account != tileStats.Account {
				t.Errorf("SetData() failed to set Account = %v, want %v", original.Account, tileStats.Account)
			}
			if original.TilesetId != tileStats.TilesetId {
				t.Errorf("SetData() failed to set TilesetId = %v, want %v", original.TilesetId, tileStats.TilesetId)
			}
		})

		t.Run("GetData", func(t *testing.T) {
			data := tileStats.GetData()
			if len(data) == 0 {
				t.Error("GetData() returned empty data")
			}

			expected := tileStats.ToJson()
			if !bytes.Equal(data, expected) {
				t.Error("GetData() does not match ToJson() output")
			}
		})
	})

	t.Run("CreateTileStats", func(t *testing.T) {
		t.Run("ValidJSON", func(t *testing.T) {
			jsonStr := `{
				"account": "test-account",
				"tilesetid": "test-tileset",
				"layerCount": 1,
				"layers": [
					{
						"account": "test-account",
						"tilesetid": "test-tileset",
						"layer": "test-layer",
						"geometry": "polygon",
						"count": 100,
						"attributes": [
							{
								"attribute": "name",
								"type": "string",
								"values": ["test1", "test2"]
							}
						]
					}
				]
			}`

			tileStats := CreateTileStats([]byte(jsonStr))
			if tileStats == nil {
				t.Fatal("CreateTileStats() returned nil for valid JSON")
			}

			if tileStats.Account != "test-account" {
				t.Errorf("Account = %v, want %v", tileStats.Account, "test-account")
			}
			if tileStats.TilesetId != "test-tileset" {
				t.Errorf("TilesetId = %v, want %v", tileStats.TilesetId, "test-tileset")
			}
			if tileStats.LayerCount != 1 {
				t.Errorf("LayerCount = %v, want %v", tileStats.LayerCount, 1)
			}
			if len(tileStats.Layers) != 1 {
				t.Errorf("Layers count = %v, want %v", len(tileStats.Layers), 1)
			}
		})

		t.Run("InvalidJSON", func(t *testing.T) {
			invalidJSON := `{ invalid json }`
			tileStats := CreateTileStats([]byte(invalidJSON))
			if tileStats != nil {
				t.Error("CreateTileStats() should return nil for invalid JSON")
			}
		})

		t.Run("EmptyJSON", func(t *testing.T) {
			tileStats := CreateTileStats([]byte(""))
			if tileStats != nil {
				t.Error("CreateTileStats() should return nil for empty JSON")
			}
		})
	})
}

func TestLayerAtrribute(t *testing.T) {
	t.Run("NewLayerAtrribute", func(t *testing.T) {
		layer := NewLayerAtrribute("test-tileset", "test-layer")
		if layer == nil {
			t.Fatal("NewLayerAtrribute() returned nil")
		}
		if layer.TilesetId != "test-tileset" {
			t.Errorf("TilesetId = %v, want %v", layer.TilesetId, "test-tileset")
		}
		if layer.Layer != "test-layer" {
			t.Errorf("Layer = %v, want %v", layer.Layer, "test-layer")
		}
		// Note: NewLayerAtrribute doesn't initialize Attributes slice
		// This is expected behavior based on the actual implementation
	})

	t.Run("LayerAtrributeFields", func(t *testing.T) {
		layer := &LayerAtrribute{
			Account:   "test-account",
			TilesetId: "test-tileset",
			Layer:     "test-layer",
			Geometry:  "polygon",
			Count:     100,
			Attributes: []*Attribute{
				{
					Attr:   "name",
					Type:   "string",
					Values: []interface{}{"test1", "test2"},
					Min:    0,
					Max:    100,
				},
			},
		}

		if layer.Account != "test-account" {
			t.Errorf("Account = %v, want %v", layer.Account, "test-account")
		}
		if layer.Geometry != "polygon" {
			t.Errorf("Geometry = %v, want %v", layer.Geometry, "polygon")
		}
		if layer.Count != 100 {
			t.Errorf("Count = %v, want %v", layer.Count, 100)
		}
		if len(layer.Attributes) != 1 {
			t.Errorf("Attributes count = %v, want %v", len(layer.Attributes), 1)
		}
	})

	t.Run("AttributeJSON", func(t *testing.T) {
		attr := &Attribute{
			Attr:   "test-attr",
			Type:   "string",
			Values: []interface{}{"value1", "value2"},
			Min:    0,
			Max:    100,
		}

		data, err := json.Marshal(attr)
		if err != nil {
			t.Fatalf("Failed to marshal Attribute: %v", err)
		}

		var decoded Attribute
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal Attribute: %v", err)
		}

		if decoded.Attr != attr.Attr {
			t.Errorf("Decoded Attr = %v, want %v", decoded.Attr, attr.Attr)
		}
		if decoded.Type != attr.Type {
			t.Errorf("Decoded Type = %v, want %v", decoded.Type, attr.Type)
		}
	})
}

func TestTileStatsCache(t *testing.T) {
	mockStore := &mockTileStore{}
	cache := NewTileStatsCache(mockStore)

	t.Run("Save", func(t *testing.T) {
		tileStats := &TileStats{
			TilesetId: "test-tileset",
			StoreID:   "test-id",
		}

		err := cache.Save(tileStats)
		if err != nil {
			t.Errorf("Save() returned error: %v", err)
		}

		if !mockStore.saved {
			t.Error("Save() did not call store.Save()")
		}
	})

	t.Run("Load", func(t *testing.T) {
		tileStats := &TileStats{
			StoreID: "test-id",
		}

		err := cache.Load(tileStats)
		if err != nil {
			t.Errorf("Load() returned error: %v", err)
		}

		if !mockStore.loaded {
			t.Error("Load() did not call store.Load()")
		}
	})
}

func TestTileJSONEdgeCases(t *testing.T) {
	t.Run("EmptyTileJSON", func(t *testing.T) {
		tileJSON := &TileJSON{}

		if tileJSON.GetExtension() != "json" {
			t.Error("GetExtension() should return 'json' even for empty TileJSON")
		}

		if tileJSON.GetFileName() != "source" {
			t.Error("GetFileName() should return 'source' even for empty TileJSON")
		}

		if tileJSON.GetID() != "" {
			t.Error("GetID() should return empty string for empty StoreID")
		}

		// Empty hash should still work
		hash := tileJSON.Hash()
		if len(hash) != 16 { // MD5 produces 16 bytes
			t.Errorf("Hash() length = %d, want 16", len(hash))
		}
	})

	t.Run("NilSetData", func(t *testing.T) {
		tileJSON := &TileJSON{}
		tileJSON.SetData(nil) // Should not panic
		// No assertions, just ensuring no panic
	})

	t.Run("EmptySetData", func(t *testing.T) {
		tileJSON := &TileJSON{}
		tileJSON.SetData([]byte("")) // Should not panic
		// No assertions, just ensuring no panic
	})

	t.Run("InvalidSetData", func(t *testing.T) {
		tileJSON := &TileJSON{Name: "original"}
		invalidJSON := `{ invalid json }`
		tileJSON.SetData([]byte(invalidJSON))

		// Original values should remain unchanged
		if tileJSON.Name != "original" {
			t.Error("SetData() with invalid JSON should not modify TileJSON")
		}
	})

	t.Run("EmptyTileStats", func(t *testing.T) {
		tileStats := &TileStats{}

		if tileStats.GetExtension() != "json" {
			t.Error("GetExtension() should return 'json' even for empty TileStats")
		}

		if tileStats.GetFileName() != "tilestats" {
			t.Error("GetFileName() should return 'tilestats' even for empty TileStats")
		}

		if tileStats.GetID() != "" {
			t.Error("GetID() should return empty string for empty StoreID")
		}

		// Empty hash should still work
		hash := tileStats.Hash()
		if len(hash) != 16 { // MD5 produces 16 bytes
			t.Errorf("Hash() length = %d, want 16", len(hash))
		}
	})

	t.Run("Constants", func(t *testing.T) {
		if RASTER_DEM != "raster-dem" {
			t.Errorf("RASTER_DEM = %v, want %v", RASTER_DEM, "raster-dem")
		}
		if RASTER != "raster" {
			t.Errorf("RASTER = %v, want %v", RASTER, "raster")
		}
		if VECTOR != "vector" {
			t.Errorf("VECTOR = %v, want %v", VECTOR, "vector")
		}
		if MAPBOX_STATELLITE != "mapbox.satellite" {
			t.Errorf("MAPBOX_STATELLITE = %v, want %v", MAPBOX_STATELLITE, "mapbox.satellite")
		}
	})
}

// mockTileStore implements Store interface for testing
type mockTileStore struct {
	saved  bool
	loaded bool
}

func (m *mockTileStore) Save(r Resource) error {
	m.saved = true
	return nil
}

func (m *mockTileStore) Load(r Resource) error {
	m.loaded = true
	return nil
}

// strPtr is a helper function to get pointer to string
func strPtr(s string) *string {
	return &s
}
