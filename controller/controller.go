package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/effects"
	"github.com/andrepxx/go-dsp-guitar/filter"
	"github.com/andrepxx/go-dsp-guitar/hwio"
	"github.com/andrepxx/go-dsp-guitar/level"
	"github.com/andrepxx/go-dsp-guitar/metronome"
	"github.com/andrepxx/go-dsp-guitar/path"
	"github.com/andrepxx/go-dsp-guitar/persistence"
	"github.com/andrepxx/go-dsp-guitar/resample"
	"github.com/andrepxx/go-dsp-guitar/signal"
	"github.com/andrepxx/go-dsp-guitar/spatializer"
	"github.com/andrepxx/go-dsp-guitar/tuner"
	"github.com/andrepxx/go-dsp-guitar/wave"
	"github.com/andrepxx/go-dsp-guitar/webserver"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strconv"
)

/*
 * Constants for the controller.
 */
const (
	CONFIG_PATH              = "config/config.json"
	DEFAULT_SAMPLE_RATE      = 96000
	BLOCK_SIZE               = 8192
	MORE_OUTPUTS_THAN_INPUTS = 3
)

/*
 * A data structure describing a connection between two JACK ports.
 */
type connectionStruct struct {
	From string
	To   string
}

/*
 * The configuration for the controller.
 */
type configStruct struct {
	ImpulseResponses string
	WebServer        webserver.Config
	Connections      []connectionStruct
}

/*
 * A data structure that tells whether an operation was successful or not.
 */
type webResponseStruct struct {
	Success bool
	Reason  string
}

/*
 * A data structure encoding a parameter for an effects unit.
 */
type webParameterStruct struct {
	Name               string
	Type               string
	PhysicalUnit       string
	Minimum            int32
	Maximum            int32
	NumericValue       int32
	DiscreteValueIndex int
	DiscreteValues     []string
}

/*
 * A data structure encoding an effects unit.
 */
type webUnitStruct struct {
	Type       int
	Bypass     bool
	Parameters []webParameterStruct
}

/*
 * A data structure encoding a signal chain.
 */
type webChainStruct struct {
	Units []webUnitStruct
}

/*
 * A data structure encoding the configuration of a single spatializer channel.
 */
type webSpatializerChannelStruct struct {
	Azimuth  float64
	Distance float64
	Level    float64
}

/*
 * A data structure encoding the spatializer configuration.
 */
type webSpatializerStruct struct {
	Channels []webSpatializerChannelStruct
}

/*
 * A data structure encoding the metronome configuration.
 */
type webMetronomeStruct struct {
	BeatsPerPeriod uint32
	MasterOutput   bool
	Speed          uint32
	Sounds         []string
	TickSound      string
	TockSound      string
}

/*
 * A data structure encoding the tuner configuration.
 */
type webTunerStruct struct {
	Channel int
}

/*
 * A data structure encoding the results of the analysis performed by a tuner.
 */
type webTunerResultStruct struct {
	Cents     int8
	Frequency float64
	Note      string
}

/*
 * A data structure encoding the current status of the level meter.
 */
type webLevelMeterStruct struct {
	Enabled bool
}

/*
 * A data structure encoding the results of the analysis performed by a level meter.
 */
type webLevelMeterResultStruct struct {
	ChannelName string
	Level       int32
	Peak        int32
}

/*
 * A data structure encoding the results of the analysis performed by the level meters.
 */
type webLevelMetersResultStruct struct {
	DSPLoad  int32
	Channels []webLevelMeterResultStruct
}

/*
 * A data structure encoding the entire DSP configuration.
 */
type webConfigurationStruct struct {
	FramesPerPeriod uint32
	Chains          []webChainStruct
	Tuner           webTunerStruct
	Spatializer     webSpatializerStruct
	Metronome       webMetronomeStruct
	LevelMeter      webLevelMeterStruct
	BatchProcessing bool
}

/*
 * A task for asynchronous signal processing.
 */
type processingTask struct {
	chain        signal.Chain
	inputBuffer  []float64
	outputBuffer []float64
	sampleRate   uint32
}

/*
 * The controller for the DSP.
 */
type controllerStruct struct {
	binding                 *hwio.Binding
	config                  configStruct
	effects                 []signal.Chain
	impulseResponses        filter.ImpulseResponses
	buffers                 [][]float64
	levelMeter              level.Meter
	metr                    metronome.Metronome
	metrMasterOutput        bool
	running                 bool
	sampleRate              uint32
	spat                    spatializer.Spatializer
	tuner                   tuner.Tuner
	tunerChannel            int
	processingTaskChannel   chan processingTask
	processingResultChannel chan bool
}

/*
 * The controller interface.
 */
type Controller interface {
	Operate(numChannels uint32)
}

/*
 * Marshals an object into a JSON representation or an error.
 * Returns the appropriate MIME type and binary representation.
 */
func (this *controllerStruct) createJSON(obj interface{}) (string, []byte) {
	buffer, err := json.MarshalIndent(obj, "", "\t")

	/*
	 * Check if we got an error during marshalling.
	 */
	if err != nil {
		conf := this.config
		confServer := conf.WebServer
		contentType := confServer.ErrorMime
		errString := err.Error()
		bufString := bytes.NewBufferString(errString)
		bufBytes := bufString.Bytes()
		return contentType, bufBytes
	} else {
		return "application/json; charset=utf-8", buffer
	}

}

/*
 * Adds a new unit to a rack.
 */
func (this *controllerStruct) addUnitHandler(request webserver.HttpRequest) webserver.HttpResponse {
	unitTypeString := request.Params["type"]
	unitType64, errUnitType := strconv.ParseUint(unitTypeString, 10, 64)
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	webResponse := webResponseStruct{}

	/*
	 * Check if unit type and chain ID are valid.
	 */
	if errUnitType != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode unit type.",
		}

	} else if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else {
		unitType := int(unitType64)
		chainId := int(chainId64)
		fx := this.effects
		nChains := len(fx)

		/*
		 * Check if chain ID is out of range.
		 */
		if (chainId < 0) || (chainId >= nChains) {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Chain ID out of range.",
			}

		} else {
			_, err := fx[chainId].AppendUnit(unitType)

			/*
			 * Check if unit was successfully appended.
			 */
			if err != nil {
				reason := err.Error()

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  reason,
				}

			} else {

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Returns the current rack configuration.
 */
func (this *controllerStruct) getConfigurationHandler(request webserver.HttpRequest) webserver.HttpResponse {
	fx := this.effects
	numChannels := len(fx)
	framesPerPeriod := uint32(0)

	/*
	 * If we are bound to a hardware interface, query frames per period.
	 */
	if this.binding != nil {
		framesPerPeriod = hwio.FramesPerPeriod()
	}

	webChains := make([]webChainStruct, numChannels)
	spatChannels := make([]webSpatializerChannelStruct, numChannels)
	paramTypes := effects.ParameterTypes()

	/*
	 * Iterate over the channels and the associated signal chains.
	 */
	for idChannel, chain := range fx {
		numUnits := chain.Length()
		webUnits := make([]webUnitStruct, numUnits)

		/*
		 * Iterate over the units in each channel.
		 */
		for idUnit := 0; idUnit < numUnits; idUnit++ {
			unitType, _ := chain.UnitType(idUnit)
			bypass, _ := chain.GetBypass(idUnit)
			params, _ := chain.Parameters(idUnit)
			numParams := len(params)
			webParams := make([]webParameterStruct, numParams)

			/*
			 * Iterate over the parameters and copy all values.
			 */
			for idParam, param := range params {
				paramTypeId := param.Type
				paramType := paramTypes[paramTypeId]
				webParams[idParam].Name = param.Name
				webParams[idParam].Type = paramType
				webParams[idParam].PhysicalUnit = param.PhysicalUnit
				webParams[idParam].Minimum = param.Minimum
				webParams[idParam].Maximum = param.Maximum
				webParams[idParam].NumericValue = param.NumericValue
				webParams[idParam].DiscreteValueIndex = param.DiscreteValueIndex
				nValues := len(param.DiscreteValues)
				discreteValuesSource := param.DiscreteValues
				discreteValuesTarget := make([]string, nValues)
				copy(discreteValuesTarget, discreteValuesSource)
				webParams[idParam].DiscreteValues = discreteValuesTarget
			}

			webUnits[idUnit].Type = unitType
			webUnits[idUnit].Bypass = bypass
			webUnits[idUnit].Parameters = webParams
		}

		webChains[idChannel].Units = webUnits
		spat := this.spat

		/*
		 * Check if spatializer exists.
		 */
		if spat != nil {
			idChannel32 := uint32(idChannel)
			azimuth, _ := spat.GetAzimuth(idChannel32)
			spatChannels[idChannel].Azimuth = azimuth
			distance, _ := spat.GetDistance(idChannel32)
			spatChannels[idChannel].Distance = distance
			level, _ := spat.GetLevel(idChannel32)
			spatChannels[idChannel].Level = level
		}

	}

	tunerChannel := this.tunerChannel

	/*
	 * Create tuner structure.
	 */
	tuner := webTunerStruct{
		Channel: tunerChannel,
	}

	/*
	 * Create spatializer structure.
	 */
	spat := webSpatializerStruct{
		Channels: spatChannels,
	}

	currentMetronome := this.metr
	irs := this.impulseResponses
	beatsPerPeriod := uint32(0)
	speed := uint32(0)
	preSounds := irs.Names()
	numSounds := len(preSounds)
	numSoundsInc := numSounds + 1
	sounds := make([]string, numSoundsInc)
	sounds[0] = "- NONE -"
	copy(sounds[1:], preSounds)
	tickSound := ""
	tockSound := ""
	metrMasterOutput := this.metrMasterOutput

	/*
	 * Check if we have a metronome.
	 */
	if currentMetronome != nil {
		beatsPerPeriod = currentMetronome.BeatsPerPeriod()
		speed = currentMetronome.Speed()
		tickSound, _ = currentMetronome.Tick()
		tockSound, _ = currentMetronome.Tock()
	}

	/*
	 * Create metronome structure.
	 */
	metr := webMetronomeStruct{
		BeatsPerPeriod: beatsPerPeriod,
		MasterOutput:   metrMasterOutput,
		Speed:          speed,
		Sounds:         sounds,
		TickSound:      tickSound,
		TockSound:      tockSound,
	}

	levelMeter := this.levelMeter
	levelMeterEnabled := levelMeter.Enabled()

	/*
	 * Create level meters structure.
	 */
	meter := webLevelMeterStruct{
		Enabled: levelMeterEnabled,
	}

	batchProcessing := (this.binding == nil)

	/*
	 * Create configuration structure.
	 */
	cfg := webConfigurationStruct{
		Chains:          webChains,
		FramesPerPeriod: framesPerPeriod,
		Tuner:           tuner,
		Spatializer:     spat,
		Metronome:       metr,
		LevelMeter:      meter,
		BatchProcessing: batchProcessing,
	}

	mimeType, buffer := this.createJSON(cfg)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Returns the results of the level analysis of the channels.
 */
func (this *controllerStruct) getLevelAnalysisHandler(request webserver.HttpRequest) webserver.HttpResponse {
	dspLoad := hwio.DSPLoad()
	dspLoad64 := float64(dspLoad)
	dspLoadRounded := math.Round(dspLoad64)
	dspLoad32 := int32(dspLoadRounded)
	levelMeter := this.levelMeter
	channelCount := levelMeter.ChannelCount()
	results := make([]webLevelMeterResultStruct, channelCount)

	/*
	 * Iterate over all channels and obtain results.
	 */
	for i := range results {
		channelId := uint32(i)
		channelName, err := levelMeter.ChannelName(channelId)

		/*
		 * Check if channel name could be obtained
		 */
		if err == nil {
			result, err := levelMeter.Analyze(channelId)

			/*
			 * Check if level analysis was successful.
			 */
			if err == nil {
				level := result.Level()
				peak := result.Peak()

				/*
				 * Fill in web result data structure.
				 */
				r := webLevelMeterResultStruct{
					ChannelName: channelName,
					Level:       level,
					Peak:        peak,
				}

				results[i] = r
			}

		}

	}

	/*
	 * Create level meters result structure.
	 */
	result := webLevelMetersResultStruct{
		DSPLoad:  dspLoad32,
		Channels: results,
	}

	mimeType, buffer := this.createJSON(result)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Returns a list of all supported types of effects units.
 */
func (this *controllerStruct) getUnitTypesHandler(request webserver.HttpRequest) webserver.HttpResponse {
	unitTypes := effects.UnitTypes()
	mimeType, buffer := this.createJSON(unitTypes)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Perform a pitch analysis via the tuner and return the results.
 */
func (this *controllerStruct) getTunerAnalysisHandler(request webserver.HttpRequest) webserver.HttpResponse {
	currentTuner := this.tuner
	analysis, err := currentTuner.Analyze()
	response := webserver.HttpResponse{}

	/*
	 * Check if analysis was successful.
	 */
	if err != nil {
		message := err.Error()
		reason := "Failed to perform analysis: " + message

		/*
		 * Indicate failure.
		 */
		errResponse := webResponseStruct{
			Success: false,
			Reason:  reason,
		}

		mimeType, buffer := this.createJSON(errResponse)

		/*
		 * Create HTTP response.
		 */
		response = webserver.HttpResponse{
			Header: map[string]string{"Content-type": mimeType},
			Body:   buffer,
		}

	} else {
		cents := analysis.Cents()
		frequency := analysis.Frequency()
		note := analysis.Note()

		/*
		 * Fill the results of the tuner into a data structure.
		 */
		result := webTunerResultStruct{
			Cents:     cents,
			Frequency: frequency,
			Note:      note,
		}

		mimeType, buffer := this.createJSON(result)

		/*
		 * Create HTTP response.
		 */
		response = webserver.HttpResponse{
			Header: map[string]string{"Content-type": mimeType},
			Body:   buffer,
		}

	}

	return response
}

/*
 * Moves a unit down in a rack.
 */
func (this *controllerStruct) moveDownHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	unitIdString := request.Params["unit"]
	unitId64, errUnitId := strconv.ParseUint(unitIdString, 10, 64)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain and unit ID are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errUnitId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode unit ID.",
		}

	} else {
		chainId := int(chainId64)
		unitId := int(unitId64)
		fx := this.effects
		nChains := len(fx)

		/*
		 * Check if chain ID is out of range.
		 */
		if (chainId < 0) || (chainId >= nChains) {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Chain ID out of range.",
			}

		} else {
			err := fx[chainId].MoveDown(unitId)

			/*
			 * Check if unit was successfully moved downwards.
			 */
			if err != nil {
				reason := err.Error()

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  reason,
				}

			} else {

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Moves a unit up in a rack.
 */
func (this *controllerStruct) moveUpHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	unitIdString := request.Params["unit"]
	unitId64, errUnitId := strconv.ParseUint(unitIdString, 10, 64)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain and unit ID are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errUnitId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode unit ID.",
		}

	} else {
		chainId := int(chainId64)
		unitId := int(unitId64)
		fx := this.effects
		nChains := len(fx)

		/*
		 * Check if chain ID is out of range.
		 */
		if (chainId < 0) || (chainId >= nChains) {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Chain ID out of range.",
			}

		} else {
			err := fx[chainId].MoveUp(unitId)

			/*
			 * Check if unit was successfully moved upwards.
			 */
			if err != nil {
				reason := err.Error()

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  reason,
				}

			} else {

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Restore (import) current configuration from JSON file.
 */
func (this *controllerStruct) persistenceRestoreHandler(request webserver.HttpRequest) webserver.HttpResponse {
	patchFiles := request.Files["patchfile"]
	webResponse := webResponseStruct{}

	/*
	 * Make sure that patch files are not nil.
	 */
	if patchFiles == nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Field 'patchfile' not defined as a multipart field.",
		}

	} else {
		numPatchFiles := len(patchFiles)

		/*
		 * Make sure that exactly one patch file is sent in request.
		 */
		if numPatchFiles == 0 {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "No patch file sent in request.",
			}

		} else if numPatchFiles != 1 {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Multiple patch files sent in request.",
			}

		} else {
			patchFile := patchFiles[0]
			patchBytes, err := ioutil.ReadAll(patchFile)

			/*
			 * Check if patch file could be successfully read.
			 */
			if err != nil {

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  "Failed to read patch file.",
				}

			} else {
				configuration := persistence.Configuration{}
				err := json.Unmarshal(patchBytes, &configuration)

				/*
				 * Check if unmarshalling was successful.
				 */
				if err != nil {
					msg := err.Error()

					/*
					 * Indicate failure.
					 */
					webResponse = webResponseStruct{
						Success: false,
						Reason:  "Error during unmarshalling: " + msg,
					}

				} else {
					fileFormat := configuration.FileFormat
					fileType := fileFormat.Type
					fileVersion := fileFormat.Version
					majorVersion := fileVersion.Major
					minorVersion := fileVersion.Minor

					/*
					 * Ensure that file format is compatible.
					 */
					if fileType != "patch" {

						/*
						 * Indicate failure.
						 */
						webResponse = webResponseStruct{
							Success: false,
							Reason:  "Uploaded file is not a patch file.",
						}

					} else if majorVersion != 1 || minorVersion < 0 {

						/*
						 * Indicate failure.
						 */
						webResponse = webResponseStruct{
							Success: false,
							Reason:  "Incompatible version of file format.",
						}

					} else {

						/*
						 * If we are bound to a hardware interface, restore frames per period.
						 */
						if this.binding != nil {
							framesPerPeriod := configuration.FramesPerPeriod
							hwio.SetFramesPerPeriod(framesPerPeriod)
						}

						channels := configuration.Channels
						numChannels := len(channels)
						signalChains := this.effects
						numChains := len(signalChains)

						/*
						 * Verify that the configuration file does not contain
						 * more channels than we have.
						 */
						if numChannels > numChains {
							numChannelsString := string(numChannels)
							numChainsString := string(numChains)
							warningMessage := "WARNING: Restored file contains "
							warningMessage += numChannelsString
							warningMessage += " channels, but we currently have only "
							warningMessage += numChainsString
							warningMessage += ". Restore may be incomplete."

							/*
							 * Indicate failure.
							 */
							webResponse = webResponseStruct{
								Success: false,
								Reason:  warningMessage,
							}

						}

						spat := this.spat
						unitTypes := effects.UnitTypes()

						/*
						 * Restore each channel.
						 */
						for channelId, channel := range channels {
							signalChain := signalChains[channelId]
							numUnits := signalChain.Length()

							/*
							 * Remove all units from the signal chain.
							 */
							for numUnits > 0 {
								unitId := numUnits - 1
								signalChain.RemoveUnit(unitId)
								numUnits = signalChain.Length()
							}

							units := channel.Units

							/*
							 * Restore each processing unit.
							 */
							for _, unit := range units {
								unitType := unit.Type
								unitTypeId := int(-1)
								unitTypeFound := false

								/*
								 * Search for the right unit type.
								 */
								for id, currentUnitType := range unitTypes {

									/*
									 * If we found the correct unit type,
									 * store its ID.
									 */
									if unitType == currentUnitType {
										unitTypeId = id
										unitTypeFound = true
									}

								}

								/*
								 * If we found the unit type, restore the unit.
								 */
								if unitTypeFound {
									signalChain.AppendUnit(unitTypeId)
									numUnits := signalChain.Length()
									lastUnitId := numUnits - 1

									/*
									 * Restore each discrete parameter.
									 */
									for _, param := range unit.DiscreteParams {
										key := param.Key
										value := param.Value
										signalChain.SetDiscreteValue(lastUnitId, key, value)
									}

									/*
									 * Restore each numeric parameter.
									 */
									for _, param := range unit.NumericParams {
										key := param.Key
										value := param.Value
										signalChain.SetNumericValue(lastUnitId, key, value)
									}

									bypass := unit.Bypass
									signalChain.SetBypass(lastUnitId, bypass)
								}

							}

							channelId32 := uint32(channelId)
							persistedSpat := channel.Spatializer
							azimuth := persistedSpat.Azimuth
							distance := persistedSpat.Distance
							level := persistedSpat.Level
							spat.SetAzimuth(channelId32, azimuth)
							spat.SetDistance(channelId32, distance)
							spat.SetLevel(channelId32, level)
						}

						irs := this.impulseResponses
						sampleRate := this.sampleRate
						metr := this.metr
						persistedMetr := configuration.Metronome
						masterOutput := persistedMetr.Master
						this.metrMasterOutput = masterOutput
						beatsPerPeriod := persistedMetr.BeatsPerPeriod
						metr.SetBeatsPerPeriod(beatsPerPeriod)
						speed := persistedMetr.Speed
						metr.SetSpeed(speed)
						tickSound := persistedMetr.TickSound

						/*
						 * Check if we should disable the tick sound.
						 */
						if tickSound == "- NONE -" {
							metr.SetTick(tickSound, nil)
						} else {
							flt := irs.CreateFilter(tickSound, sampleRate)

							/*
							 * Check if filter was successfully loaded.
							 */
							if flt != nil {
								coeffs := flt.Coefficients()
								metr.SetTick(tickSound, coeffs)
							}

						}

						tockSound := persistedMetr.TockSound

						/*
						 * Check if we should disable the tock sound.
						 */
						if tockSound == "- NONE -" {
							metr.SetTock(tockSound, nil)
						} else {
							flt := irs.CreateFilter(tockSound, sampleRate)

							/*
							 * Check if filter was successfully loaded.
							 */
							if flt != nil {
								coeffs := flt.Coefficients()
								metr.SetTock(tockSound, coeffs)
							}

						}

						/*
						 * Indicate success.
						 */
						webResponse = webResponseStruct{
							Success: true,
							Reason:  "",
						}

					}

				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Save (export) current configuration to JSON file.
 */
func (this *controllerStruct) persistenceSaveHandler(request webserver.HttpRequest) webserver.HttpResponse {
	cfg := this.config
	svr := cfg.WebServer
	appName := svr.Name
	framesPerPeriod := uint32(BLOCK_SIZE)

	/*
	 * If we are bound to a hardware interface, query frames per period.
	 */
	if this.binding != nil {
		framesPerPeriod = hwio.FramesPerPeriod()
	}

	/*
	 * Create file format version.
	 */
	version := persistence.Version{
		Major: 1,
		Minor: 0,
	}

	/*
	 * Create file format.
	 */
	fileFormat := persistence.FileFormat{
		Application: appName,
		Type:        "patch",
		Version:     version,
	}

	channels := []persistence.Channel{}
	spat := this.spat
	unitTypes := effects.UnitTypes()

	/*
	 * Iterate over the signal chains.
	 */
	for chainId, chain := range this.effects {
		numUnits := chain.Length()
		units := make([]persistence.Unit, numUnits)

		/*
		 * Iterate over all units in the current chain.
		 */
		for unitId := 0; unitId < numUnits; unitId++ {
			bypass, _ := chain.GetBypass(unitId)
			unitType, _ := chain.UnitType(unitId)
			unitTypeString := unitTypes[unitType]
			discreteParams := []persistence.DiscreteParam{}
			numericParams := []persistence.NumericParam{}
			params, _ := chain.Parameters(unitId)

			/*
			 * Iterate over all parameters.
			 */
			for _, param := range params {
				paramName := param.Name
				paramType := param.Type

				/*
				 * Handle both discrete and numeric parameters.
				 */
				switch paramType {
				case effects.PARAMETER_TYPE_DISCRETE:
					idx := param.DiscreteValueIndex
					discreteValues := param.DiscreteValues
					discreteValue := discreteValues[idx]

					/*
					 * Create description for discrete parameter.
					 */
					discreteParam := persistence.DiscreteParam{
						Key:   paramName,
						Value: discreteValue,
					}

					discreteParams = append(discreteParams, discreteParam)
				case effects.PARAMETER_TYPE_NUMERIC:
					numericValue := param.NumericValue

					/*
					 * Create description for numeric parameter.
					 */
					numericParam := persistence.NumericParam{
						Key:   paramName,
						Value: numericValue,
					}

					numericParams = append(numericParams, numericParam)
				}

			}

			/*
			 * Create data structure describing a signal processing unit.
			 */
			unit := persistence.Unit{
				Type:           unitTypeString,
				Bypass:         bypass,
				DiscreteParams: discreteParams,
				NumericParams:  numericParams,
			}

			units[unitId] = unit
		}

		chainId32 := uint32(chainId)
		azimuth, _ := spat.GetAzimuth(chainId32)
		distance, _ := spat.GetDistance(chainId32)
		level, _ := spat.GetLevel(chainId32)

		/*
		 * Create data structure describing spatializer settings for this channel.
		 */
		pSpat := persistence.Spatializer{
			Azimuth:  azimuth,
			Distance: distance,
			Level:    level,
		}

		/*
		 * Create data structure describing audio channel.
		 */
		channel := persistence.Channel{
			Units:       units,
			Spatializer: pSpat,
		}

		channels = append(channels, channel)
	}

	metrMasterOutput := this.metrMasterOutput
	metr := this.metr
	beatsPerPeriod := uint32(0)
	speed := uint32(0)
	tickSound := ""
	tockSound := ""

	/*
	 * Check if we have a metronome.
	 */
	if metr != nil {
		beatsPerPeriod = metr.BeatsPerPeriod()
		speed = metr.Speed()
		tickSound, _ = metr.Tick()
		tockSound, _ = metr.Tock()
	}

	/*
	 * Create metronome information.
	 */
	metrP := persistence.Metronome{
		Master:         metrMasterOutput,
		BeatsPerPeriod: beatsPerPeriod,
		Speed:          speed,
		TickSound:      tickSound,
		TockSound:      tockSound,
	}

	/*
	 * Create configuration.
	 */
	configuration := persistence.Configuration{
		FileFormat:      fileFormat,
		FramesPerPeriod: framesPerPeriod,
		Channels:        channels,
		Metronome:       metrP,
	}

	mimeType, buffer := this.createJSON(configuration)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Cause processing of a file in batch mode.
 */
func (this *controllerStruct) processHandler(request webserver.HttpRequest) webserver.HttpResponse {
	this.running = false

	/*
	 * Indicate success.
	 */
	webResponse := webResponseStruct{
		Success: true,
		Reason:  "",
	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Removes a unit from a rack.
 */
func (this *controllerStruct) removeUnitHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	unitIdString := request.Params["unit"]
	unitId64, errUnitId := strconv.ParseUint(unitIdString, 10, 64)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain and unit ID are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errUnitId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode unit ID.",
		}

	} else {
		chainId := int(chainId64)
		unitId := int(unitId64)
		fx := this.effects
		nChains := len(fx)

		/*
		 * Check if chain ID is out of range.
		 */
		if (chainId < 0) || (chainId >= nChains) {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Chain ID out of range.",
			}

		} else {
			err := fx[chainId].RemoveUnit(unitId)

			/*
			 * Check if unit was successfully removed.
			 */
			if err != nil {
				reason := err.Error()

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  reason,
				}

			} else {

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets the azimuth of a channel in the spatializer.
 */
func (this *controllerStruct) setAzimuthHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	valueString := request.Params["value"]
	valueInt, errValue := strconv.ParseInt(valueString, 10, 64)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain ID and azimuth value are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errValue != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode azimuth value.",
		}

	} else {
		chainId32 := uint32(chainId64)
		value := float64(valueInt)
		spat := this.spat
		err := spat.SetAzimuth(chainId32, value)

		/*
		 * Check if azimuth was set successfully.
		 */
		if err != nil {
			reason := err.Error()

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  reason,
			}

		} else {

			/*
			 * Indicate success.
			 */
			webResponse = webResponseStruct{
				Success: true,
				Reason:  "",
			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Enables or disables bypass for an effects unit.
 */
func (this *controllerStruct) setBypassHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	unitIdString := request.Params["unit"]
	unitId64, errUnitId := strconv.ParseUint(unitIdString, 10, 64)
	valueString := request.Params["value"]
	value, errValue := strconv.ParseBool(valueString)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain ID, unit ID and value are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errUnitId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode unit ID.",
		}

	} else if errValue != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode value.",
		}

	} else {
		chainId := int(chainId64)
		unitId := int(unitId64)
		fx := this.effects
		nChains := len(fx)

		/*
		 * Check if chain ID is out of range.
		 */
		if (chainId < 0) || (chainId >= nChains) {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Chain ID out of range.",
			}

		} else {
			err := fx[chainId].SetBypass(unitId, value)

			/*
			 * Check if bypass value was successfully set.
			 */
			if err != nil {
				reason := err.Error()

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  reason,
				}

			} else {

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets a discrete value as a parameter in an effects unit.
 */
func (this *controllerStruct) setDiscreteValueHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	unitIdString := request.Params["unit"]
	unitId64, errUnitId := strconv.ParseUint(unitIdString, 10, 64)
	param := request.Params["param"]
	value := request.Params["value"]
	webResponse := webResponseStruct{}

	/*
	 * Check if chain ID, unit ID and value are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errUnitId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode unit ID.",
		}

	} else {
		chainId := int(chainId64)
		unitId := int(unitId64)
		fx := this.effects
		nChains := len(fx)

		/*
		 * Check if chain ID is out of range.
		 */
		if (chainId < 0) || (chainId >= nChains) {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Chain ID out of range.",
			}

		} else {
			err := fx[chainId].SetDiscreteValue(unitId, param, value)

			/*
			 * Check if bypass value was successfully set.
			 */
			if err != nil {
				reason := err.Error()

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  reason,
				}

			} else {

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets the distance of a channel in the spatializer.
 */
func (this *controllerStruct) setDistanceHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	valueString := request.Params["value"]
	value, errDistance := strconv.ParseFloat(valueString, 64)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain ID and distance value are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errDistance != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode distance value.",
		}

	} else {
		chainId32 := uint32(chainId64)
		spat := this.spat
		err := spat.SetDistance(chainId32, value)

		/*
		 * Check if distance was set successfully.
		 */
		if err != nil {
			reason := err.Error()

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  reason,
			}

		} else {

			/*
			 * Indicate success.
			 */
			webResponse = webResponseStruct{
				Success: true,
				Reason:  "",
			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets the frames per period for the hardware interface.
 */
func (this *controllerStruct) setFramesPerPeriodHandler(request webserver.HttpRequest) webserver.HttpResponse {
	valueString := request.Params["value"]
	value64, err := strconv.ParseUint(valueString, 10, 32)
	webResponse := webResponseStruct{}

	/*
	 * Check if value is valid.
	 */
	if err != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to parse frame count.",
		}

	} else {
		value32 := uint32(value64)
		hwio.SetFramesPerPeriod(value32)

		/*
		 * Indicate success.
		 */
		webResponse = webResponseStruct{
			Success: true,
			Reason:  "",
		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets the level of a channel in the spatializer.
 */
func (this *controllerStruct) setLevelHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	valueString := request.Params["value"]
	value, errDistance := strconv.ParseFloat(valueString, 64)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain ID and distance value are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errDistance != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode level value.",
		}

	} else {
		chainId32 := uint32(chainId64)
		spat := this.spat
		err := spat.SetLevel(chainId32, value)

		/*
		 * Check if distance was set successfully.
		 */
		if err != nil {
			reason := err.Error()

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  reason,
			}

		} else {

			/*
			 * Indicate success.
			 */
			webResponse = webResponseStruct{
				Success: true,
				Reason:  "",
			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets the level of a channel in the spatializer.
 */
func (this *controllerStruct) setLevelMeterEnabledHandler(request webserver.HttpRequest) webserver.HttpResponse {
	valueString := request.Params["value"]
	value, err := strconv.ParseBool(valueString)
	webResponse := webResponseStruct{}

	/*
	 * Check if boolean value is valud.
	 */
	if err != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: true,
			Reason:  "Failed to decode boolean value.",
		}

	} else {
		meter := this.levelMeter
		meter.SetEnabled(value)

		/*
		 * If level meters should be disabled, clear buffers as well.
		 */
		if !value {
			buffers := this.buffers

			/*
			 * Iterate over all buffers.
			 */
			for _, buffer := range buffers {

				/*
				 * Clear the buffer.
				 */
				for i := range buffer {
					buffer[i] = 0.0
				}

			}

		}

		/*
		 * Indicate success.
		 */
		webResponse = webResponseStruct{
			Success: true,
			Reason:  "",
		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets a value for the metronome.
 */
func (this *controllerStruct) setMetronomeValueHandler(request webserver.HttpRequest) webserver.HttpResponse {
	metr := this.metr
	webResponse := webResponseStruct{}

	/*
	 * Check if we have a metronome.
	 */
	if metr != nil {
		param := request.Params["param"]
		value := request.Params["value"]

		/*
		 * Check which parameter should be edited.
		 */
		switch param {
		case "beats-per-period":
			rawValue, err := strconv.ParseUint(value, 10, 32)

			/*
			 * Check if value failed to parse.
			 */
			if err != nil {

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  "Failed to decode metronome beats per minute.",
				}

			} else {
				value := uint32(rawValue)
				metr.SetBeatsPerPeriod(value)

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		case "master-output":
			value, err := strconv.ParseBool(value)

			/*
			 * Check if value failed to parse.
			 */
			if err != nil {

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  "Failed to decode metronome master output flag.",
				}

			} else {
				this.metrMasterOutput = value

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		case "speed":
			rawValue, err := strconv.ParseUint(value, 10, 32)

			/*
			 * Check if value failed to parse.
			 */
			if err != nil {

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  "Failed to decode metronome speed.",
				}

			} else {
				value := uint32(rawValue)
				metr.SetSpeed(value)

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		case "tick-sound":
			irs := this.impulseResponses

			/*
			 * Check if we should disable the tick sound.
			 */
			if value == "- NONE -" {
				metr.SetTick(value, nil)

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			} else {
				sampleRate := this.sampleRate
				flt := irs.CreateFilter(value, sampleRate)

				/*
				 * Check if filter was successfully loaded.
				 */
				if flt == nil {

					/*
					 * Indicate failure.
					 */
					webResponse = webResponseStruct{
						Success: false,
						Reason:  "Failed to load impulse response for metronome tick sound.",
					}

				} else {
					coeffs := flt.Coefficients()
					metr.SetTick(value, coeffs)

					/*
					 * Indicate success.
					 */
					webResponse = webResponseStruct{
						Success: true,
						Reason:  "",
					}

				}

			}

		case "tock-sound":
			irs := this.impulseResponses

			/*
			 * Check if we should disable the tock sound.
			 */
			if value == "- NONE -" {
				metr.SetTock(value, nil)

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			} else {
				sampleRate := this.sampleRate
				flt := irs.CreateFilter(value, sampleRate)

				/*
				 * Check if filter was successfully loaded.
				 */
				if flt == nil {

					/*
					 * Indicate failure.
					 */
					webResponse = webResponseStruct{
						Success: false,
						Reason:  "Failed to load impulse response for metronome tick sound.",
					}

				} else {
					coeffs := flt.Coefficients()
					metr.SetTock(value, coeffs)

					/*
					 * Indicate success.
					 */
					webResponse = webResponseStruct{
						Success: true,
						Reason:  "",
					}

				}

			}

		default:
			reason := fmt.Sprintf("Unknown metronome parameter: '%s'", param)

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  reason,
			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets a value for the tuner.
 */
func (this *controllerStruct) setTunerValueHandler(request webserver.HttpRequest) webserver.HttpResponse {
	currentTuner := this.tuner
	webResponse := webResponseStruct{}

	/*
	 * Check if we have a tuner.
	 */
	if currentTuner != nil {
		param := request.Params["param"]
		value := request.Params["value"]

		/*
		 * Check which parameter should be edited.
		 */
		switch param {
		case "channel":
			rawValue, err := strconv.ParseInt(value, 10, 64)

			/*
			 * Check if value failed to parse.
			 */
			if err != nil {

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  "Failed to decode tuner channel.",
				}

			} else {
				this.tunerChannel = int(rawValue)

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}
		default:
			reason := fmt.Sprintf("Unknown tuner parameter: '%s'", param)

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  reason,
			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Sets a numeric value as a parameter in an effects unit.
 */
func (this *controllerStruct) setNumericValueHandler(request webserver.HttpRequest) webserver.HttpResponse {
	chainIdString := request.Params["chain"]
	chainId64, errChainId := strconv.ParseUint(chainIdString, 10, 64)
	unitIdString := request.Params["unit"]
	unitId64, errUnitId := strconv.ParseUint(unitIdString, 10, 64)
	param := request.Params["param"]
	valueString := request.Params["value"]
	value64, errValue := strconv.ParseInt(valueString, 10, 32)
	webResponse := webResponseStruct{}

	/*
	 * Check if chain ID, unit ID and value are valid.
	 */
	if errChainId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode chain ID.",
		}

	} else if errUnitId != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode unit ID.",
		}

	} else if errValue != nil {

		/*
		 * Indicate failure.
		 */
		webResponse = webResponseStruct{
			Success: false,
			Reason:  "Failed to decode value.",
		}

	} else {
		chainId := int(chainId64)
		unitId := int(unitId64)
		value := int32(value64)
		fx := this.effects
		nChains := len(fx)

		/*
		 * Check if chain ID is out of range.
		 */
		if (chainId < 0) || (chainId >= nChains) {

			/*
			 * Indicate failure.
			 */
			webResponse = webResponseStruct{
				Success: false,
				Reason:  "Chain ID out of range.",
			}

		} else {
			err := fx[chainId].SetNumericValue(unitId, param, value)

			/*
			 * Check if bypass value was successfully set.
			 */
			if err != nil {
				reason := err.Error()

				/*
				 * Indicate failure.
				 */
				webResponse = webResponseStruct{
					Success: false,
					Reason:  reason,
				}

			} else {

				/*
				 * Indicate success.
				 */
				webResponse = webResponseStruct{
					Success: true,
					Reason:  "",
				}

			}

		}

	}

	mimeType, buffer := this.createJSON(webResponse)

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": mimeType},
		Body:   buffer,
	}

	return response
}

/*
 * Handles CGI requests that could not be dispatched to other CGIs.
 */
func (this *controllerStruct) errorHandler(request webserver.HttpRequest) webserver.HttpResponse {
	conf := this.config
	confServer := conf.WebServer
	contentType := confServer.ErrorMime
	msgBuf := bytes.NewBufferString("This CGI call is not implemented.")
	msgBytes := msgBuf.Bytes()

	/*
	 * Create HTTP response.
	 */
	response := webserver.HttpResponse{
		Header: map[string]string{"Content-type": contentType},
		Body:   msgBytes,
	}

	return response
}

/*
 * Dispatch CGI requests to the corresponding CGI handlers.
 */
func (this *controllerStruct) dispatch(request webserver.HttpRequest) webserver.HttpResponse {
	cgi := request.Params["cgi"]
	response := webserver.HttpResponse{}

	/*
	 * Find the right CGI to handle the request.
	 */
	switch cgi {
	case "add-unit":
		response = this.addUnitHandler(request)
	case "get-configuration":
		response = this.getConfigurationHandler(request)
	case "get-level-analysis":
		response = this.getLevelAnalysisHandler(request)
	case "get-unit-types":
		response = this.getUnitTypesHandler(request)
	case "get-tuner-analysis":
		response = this.getTunerAnalysisHandler(request)
	case "move-down":
		response = this.moveDownHandler(request)
	case "move-up":
		response = this.moveUpHandler(request)
	case "persistence-restore":
		response = this.persistenceRestoreHandler(request)
	case "persistence-save":
		response = this.persistenceSaveHandler(request)
	case "process":
		response = this.processHandler(request)
	case "remove-unit":
		response = this.removeUnitHandler(request)
	case "set-azimuth":
		response = this.setAzimuthHandler(request)
	case "set-bypass":
		response = this.setBypassHandler(request)
	case "set-discrete-value":
		response = this.setDiscreteValueHandler(request)
	case "set-distance":
		response = this.setDistanceHandler(request)
	case "set-frames-per-period":
		response = this.setFramesPerPeriodHandler(request)
	case "set-level":
		response = this.setLevelHandler(request)
	case "set-level-meter-enabled":
		response = this.setLevelMeterEnabledHandler(request)
	case "set-metronome-value":
		response = this.setMetronomeValueHandler(request)
	case "set-tuner-value":
		response = this.setTunerValueHandler(request)
	case "set-numeric-value":
		response = this.setNumericValueHandler(request)
	default:
		response = this.errorHandler(request)
	}

	return response
}

/*
 * Perform asynchronous signal processing.
 */
func (this *controllerStruct) processAsync() {
	requests := this.processingTaskChannel
	responses := this.processingResultChannel

	/*
	 * Process tasks as long as channel is open.
	 */
	for task := range requests {
		chain := task.chain
		inputBuffer := task.inputBuffer
		outputBuffer := task.outputBuffer
		sampleRate := task.sampleRate
		chain.Process(inputBuffer, outputBuffer, sampleRate)
		responses <- true
	}

	close(responses)
}

/*
 * Process audio data.
 */
func (this *controllerStruct) process(inputBuffers [][]float64, outputBuffers [][]float64, sampleRate uint32) {
	nIn := len(inputBuffers)
	nOut := len(outputBuffers)
	nMinOut := nIn + (spatializer.OUTPUT_COUNT + metronome.OUTPUT_COUNT)
	buffers := this.buffers
	levelMeter := this.levelMeter
	levelMeterEnabled := false

	/*
	 * Check if there is a level meter and if it is enabled.
	 */
	if levelMeter != nil {
		levelMeterEnabled = levelMeter.Enabled()
	}

	tunerChannel := this.tunerChannel

	/*
	 * Check if an input channel should be passed to the tuner.
	 */
	if (tunerChannel >= 0) && (tunerChannel < nIn) {
		tunerInput := inputBuffers[tunerChannel]
		currentTuner := this.tuner
		currentTuner.Process(tunerInput, sampleRate)
	}

	/*
	 * Ensure that there are at least as many outputs as inputs registered.
	 */
	if (nOut >= nIn) && (nIn >= 0) {

		/*
		 * Start processing for each input channel.
		 */
		for i := 0; i < nIn; i++ {

			/*
			 * Create a new signal processing task.
			 */
			task := processingTask{
				chain:        this.effects[i],
				inputBuffer:  inputBuffers[i],
				outputBuffer: outputBuffers[i],
				sampleRate:   sampleRate,
			}

			this.processingTaskChannel <- task
		}

		/*
		 * Wait for processing of each channel to finish.
		 */
		for i := 0; i < nIn; i++ {
			<-this.processingResultChannel
		}

		/*
		 * If level meter is enabled, save input and output buffers.
		 */
		if levelMeterEnabled {
			copy(buffers[0:nIn], inputBuffers)
			uBound := 2 * nIn
			copy(buffers[nIn:uBound], outputBuffers)
		}

	}

	/*
	 * Check if there are enough output channels for a spatializer and a metronome.
	 */
	if nOut >= nMinOut {
		lastIdx := nOut - 1
		auxBuffer := outputBuffers[lastIdx]
		metr := this.metr

		/*
		 * Check if there is a metronome.
		 */
		if metr == nil {
			auxBuffer = nil
		} else {
			metr.Process(auxBuffer)

			/*
			 * If there level meter is enabled, save auxiliary buffer.
			 */
			if levelMeterEnabled {
				idx := 2 * nIn
				buffers[idx] = auxBuffer
			}

		}

		spat := this.spat

		/*
		 * Check if there is a spatializer.
		 */
		if spat != nil {

			/*
			 * Check if metronome output should be excluded from the master output.
			 */
			if !this.metrMasterOutput {
				auxBuffer = nil
			}

			uBound := nIn + spatializer.OUTPUT_COUNT
			spatializerInputs := outputBuffers[0:nIn]
			spatializerOutputs := outputBuffers[nIn:uBound]
			spat.Process(spatializerInputs, auxBuffer, spatializerOutputs)
			lBoundBuf := (2 * nIn) + 1
			uBoundBuf := lBoundBuf + spatializer.OUTPUT_COUNT

			/*
			 * If level Meter is enabled, save spatializer output.
			 */
			if levelMeterEnabled {
				copy(buffers[lBoundBuf:uBoundBuf], spatializerOutputs)
			}

		}

	}

	/*
	 * Feed buffers to level meter, if enabled.
	 */
	if levelMeterEnabled {
		levelMeter.Process(buffers, sampleRate)
	}

}

/*
 * This is called when the hardware changes the sample rate.
 */
func (this *controllerStruct) sampleRateListener(rate uint32) {
	this.sampleRate = rate
	spat := this.spat
	spat.SetSampleRate(rate)
	metr := this.metr
	metr.SetSampleRate(rate)
}

/*
 * Get input from the user.
 */
func (this *controllerStruct) getInput(scanner *bufio.Scanner, prompt string) string {
	fmt.Printf("%s", prompt)
	scanner.Scan()
	s := scanner.Text()
	return s
}

/*
 * Process files for batch processing.
 */
func (this *controllerStruct) processFiles(scanner *bufio.Scanner, targetRate uint32) {
	effects := this.effects
	numChannels := len(effects)
	fmt.Printf("Web interface initiated batch processing for %d channels.\n", numChannels)
	inputs := make([][]float64, numChannels)
	sampleRates := make([]uint32, numChannels)
	outputFormat := uint16(wave.AUDIO_PCM)
	validFormat := false

	/*
	 * Query the user for a target format.
	 */
	for !validFormat {
		targetFormat := this.getInput(scanner, "Please enter target format ('lpcm' or 'float'): ")

		/*
		 * Find out about the target format.
		 */
		switch targetFormat {
		case "lpcm":
			outputFormat = wave.AUDIO_PCM
			validFormat = true
		case "float":
			outputFormat = wave.AUDIO_IEEE_FLOAT
			validFormat = true
		}

	}

	bitDepth := uint16(wave.DEFAULT_BIT_DEPTH)
	validBitDepth := false

	/*
	 * Query the user for a target bit depth.
	 */
	for !validBitDepth {

		/*
		 * Different formats support different bit depths.
		 */
		switch outputFormat {
		case wave.AUDIO_PCM:
			targetBitDepthString := this.getInput(scanner, "Please enter target bit depth (8 or 16 or 24 or 32): ")
			targetBitDepth64, _ := strconv.ParseUint(targetBitDepthString, 10, 64)

			/*
			 * Check if the target bit depth is valid.
			 */
			if targetBitDepth64 == 8 || targetBitDepth64 == 16 || targetBitDepth64 == 24 || targetBitDepth64 == 32 {
				bitDepth = uint16(targetBitDepth64)
				validBitDepth = true
			}

		case wave.AUDIO_IEEE_FLOAT:
			targetBitDepthString := this.getInput(scanner, "Please enter target bit depth (32 or 64): ")
			targetBitDepth64, _ := strconv.ParseUint(targetBitDepthString, 10, 64)

			/*
			 * Check if the target bit depth is valid.
			 */
			if targetBitDepth64 == 32 || targetBitDepth64 == 64 {
				bitDepth = uint16(targetBitDepth64)
				validBitDepth = true
			}

		default:
			fmt.Printf("WARNING! Unrecognized format code: %#04x\n - Continuing with default bit depth: %d (This should not happen!)\n", outputFormat, bitDepth)
			validBitDepth = true
		}

	}

	/*
	 * Query file name and channel number for each input.
	 */
	for fileId := 0; fileId < numChannels; fileId++ {
		fmt.Printf("%s\n", "Enter name/path of the wave file for input.")
		prompt := fmt.Sprintf("File for input %d: ", fileId)
		fileName := this.getInput(scanner, prompt)
		fileName = path.Sanitize(fileName)

		/*
		 * Abort if file name is empty.
		 */
		if fileName == "" {
			fmt.Printf("Leaving channel %d empty.\n", fileId)
			inputs[fileId] = make([]float64, 0)
			sampleRates[fileId] = DEFAULT_SAMPLE_RATE
		} else {
			buf, err := ioutil.ReadFile(fileName)

			/*
			 * Check if file could be read.
			 */
			if err != nil {
				fmt.Printf("Failed to read wave file. Leaving channel %d empty.\n", fileId)
				inputs[fileId] = make([]float64, 0)
				sampleRates[fileId] = DEFAULT_SAMPLE_RATE
			} else {
				f, err := wave.FromBuffer(buf)

				/*
				 * Check if file could be parsed.
				 */
				if err != nil {
					msg := err.Error()
					fmt.Printf("Failed to parse wave file: %s\n", msg)
					inputs[fileId] = make([]float64, 0)
					sampleRates[fileId] = DEFAULT_SAMPLE_RATE
				} else {
					numChannels := f.ChannelCount()

					/*
					 * If file contains only one channel, take first,
					 * otherwise ask which one to use.
					 */
					if numChannels == 1 {
						c, err := f.Channel(0)

						/*
						 * Check if channel could be loaded.
						 */
						if err != nil {
							inputs[fileId] = make([]float64, 0)
							sampleRates[fileId] = DEFAULT_SAMPLE_RATE
						} else {
							inputs[fileId] = c.Floats()
							sampleRates[fileId] = f.SampleRate()
						}

					} else {
						loadedChan := false

						/*
						 * Do this until the channel has been loaded.
						 */
						for !loadedChan {
							uBound := numChannels - 1
							prompt := fmt.Sprintf("File contains %d channels. Which channel [%d, %d] to use? ", numChannels, 0, uBound)
							channelString := this.getInput(scanner, prompt)
							n, err := strconv.ParseUint(channelString, 10, 16)

							/*
							 * If input is valid, load this channel.
							 */
							if err != nil {
								fmt.Printf("%s\n", "Not a valid channel number.")
							} else {
								id := uint16(n)
								c, err := f.Channel(id)

								/*
								 * Check if channel could be loaded.
								 */
								if err != nil {
									msg := err.Error()
									fmt.Printf("Failed to load channel: %s\n", msg)
									inputs[fileId] = make([]float64, 0)
									sampleRates[fileId] = DEFAULT_SAMPLE_RATE
								} else {
									inputs[fileId] = c.Floats()
									sampleRates[fileId] = f.SampleRate()
									loadedChan = true
								}

							}

						}

					}

				}

			}

		}

	}

	/*
	 * Resample all inputs to the target sample rate.
	 */
	for i, input := range inputs {
		sampleRate := sampleRates[i]

		/*
		 * Check if resampling is necessary.
		 */
		if sampleRate != targetRate {
			fmt.Printf("Resampling input channel %d from %d Hz to %d Hz, please wait ...\n", i, sampleRate, targetRate)
			inputs[i] = resample.Time(input, sampleRate, targetRate)
			runtime.GC()
		}

	}

	maxLength := int(0)

	/*
	 * Find the length of the longest input stream.
	 */
	for _, input := range inputs {
		size := len(input)

		/*
		 * If we found a longer input stream, store its length.
		 */
		if size > maxLength {
			maxLength = size
		}

	}

	/*
	 * Length must be a multiple of the block size.
	 */
	if (maxLength % BLOCK_SIZE) != 0 {
		maxLength = BLOCK_SIZE * ((maxLength / BLOCK_SIZE) + 1)
	}

	/*
	 * Extend each input stream to equal length.
	 */
	for i, input := range inputs {
		size := len(input)

		/*
		 * If size of input stream doesn't already match, extend it.
		 */
		if size != maxLength {
			inputNew := make([]float64, maxLength)
			copy(inputNew, input)
			inputs[i] = inputNew
			runtime.GC()
		}

	}

	numInputs := len(inputs)
	numOutputs := numInputs + MORE_OUTPUTS_THAN_INPUTS
	outputs := make([][]float64, numOutputs)
	inputBuffers := make([][]float64, numInputs)
	outputBuffers := make([][]float64, numOutputs)

	/*
	 * Create each inner output buffer.
	 */
	for i := 0; i < numOutputs; i++ {
		outputs[i] = make([]float64, maxLength)
		outputBuffers[i] = make([]float64, BLOCK_SIZE)
	}

	/*
	 * Create each inner input buffer.
	 */
	for i := 0; i < numInputs; i++ {
		inputBuffers[i] = make([]float64, BLOCK_SIZE)
	}

	numBlocks := maxLength / BLOCK_SIZE
	numBlocksFloat := float64(numBlocks)
	fmt.Printf("%s\n", "Processing audio data ...")
	oldPercents := int(0)

	/*
	 * Process each block.
	 */
	for block := 0; block < numBlocks; block++ {
		blockFloat := float64(block)
		percents := int((100.0 * blockFloat) / numBlocksFloat)

		/*
		 * Check if percentage changed.
		 */
		if percents != oldPercents {
			fmt.Printf(" %d", percents)
			oldPercents = percents
		}

		offsetStart := BLOCK_SIZE * block
		offsetEnd := offsetStart + BLOCK_SIZE

		/*
		 * Copy part of each input stream into the input buffers.
		 */
		for i, input := range inputs {
			copy(inputBuffers[i], input[offsetStart:offsetEnd])
		}

		this.process(inputBuffers, outputBuffers, targetRate)

		/*
		 * Copy the output buffers into the right place in the output streams.
		 */
		for i, output := range outputs {
			copy(output[offsetStart:offsetEnd], outputBuffers[i])
		}

	}

	fmt.Printf("\n")

	/*
	 * Discard the input streams to free memory.
	 */
	for i := 0; i < numInputs; i++ {
		inputs[i] = nil
	}

	runtime.GC()

	/*
	 * Write each output into a wave file.
	 */
	for i, output := range outputs {
		f, err := wave.CreateEmpty(targetRate, outputFormat, bitDepth, 1)

		/*
		 * Check whether we were able to create a wave file.
		 */
		if err != nil {
			msg := err.Error()
			fmt.Printf("Failed to create wave file: %s", msg)
		} else {
			c, err := f.Channel(0)

			/*
			 * Check whether we were able to obtain the channel.
			 */
			if err != nil {
				msg := err.Error()
				fmt.Printf("Failed to create output %d: %s\n", i, msg)
			} else {
				c.WriteFloats(output)
				buf, err := f.Bytes()
				f = nil
				runtime.GC()

				/*
				 * Check whether we were able to serialize the channel.
				 */
				if err != nil {
					msg := err.Error()
					fmt.Printf("Failed to serialize output %d: %s\n", i, msg)
				} else {
					iLong := uint64(i)
					iString := strconv.FormatUint(iLong, 10)
					channelName := "out_" + iString

					/*
					 * Check whether output channel is "special".
					 */
					switch i {
					case numInputs:
						channelName = "master_left"
					case numInputs + 1:
						channelName = "master_right"
					case numInputs + 2:
						channelName = "metronome"
					}

					prompt := fmt.Sprintf("Output file for channel '%s': ", channelName)
					fileName := this.getInput(scanner, prompt)
					fileName = path.Sanitize(fileName)

					/*
					 * Check if file name is empty.
					 */
					if fileName == "" {
						fmt.Printf("%s\n", "Skipping output due to empty file name.")
					} else {
						fd, err := os.Create(fileName)

						/*
						 * Check if file was successfully created.
						 */
						if err != nil {
							fmt.Printf("%s\n", "Failed to create output file.")
						} else {
							_, err = fd.Write(buf)

							/*
							 * Check if buffer was written successfully.
							 */
							if err != nil {
								fmt.Printf("%s\n", "Failed to write to output file.")
							}

							err = fd.Close()
							buf = nil
							runtime.GC()

							/*
							 * Check if file was closed successfully.
							 */
							if err != nil {
								msg := err.Error()
								fmt.Printf("%s\n", "Failed to close output file.", msg)
							}

						}

					}

				}

			}

		}

	}

	/*
	 * Discard the output streams to free memory.
	 */
	for i := 0; i < numOutputs; i++ {
		outputs[i] = nil
		runtime.GC()
	}

}

/*
 * Initialize the controller.
 */
func (this *controllerStruct) initialize(nInputs uint32, useHardware bool) error {
	content, err := ioutil.ReadFile(CONFIG_PATH)

	/*
	 * Check if file could be read.
	 */
	if err != nil {
		return fmt.Errorf("Failed to open config file: '%s'", CONFIG_PATH)
	} else {
		config := configStruct{}
		err = json.Unmarshal(content, &config)
		this.config = config

		/*
		 * Check if file failed to unmarshal.
		 */
		if err != nil {
			return fmt.Errorf("Failed to decode config file: '%s'", CONFIG_PATH)
		} else {
			ir, err := filter.Import(config.ImpulseResponses)

			/*
			 * Check if impulse responses failed to load.
			 */
			if err != nil {
				return err
			} else {
				this.impulseResponses = ir
				fx := make([]signal.Chain, nInputs)

				/*
				 * Create an effects chain for each input.
				 */
				for i := uint32(0); i < nInputs; i++ {
					fx[i] = signal.CreateChain(ir)
				}

				this.effects = fx
				this.sampleRate = DEFAULT_SAMPLE_RATE
				spat := spatializer.Create(nInputs)
				this.spat = spat
				metr := metronome.Create()
				metr.SetTick("- NONE -", nil)
				metr.SetTock("- NONE -", nil)
				this.metr = metr
				this.tuner = tuner.Create()
				this.tunerChannel = -1
				numPorts := (2 * nInputs) + (1 + spatializer.OUTPUT_COUNT)
				portNames := make([]string, numPorts)

				/*
				 * Calculate names of all input ports.
				 */
				for i := uint32(0); i < nInputs; i++ {
					i64 := uint64(i)
					idString := strconv.FormatUint(i64, 10)
					portNames[i] = "in_" + idString
				}

				/*
				 * Calculate names of all output ports.
				 */
				for i := uint32(0); i < nInputs; i++ {
					i64 := uint64(i)
					idString := strconv.FormatUint(i64, 10)
					idx := nInputs + i
					portNames[idx] = "out_" + idString
				}

				/*
				 * Calculate name of metronome port.
				 */
				if metr != nil {
					idx := numPorts - 3
					portNames[idx] = "metronome"
				}

				/*
				 * Calculate name of master outputs.
				 */
				if spat != nil {
					idxLeft := numPorts - 2
					portNames[idxLeft] = "master_left"
					idxRight := numPorts - 1
					portNames[idxRight] = "master right"
				}

				buffers := make([][]float64, numPorts)
				this.buffers = buffers
				levelMeter, err := level.CreateMeter(numPorts, portNames)
				this.levelMeter = levelMeter

				/*
				 * Check if level meter was created.
				 */
				if err != nil {
					msg := err.Error()
					return fmt.Errorf("Failed to create level meter: %s", msg)
				} else {
					this.processingTaskChannel = make(chan processingTask, nInputs)
					this.processingResultChannel = make(chan bool, nInputs)

					/*
					 * Start a worker thread for each input channel.
					 */
					for i := uint32(0); i < nInputs; i++ {
						go this.processAsync()
					}

					/*
					 * If we don't use hardware I/O, we are done, otherwise register hardware binding.
					 */
					if !useHardware {
						return nil
					} else {
						this.binding, err = hwio.Register(this.process, this.sampleRateListener)

						/*
						 * Setup JACK connections.
						 */
						for _, connection := range config.Connections {
							source := connection.From
							destination := connection.To
							hwio.Connect(source, destination)
						}

						return err
					}

				}

			}

		}

	}

}

/*
 * Finalize the controller, freeing allocated ressources.
 */
func (this *controllerStruct) finalize() {
	this.running = false
	binding := this.binding
	hwio.Unregister(binding)
	ptc := this.processingTaskChannel
	close(ptc)
}

/*
 * Main routine of our controller. Performs initialization, then runs the message pump.
 */
func (this *controllerStruct) Operate(numChannels uint32) {
	batch := numChannels > 0
	err := fmt.Errorf("")

	/*
	 * If we are not in batch processing mode, acquire hardware channels.
	 */
	if !batch {
		err = this.initialize(hwio.INPUT_CHANNELS, true)
	} else {
		err = this.initialize(numChannels, false)
	}

	/*
	 * Check if initialization was successful.
	 */
	if err != nil {
		msg := err.Error()
		msgNew := "Initialization failed: " + msg
		fmt.Printf("%s\n", msgNew)
	} else {
		cfg := this.config
		serverCfg := cfg.WebServer
		server := webserver.CreateWebServer(serverCfg)

		/*
		 * Check if we got a web server.
		 */
		if server == nil {
			fmt.Printf("%s\n", "Web server did not enter message loop.")
		} else {
			requests := server.RegisterCgi("/cgi-bin/dsp")
			server.Run()
			in := os.Stdin
			scanner := bufio.NewScanner(in)
			sampleRate := uint32(DEFAULT_SAMPLE_RATE)
			sampleRates := filter.SampleRates()

			/*
			 * If we are in batch mode, prepare file processing.
			 */
			if batch {
				sampleRate64 := uint64(0)
				correctRate := false

				/*
				 * Ask user to enter sample rate.
				 */
				for !correctRate {
					sampleRateString := this.getInput(scanner, "Target sample rate: ")
					sampleRate64, err = strconv.ParseUint(sampleRateString, 10, 64)
					sampleRate = uint32(sampleRate64)

					/*
					 * Check if sample rate could be parsed.
					 */
					if err == nil {

						/*
						 * Iterate over the supported sample rates.
						 */
						for _, currentRate := range sampleRates {

							/*
							 * Check if sample rate is supported.
							 */
							if currentRate == sampleRate {
								correctRate = true
							}

						}

					}

					/*
					 * If rate is not supported, output error message.
					 */
					if !correctRate {
						fmt.Printf("%s\n", "Sample rate not supported.")
					}

				}

			}

			this.sampleRate = sampleRate
			tlsPort := serverCfg.TLSPort
			fmt.Printf("Web interface ready: https://localhost:%s/\n", tlsPort)

			/*
			 * We should not terminate.
			 */
			for {
				this.running = true

				/*
				 * This is the actual message pump.
				 */
				for this.running {
					request := <-requests
					response := this.dispatch(request)
					respond := request.Respond
					respond <- response
				}

				/*
				 * If we are in batch mode, process files.
				 */
				if batch {
					this.processFiles(scanner, sampleRate)
				}

			}

		}

		this.finalize()
	}

}

/*
 * Creates a new controller.
 */
func CreateController() Controller {
	controller := controllerStruct{}
	return &controller
}

/*
 * Returns version information.
 */
func Version() (string, error) {
	content, err := ioutil.ReadFile(CONFIG_PATH)

	/*
	 * Check if file could be read.
	 */
	if err != nil {
		return "", fmt.Errorf("Failed to open config file: '%s'", CONFIG_PATH)
	} else {
		config := configStruct{}
		err = json.Unmarshal(content, &config)

		/*
		 * Check if file failed to unmarshal.
		 */
		if err != nil {
			return "", fmt.Errorf("Failed to decode config file: '%s'", CONFIG_PATH)
		} else {
			svr := config.WebServer
			version := svr.Name
			return version, nil
		}

	}

}
