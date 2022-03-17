"use strict";

/*
 * A class for storing global state required by the application.
 */
function Globals() {
	this.cgi = '/cgi-bin/dsp';
	this.mimeDefault = 'application/x-www-form-urlencoded';
	this.unitTypes = [];
}

/*
 * The global state object.
 */
const globals = new Globals();

/*
 * A class implementing data storage.
 */
function Storage() {
	const g_map = new WeakMap();

	/*
	 * Store a value under a key inside an element.
	 */
	this.put = function(elem, key, value) {
		let map = g_map.get(elem);

		/*
		 * Check if element is still unknown.
		 */
		if (map == null) {
			map = new Map();
			g_map.set(elem, map);
		}

		map.set(key, value);
	};

	/*
	 * Fetch a value from a key inside an element.
	 */
	this.get = function(elem, key, value) {
		const map = g_map.get(elem);

		/*
		 * Check if element is unknown.
		 */
		if (map == null) {
			return null;
		} else {
			const value = map.get(key);
			return value;
		}

	};

	/*
	 * Check if a certain key exists inside an element.
	 */
	this.has = function(elem, key) {
		const map = g_map.get(elem);

		/*
		 * Check if element is unknown.
		 */
		if (map == null) {
			return false;
		} else {
			const value = map.has(key);
			return value;
		}

	};

	/*
	 * Remove a certain key from an element.
	 */
	this.remove = function(elem, key) {
		const map = g_map.get(elem);

		/*
		 * Check if element is known.
		 */
		if (map != null) {
			map.delete(key);

			/*
			 * If inner map is now empty, remove it from outer map.
			 */
			if (map.size === 0) {
				g_map.delete(elem);
			}

		}

	};

}

/*
 * The global storage object.
 */
const storage = new Storage();

/*
 * A class supposed to make life a little easier.
 */
function Helper() {

	/*
	 * Blocks or unblocks the site for user interactions.
	 */
	this.blockSite = function(blocked) {
		const blocker = document.getElementById('blocker');
		let displayStyle = '';

		/*
		 * If we should block the site, display blocker, otherwise hide it.
		 */
		if (blocked) {
			displayStyle = 'block';
		} else {
			displayStyle = 'none';
		}

		/*
		 * Apply style if the site has a blocker.
		 */
		if (blocker !== null) {
			blocker.style.display = displayStyle;
		}

	};

	/*
	 * Removes all child nodes from an element.
	 */
	this.clearElement = function(elem) {

		/*
		 * As long as the element has child nodes, remove one.
		 */
		while (elem.hasChildNodes()) {
			const child = elem.firstChild;
			elem.removeChild(child);
		}

	};

	/*
	 * Parse JSON string into an object without raising exceptions.
	 */
	this.parseJSON = function(jsonString) {

		/*
		 * Try to parse JSON structure.
		 */
		try {
			const obj = JSON.parse(jsonString);
			return obj;
		} catch (ex) {
			return null;
		}

	};

}

/*
 * The (global) helper object.
 */
const helper = new Helper();

/*
 * A class implementing an Ajax engine.
 */
function Ajax() {

	/*
	 * Sends an Ajax request to the server.
	 *
	 * Parameters:
	 * - method (string): The request method (e. g. 'GET', 'POST', ...).
	 * - url (string): The request URL.
	 * - data (string): Data to be passed along the request (e. g. form data).
	 * - mimeType (string): Content type (MIME type) of the content sent to the server.
	 * - callback (function): The function to be called when a response is
	 *	                  returned from the server.
	 * - block (boolean): Whether the site should be blocked.
	 *
	 * Returns: Nothing.
	 */
	this.request = function(method, url, data, mimeType, callback, block) {
		const xhr = new XMLHttpRequest();

		/*
		 * Event handler for ReadyStateChange event.
		 */
		xhr.onreadystatechange = function() {
			helper.blockSite(block);

			/*
			 * If we got a response, pass the response text to
			 * the callback function.
			 */
			if (this.readyState === 4) {

				/*
				 * If we blocked the site on the request,
				 * unblock it on the response.
				 */
				if (block) {
					helper.blockSite(false);
				}

				/*
				 * Check if callback is registered.
				 */
				if (callback !== null) {
					const content = xhr.responseText;
					callback(content);
				}

			}

		};

		xhr.open(method, url, true);

		/*
		 * Set MIME type if requested.
		 */
		if (mimeType !== null) {
			xhr.setRequestHeader('Content-type', mimeType);
		}

		xhr.send(data);
	};

}

/*
 * The (global) Ajax engine.
 */
const ajax = new Ajax();

/*
 * A class implementing a key-value-pair.
 */
function KeyValuePair(key, value) {
	const g_key = key;
	const g_value = value;

	/*
	 * Returns the key stored in this key-value pair.
	 */
	this.getKey = function() {
		return g_key;
	};

	/*
	 * Returns the value stored in this key-value pair.
	 */
	this.getValue = function() {
		return g_value;
	};

}

/*
 * A class implementing a JSON request.
 */
function Request() {
	const g_keyValues = [];

	/*
	 * Append a key-value-pair to a request.
	 */
	this.append = function(key, value) {
		const kv = new KeyValuePair(key, value);
		g_keyValues.push(kv);
	};

	/*
	 * Returns the URL encoded data for this request.
	 */
	this.getData = function() {
		const numPairs = g_keyValues.length;
		let s = '';

		/*
		 * Iterate over the key-value pairs.
		 */
		for (let i = 0; i < numPairs; i++) {
			const keyValue = g_keyValues[i];
			const key = keyValue.getKey();
			const keyEncoded = encodeURIComponent(key);
			const value = keyValue.getValue();
			const valueEncoded = encodeURIComponent(value);

			/*
			 * If this is not the first key-value pair, we need a separator.
			 */
			if (i > 0) {
				s += '&';
			}

			s += keyEncoded + '=' + valueEncoded;
		}

		return s;
	};

}

/*
 * This class implements helper functions to build a user interface.
 */
function UI() {

	/*
	 * Strings for the user interface.
	 */
	const strings = {
		'add': 'Add',
		'add_unit': 'Add unit',
		'auto_wah': 'Auto wah',
		'auto_yoy': 'Auto yoy',
		'azimuth': 'Azimuth',
		'bandpass': 'Bandpass',
		'batch_processing': 'Batch processing',
		'beats_per_period': 'Beats per period',
		'bias': 'Bias',
		'boost': 'Boost',
		'bpm': 'BPM',
		'bypass': 'Bypass',
		'cabinet': 'Cabinet',
		'cents': 'Cents',
		'channel': 'Channel',
		'chorus': 'Chorus',
		'compressor': 'Compressor',
		'delay': 'Delay',
		'delay_time': 'Delay time',
		'depth': 'Depth',
		'distance': 'Distance',
		'distortion': 'Distortion',
		'drive': 'Drive',
		'dsp_load': 'DSP load',
		'enabled': 'Enabled',
		'excess': 'Excess',
		'feedback': 'Feedback',
		'file_transfer_instructions': 'Right-click here and select \'Save link / target as ...\' to save current patch. Drop patch file here to restore patch.',
		'filter_1': 'Filter 1',
		'filter_2': 'Filter 2',
		'filter_3': 'Filter 3',
		'filter_4': 'Filter 4',
		'filter_5': 'Filter 5',
		'filter_6': 'Filter 6',
		'filter_7': 'Filter 7',
		'filter_8': 'Filter 8',
		'filter_order': 'Filter order',
		'flanger': 'Flanger',
		'follow': 'Follow',
		'frames_per_period': 'Frames per period',
		'frequency': 'Frequency',
		'frequency_1': 'Frequency 1',
		'frequency_2': 'Frequency 2',
		'from_input': 'From: Input',
		'fuzz': 'Fuzz',
		'gain': 'Gain',
		'gain_limit': 'Gain limit',
		'high': 'High',
		'hold_time': 'Hold time',
		'input_amplitude': 'Input amplitude',
		'input_gain': 'Input gain',
		'latency': 'Latency',
		'level': 'Level',
		'level_1': 'Level 1',
		'level_2': 'Level 2',
		'level_3': 'Level 3',
		'level_4': 'Level 4',
		'level_5': 'Level 5',
		'level_6': 'Level 6',
		'level_7': 'Level 7',
		'level_8': 'Level 8',
		'level_clean': 'Level clean',
		'level_dist': 'Level dist',
		'level_hysteresis': 'Level hysteresis',
		'level_octave_down_first': 'Level octave down (I)',
		'level_octave_down_second': 'Level octave down (II)',
		'level_octave_up': 'Level octave up',
		'low': 'Low',
		'master': 'Master',
		'metronome': 'Metronome',
		'middle': 'Middle',
		'mix': 'Mix',
		'move_down': 'Move down',
		'move_up': 'Move up',
		'noise_gate': 'Noise gate',
		'note': 'Note',
		'octaver': 'Octaver',
		'overdrive': 'Overdrive',
		'oversampling': 'Oversampling',
		'persistence': 'Persistence',
		'phase': 'Phase',
		'phaser': 'Phaser',
		'polarity': 'Polarity',
		'power_amp': 'Power amp',
		'presence': 'Presence',
		'process_now': 'Process now',
		'remove': 'Remove',
		'reverb': 'Reverb',
		'ring_modulator': 'Ring modulator',
		'signal_amplitude': 'Signal amplitude',
		'signal_frequency': 'Signal frequency',
		'signal_gain': 'Signal gain',
		'signal_generator': 'Signal generator',
		'signal_levels': 'Signal levels',
		'signal_type': 'Signal type',
		'spatializer': 'Spatializer',
		'speed': 'Speed',
		'target_level': 'Target level',
		'threshold_close': 'Threshold close',
		'threshold_open': 'Threshold open',
		'tick_sound': 'Tick sound',
		'tock_sound': 'Tock sound',
		'to_output': 'To: Output',
		'tone_stack': 'Tone stack',
		'tremolo': 'Tremolo',
		'tuner': 'Tuner',
		'type': 'Type',
		'valve': 'Valve'
	};

	/*
	 * Obtains a string for the user interface.
	 */
	this.getString = function(key) {

		/*
		 * Check whether key is defined in strings.
		 */
		if (key in strings) {
			const s = strings[key];
			return s;
		} else {
			return key;
		}

	};

	/*
	 * Creates a turnable knob.
	 */
	this.createKnob = function(params) {
		const label = params.label;
		const physicalUnit = params.physicalUnit;
		const valueMin = params.valueMin;
		const valueMax = params.valueMax;
		const valueDefault = params.valueDefault;
		const valueWidth = params.valueWidth;
		const valueHeight = params.valueHeight;
		const valueAngle = params.angle;
		const valueAngleArc = (valueAngle / 180.0) * Math.PI;
		const valueCursor = params.cursor;
		const valueReadonly = params.readonly;
		const colorScheme = params.colorScheme;
		const angleArc = (valueAngle / 180.0) * Math.PI;
		const halfAngleArc = 0.5 * valueAngleArc;
		const paramDiv = document.createElement('div');
		paramDiv.classList.add('paramdiv');
		const labelDiv = document.createElement('div');
		labelDiv.classList.add('knoblabel');
		const labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		paramDiv.appendChild(labelDiv);
		const knobDiv = document.createElement('div');
		knobDiv.classList.add('knobdiv');
		let fgColor = '#ff8800';
		let bgColor = '#181818';
		let labelColor = '#666666';

		/*
		 * Check if we want a blue or green color scheme.
		 */
		if (colorScheme === 'blue') {
			fgColor = '#8888ff';
			bgColor = '#181830';
			labelColor = '#4444cc';
		} else if (colorScheme === 'green') {
			fgColor = '#88ff88';
			bgColor = '#181830';
			labelColor = '#ffffff';
		}

		const knobElem = pureknob.createKnob(valueHeight, valueWidth);
		knobElem.setProperty('angleStart', -halfAngleArc);
		knobElem.setProperty('angleEnd', halfAngleArc);
		knobElem.setProperty('colorBG', bgColor);
		knobElem.setProperty('colorFG', fgColor);
		knobElem.setProperty('colorLabel', labelColor);
		knobElem.setProperty('label', physicalUnit);
		knobElem.setProperty('needle', valueCursor);
		knobElem.setProperty('readonly', valueReadonly);
		knobElem.setProperty('valMin', valueMin);
		knobElem.setProperty('valMax', valueMax);
		knobElem.setValue(valueDefault);
		const knobNode = knobElem.node();
		knobDiv.appendChild(knobNode);
		paramDiv.appendChild(knobDiv);

		/*
		 * Create knob.
		 */
		const knob = {
			'div': paramDiv,
			'node': knobNode,
			'obj': knobElem
		};

		return knob;
	};

	/*
	 * Creates a drop down menu.
	 */
	this.createDropDown = function(params) {
		const label = params.label;
		const options = params.options;
		const numOptions = options.length;
		const selectedIndex = params.selectedIndex;
		const paramDiv = document.createElement('div');
		paramDiv.classList.add('paramdiv');

		/*
		 * Check if we should apply a label.
		 */
		if (label !== null) {
			const labelDiv = document.createElement('div');
			labelDiv.classList.add('dropdownlabel');
			const labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			paramDiv.appendChild(labelDiv);
		}

		const selectElem = document.createElement('select');
		selectElem.classList.add('dropdown');

		/*
		 * Add all options.
		 */
		for (let i = 0; i < numOptions; i++) {
			const optionElem = document.createElement('option');
			optionElem.text = options[i];
			selectElem.add(optionElem);
		}

		selectElem.selectedIndex = selectedIndex;
		paramDiv.appendChild(selectElem);

		/*
		 * Create dropdown.
		 */
		const dropdown = {
			'div': paramDiv,
			'input': selectElem
		};

		return dropdown;
	};

	/*
	 * Creates a button.
	 */
	this.createButton = function(params) {
		const caption = params.caption;
		const active = params.active;
		const elem = document.createElement('button');
		elem.classList.add('button');

		/*
		 * Check whether the button should be active.
		 */
		if (active) {
			elem.classList.add('buttonactive');
		} else {
			elem.classList.add('buttonnormal');
		}

		const captionNode = document.createTextNode(caption);
		elem.appendChild(captionNode);

		/*
		 * Create button.
		 */
		const button = {
			'input': elem
		};

		return button;
	};

	/*
	 * Creates a unit.
	 */
	this.createUnit = function(params) {
		const typeString = params.type;
		const buttonsParam = params.buttons;
		const numButtonsParam = buttonsParam.length;
		const buttons = [];

		/*
		 * Iterate over the buttons.
		 */
		for (let i = 0; i < numButtonsParam; i++) {
			const buttonParam = buttonsParam[i];
			const label = buttonParam.label;
			const active = buttonParam.active;

			/*
			 * Parameters for the button.
			 */
			const params = {
				'caption': label,
				'active': active
			};

			const button = this.createButton(params);
			buttons.push(button);
		}

		const unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		const headerDiv = document.createElement('div');
		headerDiv.classList.add('headerdiv');
		const numButtons = buttons.length;

		/*
		 * Add buttons to header.
		 */
		for (let i = 0; i < numButtons; i++) {
			const button = buttons[i];
			headerDiv.appendChild(button.input);
		}

		const labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		const typeNode = document.createTextNode(typeString);
		labelDiv.appendChild(typeNode);
		headerDiv.appendChild(labelDiv);
		unitDiv.appendChild(headerDiv);
		const controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);

		/*
		 * Create unit.
		 */
		const unit = {
			'div': unitDiv,
			'controls': controlsDiv,
			'buttons': buttons,
			'expanded': false
		};

		/*
		 * Adds a control to a unit.
		 */
		unit.addControl = function(control) {
			const controlDiv = control.div;
			this.controls.appendChild(controlDiv);
		};

		/*
		 * Adds a row with controls to a unit.
		 */
		unit.addControlRow = function(controls) {
			const rowDiv = document.createElement('div');
			const numControls = controls.length;

			/*
			 * Insert controls into the row.
			 */
			for (let i = 0; i < numControls; i++) {
				const control = controls[i];
				const controlDiv = control.div;
				rowDiv.appendChild(controlDiv);
			}

			this.controls.appendChild(rowDiv);
		};

		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			const controlsDiv = this.controls;
			let displayValue = '';

			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value) {
				displayValue = 'block';
			} else {
				displayValue = 'none';
			}

			controlsDiv.style.display = displayValue;
			this.expanded = value;
		};

		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		};

		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			const state = this.getExpanded();
			this.setExpanded(!state);
		};

		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(e) {
			const unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		};

		storage.put(labelDiv, 'unit', unit);
		return unit;
	};

	/*
	 * Checks whether a certain combination of unit type and parameter name requires special handling.
	 */
	this.isSpecialParameter = function(unitType, paramName) {

		/*
		 * Discern unit type.
		 */
		switch (unitType) {
			case 'power_amp':

				/*
				 * Only parameters of the form 'level_N' and 'filter_N',
				 * where N is a number, require special handling.
				 */
				if (paramName.startsWith('level_')) {
					return true;
				} else if (paramName.startsWith('filter_')) {
					const suffix = paramName.substring(7);
					const isNumeric = isFinite(suffix);
					return isNumeric;
				}

			default:
				return false;
		}

	};

	/*
	 * Renders a unit given chain and unit ID, as well as a description returned from the server.
	 */
	this.renderUnit = function(chainId, unitId, description) {
		const bypassButtonLabel = ui.getString('bypass');
		const moveUpButtonLabel = ui.getString('move_up');
		const moveDownButtonLabel = ui.getString('move_down');
		const removeButtonLabel = ui.getString('remove');
		const unitTypes = globals.unitTypes;
		const unitTypeId = description.Type;
		const unitType = unitTypes[unitTypeId];
		const unitTypeString = ui.getString(unitType);
		const bypassActive = description.Bypass;

		/*
		 * Buttons for this unit.
		 */
		const buttons = [
			{
				'label': bypassButtonLabel,
				'active': bypassActive
			},
			{
				'label': moveUpButtonLabel,
				'active': false
			},
			{
				'label': moveDownButtonLabel,
				'active': false
			},
			{
				'label': removeButtonLabel,
				'active': false
			}
		];

		/*
		 * Parameters for the unit UI element.
		 */
		const paramsUnit = {
			'type': unitTypeString,
			'buttons': buttons
		};

		const unit = ui.createUnit(paramsUnit);
		const btnBypass = unit.buttons[0].input;
		storage.put(btnBypass, 'chain', chainId);
		storage.put(btnBypass, 'unit', unitId);
		storage.put(btnBypass, 'active', bypassActive);

		/*
		 * This is invoked when someone clicks on the 'bypass' button.
		 */
		btnBypass.onclick = function(e) {
			const chainId = storage.get(this, 'chain');
			const unitId = storage.get(this, 'unit');
			const active = !storage.get(this, 'active');

			/*
			 * Check whether the control should be active.
			 */
			if (active) {
				this.classList.remove('buttonnormal');
				this.classList.add('buttonactive');
			} else {
				this.classList.remove('buttonactive');
				this.classList.add('buttonnormal');
			}

			storage.put(this, 'active', active);
			handler.setBypass(chainId, unitId, active);
		};

		const btnMoveUp = unit.buttons[1].input;
		storage.put(btnMoveUp, 'chain', chainId);
		storage.put(btnMoveUp, 'unit', unitId);

		/*
		 * This is invoked when someone clicks on the 'move up' button.
		 */
		btnMoveUp.onclick = function(e) {
			const chainId = storage.get(this, 'chain');
			const unitId = storage.get(this, 'unit');
			handler.moveUp(chainId, unitId);
		};

		const btnMoveDown = unit.buttons[2].input;
		storage.put(btnMoveDown, 'chain', chainId);
		storage.put(btnMoveDown, 'unit', unitId);

		/*
		 * This is invoked when someone clicks on the 'move down' button.
		 */
		btnMoveDown.onclick = function(e) {
			const chainId = storage.get(this, 'chain');
			const unitId = storage.get(this, 'unit');
			handler.moveDown(chainId, unitId);
		};

		const btnRemove = unit.buttons[3].input;
		btnRemove.classList.add('buttonremove');
		storage.put(btnRemove, 'chain', chainId);
		storage.put(btnRemove, 'unit', unitId);

		/*
		 * This is invoked when someone clicks on the 'remove' button.
		 */
		btnRemove.onclick = function(e) {
			const chainId = storage.get(this, 'chain');
			const unitId = storage.get(this, 'unit');
			handler.removeUnit(chainId, unitId);
		};

		const unitParams = description.Parameters;
		const numParams = unitParams.length;

		/*
		 * Iterate over the parameters and add all 'ordinary' (non-special) ones to the unit.
		 */
		for (let i = 0; i < numParams; i++) {
			const param = unitParams[i];
			const paramType = param.Type;
			const paramName = param.Name;
			const paramPhysicalUnit = param.PhysicalUnit;
			const paramMinimum = param.Minimum;
			const paramMaximum = param.Maximum;
			const paramNumericValue = param.NumericValue;
			const paramDiscreteValues = param.DiscreteValues;
			const paramDiscreteValueIndex = param.DiscreteValueIndex;
			const isSpecial = this.isSpecialParameter(unitType, paramName);

			/*
			 * Only handle 'ordinary' (non-special) parameters on the first pass.
			 */
			if (!isSpecial) {
				const isFloating = (i !== 0);
				const label = ui.getString(paramName);

				/*
				 * Handle numeric parameter.
				 */
				if (paramType === 'numeric') {

					/*
					 * Parameters for the knob.
					 */
					const params = {
						'label': label,
						'physicalUnit': paramPhysicalUnit,
						'valueMin': paramMinimum,
						'valueMax': paramMaximum,
						'valueDefault': paramNumericValue,
						'valueWidth': 150,
						'valueHeight': 150,
						'angle': 270,
						'cursor': false,
						'colorScheme': 'default',
						'readonly': false
					};

					const knob = ui.createKnob(params);
					unit.addControl(knob);
					const knobNode = knob.node;
					storage.put(knobNode, 'chain', chainId);
					storage.put(knobNode, 'unit', unitId);
					storage.put(knobNode, 'param', paramName);

					/*
					 * This is called when a numeric value changes.
					 */
					const knobHandler = function(knob, value) {
						const knobNode = knob.node();
						const chain = storage.get(knobNode, 'chain');
						const unit = storage.get(knobNode, 'unit');
						const param = storage.get(knobNode, 'param');
						handler.setNumericValue(chain, unit, param, value);
					};

					const knobObj = knob.obj;
					knobObj.addListener(knobHandler);
				}

				/*
				 * Handle discrete parameter.
				 */
				if (paramType === 'discrete') {

					/*
					 * Parameters for the drop down menu.
					 */
					const params = {
						'label': label,
						'options': paramDiscreteValues,
						'selectedIndex': paramDiscreteValueIndex
					};

					const dropDown = ui.createDropDown(params);
					const dropDownInput = dropDown.input;
					storage.put(dropDownInput, 'chain', chainId);
					storage.put(dropDownInput, 'unit', unitId);
					storage.put(dropDownInput, 'param', paramName);

					/*
					 * This is called when a discrete value changes.
					 */
					dropDownInput.onchange = function(e) {
						const chain = storage.get(this, 'chain');
						const unit = storage.get(this, 'unit');
						const param = storage.get(this, 'param');
						const idx = this.selectedIndex;
						const option = this.options[idx];
						const value = option.text;
						handler.setDiscreteValue(chain, unit, param, value);
					};

					unit.addControl(dropDown);
				}

			}

		}

		/*
		 * Iterate over the parameters and add all special discrete ones to the unit.
		 */
		for (let i = 0; i < numParams; i++) {
			const param = unitParams[i];
			const paramType = param.Type;
			const paramName = param.Name;
			const paramDiscreteValues = param.DiscreteValues;
			const paramDiscreteValueIndex = param.DiscreteValueIndex;
			const isSpecial = this.isSpecialParameter(unitType, paramName);

			/*
			 * Only handle special discrete parameters on the second pass.
			 */
			if (isSpecial & (paramType === 'discrete')) {
				const label = ui.getString(paramName);

				/*
				 * Parameters for the drop down menu.
				 */
				const params = {
					'label': label,
					'options': paramDiscreteValues,
					'selectedIndex': paramDiscreteValueIndex
				};

				const dropDown = ui.createDropDown(params);
				const dropDownInput = dropDown.input;
				storage.put(dropDownInput, 'chain', chainId);
				storage.put(dropDownInput, 'unit', unitId);
				storage.put(dropDownInput, 'param', paramName);

				/*
				 * This is called when a discrete value changes.
				 */
				dropDownInput.onchange = function(e) {
					const chain = storage.get(this, 'chain');
					const unit = storage.get(this, 'unit');
					const param = storage.get(this, 'param');
					const idx = this.selectedIndex;
					const option = this.options[idx];
					const value = option.text;
					handler.setDiscreteValue(chain, unit, param, value);
				};

				const controlRow = [];
				controlRow.push(dropDown);
				unit.addControlRow(controlRow);
			}

		}

		/*
		 * Iterate over the parameters and add all special numeric ones to the unit.
		 */
		for (let i = 0; i < numParams; i++) {
			const param = unitParams[i];
			const paramType = param.Type;
			const paramName = param.Name;
			const paramPhysicalUnit = param.PhysicalUnit;
			const paramMinimum = param.Minimum;
			const paramMaximum = param.Maximum;
			const paramNumericValue = param.NumericValue;
			const isSpecial = this.isSpecialParameter(unitType, paramName);

			/*
			 * Only handle special numeric parameters on the third pass.
			 */
			if (isSpecial & (paramType === 'numeric')) {
				const label = ui.getString(paramName);

				/*
				 * Parameters for the knob.
				 */
				const params = {
					'label': label,
					'physicalUnit': paramPhysicalUnit,
					'valueMin': paramMinimum,
					'valueMax': paramMaximum,
					'valueDefault': paramNumericValue,
					'valueWidth': 150,
					'valueHeight': 150,
					'angle': 270,
					'cursor': false,
					'colorScheme': 'default',
					'readonly': false
				};

				const knob = ui.createKnob(params);
				unit.addControl(knob);
				const knobNode = knob.node;
				storage.put(knobNode, 'chain', chainId);
				storage.put(knobNode, 'unit', unitId);
				storage.put(knobNode, 'param', paramName);

				/*
				 * This is called when a numeric value changes.
				 */
				const knobHandler = function(knob, value) {
					const knobNode = knob.node();
					const chain = storage.get(knobNode, 'chain');
					const unit = storage.get(knobNode, 'unit');
					const param = storage.get(knobNode, 'param');
					handler.setNumericValue(chain, unit, param, value);
				};

				const knobObj = knob.obj;
				knobObj.addListener(knobHandler);
			}

		}

		return unit;
	};

	/*
	 * Renders a signal chain, given its ID and a chain description returned from the server.
	 */
	this.renderSignalChain = function(id, description) {
		const idString = id.toString();
		const chainDiv = document.createElement('div');
		const beginDiv = document.createElement('div');
		beginDiv.classList.add('contentdiv');
		beginDiv.classList.add('iodiv');
		const beginHeaderDiv = document.createElement('div');
		beginHeaderDiv.classList.add('headerdiv');
		const beginLabelDiv = document.createElement('div');
		beginLabelDiv.classList.add('labeldiv');
		const labelFromInput = ui.getString('from_input');
		const beginLabelText = labelFromInput + ' ' + idString;
		const beginLabelNode = document.createTextNode(beginLabelText);
		beginLabelDiv.appendChild(beginLabelNode);
		beginHeaderDiv.appendChild(beginLabelDiv);
		beginDiv.appendChild(beginHeaderDiv);
		chainDiv.appendChild(beginDiv);
		const units = description.Units;
		const numUnits = units.length;

		/*
		 * Iterate over the units in this chain.
		 */
		for (let i = 0; i < numUnits; i++) {
			const unit = units[i];
			const result = this.renderUnit(id, i, unit);
			const unitDiv = result.div;
			chainDiv.appendChild(unitDiv);
		}

		const labelDropdown = ui.getString('add_unit');
		const labelButton = ui.getString('add');
		const unitTypes = globals.unitTypes;
		const numUnitTypes = unitTypes.length;
		const unitTypeNames = [];

		/*
		 * Look up the name of the unit types.
		 */
		for (let i = 0; i < numUnitTypes; i++) {
			const unitType = unitTypes[i];
			const unitTypeName = ui.getString(unitType);
			unitTypeNames.push(unitTypeName);
		}

		/*
		 * Parameters for the drop down menu.
		 */
		var paramsDropDown = {
			'label': labelDropdown,
			'options': unitTypeNames,
			'selectedIndex': 0
		};

		var dropDown = ui.createDropDown(paramsDropDown);

		/*
		 * Parameters for the 'create' button.
		 */
		var paramsButton = {
			'caption': labelButton,
			'active': false
		};

		var button = ui.createButton(paramsButton);
		var buttonElem = button.input;

		/*
		 * What happens when we click on the 'add' button.
		 */
		buttonElem.onclick = function(e) {
			const chainId = storage.get(this, 'chain');
			const dropdown = storage.get(this, 'dropdown');
			const unitType = dropdown.selectedIndex;
			handler.addUnit(unitType, chainId);
		};

		storage.put(buttonElem, 'chain', id);
		storage.put(buttonElem, 'dropdown', dropDown.input);
		const dropDownDiv = document.createElement('div');
		dropDownDiv.classList.add('contentdiv');
		dropDownDiv.classList.add('addunitdiv');
		dropDownDiv.appendChild(dropDown.div);
		dropDownDiv.appendChild(buttonElem);
		chainDiv.appendChild(dropDownDiv);
		const endDiv = document.createElement('div');
		endDiv.classList.add('contentdiv');
		endDiv.classList.add('iodiv');
		const endHeaderDiv = document.createElement('div');
		endHeaderDiv.classList.add('headerdiv');
		const endLabelDiv = document.createElement('div');
		endLabelDiv.classList.add('labeldiv');
		const labelToOutput = ui.getString('to_output');
		const endLabelText = labelToOutput + ' ' + idString;
		const endLabelNode = document.createTextNode(endLabelText);
		endLabelDiv.appendChild(endLabelNode);
		endHeaderDiv.appendChild(endLabelDiv);
		endDiv.appendChild(endHeaderDiv);
		chainDiv.appendChild(endDiv);

		/*
		 * This object represents the signal chain.
		 */
		const chain = {
			'div': chainDiv
		};

		return chain;
	}

	/*
	 * Renders the signal chains given a configuration returned from the server.
	 */
	this.renderSignalChains = function(configuration) {
		const elem = document.getElementById('signal_chains');
		helper.clearElement(elem);
		const chains = configuration.Chains;
		const numChains = chains.length;

		/*
		 * Iterate over the signal chains.
		 */
		for (let i = 0; i < numChains; i++) {
			const chain = chains[i];
			const result = this.renderSignalChain(i, chain);
			const chainDiv = result.div;
			elem.append(chainDiv);
			const spacerDiv = document.createElement('div');
			spacerDiv.classList.add('spacerdiv');
			elem.appendChild(spacerDiv);
		}

	};

	/*
	 * Renders the persistence area given a configuration returned from the server.
	 */
	this.renderPersistence = function(configuration) {
		const elem = document.getElementById('persistence');
		helper.clearElement(elem);
		const unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		unitDiv.classList.add('masterunitdiv');
		const headerDiv = document.createElement('div');
		const labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		labelDiv.classList.add('io');
		const label = ui.getString('persistence');
		const labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		headerDiv.appendChild(labelDiv);
		headerDiv.classList.add('headerdiv');
		unitDiv.appendChild(headerDiv);
		const controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		elem.appendChild(unitDiv);
		const uploadAreaDiv = document.createElement('div');
		uploadAreaDiv.classList.add('uploadarea');
		uploadAreaDiv.addEventListener('dragend', handler.dragLeave);
		uploadAreaDiv.addEventListener('dragenter', handler.dragEnter);
		uploadAreaDiv.addEventListener('dragleave', handler.dragLeave);
		uploadAreaDiv.addEventListener('dragover', handler.absorbEvent);
		uploadAreaDiv.addEventListener('drop', handler.uploadFile);
		const downloadAnchor = document.createElement('a');
		const cgi = globals.cgi;
		const downloadTarget = cgi + '?cgi=persistence-save';
		downloadAnchor.setAttribute('href', downloadTarget);
		downloadAnchor.setAttribute('target', '_blank');
		downloadAnchor.classList.add('link');
		downloadAnchor.classList.add('auto');
		const instructionsString = ui.getString('file_transfer_instructions');
		const instructionsNode = document.createTextNode(instructionsString);
		downloadAnchor.appendChild(instructionsNode);
		uploadAreaDiv.appendChild(downloadAnchor);
		controlsDiv.appendChild(uploadAreaDiv);

		/*
		 * Create unit object.
		 */
		const unit = {
			'controls': controlsDiv,
			'expanded': false
		};

		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			const controlsDiv = this.controls;
			let displayValue = '';

			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value) {
				displayValue = 'block';
			} else {
				displayValue = 'none';
			}

			controlsDiv.style.display = displayValue;
			this.expanded = value;
		};

		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		};

		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			const state = this.getExpanded();
			this.setExpanded(!state);
		};

		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(e) {
			const unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		};

		storage.put(labelDiv, 'unit', unit);
	};

	/*
	 * Renders the latency configuration given a configuration returned from the server.
	 */
	this.renderLatency = function(configuration) {
		const batchProcessing = configuration.BatchProcessing;
		const elem = document.getElementById('latency');
		helper.clearElement(elem);

		/*
		 * Only display latency controls if batch processing is disabled on the server.
		 */
		if (batchProcessing === false) {
			const unitDiv = document.createElement('div');
			unitDiv.classList.add('contentdiv');
			unitDiv.classList.add('masterunitdiv');
			const headerDiv = document.createElement('div');
			const labelDiv = document.createElement('div');
			labelDiv.classList.add('labeldiv');
			labelDiv.classList.add('active');
			labelDiv.classList.add('io');
			const label = ui.getString('latency');
			const labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			headerDiv.appendChild(labelDiv);
			headerDiv.classList.add('headerdiv');
			unitDiv.appendChild(headerDiv);
			const controlsDiv = document.createElement('div');
			controlsDiv.classList.add('controlsdiv');
			unitDiv.appendChild(controlsDiv);
			elem.appendChild(unitDiv);
			const labelFramesPerPeriod = ui.getString('frames_per_period');
			const framesPerPeriod = configuration.FramesPerPeriod;
			const dropdownRow = document.createElement('div');
			const values = [16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192];
			let valueIdx = 0;

			/*
			 * Iterate over all possible values.
			 */
			for (let i = 0; i < values.length; i++) {
				const currentValue = values[i];

				/*
				 * If we have a match, store index.
				 */
				if (framesPerPeriod === currentValue) {
					valueIdx = i;
				}

			}

			/*
			 * Parameters for the frames per period drop down menu.
			 */
			const paramsFramesPerPeriod = {
				'label': labelFramesPerPeriod,
				'options': values,
				'selectedIndex': valueIdx
			};

			const dropDownFramesPerPeriod = ui.createDropDown(paramsFramesPerPeriod);
			const dropDownFramesPerPeriodElem = dropDownFramesPerPeriod.input;

			/*
			 * This is called when the period size changes.
			 */
			dropDownFramesPerPeriodElem.onchange = function(e) {
				const idx = this.selectedIndex;
				const option = this.options[idx];
				const value = option.text;
				handler.setFramesPerPeriod(value);
			};

			dropdownRow.appendChild(dropDownFramesPerPeriod.div);
			controlsDiv.appendChild(dropdownRow);

			/*
			 * Create unit object.
			 */
			const unit = {
				'controls': controlsDiv,
				'expanded': false
			};

			/*
			 * Expands or collapses a unit.
			 */
			unit.setExpanded = function(value) {
				const controlsDiv = this.controls;
				let displayValue = '';

				/*
				 * Check whether we should expand or collapse the unit.
				 */
				if (value) {
					displayValue = 'block';
				} else {
					displayValue = 'none';
				}

				controlsDiv.style.display = displayValue;
				this.expanded = value;
			};

			/*
			 * Returns whether a unit is expanded.
			 */
			unit.getExpanded = function() {
				return this.expanded;
			};

			/*
			 * Toggles a unit between expanded and collapsed state.
			 */
			unit.toggleExpanded = function() {
				const state = this.getExpanded();
				this.setExpanded(!state);
			};

			/*
			 * This is called when a user clicks on the label div.
			 */
			labelDiv.onclick = function(e) {
				const unit = storage.get(this, 'unit');
				unit.toggleExpanded();
			};

			storage.put(labelDiv, 'unit', unit);
		}

	};

	/*
	 * Updates the tuner display based on information returned from the server.
	 */
	this.updateTuner = function(result) {
		const cents = result.Cents;
		const frequency = result.Frequency;
		const note = result.Note;
		const centsDiv = document.querySelector('.tunercentsknob');
		const centsKnob = storage.get(centsDiv, 'knob');
		centsKnob.setValue(cents);
		const frequencyDiv = document.querySelector('.tunerfrequencydiv');
		const frequencyString = frequency.toFixed(4);
		frequencyDiv.innerHTML = frequencyString;
		const noteDiv = document.querySelector('.tunernotediv');
		const noteString = note.toString();
		noteDiv.innerHTML = noteString;
	};

	/*
	 * Renders the tuner given a configuration returned from the server.
	 */
	this.renderTuner = function(configuration) {
		const batchProcessing = configuration.BatchProcessing;
		const elem = document.getElementById('tuner');
		helper.clearElement(elem);

		/*
		 * Only display tuner if batch processing is disabled on the server.
		 */
		if (batchProcessing === false) {
			const chainsConfiguration = configuration.Chains;
			const numChannels = chainsConfiguration.length;
			const tunerConfiguration = configuration.Tuner;
			const unitDiv = document.createElement('div');
			unitDiv.classList.add('contentdiv');
			unitDiv.classList.add('masterunitdiv');
			const headerDiv = document.createElement('div');
			const labelDiv = document.createElement('div');
			labelDiv.classList.add('labeldiv');
			labelDiv.classList.add('active');
			labelDiv.classList.add('io');
			const label = ui.getString('tuner');
			const labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			headerDiv.appendChild(labelDiv);
			headerDiv.classList.add('headerdiv');
			unitDiv.appendChild(headerDiv);
			const controlsDiv = document.createElement('div');
			controlsDiv.classList.add('controlsdiv');
			unitDiv.appendChild(controlsDiv);
			elem.appendChild(unitDiv);
			const centsString = ui.getString('cents');
			const frequencyString = ui.getString('frequency');
			const noteString = ui.getString('note');
			const centsValue = tunerConfiguration.BeatsPerPeriod;

			/*
			 * Parameters for the cents knob.
			 */
			const centsParams = {
				'label': centsString,
				'physicalUnit': null,
				'valueMin': -50,
				'valueMax': 50,
				'valueDefault': 0,
				'valueWidth': 150,
				'valueHeight': 150,
				'angle': 270,
				'cursor': true,
				'colorScheme': 'green',
				'readonly': true
			};

			const centsKnob = ui.createKnob(centsParams);
			const centsKnobNode = centsKnob.node;
			centsKnobNode.classList.add('tunercentsknob');
			const centsKnobObj = centsKnob.obj;
			storage.put(centsKnobNode, 'knob', centsKnobObj);
			const centsKnobDiv = centsKnob.div;
			controlsDiv.appendChild(centsKnobDiv);
			const frequencyRow = document.createElement('div');
			const labelFrequency = ui.getString('frequency');
			const frequencyLabelDiv = document.createElement('div');
			frequencyLabelDiv.classList.add('labeldiv');
			const frequencyLabelNode = document.createTextNode(labelFrequency);
			frequencyLabelDiv.appendChild(frequencyLabelNode);
			frequencyRow.appendChild(frequencyLabelDiv);
			const frequencyValueDiv = document.createElement('div');
			frequencyValueDiv.classList.add('tunerfrequencydiv');
			frequencyRow.appendChild(frequencyValueDiv);
			controlsDiv.appendChild(frequencyRow);
			const noteRow = document.createElement('div');
			const labelNote = ui.getString('note');
			const noteLabelDiv = document.createElement('div');
			noteLabelDiv.classList.add('labeldiv');
			const noteLabelNode = document.createTextNode(labelNote);
			noteLabelDiv.appendChild(noteLabelNode);
			noteRow.appendChild(noteLabelDiv);
			const noteNameDiv = document.createElement('div');
			noteNameDiv.classList.add('tunernotediv');
			noteRow.appendChild(noteNameDiv);
			controlsDiv.appendChild(noteRow);
			const channelRow = document.createElement('div');
			const labelChannel = ui.getString('channel');
			const channels = ['- NONE -'];

			/*
			 * Append indices for all channels.
			 */
			for (let i = 0; i < numChannels; i++) {
				const idxString = i.toString();
				channels.push(idxString);
			}

			const channelIdx = tunerConfiguration.Channel;
			const channelIdxInc = channelIdx + 1;

			/*
			 * Parameters for the channel drop down menu.
			 */
			const paramsChannel = {
				'label': labelChannel,
				'options': channels,
				'selectedIndex': channelIdxInc
			};

			const dropDownChannel = ui.createDropDown(paramsChannel);
			const dropDownChannelElem = dropDownChannel.input;

			/*
			 * This is called when the channel number changes.
			 */
			dropDownChannelElem.onchange = function(e) {
				const idx = this.selectedIndex;
				const option = this.options[idx];
				let value = option.text;
				const interval = storage.get(this, 'interval');
				window.clearInterval(interval);

				/*
				 * This gets executed whenever the timer ticks.
				 */
				const callback = function() {
					handler.refreshTuner();
				};

				/*
				 * Handle special case of no channel and register timer
				 * for updating readings for the UI.
				 */
				if (value === '- NONE -') {
					value = '-1';
				} else {
					const intervalNew = window.setInterval(callback, 250);
					storage.put(this, 'interval', intervalNew);
				}

				/*
				 * Set input channel when this is an actual event.
				 */
				if (e !== null) {
					handler.setTunerValue('channel', value);
				}

			};

			dropDownChannelElem.onchange(null);
			const dropDownChannelDiv = dropDownChannel.div;
			channelRow.appendChild(dropDownChannelDiv);
			controlsDiv.appendChild(channelRow);

			/*
			 * Create unit object.
			 */
			const unit = {
				'controls': controlsDiv,
				'expanded': false
			};

			/*
			 * Expands or collapses a unit.
			 */
			unit.setExpanded = function(value) {
				const controlsDiv = this.controls;
				let displayValue = '';

				/*
				 * Check whether we should expand or collapse the unit.
				 */
				if (value) {
					displayValue = 'block';
				} else {
					displayValue = 'none';
				}

				controlsDiv.style.display = displayValue;
				this.expanded = value;
			};

			/*
			 * Returns whether a unit is expanded.
			 */
			unit.getExpanded = function() {
				return this.expanded;
			};

			/*
			 * Toggles a unit between expanded and collapsed state.
			 */
			unit.toggleExpanded = function() {
				const state = this.getExpanded();
				this.setExpanded(!state);
			};

			/*
			 * This is called when a user clicks on the label div.
			 */
			labelDiv.onclick = function(e) {
				const unit = storage.get(this, 'unit');
				unit.toggleExpanded();
			}

			storage.put(labelDiv, 'unit', unit);
		}

	}

	/*
	 * Renders the spatializer given a configuration returned from the server.
	 */
	this.renderSpatializer = function(configuration) {
		const spatializer = configuration.Spatializer;
		const channels = spatializer.Channels;
		const numChannels = channels.length;
		const elem = document.getElementById('spatializer');
		helper.clearElement(elem);
		const unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		unitDiv.classList.add('masterunitdiv');
		const headerDiv = document.createElement('div');
		const labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		labelDiv.classList.add('io');
		const label = ui.getString('spatializer');
		const labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		headerDiv.appendChild(labelDiv);
		headerDiv.classList.add('headerdiv');
		unitDiv.appendChild(headerDiv);
		const controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		elem.appendChild(unitDiv);

		/*
		 * Iterate over the channels.
		 */
		for (let i = 0; i < numChannels; i++) {
			const iString = i.toString();
			const channel = channels[i];
			const azimuth = channel.Azimuth;
			const distance = 10 * channel.Distance;
			const level = 100 * channel.Level;
			const azimuthString = ui.getString('azimuth');
			const azimuthLabel = azimuthString + ' ' + iString;

			/*
			 * Parameters for the azimuth knob.
			 */
			const azimuthParams = {
				'label': azimuthLabel,
				'physicalUnit': '°',
				'valueMin': -90,
				'valueMax': 90,
				'valueDefault': azimuth,
				'valueWidth': 150,
				'valueHeight': 150,
				'angle': 180,
				'cursor': true,
				'colorScheme': 'blue',
				'readonly': false
			};

			const azimuthKnob = ui.createKnob(azimuthParams);
			const azimuthKnobDiv = azimuthKnob.div;
			controlsDiv.appendChild(azimuthKnobDiv);
			const distanceString = ui.getString('distance');
			const distanceLabel = distanceString + ' ' + iString;

			/*
			 * Parameters for the distance knob.
			 */
			const distanceParams = {
				'label': distanceLabel,
				'physicalUnit': 'dm',
				'valueMin': 0,
				'valueMax': 100,
				'valueDefault': distance,
				'valueWidth': 150,
				'valueHeight': 150,
				'angle': 270,
				'cursor': false,
				'colorScheme': 'blue',
				'readonly': false
			};

			const distanceKnob = ui.createKnob(distanceParams);
			const distanceKnobDiv = distanceKnob.div;
			controlsDiv.append(distanceKnobDiv);
			const levelString = ui.getString('level');
			const levelLabel = levelString + ' ' + iString;

			/*
			 * Parameters for the level knob.
			 */
			const levelParams = {
				'label': levelLabel,
				'physicalUnit': '%',
				'valueMin': 0,
				'valueMax': 100,
				'valueDefault': level,
				'valueWidth': 150,
				'valueHeight': 150,
				'angle': 270,
				'cursor': false,
				'colorScheme': 'blue',
				'readonly': false
			};

			const levelKnob = ui.createKnob(levelParams);
			const levelKnobDiv = levelKnob.div;
			controlsDiv.append(levelKnobDiv);
			const azimuthKnobNode = azimuthKnob.node;
			const distanceKnobNode = distanceKnob.node;
			const levelKnobNode = levelKnob.node;
			storage.put(azimuthKnobNode, 'channel', i);
			storage.put(distanceKnobNode, 'channel', i);
			storage.put(levelKnobNode, 'channel', i);

			/*
			 * This gets executed when the azimuth value changes.
			 */
			const azimuthHandler = function(knob, value) {
				const node = knob.node();
				const channel = storage.get(node, 'channel');
				handler.setAzimuth(channel, value);
			};

			/*
			 * This gets executed when the distance value changes.
			 */
			const distanceHandler = function(knob, value) {
				const node = knob.node();
				const channel = storage.get(node, 'channel');
				const distanceValue = (0.1 * value).toFixed(1);
				handler.setDistance(channel, distanceValue);
			};

			/*
			 * This gets executed when the level value changes.
			 */
			const levelHandler = function(knob, value) {
				const node = knob.node();
				const channel = storage.get(node, 'channel');
				const levelValue = (0.01 * value).toFixed(2);
				handler.setLevel(channel, levelValue);
			};

			const azimuthKnobObj = azimuthKnob.obj;
			azimuthKnobObj.addListener(azimuthHandler);
			const distanceKnobObj = distanceKnob.obj;
			distanceKnobObj.addListener(distanceHandler);
			const levelKnobObj = levelKnob.obj;
			levelKnobObj.addListener(levelHandler);
		}

		/*
		 * Create unit object.
		 */
		const unit = {
			'controls': controlsDiv,
			'expanded': false
		};

		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			const controlsDiv = this.controls;
			let displayValue = '';

			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value) {
				displayValue = 'block';
			} else {
				displayValue = 'none';
			}

			controlsDiv.style.display = displayValue;
			this.expanded = value;
		};

		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		};

		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			const state = this.getExpanded();
			this.setExpanded(!state);
		};

		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(e) {
			const unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		};

		storage.put(labelDiv, 'unit', unit);
	};

	/*
	 * Renders the metronome given a configuration returned from the server.
	 */
	this.renderMetronome = function(configuration) {
		const metronomeConfiguration = configuration.Metronome;
		const masterOutput = metronomeConfiguration.MasterOutput;
		const elem = document.getElementById('metronome');
		helper.clearElement(elem);
		const unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		unitDiv.classList.add('masterunitdiv');
		const headerDiv = document.createElement('div');
		const masterString = ui.getString('master');

		/*
		 * Parameters for metronome button.
		 */
		const paramsButton = {
			caption: masterString,
			active: masterOutput
		};

		const button = ui.createButton(paramsButton);
		const buttonElem = button.input;
		storage.put(buttonElem, 'active', masterOutput);

		/*
		 * This is called when the user clicks on the 'master' button of the metronome.
		 */
		buttonElem.onclick = function(e) {
			const active = !storage.get(this, 'active');

			/*
			 * Check whether the control should be active.
			 */
			if (active) {
				this.classList.remove('buttonnormal');
				this.classList.add('buttonactive');
			} else {
				this.classList.remove('buttonactive');
				this.classList.add('buttonnormal');
			}

			storage.put(this, 'active', active);
			handler.setMetronomeValue('master-output', active);
		};

		headerDiv.appendChild(buttonElem);
		const labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		labelDiv.classList.add('io');
		const label = ui.getString('metronome');
		const labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		headerDiv.appendChild(labelDiv);
		headerDiv.classList.add('headerdiv');
		unitDiv.appendChild(headerDiv);
		const controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		elem.appendChild(unitDiv);
		const beatsString = ui.getString('beats_per_period');
		const beatsValue = metronomeConfiguration.BeatsPerPeriod;

		/*
		 * Parameters for the beats per period knob.
		 */
		var beatsParams = {
			'label': beatsString,
			'physicalUnit': '',
			'valueMin': 1,
			'valueMax': 16,
			'valueDefault': beatsValue,
			'valueWidth': 150,
			'valueHeight': 150,
			'angle': 270,
			'cursor': false,
			'colorScheme': 'blue',
			'readonly': false
		};

		const beatsKnob = ui.createKnob(beatsParams);
		const beatsKnobDiv = beatsKnob.div;
		controlsDiv.appendChild(beatsKnobDiv);
		const speedString = ui.getString('speed');
		const bpmString = ui.getString('bpm');
		const speedValue = metronomeConfiguration.Speed;

		/*
		 * Parameters for the speed knob.
		 */
		var speedParams = {
			'label': speedString,
			'physicalUnit': bpmString,
			'valueMin': 40,
			'valueMax': 360,
			'valueDefault': speedValue,
			'valueWidth': 150,
			'valueHeight': 150,
			'angle': 270,
			'cursor': false,
			'colorScheme': 'blue',
			'readonly': false
		};

		const speedKnob = ui.createKnob(speedParams);
		const speedKnobDiv = speedKnob.div;
		controlsDiv.appendChild(speedKnobDiv);

		/*
		 * This gets executed when the beats per period value changes.
		 */
		const beatsHandler = function(knob, value) {
			handler.setMetronomeValue('beats-per-period', value);
		};

		/*
		 * This gets executed when the beats per period value changes.
		 */
		const speedHandler = function(knob, value) {
			handler.setMetronomeValue('speed', value);
		};

		const beatsKnobObj = beatsKnob.obj;
		beatsKnobObj.addListener(beatsHandler);
		const speedKnobObj = speedKnob.obj;
		speedKnobObj.addListener(speedHandler);
		const sounds = metronomeConfiguration.Sounds;
		const numSounds = sounds.length;
		const tickSound = metronomeConfiguration.TickSound;
		const tockSound = metronomeConfiguration.TockSound;
		let tickIdx = 0;
		let tockIdx = 0;

		/*
		 * Iterate over all sounds and find the tick and tock sound.
		 */
		for (let i = 0; i < numSounds; i++) {
			const sound = sounds[i];

			/*
			 * If we found the tick sound, store index.
			 */
			if (sound === tickSound) {
				tickIdx = i;
			}

			/*
			 * If we found the tock sound, store index.
			 */
			if (sound === tockSound) {
				tockIdx = i;
			}

		}

		const labelTick = ui.getString('tick_sound');
		const labelTock = ui.getString('tock_sound');

		/*
		 * Parameters for the tick sound drop down menu.
		 */
		const paramsTick = {
			'label': labelTick,
			'options': sounds,
			'selectedIndex': tickIdx
		};

		/*
		 * Parameters for the tock sound drop down menu.
		 */
		const paramsTock = {
			'label': labelTock,
			'options': sounds,
			'selectedIndex': tockIdx
		};

		const dropDownTick = ui.createDropDown(paramsTick);
		const dropDownTock = ui.createDropDown(paramsTock);
		const dropDownTickElem = dropDownTick.input;
		const dropDownTockElem = dropDownTock.input;

		/*
		 * This is called when the tick sound changes.
		 */
		dropDownTickElem.onchange = function(e) {
			const idx = this.selectedIndex;
			const option = this.options[idx];
			const value = option.text;
			handler.setMetronomeValue('tick-sound', value);
		};

		/*
		 * This is called when the tock sound changes.
		 */
		dropDownTockElem.onchange = function(e) {
			const idx = this.selectedIndex;
			const option = this.options[idx];
			const value = option.text;
			handler.setMetronomeValue('tock-sound', value);
		};

		const controlRowTick = document.createElement('div');
		controlRowTick.appendChild(dropDownTick.div);
		controlsDiv.appendChild(controlRowTick);
		const controlRowTock = document.createElement('div');
		controlRowTock.appendChild(dropDownTock.div);
		controlsDiv.appendChild(controlRowTock);

		/*
		 * Create unit object.
		 */
		const unit = {
			'controls': controlsDiv,
			'expanded': false
		};

		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			const controlsDiv = this.controls;
			let displayValue = '';

			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value) {
				displayValue = 'block';
			} else {
				displayValue = 'none';
			}

			controlsDiv.style.display = displayValue;
			this.expanded = value;
		};

		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		};

		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			const state = this.getExpanded();
			this.setExpanded(!state);
		};

		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(e) {
			const unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		};

		storage.put(labelDiv, 'unit', unit);
	};

	/*
	 * Renders the signal level analysis section given a configuration returned from the server.
	 */
	this.renderSignalLevels = function(configuration) {
		const batchProcessing = configuration.BatchProcessing;
		const elem = document.getElementById('levels');
		helper.clearElement(elem);

		/*
		 * Only display levels if batch processing is disabled on the server.
		 */
		if (batchProcessing === false) {
			const unitDiv = document.createElement('div');
			unitDiv.classList.add('contentdiv');
			unitDiv.classList.add('masterunitdiv');
			const headerDiv = document.createElement('div');
			const enabledString = ui.getString('enabled');
			const levelMeter = configuration.LevelMeter;
			const enabled = levelMeter.Enabled;

			/*
			 * Parameters for metronome button.
			 */
			var paramsButton = {
				caption: enabledString,
				active: enabled
			};

			const button = ui.createButton(paramsButton);
			const buttonElem = button.input;
			storage.put(buttonElem, 'active', enabled);

			/*
			 * This is called when the user clicks on the 'enabled' button of the level meter.
			 */
			buttonElem.onclick = function(e) {
				const active = !storage.get(this, 'active');

				/*
				 * Check whether the control should be active.
				 */
				if (active) {
					this.classList.remove('buttonnormal');
					this.classList.add('buttonactive');
				} else {
					this.classList.remove('buttonactive');
					this.classList.add('buttonnormal');
				}

				storage.put(this, 'active', active);
				handler.setLevelMeterEnabled(active);
			};

			headerDiv.appendChild(buttonElem);
			const labelDiv = document.createElement('div');
			labelDiv.classList.add('labeldiv');
			labelDiv.classList.add('active');
			labelDiv.classList.add('io');
			const label = ui.getString('signal_levels');
			const labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			headerDiv.appendChild(labelDiv);
			headerDiv.classList.add('headerdiv');
			unitDiv.appendChild(headerDiv);
			const controlsDiv = document.createElement('div');
			controlsDiv.classList.add('controlsdiv');
			unitDiv.appendChild(controlsDiv);
			elem.appendChild(unitDiv);

			/*
			 * This is called when the unit is collapsed or expanded.
			 */
			var expansionListener = function(unit, value) {

				/*
				 * Check if unit shall expand or contract.
				 *
				 * If unit gets expanded, register timer, which requests
				 * 'get-level-analysis' CGI regularly (every 250 ms).
				 *
				 * If unit gets contracted, unregister timer.
				 */
				if (value) {

					/*
					 * This is executed on each timer tick.
					 *
					 * Query 'get-level-analysis' CGI.
					 *
					 * On response, check if all channel names are already known and
					 * in the same order as in the response.
					 *
					 * If not, clear controls DIV, generate new channel controls and
					 * store new channel names.
					 *
					 * Update controls to represent new level and peak values obtained
					 * from the response.
					 */
					const callback = function() {

						/*
						 * This is called when the server returns a response.
						 */
						const responseListener = function(response) {
							let dspLoadControl = unit.dspLoadControl;
							let channelNames = unit.channelNames;
							const numNames = channelNames.length;
							let channelControls = unit.channelControls;
							let mismatch = (dspLoadControl === null);
							const channels = response.Channels;
							const numChannels = channels.length;

							/*
							 * Iterate over all channels in the response.
							 */
							for (let i = 0; i < numChannels; i++) {
								const channel = channels[i];
								const channelNameResponse = channel.ChannelName;

								/*
								 * If one of the channels does not match,
								 * report mismatch.
								 */
								if (numNames <= i) {
									mismatch = true;
								} else {
									const channelNameControl = channelNames[i];

									/*
									 * Check if name of the response matches name
									 * of the control.
									 */
									if (channelNameResponse !== channelNameControl) {
										mismatch = true;
									}

								}

							}

							/*
							 * If the channel mapping has changed, create new controls.
							 */
							if (mismatch) {
								const controlsDiv = unit.controls;
								helper.clearElement(controlsDiv);
								const dspLoadString = ui.getString('dsp_load');
								const dspLoadLabelDiv = document.createElement('div');
								const dspLoadLabelNode = document.createTextNode(dspLoadString);
								dspLoadLabelDiv.appendChild(dspLoadLabelNode);
								dspLoadControl = pureknob.createBarGraph(400, 40);
								dspLoadControl.setProperty('colorFG', '#ff4444');
								dspLoadControl.setProperty('colorMarkers', '#ffffff');
								dspLoadControl.setProperty('markerStart', 0);
								dspLoadControl.setProperty('markerEnd', 100);
								dspLoadControl.setProperty('markerStep', 25);
								dspLoadControl.setProperty('valMin', 0);
								dspLoadControl.setProperty('valMax', 100);
								dspLoadControl.setValue(0);
								const node = dspLoadControl.node();
								const nodeWrapper = document.createElement('div');
								nodeWrapper.appendChild(node);
								const container = document.createElement('div');
								container.appendChild(dspLoadLabelDiv);
								container.appendChild(nodeWrapper);
								controlsDiv.appendChild(container);
								channelNames = [];
								channelControls = [];

								/*
								 * Iterate over all channels in the response.
								 */
								for (let i = 0; i < numChannels; i++) {
									const channel = channels[i];
									const channelName = channel.ChannelName;
									const channelControl = pureknob.createBarGraph(400, 40);
									channelControl.setProperty('colorFG', '#44ff44');
									channelControl.setProperty('colorMarkers', '#ffffff');
									channelControl.setProperty('markerStart', -60);
									channelControl.setProperty('markerEnd', 0);
									channelControl.setProperty('markerStep', 10);
									channelControl.setProperty('valMin', -145);
									channelControl.setProperty('valMax', 0);
									channelControl.setValue(-145);
									channelNames.push(channelName);
									channelControls.push(channelControl);
									const channelNameDiv = document.createElement('div');
									const channelNameNode = document.createTextNode(channelName);
									channelNameDiv.appendChild(channelNameNode);
									const node = channelControl.node();
									const nodeWrapper = document.createElement('div');
									nodeWrapper.appendChild(node);
									const container = document.createElement('div');
									container.appendChild(channelNameDiv);
									container.appendChild(nodeWrapper);
									controlsDiv.appendChild(container);
								}

							}

							/*
							 * Display DSP load.
							 */
							if (dspLoadControl !== null) {
								const dspLoad = response.DSPLoad;
								dspLoadControl.setValue(dspLoad);
							}

							/*
							 * Iterate over all channels in the response.
							 */
							for (let i = 0; i < numChannels; i++) {
								const channel = channels[i];
								const channelLevel = channel.Level;
								const channelPeak = channel.Peak;
								const channelControl = channelControls[i];
								channelControl.setValue(channelLevel);
								channelControl.setPeaks([channelPeak]);
							}

							unit.dspLoadControl = dspLoadControl;
							unit.channelNames = channelNames;
							unit.channelControls = channelControls;
						};

						handler.getLevelAnalysis(responseListener);
					};

					const timer = window.setInterval(callback, 200);
					unit.timer = timer;
				} else {
					const timer = unit.timer;

					/*
					 * If a timer is registered, clear it.
					 */
					if (timer !== null) {
						window.clearInterval(timer);
					}

					unit.timer = null;
				}

			};

			/*
			 * Create unit object.
			 */
			const unit = {
				'channelNames': [],
				'channelControls': [],
				'controls': controlsDiv,
				'expanded': false,
				'listeners': [expansionListener],
				'timer': null
			};

			/*
			 * Expands or collapses a unit.
			 */
			unit.setExpanded = function(value) {
				const controlsDiv = this.controls;
				let displayValue = '';

				/*
				 * Check whether we should expand or collapse the unit.
				 */
				if (value) {
					displayValue = 'block';
				} else {
					displayValue = 'none';
				}

				controlsDiv.style.display = displayValue;
				this.expanded = value;
				const listeners = this.listeners;

				/*
				 * Check if there are listeners resistered.
				 */
				if (listeners !== null) {
					var numListeners = listeners.length;

					/*
					 * Invoke each listener.
					 */
					for (let i = 0; i < numListeners; i++) {
						const listener = listeners[i];
						listener(this, value);
					}

				}

			};

			/*
			 * Returns whether a unit is expanded.
			 */
			unit.getExpanded = function() {
				return this.expanded;
			};

			/*
			 * Toggles a unit between expanded and collapsed state.
			 */
			unit.toggleExpanded = function() {
				const state = this.getExpanded();
				this.setExpanded(!state);
			};

			/*
			 * This is called when a user clicks on the label div.
			 */
			labelDiv.onclick = function(e) {
				const unit = storage.get(this, 'unit');
				unit.toggleExpanded();
			};

			storage.put(labelDiv, 'unit', unit);
		}

	};

	/*
	 * Renders the 'processing' button given a configuration returned from the server.
	 */
	this.renderProcessing = function(configuration) {
		const batchProcessing = configuration.BatchProcessing;
		const elem = document.getElementById('processing');
		helper.clearElement(elem);

		/*
		 * Only display processing button if batch processing is enabled on the server.
		 */
		if (batchProcessing === true) {
			const unitDiv = document.createElement('div');
			unitDiv.classList.add('contentdiv');
			unitDiv.classList.add('masterunitdiv');
			const headerDiv = document.createElement('div');
			const processString = ui.getString('process_now');

			/*
			 * Parameters for process button.
			 */
			var paramsButton = {
				caption: processString,
				active: false
			};

			const button = ui.createButton(paramsButton);
			const buttonElem = button.input;
			storage.put(buttonElem, 'active', batchProcessing);

			/*
			 * This is called when the user clicks on the 'process' button.
			 */
			buttonElem.onclick = function(e) {
				const active = storage.get(this, 'active');

				/*
				 * Trigger batch processing if the control is active.
				 */
				if (active) {
					handler.process();
				}

			};

			headerDiv.appendChild(buttonElem);
			const labelDiv = document.createElement('div');
			labelDiv.classList.add('labeldiv');
			labelDiv.classList.add('active');
			labelDiv.classList.add('io');
			const label = ui.getString('batch_processing');
			const labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			headerDiv.appendChild(labelDiv);
			headerDiv.classList.add('headerdiv');
			unitDiv.appendChild(headerDiv);
			const controlsDiv = document.createElement('div');
			controlsDiv.classList.add('controlsdiv');
			unitDiv.appendChild(controlsDiv);
			elem.appendChild(unitDiv);

			/*
			 * Create unit object.
			 */
			const unit = {
				'controls': controlsDiv,
				'expanded': false
			};

			/*
			 * Expands or collapses a unit.
			 */
			unit.setExpanded = function(value) {
				const controlsDiv = this.controls;
				let displayValue = '';

				/*
				 * Check whether we should expand or collapse the unit.
				 */
				if (value) {
					displayValue = 'block';
				} else {
					displayValue = 'none';
				}

				controlsDiv.style.display = displayValue;
				this.expanded = value;
			};

			/*
			 * Returns whether a unit is expanded.
			 */
			unit.getExpanded = function() {
				return this.expanded;
			};

			/*
			 * Toggles a unit between expanded and collapsed state.
			 */
			unit.toggleExpanded = function() {
				const state = this.getExpanded();
				this.setExpanded(!state);
			};

			/*
			 * This is called when a user clicks on the label div.
			 */
			labelDiv.onclick = function(e) {
				const unit = storage.get(this, 'unit');
				unit.toggleExpanded();
			};

			storage.put(labelDiv, 'unit', unit);
		}

	};

}

const ui = new UI();

/*
 * This class implements all handler functions for user interaction.
 */
function Handler() {
	const self = this;

	/*
	 * This is called when a new effects unit should be added.
	 */
	this.addUnit = function(unitType, chain) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt, otherwise refresh rack.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Adding new unit failed: ' + reason;
					console.log(msg);
				} else {
					self.refresh();
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const unitTypeString = unitType.toString();
		const chainString = chain.toString();
		const request = new Request();
		request.append('cgi', 'add-unit');
		request.append('type', unitTypeString);
		request.append('chain', chainString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a new level analysis should be obtained.
	 */
	this.getLevelAnalysis = function(callback) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const levels = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (levels !== null) {
				callback(levels);
			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const request = new Request();
		request.append('cgi', 'get-level-analysis');
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, false);
	};

	/*
	 * This is called when a unit should be moved down the chain.
	 */
	this.moveDown = function(chain, unit) {

		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Moving unit down failed: ' + reason;
					console.log(msg);
				} else {
					self.refresh();
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const unitString = unit.toString();
		const request = new Request();
		request.append('cgi', 'move-down');
		request.append('chain', chainString);
		request.append('unit', unitString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a unit should be moved up the chain.
	 */
	this.moveUp = function(chain, unit) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Moving unit up failed: ' + reason;
					console.log(msg);
				} else {
					self.refresh();
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const unitString = unit.toString();
		const request = new Request();
		request.append('cgi', 'move-up');
		request.append('chain', chainString);
		request.append('unit', unitString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a unit should be removed from a chain.
	 */
	this.removeUnit = function(chain, unit) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Removing unit failed: ' + reason;
					console.log(msg);
				} else {
					self.refresh();
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const unitString = unit.toString();
		const request = new Request();
		request.append('cgi', 'remove-unit');
		request.append('chain', chainString);
		request.append('unit', unitString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a new azimuth value should be set.
	 */
	this.setAzimuth = function(chain, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting azimuth value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const valueString = value.toString()
		const request = new Request();
		request.append('cgi', 'set-azimuth');
		request.append('chain', chainString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a unit should be bypassed or bypass should be disabled for a unit.
	 */
	this.setBypass = function(chain, unit, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting bypass value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const unitString = unit.toString();
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-bypass');
		request.append('chain', chainString);
		request.append('unit', unitString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a new distance value should be set.
	 */
	this.setDistance = function(chain, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting distance value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-distance');
		request.append('chain', chainString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a discrete value should be set.
	 */
	this.setDiscreteValue = function(chain, unit, param, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting discrete value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const unitString = unit.toString();
		const paramString = param.toString();
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-discrete-value');
		request.append('chain', chainString);
		request.append('unit', unitString);
		request.append('param', paramString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when period size should be changed.
	 */
	this.setFramesPerPeriod = function(value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting frames per period failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-frames-per-period');
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a new level value should be set.
	 */
	this.setLevel = function(chain, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting level value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-level');
		request.append('chain', chainString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when the level meter should be enabled or disabled.
	 */
	this.setLevelMeterEnabled = function(value) {

		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const action = 'Enabling';

					/*
					 * Check if we should disable the level meter.
					 */
					if (!value) {
						action = 'Disabling';
					}

					const msg = action + ' level meter failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-level-meter-enabled');
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a metronome value should be changed.
	 */
	this.setMetronomeValue = function(param, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting metronome value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const paramString = param.toString();
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-metronome-value');
		request.append('param', paramString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a tuner value should be changed.
	 */
	this.setTunerValue = function(param, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting tuner value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const paramString = param.toString();
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-tuner-value');
		request.append('param', paramString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a numeric value should be set.
	 */
	this.setNumericValue = function(chain, unit, param, value) {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const webResponse = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse !== null) {

				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success !== true) {
					const reason = webResponse.Reason;
					const msg = 'Setting numeric value failed: ' + reason;
					console.log(msg);
				}

			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const chainString = chain.toString();
		const unitString = unit.toString();
		const paramString = param.toString();
		const valueString = value.toString();
		const request = new Request();
		request.append('cgi', 'set-numeric-value');
		request.append('chain', chainString);
		request.append('unit', unitString);
		request.append('param', paramString);
		request.append('value', valueString);
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when the configuration needs to be refreshed.
	 */
	this.refresh = function() {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const configuration = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (configuration !== null) {
				ui.renderSignalChains(configuration);
				ui.renderPersistence(configuration);
				ui.renderLatency(configuration);
				ui.renderTuner(configuration);
				ui.renderSpatializer(configuration);
				ui.renderMetronome(configuration);
				ui.renderSignalLevels(configuration);
				ui.renderProcessing(configuration);
			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const request = new Request();
		request.append('cgi', 'get-configuration');
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

	/*
	 * This is called when a new analysis should be performed by the tuner.
	 */
	this.refreshTuner = function() {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const analysis = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (analysis !== null) {
				ui.updateTuner(analysis);
			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const request = new Request();
		request.append('cgi', 'get-tuner-analysis');
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, false);
	};

	/*
	 * This is called when the user clicks on the 'process' button.
	 */
	this.process = function() {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			helper.blockSite(true);
		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const request = new Request();
		request.append('cgi', 'process');
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, false);
	};

	/*
	 * This is used to prevent the default action from occuring.
	 */
	this.absorbEvent = function(e) {
		e.stopPropagation();
		e.preventDefault();
		return false;
	};

	/*
	 * This is used to prevent the default action from occuring.
	 */
	this.dragEnter = function(e) {
		e.stopPropagation();
		e.preventDefault();
		const elem = e.target;
		elem.classList.add('dragover');
		return false;
	};

	/*
	 * This is used to prevent the default action from occuring.
	 */
	this.dragLeave = function(e) {
		e.stopPropagation();
		e.preventDefault();
		const elem = e.target;
		elem.classList.remove('dragover');
		return false;
	};

	/*
	 * This is called when the user drops a patch file into the upload area.
	 */
	this.uploadFile = function(e) {
		e.stopPropagation();
		e.preventDefault();
		const transfer = e.dataTransfer;
		const files = transfer.files;
		const numFiles = files.length;

		/*
		 * Check if there is a file.
		 */
		if (numFiles > 0) {
			const file = files[0];

			/*
			 * This gets called when the server returns a response.
			 */
			var responseHandler = function(response) {
				handler.refresh();
			};

			const url = globals.cgi;
			const data = new FormData();
			data.append('cgi', 'persistence-restore');
			data.append('patchfile', file);
			ajax.request('POST', url, data, null, responseHandler, true);
		}

		return false;
	};

	/*
	 * This is called when the user interface initializes.
	 */
	this.initialize = function() {

		/*
		 * This gets called when the server returns a response.
		 */
		const responseHandler = function(response) {
			const unitTypes = helper.parseJSON(response);

			/*
			 * Check if the response is valid JSON.
			 */
			if (unitTypes !== null) {
				const numUnitTypes = unitTypes.length;

				/*
				 * Iterate over the unit types and add them to the global list.
				 */
				for (let i = 0; i < numUnitTypes; i++) {
					const t = unitTypes[i];
					globals.unitTypes.push(t);
				}

				self.refresh();
			}

		};

		const url = globals.cgi;
		const mimeType = globals.mimeDefault;
		const request = new Request();
		request.append('cgi', 'get-unit-types');
		const requestBody = request.getData();
		ajax.request('POST', url, requestBody, mimeType, responseHandler, true);
	};

}

/*
 * The (global) event handlers.
 */
const handler = new Handler();
document.addEventListener('DOMContentLoaded', handler.initialize);

