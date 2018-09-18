package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "image/jpeg"

	_ "image/gif"

	_ "image/png"
)

func main() {
	var dir string
	var validFiles []*os.File
	var input string

	flag.StringVar(&dir, "dir", "", "Enter the image directory path")
	flag.Parse()
	//Making sure that the user wants to work in the current directory. No directory passed.
	fmt.Printf("Directory passed: %s\n", dir)

	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Using directory: %s\n", dir)
		fmt.Println("Do you want to use the current directory as root? (y/n)")
		fmt.Scanf("%s\n", &input)
		if input != "y" && input != "Y" {
			fmt.Println("Aborting...")
			return
		}
	}

	var dirFiles []string
	fmt.Println("Do you want to include all of the subdirectories? (y/n)")
	fmt.Scanf("%s\n", &input)
	if input == "y" || input == "Y" {
		fmt.Println("Drilling down...")
		_ = filepath.Walk(dir, digDown(&dirFiles))

	} else if input == "n" || input == "N" {
		fmt.Println("Using the current directory only.")
		_ = filepath.Walk(dir, singleDir(&dirFiles, dir))
	} else {
		fmt.Println("Unrecognized command.")
		fmt.Println("Aborting...")
	}

	//Parsing out the files to be renamed
	for _, file := range dirFiles {
		//opening all of the files.
		img, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
			continue //the error prints out but skips a problematic file
		}

		//Selecting for image files. Non-image (all but .jpeg, .png, and .gif) files will throw an error.
		_, _, err = image.DecodeConfig(img)
		if err != nil {
			_, n := filepath.Split(img.Name())
			fmt.Printf("%s was skipped:\t\t %s\n", n, err)
			continue //The error will be printed out and the file will be skipped.
		}
		//checking for the "(..#x#.." condition. This would likely signify that the file was already renamed.
		label := strings.SplitAfter(img.Name(), "(")
		lastLabel := strings.Join(label[len(label)-1:], "")
		if func() bool {
			x := strings.Index(lastLabel, "x")
			if x != -1 && x != 0 && x != len(lastLabel) {
				if "0" <= string(lastLabel[x-1]) && string(lastLabel[x-1]) <= "9" {
					if "0" <= string(lastLabel[x+1]) && string(lastLabel[x+1]) <= "9" {
						return true
					}
				}
			}
			return false
		}() {
			_, n := filepath.Split(img.Name())
			fmt.Printf("%s was skipped:\t\t (already renamed)\n", n)
			continue
		}
		//Creating a table of valid files to be renamed.
		validFiles = append(validFiles, img)
		img.Close()

	}
	//Quitting the program if there're no valid files to rename.
	if len(validFiles) == 0 {
		fmt.Println("No files to rename.")
		fmt.Println("Quitting...")
		return
	}
	//Counting and referencing files to be renamed by their current names.
	fmt.Printf("\nThe following %d files will be renamed:\n", len(validFiles))
	for _, file := range validFiles {
		_, n := filepath.Split(file.Name())
		fmt.Println(n)
	}
	fmt.Println("Continue? (y/n)") //Asking user to make sure they want to continue
	fmt.Scanf("%s\n", &input)

	switch input {
	case "y", "Y":
		fmt.Println("Renaming Files...") //running through the same routine as above.
		for _, file := range dirFiles {

			// img, err := os.Open(dir + file.Name())
			img, err := os.Open(file)
			if err != nil {

				fmt.Println(err)
				continue
			}

			conf, form, err := image.DecodeConfig(img)
			if err != nil {
				continue //not being verbose about skipping non-image files though.
			}

			//not being verbose about skipping the already renamed files.
			label := strings.SplitAfter(img.Name(), "(")
			lastLabel := strings.Join(label[len(label)-1:], "")
			if func() bool {
				x := strings.Index(lastLabel, "x")
				if x != -1 && x != 0 && x != len(lastLabel) {
					if "0" <= string(lastLabel[x-1]) && string(lastLabel[x-1]) <= "9" {
						if "0" <= string(lastLabel[x+1]) && string(lastLabel[x+1]) <= "9" {
							return true
						}
					}
				}
				return false
			}() {
				continue
			}
			//Actual renaming.
			nameBits := strings.Split(img.Name(), ".")                                         //stripping file's extention
			newNameRaw := strings.Join(nameBits[0:len(nameBits)-1], "")                        //reconstructing file's name sans extention
			newName := fmt.Sprintf("%s_(%dx%d).%s", newNameRaw, conf.Width, conf.Height, form) //adding res. suffix
			fmt.Println(newName)                                                               //Printing out new names.
			img.Close()                                                                        //Closing open files for renaming.
			err = os.Rename(img.Name(), newName)                                               //actual renaming process
			if err != nil {
				fmt.Println(err)
			}
		}

	case "n", "N":
		fmt.Println("Aborting...")
	default:
		fmt.Println("Unrecognized command.")
		fmt.Println("Aborting...")

	}
}

func digDown(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		*files = append(*files, path)
		return nil
	}
}

func singleDir(files *[]string, dir string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if info.IsDir() == true && path != dir {
			return filepath.SkipDir
		}

		*files = append(*files, path)
		return nil
	}
}
