package server

type TrackingType uint8

const (
	NullTrackingType TrackingType = iota
	FaceTrackingType
	HandTrackingType
)

type TrackingData struct {
	facem FaceTracking
	handm HandTracking
}

type anyTracking struct {
	Type uint8 `json:"type"`
}

type Category struct {
	Index        int     `json:"index"`
	Score        float32 `json:"score"`
	CategoryName string  `json:"category_name"`
	DisplayName  string  `json:"display_name"`
}

type Landmark struct {
	X             float32 `json:"x"`
	Y             float32 `json:"y"`
	Z             float32 `json:"z"`
	HasVisibility bool    `json:"has_visibility"`
	Visibility    float32 `json:"visibility"`
	HasPresence   bool    `json:"has_presence"`
	Presence      float32 `json:"presence"`
	Name          string  `json:"name"`
}

// FACE TRACKING

type FaceTracking struct {
	Type        uint8      `json:"type"`
	Status      int        `json:"status"`
	Timestamp   int        `json:"timestamp"`
	Blendshapes []Category `json:"blendshapes"`
	Landmarks   []Landmark `json:"landmarks"`
	Matrixes    []Matrix   `json:"matrixes"`
}

type Matrix struct {
	Rows uint32    `json:"rows"`
	Cols uint32    `json:"cols"`
	Data []float32 `json:"data"`
}

// HAND TRACKING

type HandTracking struct {
	Type      uint8  `json:"type"`
	Status    int    `json:"status"`
	Timestamp int    `json:"timestamp"`
	Hand      []Hand `json:"hands"`
}

type Hand struct {
	Handedness     []Category `json:"handedness"`
	Landmarks      []Landmark `json:"landmarks"`
	WorldLandmarks []Landmark `json:"world_landmarks"`
}
