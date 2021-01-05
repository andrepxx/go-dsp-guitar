package path

import (
	"testing"
)

/*
 * Test sanitation of file system paths.
 */
func TestPaths(t *testing.T) {

	/*
	 * Input values.
	 */
	in := []string{
		"/neither/leading/nor/trailing/space",
		" /single/leading/space",
		"  /multiple/leading/spaces",
		"/single/trailing/space ",
		"/multiple/trailing/spaces  ",
		" /single/leading/and/trailing/space ",
		"  /multiple/leading/and/trailing/spaces  ",
		"  /more/leading/than/trailing/spaces ",
		" /more/trailing/than/leading/spaces  ",
		"'/neither/leading/nor/trailing/space'",
		" '/single/leading/space'",
		"  '/multiple/leading/spaces'",
		"'/single/trailing/space' ",
		"'/multiple/trailing/spaces'  ",
		" '/single/leading/and/trailing/space' ",
		"  '/multiple/leading/and/trailing/spaces'  ",
		"  '/more/leading/than/trailing/spaces' ",
		" '/more/trailing/than/leading/spaces'  ",
		"\"/neither/leading/nor/trailing/space\"",
		" \"/single/leading/space\"",
		"  \"/multiple/leading/spaces\"",
		"\"/single/trailing/space\" ",
		"\"/multiple/trailing/spaces\"  ",
		" \"/single/leading/and/trailing/space\" ",
		"  \"/multiple/leading/and/trailing/spaces\"  ",
		"  \"/more/leading/than/trailing/spaces\" ",
		" \"/more/trailing/than/leading/spaces\"  ",
		"",
		" ",
		"''",
		" ''",
		"'' ",
		" '' ",
		"\"\"",
		" \"\"",
		"\"\" ",
		" \"\" ",
		"x",
	}

	/*
	 * Expected output.
	 */
	out := []string{
		"/neither/leading/nor/trailing/space",
		"/single/leading/space",
		"/multiple/leading/spaces",
		"/single/trailing/space",
		"/multiple/trailing/spaces",
		"/single/leading/and/trailing/space",
		"/multiple/leading/and/trailing/spaces",
		"/more/leading/than/trailing/spaces",
		"/more/trailing/than/leading/spaces",
		"/neither/leading/nor/trailing/space",
		"/single/leading/space",
		"/multiple/leading/spaces",
		"/single/trailing/space",
		"/multiple/trailing/spaces",
		"/single/leading/and/trailing/space",
		"/multiple/leading/and/trailing/spaces",
		"/more/leading/than/trailing/spaces",
		"/more/trailing/than/leading/spaces",
		"/neither/leading/nor/trailing/space",
		"/single/leading/space",
		"/multiple/leading/spaces",
		"/single/trailing/space",
		"/multiple/trailing/spaces",
		"/single/leading/and/trailing/space",
		"/multiple/leading/and/trailing/spaces",
		"/more/leading/than/trailing/spaces",
		"/more/trailing/than/leading/spaces",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"x",
	}

	/*
	 * Test the function for each input value.
	 */
	for i, val := range in {
		expected := out[i]
		result := Sanitize(val)

		/*
		 * Check if we got the expected result.
		 */
		if result != expected {
			t.Errorf("Sanitization of string number %d failed: Input '%s', expected '%s', got '%s'.", i, val, result, expected)
		}

	}

}
