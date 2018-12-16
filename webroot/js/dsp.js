"use strict";

/*
 * A class for storing global state required by the application.
 */
function Globals() {
	this.cgi = '/cgi-bin/dsp';
	this.unitTypes = new Array();
}

/*
 * The global state object.
 */
var globals = new Globals();

/*
 * A class implementing data storage.
 */
function Storage() {
	var g_map = new WeakMap();
	
	/*
	 * Store a value under a key inside an element.
	 */
	this.put = function(elem, key, value) {
		var map = g_map.get(elem);
		
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
		var map = g_map.get(elem);
		
		/*
		 * Check if element is unknown.
		 */
		if (map == null) {
			return null;
		} else {
			var value = map.get(key);
			return value;
		}
		
	};
	
	/*
	 * Check if a certain key exists inside an element.
	 */
	this.has = function(elem, key) {
		var map = g_map.get(elem);
		
		/*
		 * Check if element is unknown.
		 */
		if (map == null) {
			return false;
		} else {
			var value = map.has(key);
			return value;
		}
		
	};
	
	/*
	 * Remove a certain key from an element.
	 */
	this.remove = function(elem, key) {
		var map = g_map.get(elem);
		
		/*
		 * Check if element is known.
		 */
		if (map != null) {
			map.delete(key);
			
			/*
			 * If inner map is now empty, remove it from outer map.
			 */
			if (map.size == 0) {
				g_map.delete(elem);
			}
			
		}
		
	};
	
}

var storage = new Storage();

/*
 * A class supposed to make life a little easier.
 */
function Helper() {
	
	/*
	 * Blocks or unblocks the site for user interactions.
	 */
	this.blockSite = function(blocked) {
		var blocker = document.getElementById('blocker');
		var displayStyle = '';
		
		/*
		 * If we should block the site, display blocker, otherwise hide it.
		 */
		if (blocked)
			displayStyle = 'block';
		else
			displayStyle = 'none';
		
		/*
		 * Apply style if the site has a blocker.
		 */
		if (blocker != null)
			blocker.style.display = displayStyle;
		
	};
	
	/*
	 * Removes all child nodes from an element.
	 */
	this.clearElement = function(elem) {
		
		/*
		 * As long as the element has child nodes, remove one.
		 */
		while (elem.hasChildNodes()) {
			var child = elem.firstChild;
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
			var obj = JSON.parse(jsonString);
			return obj;
		} catch (ex) {
			return null;
		}
		
	};
	
}

/*
 * The (global) helper object.
 */
var helper = new Helper();

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
	 * - callback (function): The function to be called when a response is
	 *	returned from the server.
	 * - block (boolean): Whether the site should be blocked.
	 *
	 * Returns: Nothing.
	 */
	this.request = function(method, url, data, callback, block) {
		var xhr = new XMLHttpRequest();
		
		/*
		 * Event handler for ReadyStateChange event.
		 */
		xhr.onreadystatechange = function() {
			helper.blockSite(block);
			
			/*
			 * If we got a response, pass the response text to
			 * the callback function.
			 */
			if (this.readyState == 4) {
				
				/*
				 * If we blocked the site on the request,
				 * unblock it on the response.
				 */
				if (block)
					helper.blockSite(false);
				
				/*
				 * Check if callback is registered.
				 */
				if (callback != null) {
					var content = xhr.responseText;
					callback(content);
				}
				
			}
			
		};
		
		xhr.open(method, url, true);
		xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
		xhr.send(data);
	};
	
}

/*
 * The (global) Ajax engine.
 */
var ajax = new Ajax();

/*
 * A class implementing a key-value-pair.
 */
function KeyValuePair(key, value) {
	var g_key = key;
	var g_value = value;
	
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
	var g_keyValues = Array();
	
	/*
	 * Append a key-value-pair to a request.
	 */
	this.append = function(key, value) {
		var kv = new KeyValuePair(key, value);
		g_keyValues.push(kv);
	}
	
	/*
	 * Returns the URL encoded data for this request.
	 */
	this.getData = function() {
		var numPairs = g_keyValues.length;
		var s = '';
		
		/*
		 * Iterate over the key-value pairs.
		 */
		for (var i = 0; i < numPairs; i++) {
			var keyValue = g_keyValues[i];
			var key = keyValue.getKey();
			var keyEncoded = encodeURIComponent(key);
			var value = keyValue.getValue();
			var valueEncoded = encodeURIComponent(value);
			
			/*
			 * If this is not the first key-value pair, we need a separator.
			 */
			if (i > 0)
				s += '&';
			
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
	var strings = {
		'add': 'Add',
		'add_unit': 'Add unit',
		'auto_wah': 'Auto wah',
		'azimuth': 'Azimuth',
		'bandpass': 'Bandpass',
		'batch_processing': 'Batch processing',
		'beats_per_period': 'Beats per period',
		'bias': 'Bias',
		'boost': 'Boost',
		'bypass': 'Bypass',
		'cents': 'Cents',
		'channel': 'Channel',
		'chorus': 'Chorus',
		'delay': 'Delay',
		'delay_time': 'Delay time',
		'depth': 'Depth',
		'distance': 'Distance',
		'distortion': 'Distortion',
		'drive': 'Drive',
		'dsp_load': 'DSP load',
		'excess': 'Excess',
		'feedback': 'Feedback',
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
		'move_down': 'Move down',
		'move_up': 'Move up',
		'noise_gate': 'Noise gate',
		'note': 'Note',
		'octaver': 'Octaver',
		'overdrive': 'Overdrive',
		'phase': 'Phase',
		'phaser': 'Phaser',
		'polarity': 'Polarity',
		'power_amp': 'Power amp',
		'presence': 'Presence',
		'process_now': 'Process now',
		'remove': 'Remove',
		'ring_modulator': 'Ring modulator',
		'signal_amplitude': 'Signal amplitude',
		'signal_frequency': 'Signal frequency',
		'signal_gain': 'Signal gain',
		'signal_generator': 'Signal generator',
		'signal_levels': 'Signal levels',
		'signal_type': 'Signal type',
		'spatializer': 'Spatializer',
		'speed': 'Speed',
		'threshold_close': 'Threshold close',
		'threshold_open': 'Threshold open',
		'tick_sound': 'Tick sound',
		'tock_sound': 'Tock sound',
		'to_output': 'To: Output',
		'tone_stack': 'Tone stack',
		'tremolo': 'Tremolo',
		'tuner': 'Tuner'
	};
	
	/*
	 * Obtains a string for the user interface.
	 */
	this.getString = function(key) {
		
		/*
		 * Check whether key is defined in strings.
		 */
		if (key in strings) {
			var s = strings[key];
			return s;
		} else
			return key;
		
	}

	/*
	 * Creates a turnable knob.
	 */
	this.createKnob = function(params) {
		var label = params.label;
		var valueMin = params.valueMin;
		var valueMax = params.valueMax;
		var valueDefault = params.valueDefault;
		var valueWidth = params.valueWidth;
		var valueHeight = params.valueHeight;
		var valueAngle = params.angle;
		var valueAngleArc = (valueAngle / 180.0) * Math.PI;
		var valueCursor = params.cursor;
		var valueReadonly = params.readonly;
		var colorScheme = params.colorScheme;
		var angleArc = (valueAngle / 180.0) * Math.PI;
		var halfAngleArc = 0.5 * valueAngleArc;
		var paramDiv = document.createElement('div');
		paramDiv.classList.add('paramdiv');
		var labelDiv = document.createElement('div');
		labelDiv.classList.add('knoblabel');
		var labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		paramDiv.appendChild(labelDiv);
		var knobDiv = document.createElement('div');
		knobDiv.classList.add('knobdiv');
		var fgColor = '#ff8800';
		var bgColor = '#181818';
		
		/*
		 * Check if we want a blue or green color scheme.
		 */
		if (colorScheme == 'blue') {
			fgColor = '#8888ff';
			bgColor = '#181830';
		} else if (colorScheme == 'green') {
			fgColor = '#88ff88';
			bgColor = '#181830';
		}
		
		var knobElem = pureknob.createKnob(valueHeight, valueWidth);
		knobElem.setProperty('angleStart', -halfAngleArc);
		knobElem.setProperty('angleEnd', halfAngleArc);
		knobElem.setProperty('colorBG', bgColor);
		knobElem.setProperty('colorFG', fgColor);
		knobElem.setProperty('needle', valueCursor);
		knobElem.setProperty('readonly', valueReadonly);
		knobElem.setProperty('valMin', valueMin);
		knobElem.setProperty('valMax', valueMax);
		knobElem.setValue(valueDefault);
		var knobNode = knobElem.node();
		knobDiv.appendChild(knobNode);
		paramDiv.appendChild(knobDiv);
		
		/*
		 * Create knob.
		 */
		var knob = {
			'div': paramDiv,
			'node': knobNode,
			'obj': knobElem
		};
		
		return knob;
	}
	
	/*
	 * Creates a drop down menu.
	 */
	this.createDropDown = function(params) {
		var label = params.label;
		var options = params.options;
		var numOptions = options.length;
		var selectedIndex = params.selectedIndex;
		var paramDiv = document.createElement('div');
		paramDiv.classList.add('paramdiv');
		
		/*
		 * Check if we should apply a label.
		 */
		if (label != null) {
			var labelDiv = document.createElement('div');
			labelDiv.classList.add('dropdownlabel');
			var labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			paramDiv.appendChild(labelDiv);
		}
		
		var selectElem = document.createElement('select');
		selectElem.classList.add('dropdown');
		
		/*
		 * Add all options.
		 */
		for (var i = 0; i < numOptions; i++) {
			var optionElem = document.createElement('option');
			optionElem.text = options[i];
			selectElem.add(optionElem);
		}
		
		selectElem.selectedIndex = selectedIndex;
		paramDiv.appendChild(selectElem);
		
		/*
		 * Create dropdown.
		 */
		var dropdown = {
			'div': paramDiv,
			'input': selectElem
		};
		
		return dropdown;
	}
	
	/*
	 * Creates a button.
	 */
	this.createButton = function(params) {
		var caption = params.caption;
		var active = params.active;
		var elem = document.createElement('button');
		elem.classList.add('button');
		
		/*
		 * Check whether the button should be active
		 */
		if (active)
			elem.classList.add('buttonactive');
		else
			elem.classList.add('buttonnormal');
		
		var captionNode = document.createTextNode(caption);
		elem.appendChild(captionNode);
		
		/*
		 * Create button.
		 */
		var button = {
			'input': elem
		};
		
		return button;
	}
	
	/*
	 * Creates a unit.
	 */
	this.createUnit = function(params) {
		var typeString = params.type;
		var buttonsParam = params.buttons;
		var numButtonsParam = buttonsParam.length;
		var buttons = [];
		
		/*
		 * Iterate over the buttons;
		 */
		for (var i = 0; i < numButtonsParam; i++) {
			var buttonParam = buttonsParam[i];
			var label = buttonParam.label;
			var active = buttonParam.active;
			
			/*
			 * Parameters for the button.
			 */
			var params = {
				'caption': label,
				'active': active
			};
			
			var button = this.createButton(params);
			buttons.push(button);
		}
		
		var unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		var headerDiv = document.createElement('div');
		headerDiv.classList.add('headerdiv');
		var numButtons = buttons.length;
		
		/*
		 * Add buttons to header.
		 */
		for (var i = 0; i < numButtons; i++) {
			var button = buttons[i];
			headerDiv.appendChild(button.input);
		}
		
		var labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		var typeNode = document.createTextNode(typeString);
		labelDiv.appendChild(typeNode);
		headerDiv.appendChild(labelDiv);
		unitDiv.appendChild(headerDiv);
		var controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		
		/*
		 * Create unit.
		 */
		var unit = {
			'div': unitDiv,
			'controls': controlsDiv,
			'buttons': buttons,
			'expanded': false
		};
		
		/*
		 * Adds a control to a unit.
		 */
		unit.addControl = function(control) {
			var controlDiv = control.div;
			this.controls.appendChild(controlDiv);
		}
		
		/*
		 * Adds a row with controls to a unit.
		 */
		unit.addControlRow = function(controls) {
			var rowDiv = document.createElement('div');
			var numControls = controls.length;
			
			/*
			 * Insert controls into the row.
			 */
			for (var i = 0; i < numControls; i++) {
				var control = controls[i];
				var controlDiv = control.div;
				rowDiv.appendChild(controlDiv);
			}
			
			this.controls.appendChild(rowDiv);
		}
		
		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			var controlsDiv = this.controls;
			var displayValue = '';
			
			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value)
				displayValue = 'block';
			else
				displayValue = 'none';
			
			controlsDiv.style.display = displayValue;
			this.expanded = value;
		}
		
		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		}
		
		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			var state = this.getExpanded();
			this.setExpanded(!state);
		}
		
		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(event) {
			var unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		}
		
		storage.put(labelDiv, 'unit', unit);
		return unit;
	}
	
	/*
	 * Checks whether a certain combination of unit type and parameter name requires special handling.
	 */
	this.isSpecialParameter = function(unitType, paramName) {
		
		/*
		 * Discern unit type.
		 */
		switch (unitType) {
			case 'power_amp':
				
				if (paramName.startsWith('level_'))
					return true;
				else if (paramName.startsWith('filter_')) {
					var suffix = paramName.substring(7);
					var isNumeric = isFinite(suffix);
					return isNumeric;
				}
				
			default:
				return false;
		}
		
	}
	
	/*
	 * Renders a unit given chain and unit ID, as well as a description returned from the server.
	 */
	this.renderUnit = function(chainId, unitId, description) {
		var bypassButtonLabel = ui.getString('bypass');
		var moveUpButtonLabel = ui.getString('move_up');
		var moveDownButtonLabel = ui.getString('move_down');
		var removeButtonLabel = ui.getString('remove');
		var unitTypes = globals.unitTypes;
		var unitTypeId = description.Type;
		var unitType = unitTypes[unitTypeId];
		var unitTypeString = ui.getString(unitType);
		var bypassActive = description.Bypass;
		
		/*
		 * Buttons for this unit.
		 */
		var buttons = [
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
		var paramsUnit = {
			'type': unitTypeString,
			'buttons': buttons
		};
		
		var unit = ui.createUnit(paramsUnit);
		var btnBypass = unit.buttons[0].input;
		storage.put(btnBypass, 'chain', chainId);
		storage.put(btnBypass, 'unit', unitId);
		storage.put(btnBypass, 'active', bypassActive);
		
		/*
		 * This is invoked when someone clicks on the 'bypass' button.
		 */
		btnBypass.onclick = function(event) {
			var chainId = storage.get(this, 'chain');
			var unitId = storage.get(this, 'unit');
			var active = !storage.get(this, 'active');
			
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
		
		var btnMoveUp = unit.buttons[1].input;
		storage.put(btnMoveUp, 'chain', chainId);
		storage.put(btnMoveUp, 'unit', unitId);
		
		/*
		 * This is invoked when someone clicks on the 'move up' button.
		 */
		btnMoveUp.onclick = function(event) {
			var chainId = storage.get(this, 'chain');
			var unitId = storage.get(this, 'unit');
			handler.moveUp(chainId, unitId);
		};
		
		var btnMoveDown = unit.buttons[2].input;
		storage.put(btnMoveDown, 'chain', chainId);
		storage.put(btnMoveDown, 'unit', unitId);
		
		/*
		 * This is invoked when someone clicks on the 'move down' button.
		 */
		btnMoveDown.onclick = function(event) {
			var chainId = storage.get(this, 'chain');
			var unitId = storage.get(this, 'unit');
			handler.moveDown(chainId, unitId);
		};
		
		var btnRemove = unit.buttons[3].input;
		btnRemove.classList.add('buttonremove');
		storage.put(btnRemove, 'chain', chainId);
		storage.put(btnRemove, 'unit', unitId);
		
		/*
		 * This is invoked when someone clicks on the 'remove' button.
		 */
		btnRemove.onclick = function(event) {
			var chainId = storage.get(this, 'chain');
			var unitId = storage.get(this, 'unit');
			handler.removeUnit(chainId, unitId);
		};
		
		var unitParams = description.Parameters;
		var numParams = unitParams.length;
		
		/*
		 * Iterate over the parameters and add all 'ordinary' (non-special) ones to the unit.
		 */
		for (var i = 0; i < numParams; i++) {
			var currentParam = unitParams[i];
			var paramType = currentParam.Type;
			var paramName = currentParam.Name;
			var isSpecial = this.isSpecialParameter(unitType, paramName);
			
			/*
			 * Only handle 'ordinary' (non-special) parameters on the first pass.
			 */
			if (!isSpecial) {
				var isFloating = (i != 0);
				var label = ui.getString(paramName);
				
				/*
				 * Handle numeric parameter.
				 */
				if (paramType == 'numeric') {
					
					/*
					 * Parameters for the knob.
					 */
					var params = {
						'label': label,
						'valueMin': currentParam.Minimum,
						'valueMax': currentParam.Maximum,
						'valueDefault': currentParam.NumericValue,
						'valueWidth': 150,
						'valueHeight': 150,
						'angle': 270,
						'cursor': false,
						'colorScheme': 'default',
						'readonly': false
					};
					
					var knob = ui.createKnob(params);
					unit.addControl(knob);
					var knobNode = knob.node;
					storage.put(knobNode, 'chain', chainId);
					storage.put(knobNode, 'unit', unitId);
					storage.put(knobNode, 'param', paramName);
					
					/*
					 * This is called when a numeric value changes.
					 */
					var knobHandler = function(knob, value) {
						var knobNode = knob.node();
						var chain = storage.get(knobNode, 'chain');
						var unit = storage.get(knobNode, 'unit');
						var param = storage.get(knobNode, 'param');
						handler.setNumericValue(chain, unit, param, value);
					};
					
					var knobObj = knob.obj;
					knobObj.addListener(knobHandler);
				}
				
				/*
				 * Handle discrete parameter.
				 */
				if (paramType == 'discrete') {
					
					/*
					 * Parameters for the drop down menu.
					 */
					var params = {
						'label': label,
						'options': currentParam.DiscreteValues,
						'selectedIndex': currentParam.DiscreteValueIndex
					};
					
					var dropDown = ui.createDropDown(params);
					var dropDownInput = dropDown.input;
					storage.put(dropDownInput, 'chain', chainId);
					storage.put(dropDownInput, 'unit', unitId);
					storage.put(dropDownInput, 'param', paramName);
					
					/*
					 * This is called when a discrete value changes.
					 */
					dropDownInput.onchange = function(event) {
						var chain = storage.get(this, 'chain');
						var unit = storage.get(this, 'unit');
						var param = storage.get(this, 'param');
						var idx = this.selectedIndex;
						var option = this.options[idx];
						var value = option.text;
						handler.setDiscreteValue(chain, unit, param, value);
					};
					
					unit.addControl(dropDown);
				}
				
			}
			
		}
		
		/*
		 * Iterate over the parameters and add all special discrete ones to the unit.
		 */
		for (var i = 0; i < numParams; i++) {
			var param = unitParams[i];
			var paramType = param.Type;
			var paramName = param.Name;
			var isSpecial = this.isSpecialParameter(unitType, paramName);
			
			/*
			 * Only handle special discrete parameters on the second pass.
			 */
			if (isSpecial & (paramType == 'discrete')) {
				var label = ui.getString(paramName);
				
				/*
				 * Parameters for the drop down menu.
				 */
				var params = {
					'label': label,
					'options': param.DiscreteValues,
					'selectedIndex': param.DiscreteValueIndex
				};
				
				var dropDown = ui.createDropDown(params);
				var dropDownInput = dropDown.input;
				storage.put(dropDownInput, 'chain', chainId);
				storage.put(dropDownInput, 'unit', unitId);
				storage.put(dropDownInput, 'param', paramName);
				
				/*
				 * This is called when a discrete value changes.
				 */
				dropDownInput.onchange = function(event) {
					var chain = storage.get(this, 'chain');
					var unit = storage.get(this, 'unit');
					var param = storage.get(this, 'param');
					var idx = this.selectedIndex;
					var option = this.options[idx];
					var value = option.text;
					handler.setDiscreteValue(chain, unit, param, value);
				};
				
				var controlRow = Array();
				controlRow.push(dropDown);
				unit.addControlRow(controlRow);
			}
			
		}

		/*
		 * Iterate over the parameters and add all special numeric ones to the unit.
		 */
		for (var i = 0; i < numParams; i++) {
			var param = unitParams[i];
			var paramType = param.Type;
			var paramName = param.Name;
			var isSpecial = this.isSpecialParameter(unitType, paramName);
			
			/*
			 * Only handle special numeric parameters on the third pass.
			 */
			if (isSpecial & (paramType == 'numeric')) {
				var label = ui.getString(paramName);
					
				/*
				 * Parameters for the knob.
				 */
				var params = {
					'label': label,
					'valueMin': param.Minimum,
					'valueMax': param.Maximum,
					'valueDefault': param.NumericValue,
					'valueWidth': 150,
					'valueHeight': 150,
					'angle': 270,
					'cursor': false,
					'colorScheme': 'default',
					'readonly': false
				};
				
				var knob = ui.createKnob(params);
				unit.addControl(knob);
				var knobNode = knob.node;
				storage.put(knobNode, 'chain', chainId);
				storage.put(knobNode, 'unit', unitId);
				storage.put(knobNode, 'param', paramName);
				
				/*
				 * This is called when a numeric value changes.
				 */
				var knobHandler = function(knob, value) {
					var knobNode = knob.node();
					var chain = storage.get(knobNode, 'chain');
					var unit = storage.get(knobNode, 'unit');
					var param = storage.get(knobNode, 'param');
					handler.setNumericValue(chain, unit, param, value);
				};
				
				var knobObj = knob.obj;
				knobObj.addListener(knobHandler);
			}
			
		}
		
		return unit;
	}
	
	/*
	 * Renders a signal chain, given its ID and a chain description returned from the server.
	 */
	this.renderSignalChain = function(id, description) {
		var idString = id.toString();
		var chainDiv = document.createElement('div');
		var beginDiv = document.createElement('div');
		beginDiv.classList.add('contentdiv');
		beginDiv.classList.add('iodiv');
		var beginHeaderDiv = document.createElement('div');
		beginHeaderDiv.classList.add('headerdiv');
		var beginLabelDiv = document.createElement('div');
		beginLabelDiv.classList.add('labeldiv');
		var labelFromInput = ui.getString('from_input');
		var beginLabelText = labelFromInput + ' ' + idString;
		var beginLabelNode = document.createTextNode(beginLabelText);
		beginLabelDiv.appendChild(beginLabelNode);
		beginHeaderDiv.appendChild(beginLabelDiv);
		beginDiv.appendChild(beginHeaderDiv);
		chainDiv.appendChild(beginDiv);
		var units = description.Units;
		var numUnits = units.length;
		
		/*
		 * Iterate over the units in this chain.
		 */
		for (var i = 0; i < numUnits; i++) {
			var unit = units[i];
			var result = this.renderUnit(id, i, unit);
			var unitDiv = result.div;
			chainDiv.appendChild(unitDiv);
		}
		
		var labelDropdown = ui.getString('add_unit');
		var labelButton = ui.getString('add');
		var unitTypes = globals.unitTypes;
		var numUnitTypes = unitTypes.length;
		var unitTypeNames = new Array();
		
		/*
		 * Look up the name of the unit types.
		 */
		for (var i = 0; i < numUnitTypes; i++) {
			var unitType = unitTypes[i];
			var unitTypeName = ui.getString(unitType);
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
		buttonElem.onclick = function(event) {
			var chainId = storage.get(this, 'chain');
			var dropdown = storage.get(this, 'dropdown');
			var unitType = dropdown.selectedIndex;
			handler.addUnit(unitType, chainId);
		}
		
		storage.put(buttonElem, 'chain', id);
		storage.put(buttonElem, 'dropdown', dropDown.input);
		var dropDownDiv = document.createElement('div');
		dropDownDiv.classList.add('contentdiv');
		dropDownDiv.classList.add('addunitdiv');
		dropDownDiv.appendChild(dropDown.div);
		dropDownDiv.appendChild(buttonElem);
		chainDiv.appendChild(dropDownDiv);
		var endDiv = document.createElement('div');
		endDiv.classList.add('contentdiv');
		endDiv.classList.add('iodiv');
		var endHeaderDiv = document.createElement('div');
		endHeaderDiv.classList.add('headerdiv');
		var endLabelDiv = document.createElement('div');
		endLabelDiv.classList.add('labeldiv');
		var labelToOutput = ui.getString('to_output');
		var endLabelText = labelToOutput + ' ' + idString;
		var endLabelNode = document.createTextNode(endLabelText);
		endLabelDiv.appendChild(endLabelNode);
		endHeaderDiv.appendChild(endLabelDiv);
		endDiv.appendChild(endHeaderDiv);
		chainDiv.appendChild(endDiv);
		
		/*
		 * This object represents the signal chain.
		 */
		var chain = {
			'div': chainDiv
		};
		
		return chain;
	}
	
	/*
	 * Renders the signal chains given a configuration returned from the server.
	 */
	this.renderSignalChains = function(configuration) {
		var elem = document.getElementById('signal_chains');
		helper.clearElement(elem);
		var chains = configuration.Chains;
		var numChains = chains.length;
		
		/*
		 * Iterate over the signal chains.
		 */
		for (var i = 0; i < numChains; i++) {
			var chain = chains[i];
			var result = this.renderSignalChain(i, chain);
			var chainDiv = result.div;
			elem.append(chainDiv);
			var spacerDiv = document.createElement('div');
			spacerDiv.classList.add('spacerdiv');
			elem.appendChild(spacerDiv);
		}
		
	}
	
	/*
	 * Renders the latency configuration given a configuration returned from the server.
	 */
	this.renderLatency = function(configuration) {
		var elem = document.getElementById('latency');
		helper.clearElement(elem);
		var unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		unitDiv.classList.add('masterunitdiv');
		var headerDiv = document.createElement('div');
		var labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		labelDiv.classList.add('io');
		var label = ui.getString('latency');
		var labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		headerDiv.appendChild(labelDiv);
		headerDiv.classList.add('headerdiv');
		unitDiv.appendChild(headerDiv);
		var controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		elem.appendChild(unitDiv);
		var labelFramesPerPeriod = ui.getString('frames_per_period');
		var framesPerPeriod = configuration.FramesPerPeriod;
		var dropdownRow = document.createElement('div');
		var values = [64, 128, 256, 512, 1024, 2048, 4096, 8192];
		var valueIdx = 0;
		
		/*
		 * Iterate over all possible values.
		 */
		for (var i = 0; i < values.length; i++) {
			var currentValue = values[i];
			
			/*
			 * If we have a match, store index.
			 */
			if (framesPerPeriod == currentValue)
				valueIdx = i;
			
		}
		
		/*
		 * Parameters for the frames per period drop down menu.
		 */
		var paramsFramesPerPeriod = {
			'label': labelFramesPerPeriod,
			'options': values,
			'selectedIndex': valueIdx
		};
		
		var dropDownFramesPerPeriod = ui.createDropDown(paramsFramesPerPeriod);
		var dropDownFramesPerPeriodElem = dropDownFramesPerPeriod.input;
		
		/*
		 * This is called when the period size changes.
		 */
		dropDownFramesPerPeriodElem.onchange = function(event) {
			var idx = this.selectedIndex;
			var option = this.options[idx];
			var value = option.text;
			handler.setFramesPerPeriod(value);
		};
		
		dropdownRow.appendChild(dropDownFramesPerPeriod.div);
		controlsDiv.appendChild(dropdownRow);
		
		/*
		 * Create unit object.
		 */
		var unit = {
			'controls': controlsDiv,
			'expanded': false
		};
		
		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			var controlsDiv = this.controls;
			var displayValue = '';
			
			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value)
				displayValue = 'block';
			else
				displayValue = 'none';
			
			controlsDiv.style.display = displayValue;
			this.expanded = value;
		}
		
		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		}
		
		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			var state = this.getExpanded();
			this.setExpanded(!state);
		}
		
		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(event) {
			var unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		}
		
		storage.put(labelDiv, 'unit', unit);
	}
	
	/*
	 * Updates the tuner display based on information returned from the server.
	 */
	this.updateTuner = function(result) {
		var cents = result.Cents;
		var frequency = result.Frequency;
		var note = result.Note;
		var centsDiv = document.querySelector('.tunercentsknob');
		var centsKnob = storage.get(centsDiv, 'knob');
		centsKnob.setValue(cents);
		var frequencyDiv = document.querySelector('.tunerfrequencydiv');
		var frequencyString = frequency.toFixed(4);
		frequencyDiv.innerHTML = frequencyString;
		var noteDiv = document.querySelector('.tunernotediv');
		var noteString = note.toString();
		noteDiv.innerHTML = noteString;
	}
	
	/*
	 * Renders the tuner given a configuration returned from the server.
	 */
	this.renderTuner = function(configuration) {
		var chainsConfiguration = configuration.Chains;
		var numChannels = chainsConfiguration.length;
		var tunerConfiguration = configuration.Tuner;
		var elem = document.getElementById('tuner');
		helper.clearElement(elem);
		var unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		unitDiv.classList.add('masterunitdiv');
		var headerDiv = document.createElement('div');
		var labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		labelDiv.classList.add('io');
		var label = ui.getString('tuner');
		var labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		headerDiv.appendChild(labelDiv);
		headerDiv.classList.add('headerdiv');
		unitDiv.appendChild(headerDiv);
		var controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		elem.appendChild(unitDiv);
		var centsString = ui.getString('cents');
		var frequencyString = ui.getString('frequency');
		var noteString = ui.getString('note');
		var centsValue = tunerConfiguration.BeatsPerPeriod;
		
		/*
		 * Parameters for the cents knob.
		 */
		var centsParams = {
			'label': centsString,
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
		
		var centsKnob = ui.createKnob(centsParams);
		var centsKnobNode = centsKnob.node;
		centsKnobNode.classList.add('tunercentsknob');
		var centsKnobObj = centsKnob.obj;
		storage.put(centsKnobNode, 'knob', centsKnobObj);
		var centsKnobDiv = centsKnob.div;
		controlsDiv.appendChild(centsKnobDiv);
		var frequencyRow = document.createElement('div');
		var labelFrequency = ui.getString('frequency');
		var frequencyLabelDiv = document.createElement('div');
		frequencyLabelDiv.classList.add('labeldiv');
		var frequencyLabelNode = document.createTextNode(labelFrequency);
		frequencyLabelDiv.appendChild(frequencyLabelNode);
		frequencyRow.appendChild(frequencyLabelDiv);
		var frequencyValueDiv = document.createElement('div');
		frequencyValueDiv.classList.add('tunerfrequencydiv');
		frequencyRow.appendChild(frequencyValueDiv);
		controlsDiv.appendChild(frequencyRow);
		var noteRow = document.createElement('div');
		var labelNote = ui.getString('note');
		var noteLabelDiv = document.createElement('div');
		noteLabelDiv.classList.add('labeldiv');
		var noteLabelNode = document.createTextNode(labelNote);
		noteLabelDiv.appendChild(noteLabelNode);
		noteRow.appendChild(noteLabelDiv);
		var noteNameDiv = document.createElement('div');
		noteNameDiv.classList.add('tunernotediv');
		noteRow.appendChild(noteNameDiv);
		controlsDiv.appendChild(noteRow);
		var channelRow = document.createElement('div');
		var labelChannel = ui.getString('channel');
		var channels = ['- NONE -'];
		
		/*
		 * Append indices for all channels.
		 */
		for (var i = 0; i < numChannels; i++) {
			var idxString = i.toString();
			channels.push(idxString);
		}
		
		var channelIdx = tunerConfiguration.Channel;
		var channelIdxInc = channelIdx + 1;
		
		/*
		 * Parameters for the channel drop down menu.
		 */
		var paramsChannel = {
			'label': labelChannel,
			'options': channels,
			'selectedIndex': channelIdxInc
		};
		
		var dropDownChannel = ui.createDropDown(paramsChannel);
		var dropDownChannelElem = dropDownChannel.input;
		
		/*
		 * This is called when the channel number changes.
		 */
		dropDownChannelElem.onchange = function(event) {
			var idx = this.selectedIndex;
			var option = this.options[idx];
			var value = option.text;
			var interval = storage.get(this, 'interval');
			window.clearInterval(interval);
			
			/*
			 * This gets executed whenever the timer ticks.
			 */
			var callback = function() {
				handler.refreshTuner();
			}
			
			/*
			 * Handle special case of no channel and register timer
			 * for updating readings for the UI.
			 */
			if (value == '- NONE -')
				value = '-1';
			else {
				var intervalNew = window.setInterval(callback, 250);
				storage.put(this, 'interval', intervalNew);
			}
			
			handler.setTunerValue('channel', value);
		};
		
		dropDownChannelElem.onchange(null);
		var dropDownChannelDiv = dropDownChannel.div;
		channelRow.appendChild(dropDownChannelDiv);
		controlsDiv.appendChild(channelRow);
		
		/*
		 * Create unit object.
		 */
		var unit = {
			'controls': controlsDiv,
			'expanded': false
		};
		
		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			var controlsDiv = this.controls;
			var displayValue = '';
			
			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value)
				displayValue = 'block';
			else
				displayValue = 'none';
			
			controlsDiv.style.display = displayValue;
			this.expanded = value;
		}
		
		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		}
		
		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			var state = this.getExpanded();
			this.setExpanded(!state);
		}
		
		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(event) {
			var unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		}
		
		storage.put(labelDiv, 'unit', unit);
	}
	
	/*
	 * Renders the spatializer given a configuration returned from the server.
	 */
	this.renderSpatializer = function(configuration) {
		var spatializer = configuration.Spatializer;
		var channels = spatializer.Channels;
		var numChannels = channels.length;
		var elem = document.getElementById('spatializer');
		helper.clearElement(elem);
		var unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		unitDiv.classList.add('masterunitdiv');
		var headerDiv = document.createElement('div');
		var labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		labelDiv.classList.add('io');
		var label = ui.getString('spatializer');
		var labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		headerDiv.appendChild(labelDiv);
		headerDiv.classList.add('headerdiv');
		unitDiv.appendChild(headerDiv);
		var controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		elem.appendChild(unitDiv);

		/*
		 * Iterate over the channels.
		 */
		for (var i = 0; i < numChannels; i++) {
			var iString = i.toString();
			var channel = channels[i];
			var azimuth = channel.Azimuth;
			var distance = 10 * channel.Distance;
			var level = 100 * channel.Level;
			var azimuthString = ui.getString('azimuth');
			var azimuthLabel = azimuthString + ' ' + iString;
	
			/*
			 * Parameters for the azimuth knob.
			 */
			var azimuthParams = {
				'label': azimuthLabel,
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
	
			var azimuthKnob = ui.createKnob(azimuthParams);
			var azimuthKnobDiv = azimuthKnob.div;
			controlsDiv.appendChild(azimuthKnobDiv);
			var distanceString = ui.getString('distance');
			var distanceLabel = distanceString + ' ' + iString;
	
			/*
			 * Parameters for the distance knob.
			 */
			var distanceParams = {
				'label': distanceLabel,
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
	
			var distanceKnob = ui.createKnob(distanceParams);
			var distanceKnobDiv = distanceKnob.div;
			controlsDiv.append(distanceKnobDiv);
			var levelString = ui.getString('level');
			var levelLabel = levelString + ' ' + iString;
	
			/*
			 * Parameters for the level knob.
			 */
			var levelParams = {
				'label': levelLabel,
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
	
			var levelKnob = ui.createKnob(levelParams);
			var levelKnobDiv = levelKnob.div;
			controlsDiv.append(levelKnobDiv);
			var azimuthKnobNode = azimuthKnob.node;
			var distanceKnobNode = distanceKnob.node;
			var levelKnobNode = levelKnob.node;
			storage.put(azimuthKnobNode, 'channel', i);
			storage.put(distanceKnobNode, 'channel', i);
			storage.put(levelKnobNode, 'channel', i);
	
			/*
			 * This gets executed when the azimuth value changes.
			 */
			var azimuthHandler = function(knob, value) {
				var node = knob.node();
				var channel = storage.get(node, 'channel');
				handler.setAzimuth(channel, value);
			};

			/*
			 * This gets executed when the distance value changes.
			 */					
			var distanceHandler = function(knob, value) {
				var node = knob.node();
				var channel = storage.get(node, 'channel');
				var distanceValue = (0.1 * value).toFixed(1);
				handler.setDistance(channel, distanceValue);
			};
			
			/*
			 * This gets executed when the level value changes.
			 */					
			var levelHandler = function(knob, value) {
				var node = knob.node();
				var channel = storage.get(node, 'channel');
				var levelValue = (0.01 * value).toFixed(2);
				handler.setLevel(channel, levelValue);
			};
			
			var azimuthKnobObj = azimuthKnob.obj;
			azimuthKnobObj.addListener(azimuthHandler);
			var distanceKnobObj = distanceKnob.obj;
			distanceKnobObj.addListener(distanceHandler);
			var levelKnobObj = levelKnob.obj;
			levelKnobObj.addListener(levelHandler);
		}
		
		/*
		 * Create unit object.
		 */
		var unit = {
			'controls': controlsDiv,
			'expanded': false
		};
		
		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			var controlsDiv = this.controls;
			var displayValue = '';
			
			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value)
				displayValue = 'block';
			else
				displayValue = 'none';
			
			controlsDiv.style.display = displayValue;
			this.expanded = value;
		}
		
		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		}
		
		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			var state = this.getExpanded();
			this.setExpanded(!state);
		}
		
		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(event) {
			var unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		}
		
		storage.put(labelDiv, 'unit', unit);
	}
	
	/*
	 * Renders the metronome given a configuration returned from the server.
	 */
	this.renderMetronome = function(configuration) {
		var metronomeConfiguration = configuration.Metronome;
		var masterOutput = metronomeConfiguration.MasterOutput;
		var elem = document.getElementById('metronome');
		helper.clearElement(elem);
		var unitDiv = document.createElement('div');
		unitDiv.classList.add('contentdiv');
		unitDiv.classList.add('masterunitdiv');
		var headerDiv = document.createElement('div');
		var masterString = ui.getString('master');
		
		/*
		 * Parameters for metronome button.
		 */
		var paramsButton = {
			caption: masterString,
			active: masterOutput
		};
		
		var button = ui.createButton(paramsButton);
		var buttonElem = button.input;
		storage.put(buttonElem, 'active', masterOutput);
		
		/*
		 * This is called when the user clicks on the 'master' button of the metronome.
		 */
		buttonElem.onclick = function(event) {
			var active = !storage.get(this, 'active');
			
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
		}
		
		headerDiv.appendChild(buttonElem);
		var labelDiv = document.createElement('div');
		labelDiv.classList.add('labeldiv');
		labelDiv.classList.add('active');
		labelDiv.classList.add('io');
		var label = ui.getString('metronome');
		var labelNode = document.createTextNode(label);
		labelDiv.appendChild(labelNode);
		headerDiv.appendChild(labelDiv);
		headerDiv.classList.add('headerdiv');
		unitDiv.appendChild(headerDiv);
		var controlsDiv = document.createElement('div');
		controlsDiv.classList.add('controlsdiv');
		unitDiv.appendChild(controlsDiv);
		elem.appendChild(unitDiv);
		var beatsString = ui.getString('beats_per_period');
		var beatsValue = metronomeConfiguration.BeatsPerPeriod;
		
		/*
		 * Parameters for the beats per period knob.
		 */
		var beatsParams = {
			'label': beatsString,
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
		
		var beatsKnob = ui.createKnob(beatsParams);
		var beatsKnobDiv = beatsKnob.div;
		controlsDiv.appendChild(beatsKnobDiv);
		var speedString = ui.getString('speed');
		var speedValue = metronomeConfiguration.Speed;
		
		/*
		 * Parameters for the speed knob.
		 */
		var speedParams = {
			'label': speedString,
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
		
		var speedKnob = ui.createKnob(speedParams);
		var speedKnobDiv = speedKnob.div;
		controlsDiv.appendChild(speedKnobDiv);
		
		/*
		 * This gets executed when the beats per period value changes.
		 */
		var beatsHandler = function(knob, value) {
			handler.setMetronomeValue('beats-per-period', value);
		};
		
		/*
		 * This gets executed when the beats per period value changes.
		 */
		var speedHandler = function(knob, value) {
			handler.setMetronomeValue('speed', value);
		};
		
		var beatsKnobObj = beatsKnob.obj;
		beatsKnobObj.addListener(beatsHandler);
		var speedKnobObj = speedKnob.obj;
		speedKnobObj.addListener(speedHandler);
		var sounds = metronomeConfiguration.Sounds;
		var numSounds = sounds.length;
		var tickSound = metronomeConfiguration.TickSound;
		var tockSound = metronomeConfiguration.TockSound;
		var tickIdx = 0;
		var tockIdx = 0;
		
		/*
		 * Iterate over all sounds and find the tick and tock sound.
		 */
		for (var i = 0; i < numSounds; i++) {
			var sound = sounds[i];
			
			/*
			 * If we found the tick sound, store index.
			 */
			if (sound == tickSound)
				tickIdx = i;
			
			/*
			 * If we found the tock sound, store index.
			 */
			if (sound == tockSound)
				tockIdx = i;
			
		}
		
		var labelTick = ui.getString('tick_sound');
		var labelTock = ui.getString('tock_sound');
		
		/*
		 * Parameters for the tick sound drop down menu.
		 */
		var paramsTick = {
			'label': labelTick,
			'options': sounds,
			'selectedIndex': tickIdx
		};
		
		/*
		 * Parameters for the tock sound drop down menu.
		 */
		var paramsTock = {
			'label': labelTock,
			'options': sounds,
			'selectedIndex': tockIdx
		};
		
		var dropDownTick = ui.createDropDown(paramsTick);
		var dropDownTock = ui.createDropDown(paramsTock);
		var dropDownTickElem = dropDownTick.input;
		var dropDownTockElem = dropDownTock.input;
		
		/*
		 * This is called when the tick sound changes.
		 */
		dropDownTickElem.onchange = function(event) {
			var idx = this.selectedIndex;
			var option = this.options[idx];
			var value = option.text;
			handler.setMetronomeValue('tick-sound', value);
		};
		
		/*
		 * This is called when the tock sound changes.
		 */
		dropDownTockElem.onchange = function(event) {
			var idx = this.selectedIndex;
			var option = this.options[idx];
			var value = option.text;
			handler.setMetronomeValue('tock-sound', value);
		};
		
		var controlRowTick = document.createElement('div');
		controlRowTick.appendChild(dropDownTick.div);
		controlsDiv.appendChild(controlRowTick);
		var controlRowTock = document.createElement('div');
		controlRowTock.appendChild(dropDownTock.div);
		controlsDiv.appendChild(controlRowTock);
		
		/*
		 * Create unit object.
		 */
		var unit = {
			'controls': controlsDiv,
			'expanded': false
		};
		
		/*
		 * Expands or collapses a unit.
		 */
		unit.setExpanded = function(value) {
			var controlsDiv = this.controls;
			var displayValue = '';
			
			/*
			 * Check whether we should expand or collapse the unit.
			 */
			if (value)
				displayValue = 'block';
			else
				displayValue = 'none';
			
			controlsDiv.style.display = displayValue;
			this.expanded = value;
		}
		
		/*
		 * Returns whether a unit is expanded.
		 */
		unit.getExpanded = function() {
			return this.expanded;
		}
		
		/*
		 * Toggles a unit between expanded and collapsed state.
		 */
		unit.toggleExpanded = function() {
			var state = this.getExpanded();
			this.setExpanded(!state);
		}
		
		/*
		 * This is called when a user clicks on the label div.
		 */
		labelDiv.onclick = function(event) {
			var unit = storage.get(this, 'unit');
			unit.toggleExpanded();
		}
		
		storage.put(labelDiv, 'unit', unit);
	}
	
	/*
	 * Renders the signal level analysis section given a configuration returned from the server.
	 */
	this.renderSignalLevels = function(configuration) {
		var batchProcessing = configuration.BatchProcessing;
		var elem = document.getElementById('levels');
		helper.clearElement(elem);
		
		/*
		 * Only display levels if batch processing is disabled on the server.
		 */
		if (!batchProcessing) {
			var unitDiv = document.createElement('div');
			unitDiv.classList.add('contentdiv');
			unitDiv.classList.add('masterunitdiv');
			var headerDiv = document.createElement('div');
			var labelDiv = document.createElement('div');
			labelDiv.classList.add('labeldiv');
			labelDiv.classList.add('active');
			labelDiv.classList.add('io');
			var label = ui.getString('signal_levels');
			var labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			headerDiv.appendChild(labelDiv);
			headerDiv.classList.add('headerdiv');
			unitDiv.appendChild(headerDiv);
			var controlsDiv = document.createElement('div');
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
					var callback = function() {
						
						/*
						 * This is called when the server returns a response.
						 */
						var responseListener = function(response) {
							var dspLoadControl = unit.dspLoadControl;
							var channelNames = unit.channelNames;
							var numNames = channelNames.length;
							var channelControls = unit.channelControls;
							var idx = 0;
							var mismatch = (dspLoadControl == null);
							
							/*
							 * Iterate over all categories ("Inputs", "Outputs",
							 * "Metronome", "Master", ...) in the response.
							 */
							for (var key in response) {
								var keyNative = response.hasOwnProperty(key);
								
								/*
								 * Verify that the key is an actual property of
								 * the response (and not of the prototype).
								 */
								if (keyNative) {
									var category = response[key];
									var numChannels = category.length;
									
									/*
									 * Iterate over the channels.
									 */
									for (var i = 0; i < numChannels; i++) {
										var channel = category[i];
										var channelNameResponse = channel.ChannelName;
										
										/*
										 * If one of the channels does not match,
										 * report mismatch.
										 */
										if (numNames <= idx)
											mismatch = true;
										else {
											var channelNameControls = channelNames[idx];
											
											/*
											 * Check if name of the response matches name
											 * of the control.
											 */
											if (channelNameResponse != channelNameControls)
												mismatch = true;
											
										}
										
										idx++;
									}
									
								}
								
							}
							
							/*
							 * If the channel mapping has changed, create new controls.
							 */
							if (mismatch) {
								var controlsDiv = unit.controls;
								helper.clearElement(controlsDiv);
								var dspLoadString = ui.getString('dsp_load');
								var dspLoadLabelDiv = document.createElement('div');
								var dspLoadLabelNode = document.createTextNode(dspLoadString);
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
								var node = dspLoadControl.node();
								var nodeWrapper = document.createElement('div');
								nodeWrapper.appendChild(node);
								var container = document.createElement('div');
								container.appendChild(dspLoadLabelDiv);
								container.appendChild(nodeWrapper);
								controlsDiv.appendChild(container);
								channelNames = [];
								channelControls = [];
								
								/*
								 * Iterate over all categories ("Inputs", "Outputs",
								 * "Metronome", "Master", ...) in the response.
								 */
								for (var key in response) {
									var isProperty = response.hasOwnProperty(key);
									
									/*
									 * Verify that the key is an actual property of
									 * the response (and not of the prototype).
									 */
									if (isProperty) {
										var category = response[key];
										var numChannels = category.length;
										
										/*
										 * Iterate over the channels.
										 */
										for (var i = 0; i < numChannels; i++) {
											var channel = category[i];
											var channelName = channel.ChannelName;
											var channelControl = pureknob.createBarGraph(400, 40);
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
											var channelNameDiv = document.createElement('div');
											var channelNameNode = document.createTextNode(channelName);
											channelNameDiv.appendChild(channelNameNode);
											var node = channelControl.node();
											var nodeWrapper = document.createElement('div');
											nodeWrapper.appendChild(node);
											var container = document.createElement('div');
											container.appendChild(channelNameDiv);
											container.appendChild(nodeWrapper);
											controlsDiv.appendChild(container);
										}
									
									}
									
								}
								
							}
							
							/*
							 * Display DSP load.
							 */
							if (dspLoadControl != null) {
								var dspLoad = response.DSPLoad;
								dspLoadControl.setValue(dspLoad);
							}
							
							idx = 0;
							
							/*
							 * Iterate over all categories ("Inputs", "Outputs",
							 * "Metronome", "Master", ...) in the response.
							 */
							for (var key in response) {
								var isProperty = response.hasOwnProperty(key);
								
								/*
								 * Verify that the key is an actual property of
								 * the response (and not of the prototype).
								 */
								if (isProperty) {
									var category = response[key];
									var numChannels = category.length;
									
									/*
									 * Iterate over the channels.
									 */
									for (var i = 0; i < numChannels; i++) {
										var channel = category[i];
										var channelLevel = channel.Level;
										var channelPeak = channel.Peak;
										var channelControl = channelControls[idx];
										channelControl.setValue(channelLevel);
										channelControl.setPeaks([channelPeak]);
										idx++;
									}
								
								}
								
							}
							
							unit.dspLoadControl = dspLoadControl;
							unit.channelNames = channelNames;
							unit.channelControls = channelControls;
						}
						
						handler.getLevelAnalysis(responseListener);
					};
					
					var timer = window.setInterval(callback, 200);
					unit.timer = timer;
				} else {
					var timer = unit.timer;
					
					/*
					 * If a timer is registered, clear it.
					 */
					if (timer != null)
						window.clearInterval(timer);
					
					unit.timer = null;
				}
				
			};
			
			/*
			 * Create unit object.
			 */
			var unit = {
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
				var controlsDiv = this.controls;
				var displayValue = '';
			
				/*
				 * Check whether we should expand or collapse the unit.
				 */
				if (value)
					displayValue = 'block';
				else
					displayValue = 'none';
			
				controlsDiv.style.display = displayValue;
				this.expanded = value;
				var listeners = this.listeners;
				
				/*
				 * Check if there are listeners resistered.
				 */
				if (listeners != null) {
					var numListeners = listeners.length;
					
					/*
					 * Invoke each listener.
					 */
					for (var i = 0; i < numListeners; i++) {
						var listener = listeners[i];
						listener(this, value);
					}
					
				}
				
			}
		
			/*
			 * Returns whether a unit is expanded.
			 */
			unit.getExpanded = function() {
				return this.expanded;
			}
		
			/*
			 * Toggles a unit between expanded and collapsed state.
			 */
			unit.toggleExpanded = function() {
				var state = this.getExpanded();
				this.setExpanded(!state);
			}
		
			/*
			 * This is called when a user clicks on the label div.
			 */
			labelDiv.onclick = function(event) {
				var unit = storage.get(this, 'unit');
				unit.toggleExpanded();
			}
		
			storage.put(labelDiv, 'unit', unit);
		}
		
	}
	
	/*
	 * Renders the 'processing' button given a configuration returned from the server.
	 */
	this.renderProcessing = function(configuration) {
		var batchProcessing = configuration.BatchProcessing;
		var elem = document.getElementById('processing');
		helper.clearElement(elem);
		
		/*
		 * Only display this if batch processing is enabled on the server.
		 */
		if (batchProcessing) {
			var unitDiv = document.createElement('div');
			unitDiv.classList.add('contentdiv');
			unitDiv.classList.add('masterunitdiv');
			var headerDiv = document.createElement('div');
			var processString = ui.getString('process_now');
		
			/*
			 * Parameters for process button.
			 */
			var paramsButton = {
				caption: processString,
				active: false
			};
		
			var button = ui.createButton(paramsButton);
			var buttonElem = button.input;
			storage.put(buttonElem, 'active', batchProcessing);
		
			/*
			 * This is called when the user clicks on the 'process' button.
			 */
			buttonElem.onclick = function(event) {
				var active = storage.get(this, 'active');
			
				/*
				 * Trigger batch processing if the control is active.
				 */
				if (active)
					handler.process();
			
			}
		
			headerDiv.appendChild(buttonElem);
			var labelDiv = document.createElement('div');
			labelDiv.classList.add('labeldiv');
			labelDiv.classList.add('active');
			labelDiv.classList.add('io');
			var label = ui.getString('batch_processing');
			var labelNode = document.createTextNode(label);
			labelDiv.appendChild(labelNode);
			headerDiv.appendChild(labelDiv);
			headerDiv.classList.add('headerdiv');
			unitDiv.appendChild(headerDiv);
			var controlsDiv = document.createElement('div');
			controlsDiv.classList.add('controlsdiv');
			unitDiv.appendChild(controlsDiv);
			elem.appendChild(unitDiv);
		
			/*
			 * Create unit object.
			 */
			var unit = {
				'controls': controlsDiv,
				'expanded': false
			};
		
			/*
			 * Expands or collapses a unit.
			 */
			unit.setExpanded = function(value) {
				var controlsDiv = this.controls;
				var displayValue = '';
			
				/*
				 * Check whether we should expand or collapse the unit.
				 */
				if (value)
					displayValue = 'block';
				else
					displayValue = 'none';
			
				controlsDiv.style.display = displayValue;
				this.expanded = value;
			}
		
			/*
			 * Returns whether a unit is expanded.
			 */
			unit.getExpanded = function() {
				return this.expanded;
			}
		
			/*
			 * Toggles a unit between expanded and collapsed state.
			 */
			unit.toggleExpanded = function() {
				var state = this.getExpanded();
				this.setExpanded(!state);
			}
		
			/*
			 * This is called when a user clicks on the label div.
			 */
			labelDiv.onclick = function(event) {
				var unit = storage.get(this, 'unit');
				unit.toggleExpanded();
			}
		
			storage.put(labelDiv, 'unit', unit);
		}
		
	}
	
}

var ui = new UI();

/*
 * This class implements all handler functions for user interaction.
 */
function Handler() {
	var self = this;
	
	/*
	 * This is called when a new effects unit should be added.
	 */
	this.addUnit = function(unitType, chain) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt, otherwise refresh rack.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Adding new unit failed: ' + reason;
					console.log(msg);
				} else
					self.refresh();
				
			}
			
		};
		
		var unitTypeString = unitType.toString();
		var chainString = chain.toString();
		var request = new Request();
		request.append('cgi', 'add-unit');
		request.append('type', unitTypeString);
		request.append('chain', chainString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a new level analysis should be obtained.
	 */
	this.getLevelAnalysis = function(callback) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var levels = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (levels != null)
				callback(levels);
			
		};
		
		var request = new Request();
		request.append('cgi', 'get-level-analysis');
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, false);
	}
	
	/*
	 * This is called when a unit should be moved down the chain.
	 */
	this.moveDown = function(chain, unit) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Moving unit down failed: ' + reason;
					console.log(msg);
				} else
					self.refresh();
				
			}
			
		};
		
		var chainString = chain.toString();
		var unitString = unit.toString();
		var request = new Request();
		request.append('cgi', 'move-down');
		request.append('chain', chainString);
		request.append('unit', unitString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a unit should be moved up the chain.
	 */
	this.moveUp = function(chain, unit) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Moving unit up failed: ' + reason;
					console.log(msg);
				} else
					self.refresh();
				
			}
			
		};
		
		var chainString = chain.toString();
		var unitString = unit.toString();
		var request = new Request();
		request.append('cgi', 'move-up');
		request.append('chain', chainString);
		request.append('unit', unitString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a unit should be removed from a chain.
	 */
	this.removeUnit = function(chain, unit) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Removing unit failed: ' + reason;
					console.log(msg);
				} else
					self.refresh();
				
			}
			
		};
		
		var chainString = chain.toString();
		var unitString = unit.toString();
		var request = new Request();
		request.append('cgi', 'remove-unit');
		request.append('chain', chainString);
		request.append('unit', unitString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a new azimuth value should be set.
	 */
	this.setAzimuth = function(chain, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting azimuth value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var chainString = chain.toString();
		var valueString = value.toString()
		var request = new Request();
		request.append('cgi', 'set-azimuth');
		request.append('chain', chainString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a unit should be bypassed or bypass should be disabled for a unit.
	 */
	this.setBypass = function(chain, unit, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting bypass value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var chainString = chain.toString();
		var unitString = unit.toString();
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-bypass');
		request.append('chain', chainString);
		request.append('unit', unitString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a new distance value should be set.
	 */
	this.setDistance = function(chain, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting distance value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var chainString = chain.toString();
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-distance');
		request.append('chain', chainString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a discrete value should be set.
	 */
	this.setDiscreteValue = function(chain, unit, param, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting discrete value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var chainString = chain.toString();
		var unitString = unit.toString();
		var paramString = param.toString();
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-discrete-value');
		request.append('chain', chainString);
		request.append('unit', unitString);
		request.append('param', paramString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when period size should be changed.
	 */
	this.setFramesPerPeriod = function(value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting frames per period failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-frames-per-period');
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a new level value should be set.
	 */
	this.setLevel = function(chain, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting level value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var chainString = chain.toString();
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-level');
		request.append('chain', chainString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a metronome value should be changed.
	 */
	this.setMetronomeValue = function(param, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting metronome value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var paramString = param.toString();
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-metronome-value');
		request.append('param', paramString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a tuner value should be changed.
	 */
	this.setTunerValue = function(param, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting tuner value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var paramString = param.toString();
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-tuner-value');
		request.append('param', paramString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a numeric value should be set.
	 */
	this.setNumericValue = function(chain, unit, param, value) {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var webResponse = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (webResponse != null) {
				
				/*
				 * If we were not successful, log failed attempt.
				 */
				if (webResponse.Success != true) {
					var reason = webResponse.Reason;
					var msg = 'Setting numeric value failed: ' + reason;
					console.log(msg);
				}
				
			}
			
		};
		
		var chainString = chain.toString();
		var unitString = unit.toString();
		var paramString = param.toString();
		var valueString = value.toString();
		var request = new Request();
		request.append('cgi', 'set-numeric-value');
		request.append('chain', chainString);
		request.append('unit', unitString);
		request.append('param', paramString);
		request.append('value', valueString);
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when the configuration needs to be refreshed.
	 */
	this.refresh = function() {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var configuration = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (configuration != null) {
				ui.renderSignalChains(configuration);
				ui.renderLatency(configuration);
				ui.renderTuner(configuration);
				ui.renderSpatializer(configuration);
				ui.renderMetronome(configuration);
				ui.renderSignalLevels(configuration);
				ui.renderProcessing(configuration);
			}
			
		};
		
		var request = new Request();
		request.append('cgi', 'get-configuration');
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
	/*
	 * This is called when a new analysis should be performed by the tuner.
	 */
	this.refreshTuner = function() {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var analysis = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (analysis != null) {
				ui.updateTuner(analysis);
			}
			
		};
		
		var request = new Request();
		request.append('cgi', 'get-tuner-analysis');
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, false);
	}
	
	/*
	 * This is called when the user clicks on the 'process' button.
	 */
	this.process = function() {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			helper.blockSite(true);
		};
		
		var request = new Request();
		request.append('cgi', 'process');
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, false);
	}
	
	/*
	 * This is called when the user interface initializes.
	 */
	this.initialize = function() {
		
		/*
		 * This gets called when the server returns a response.
		 */
		var responseHandler = function(response) {
			var unitTypes = helper.parseJSON(response);
			
			/*
			 * Check if the response is valid JSON.
			 */
			if (unitTypes != null) {
				var numUnitTypes = unitTypes.length;
				
				/*
				 * Iterate over the unit types and add them to the global list.
				 */
				for (var i = 0; i < numUnitTypes; i++) {
					var t = unitTypes[i];
					globals.unitTypes.push(t);
				}
				
				self.refresh();
			}
			
		};
		
		var request = new Request();
		request.append('cgi', 'get-unit-types');
		var requestBody = request.getData();
		ajax.request('POST', globals.cgi, requestBody, responseHandler, true);
	}
	
}

/*
 * The (global) event handlers.
 */
var handler = new Handler();
document.addEventListener('DOMContentLoaded', handler.initialize);

