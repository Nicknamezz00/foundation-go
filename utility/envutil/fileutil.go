package envutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func RelPathFromWDToSrc() string {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		return ""
	}

	// Get the current folder (assuming it's the folder where the executable is located)
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return ""
	}

	// Get the relative path
	relativePath, err := filepath.Rel(wd, currentDir)
	if err != nil {
		fmt.Printf("Error calculating relative path: %v\n", err)
		return ""
	}
	return relativePath
}
