package persistence

/*
 * Data structure representing version information.
 */
type Version struct {
	Major uint32
	Minor uint32
}

/*
 * Data structure representing file format information.
 */
type FileFormat struct {
	Application string
	Type        string
	Version     Version
}

/*
 * Data structure representing a discrete parameter.
 */
type DiscreteParam struct {
	Key   string
	Value string
}

/*
 * Data structure representing a numeric parameter.
 */
type NumericParam struct {
	Key   string
	Value int32
}

/*
 * Data structure representing a signal processing unit.
 */
type Unit struct {
	Type           string
	Bypass         bool
	DiscreteParams []DiscreteParam
	NumericParams  []NumericParam
}

/*
 * Data structure representing spatializer settings for a channel.
 */
type Spatializer struct {
	Azimuth  float64
	Distance float64
	Level    float64
}

/*
 * Data structure representing an audio channel.
 */
type Channel struct {
	Units       []Unit
	Spatializer Spatializer
}

/*
 * Data structure representing metronome settings.
 */
type Metronome struct {
	Master         bool
	BeatsPerPeriod uint32
	Speed          uint32
	TickSound      string
	TockSound      string
}

/*
 * Data structure representing a configuration file.
 */
type Configuration struct {
	FileFormat      FileFormat
	FramesPerPeriod uint32
	Channels        []Channel
	Metronome       Metronome
}
