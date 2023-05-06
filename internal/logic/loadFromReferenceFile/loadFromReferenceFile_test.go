package loadFromReferenceFile

// import the testing package
import "testing"

// define a test function that takes a testing pointer as an argument
// and uses the testing package methods to run the test cases
func TestGetAllGzFiles(t *testing.T) {
	// create a slice of test cases, each containing a directory path
	// and an expected slice of gz file paths or an expected error message

	actual, err := getAllGzFiles("/mnt/c/Users/ider/Downloads")
	if err == nil {
		t.Log(actual)
	} else {
		t.Error(err)
	}

}

// define a helper function that takes two slices of strings as arguments
// and returns true if they have the same length and elements, false otherwise
func equalSlices(a, b []string) bool {
	// check if the lengths of the slices are different
	if len(a) != len(b) {
		return false
	}

	// loop through the elements of the slices using indexes
	for i := range a {
		// check if the elements at the same index are different
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
