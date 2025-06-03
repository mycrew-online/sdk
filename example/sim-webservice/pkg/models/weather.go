package models

// EnvironmentalData holds the current environmental information
type EnvironmentalData struct {
	// Core Weather (Row 1)
	Temperature   float32 `json:"temperature"`   // Celsius
	Pressure      float32 `json:"pressure"`      // inHg
	WindSpeed     float32 `json:"windSpeed"`     // knots
	WindDirection float32 `json:"windDirection"` // degrees

	// Environmental Conditions (Row 2)
	Visibility      float32 `json:"visibility"`      // meters
	PrecipRate      float32 `json:"precipRate"`      // millimeters of water
	PrecipState     uint32  `json:"precipState"`     // 2=None, 4=Rain, 8=Snow
	DensityAltitude float32 `json:"densityAltitude"` // feet
	GroundAltitude  float32 `json:"groundAltitude"`  // meters
	MagVar          float32 `json:"magVar"`          // degrees
	SeaLevelPress   float32 `json:"seaLevelPress"`   // millibars
	AmbientDensity  float32 `json:"ambientDensity"`  // slugs per cubic feet

	LastUpdate string `json:"lastUpdate"`
}

// WeatherData is an alias for backward compatibility
type WeatherData = EnvironmentalData

// WeatherPreset represents a weather configuration
type WeatherPreset struct {
	Name          string  `json:"name"`
	Temperature   float32 `json:"temperature"`
	Pressure      float32 `json:"pressure"`
	WindSpeed     float32 `json:"windSpeed"`
	WindDirection float32 `json:"windDirection"`
}

// SimVarDefinition holds information about a SimConnect variable
type SimVarDefinition struct {
	DefineID  uint32
	Name      string
	Units     string
	DataType  interface{} // types.SimConnectDataType
	RequestID uint32
}
