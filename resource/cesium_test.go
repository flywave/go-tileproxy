package resource

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"testing"
)

func TestLayerJson(t *testing.T) {
	layer := &LayerJson{
		Name:        "test-layer",
		TileJson:    "2.2.0",
		Version:     "1.0.0",
		Format:      "application/octet-stream",
		Description: "Test layer for unit testing",
		Attribution: "Test Attribution",
		Available: [][]AvailableBounds{
			{
				{StartX: 0, StartY: 0, EndX: 10, EndY: 10},
				{StartX: 11, StartY: 11, EndX: 20, EndY: 20},
			},
		},
		MetadataAvailability: 1,
		Bounds:               [4]float64{-180, -90, 180, 90},
		Extensions:           []string{"test-extension"},
		Minzoom:              0,
		Maxzoom:              18,
		BVHLevels:            4,
		Projection:           "EPSG:4326",
		Scheme:               "tms",
		Tiles:                []string{"{z}/{x}/{y}.terrain?v={version}"},
		Location:             "/tmp/test",
		Stored:               false,
		StoreID:              "test-store-id",
	}

	t.Run("GetExtension", func(t *testing.T) {
		if got := layer.GetExtension(); got != "json" {
			t.Errorf("GetExtension() = %v, want %v", got, "json")
		}
	})

	t.Run("IsStored", func(t *testing.T) {
		if got := layer.IsStored(); got != false {
			t.Errorf("IsStored() = %v, want %v", got, false)
		}
	})

	t.Run("SetStored", func(t *testing.T) {
		layer.SetStored()
		if got := layer.IsStored(); got != true {
			t.Errorf("IsStored() after SetStored() = %v, want %v", got, true)
		}
	})

	t.Run("GetFileName", func(t *testing.T) {
		if got := layer.GetFileName(); got != "layer" {
			t.Errorf("GetFileName() = %v, want %v", got, "layer")
		}
	})

	t.Run("GetLocation", func(t *testing.T) {
		if got := layer.GetLocation(); got != "/tmp/test" {
			t.Errorf("GetLocation() = %v, want %v", got, "/tmp/test")
		}
	})

	t.Run("SetLocation", func(t *testing.T) {
		layer.SetLocation("/new/location")
		if got := layer.GetLocation(); got != "/new/location" {
			t.Errorf("GetLocation() after SetLocation() = %v, want %v", got, "/new/location")
		}
	})

	t.Run("GetID", func(t *testing.T) {
		if got := layer.GetID(); got != "test-store-id" {
			t.Errorf("GetID() = %v, want %v", got, "test-store-id")
		}
	})

	t.Run("SetID", func(t *testing.T) {
		layer.SetID("new-id")
		if got := layer.GetID(); got != "new-id" {
			t.Errorf("GetID() after SetID() = %v, want %v", got, "new-id")
		}
	})

	t.Run("Hash", func(t *testing.T) {
		expected := md5.Sum([]byte("new-id"))
		if got := layer.Hash(); !bytes.Equal(got, expected[:]) {
			t.Errorf("Hash() = %v, want %v", got, expected[:])
		}
	})
}

func TestLayerJsonSerialization(t *testing.T) {
	layer := &LayerJson{
		Name:       "test-layer",
		TileJson:   "2.2.0",
		Version:    "1.0.0",
		Format:     "application/octet-stream",
		Bounds:     [4]float64{-180, -90, 180, 90},
		Minzoom:    0,
		Maxzoom:    18,
		BVHLevels:  4,
		Projection: "EPSG:4326",
		Scheme:     "tms",
		Tiles:      []string{"{z}/{x}/{y}.terrain?v={version}"},
	}

	t.Run("ToJson", func(t *testing.T) {
		data := layer.ToJson()
		if len(data) == 0 {
			t.Error("ToJson() returned empty data")
		}

		// Verify JSON is valid
		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("ToJson() produced invalid JSON: %v", err)
		}

		if decoded["name"] != "test-layer" {
			t.Errorf("Decoded name = %v, want %v", decoded["name"], "test-layer")
		}
	})

	t.Run("SetData", func(t *testing.T) {
		original := &LayerJson{}
		jsonData := layer.ToJson()
		original.SetData(jsonData)

		if original.Name != layer.Name {
			t.Errorf("SetData() failed to set Name = %v, want %v", original.Name, layer.Name)
		}
		if original.Version != layer.Version {
			t.Errorf("SetData() failed to set Version = %v, want %v", original.Version, layer.Version)
		}
	})

	t.Run("GetData", func(t *testing.T) {
		data := layer.GetData()
		if len(data) == 0 {
			t.Error("GetData() returned empty data")
		}

		// Verify the data matches ToJson output
		expected := layer.ToJson()
		if !bytes.Equal(data, expected) {
			t.Error("GetData() does not match ToJson() output")
		}
	})
}

func TestCreateLayerJson(t *testing.T) {
	t.Run("ValidJSON", func(t *testing.T) {
		jsonStr := `{
			"name": "test-layer",
			"tilejson": "2.2.0",
			"version": "1.0.0",
			"format": "application/octet-stream",
			"bounds": [-180, -90, 180, 90],
			"minzoom": 0,
			"maxzoom": 18,
			"bvhlevels": 4,
			"projection": "EPSG:4326",
			"scheme": "tms",
			"tiles": ["{z}/{x}/{y}.terrain?v={version}"]
		}`

		layer := CreateLayerJson([]byte(jsonStr))
		if layer == nil {
			t.Fatal("CreateLayerJson() returned nil for valid JSON")
		}

		if layer.Name != "test-layer" {
			t.Errorf("Name = %v, want %v", layer.Name, "test-layer")
		}
		if layer.TileJson != "2.2.0" {
			t.Errorf("TileJson = %v, want %v", layer.TileJson, "2.2.0")
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := `{ invalid json }`
		layer := CreateLayerJson([]byte(invalidJSON))
		if layer != nil {
			t.Error("CreateLayerJson() should return nil for invalid JSON")
		}
	})

	t.Run("EmptyJSON", func(t *testing.T) {
		layer := CreateLayerJson([]byte(""))
		if layer != nil {
			t.Error("CreateLayerJson() should return nil for empty JSON")
		}
	})
}

func TestLayerJSONCache(t *testing.T) {
	// Mock store for testing
	mockStore := &mockStore{}
	cache := NewLayerJSONCache(mockStore)

	t.Run("Save", func(t *testing.T) {
		layer := &LayerJson{
			Name:    "test-layer",
			StoreID: "test-id",
		}

		err := cache.Save(layer)
		if err != nil {
			t.Errorf("Save() returned error: %v", err)
		}

		if !mockStore.saved {
			t.Error("Save() did not call store.Save()")
		}
	})

	t.Run("Load", func(t *testing.T) {
		layer := &LayerJson{
			StoreID: "test-id",
		}

		err := cache.Load(layer)
		if err != nil {
			t.Errorf("Load() returned error: %v", err)
		}

		if !mockStore.loaded {
			t.Error("Load() did not call store.Load()")
		}
	})
}

// mockStore implements Store interface for testing
type mockStore struct {
	saved  bool
	loaded bool
}

func (m *mockStore) Save(r Resource) error {
	m.saved = true
	return nil
}

func (m *mockStore) Load(r Resource) error {
	m.loaded = true
	return nil
}

func TestAvailableBounds(t *testing.T) {
	bounds := AvailableBounds{
		StartX: 0,
		StartY: 0,
		EndX:   10,
		EndY:   10,
	}

	t.Run("StructFields", func(t *testing.T) {
		if bounds.StartX != 0 {
			t.Errorf("StartX = %v, want %v", bounds.StartX, 0)
		}
		if bounds.StartY != 0 {
			t.Errorf("StartY = %v, want %v", bounds.StartY, 0)
		}
		if bounds.EndX != 10 {
			t.Errorf("EndX = %v, want %v", bounds.EndX, 10)
		}
		if bounds.EndY != 10 {
			t.Errorf("EndY = %v, want %v", bounds.EndY, 10)
		}
	})

	t.Run("JSONMarshal", func(t *testing.T) {
		data, err := json.Marshal(bounds)
		if err != nil {
			t.Fatalf("Failed to marshal AvailableBounds: %v", err)
		}

		expected := `{"startX":0,"startY":0,"endX":10,"endY":10}`
		if string(data) != expected {
			t.Errorf("JSON marshal result = %s, want %s", string(data), expected)
		}
	})

	t.Run("JSONUnmarshal", func(t *testing.T) {
		jsonStr := `{"startX":5,"startY":5,"endX":15,"endY":15}`
		var decoded AvailableBounds
		if err := json.Unmarshal([]byte(jsonStr), &decoded); err != nil {
			t.Fatalf("Failed to unmarshal AvailableBounds: %v", err)
		}

		if decoded != (AvailableBounds{StartX: 5, StartY: 5, EndX: 15, EndY: 15}) {
			t.Errorf("Unmarshaled AvailableBounds = %v, want %v", decoded, AvailableBounds{StartX: 5, StartY: 5, EndX: 15, EndY: 15})
		}
	})
}

func TestLayerJsonEdgeCases(t *testing.T) {
	t.Run("EmptyFields", func(t *testing.T) {
		layer := &LayerJson{}

		if layer.GetExtension() != "json" {
			t.Error("GetExtension() should return 'json' even for empty layer")
		}

		if layer.GetFileName() != "layer" {
			t.Error("GetFileName() should return 'layer' even for empty layer")
		}

		if layer.GetID() != "" {
			t.Error("GetID() should return empty string for empty StoreID")
		}

		// Empty hash should still work
		hash := layer.Hash()
		if len(hash) != 16 { // MD5 produces 16 bytes
			t.Errorf("Hash() length = %d, want 16", len(hash))
		}
	})

	t.Run("NilSetData", func(t *testing.T) {
		layer := &LayerJson{}
		layer.SetData(nil) // Should not panic
		// No assertions, just ensuring no panic
	})

	t.Run("EmptySetData", func(t *testing.T) {
		layer := &LayerJson{}
		layer.SetData([]byte("")) // Should not panic
		// No assertions, just ensuring no panic
	})

	t.Run("InvalidSetData", func(t *testing.T) {
		layer := &LayerJson{Name: "original"}
		invalidJSON := `{ invalid json }`
		layer.SetData([]byte(invalidJSON))

		// Original values should remain unchanged
		if layer.Name != "original" {
			t.Error("SetData() with invalid JSON should not modify layer")
		}
	})
}
