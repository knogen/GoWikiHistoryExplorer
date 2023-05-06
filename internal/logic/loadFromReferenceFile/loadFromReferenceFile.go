package loadFromReferenceFile

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// define a function that takes a directory path as a string argument
// and returns a slice of strings containing the paths of all gz files
// in that directory and its subdirectories, or an error if any
func getAllGzFiles(dir string) ([]string, error) {
	// create an empty slice to store the gz file paths
	var gzFiles []string

	// use the filepath.Walk function to traverse the directory tree
	// it takes a root path and a walk function as arguments
	err := filepath.Walk(dir,
		// define the walk function that takes a path, file info and error as arguments
		// and returns an error if any
		func(path string, info os.FileInfo, err error) error {
			// check if there is an error and return it if any
			if err != nil {
				return err
			}

			// check if the file has a .gz extension using filepath.Ext function
			// it returns the extension of the file name with a leading dot
			if filepath.Ext(path) == ".gz" {
				// append the gz file path to the slice
				gzFiles = append(gzFiles, path)
			}

			// return nil to continue walking
			return nil
		})

	// check if there is an error from walking and return it if any
	if err != nil {
		return nil, err
	}

	// return the slice of gz file paths and nil error
	return gzFiles, nil
}

func Main() {
	// get all gz file
	// 229：/home/ni/data/wiki/tmpdata_allref_2022new_all_0621_zh/ 这下面700多个wiki的引用文件
	filers, err := getAllGzFiles("/home/ni/data/wiki/tmpdata_allref_2022new_all_0621_zh/")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("gz file count", len(filers))

}
