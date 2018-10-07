package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	flag "github.com/spf13/pflag"
)

func main() {
	// Initializing //
	startNumber := flag.IntP("startpage", "s", 0, "The page to start printing at [Necessary, no greater than endpage]")
	endNumber := flag.IntP("endpage", "e", 0, "The page to end printing at [Necessary, no less than startpage]")
	lineNumber := flag.IntP("linenumber", "l", 72, "If this flag is used, a page will consist of a fixed number of characters, which is given by you")
	forcePage := flag.BoolP("forcepaging", "f", false, "Change page only if '-f' appears [Cannot be used with -l]")
	destinationPrinter := flag.StringP("destination", "d", "", "Choose a printer to accept the result as a task")

	// StdErr printer //
	l := log.New(os.Stderr, "", 0)

	// Data holder //
	bytes := make([]byte, 65535)
	var data string
	var resultData string

	flag.Parse()

	// Are necessary flags given? //
	if *startNumber == 0 || *endNumber == 0 {
		l.Println("Necessary flags are not given!")
		flag.Usage()
		os.Exit(1)
	}

	// Are flags value valid? //
	if (*startNumber > *endNumber) || *startNumber < 0 || *endNumber < 0 || *lineNumber <= 0 {
		l.Println("Invalid flag values!")
		flag.Usage()
		os.Exit(1)
	}

	// Are lineNumber and forcePage set at the same time? //
	if *lineNumber != 72 && *forcePage {
		l.Println("Linenumber and forcepaging cannot be set at the same time!")
		flag.Usage()
		os.Exit(1)
	}

	// Too many arguments? //
	if flag.NArg() > 1 {
		l.Println("Too many arguments!")
		flag.Usage()
		os.Exit(1)
	}

	// StdIn or File? //
	if flag.NArg() == 0 {
		// StdIn condition //
		reader := bufio.NewReader(os.Stdin)

		size, err := reader.Read(bytes)

		for size != 0 && err == nil {
			data = data + string(bytes)
			size, err = reader.Read(bytes)
		}

		// Error
		if err != io.EOF {
			l.Println("Error occured when reading from StdIn:\n", err.Error())
			os.Exit(1)
		}

	} else {
		// File condition //
		file, err := os.Open(flag.Args()[0]) // TODO TEST: is PATH needed?
		if err != nil {
			l.Println("Error occured when opening file:\n", err.Error())
			os.Exit(1)
		}

		// Read the whole file
		size, err := file.Read(bytes)

		for size != 0 && err == nil {
			data = data + string(bytes)
			size, err = file.Read(bytes)
		}

		// Error
		if err != io.EOF {
			l.Println("Error occured when reading file:\n", err.Error())
			os.Exit(1)
		}
	}

	// LineNumber or ForcePaging? //
	if *forcePage {
		// ForcePaging //
		pagedData := strings.SplitAfter(data, "\f")

		if len(pagedData) < *endNumber {
			l.Println("Invalid flag values! Too large endNumber!")
			flag.Usage()
			os.Exit(1)
		}

		resultData = strings.Join(pagedData[*startNumber-1:*endNumber], "")
	} else {
		// LineNumber //
		lines := strings.SplitAfter(data, "\n")
		if len(lines) < (*endNumber-1)*(*lineNumber)+1 {
			l.Println("Invalid flag values! Too large endNumber!")
			flag.Usage()
			os.Exit(1)
		}
		if len(lines) < *endNumber*(*lineNumber) {
			resultData = strings.Join(lines[(*startNumber)*(*lineNumber)-(*lineNumber):], "")
		} else {
			resultData = strings.Join(lines[(*startNumber)*(*lineNumber)-(*lineNumber):(*endNumber)*(*lineNumber)], "")
		}
	}

	writer := bufio.NewWriter(os.Stdout)

	// StdOut or Printer? //
	if *destinationPrinter == "" {
		// StdOut //
		fmt.Printf("%s", resultData)
	} else {
		// Printer //
		cmd := exec.Command("cat" /*, "-d"+*destinationPrinter*/)
		lpStdin, err := cmd.StdinPipe()

		if err != nil {
			l.Println("Error occured when trying to send data to lp:\n", err.Error())
			os.Exit(1)
		}
		go func() {
			defer lpStdin.Close()
			io.WriteString(lpStdin, resultData)
		}()

		out, err := cmd.CombinedOutput()
		if err != nil {
			l.Println("Error occured when sending data to lp:\n", err.Error())
			os.Exit(1)
		}

		_, err = writer.Write(out)

		if err != nil {
			l.Println("Error occured when writing information to StdOut:\n", err.Error())
			os.Exit(1)
		}
	}
}
