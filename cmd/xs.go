package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/juju/errors"
	"github.com/tealeg/xlsx"
	"github.com/yuki-eto/xlsx_searcher/semaphore"
)

type Result struct {
	XlsxPath string
	Sheet    string
	Cell     string
	Value    string
}

var (
	searchStr = ""
	resultCh  chan Result
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("please input search string")
		os.Exit(1)
	} else if len(args) < 3 {
		fmt.Println("please input path")
	}

	searchStr = os.Args[1]
	rootPath := os.Args[2]

	resultCh = make(chan Result)
	go selectResultChannel()

	s := semaphore.NewSemaphore(runtime.NumCPU())
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Trace(err)
		}
		if !strings.HasSuffix(path, ".xlsx") {
			return nil
		}
		if strings.HasPrefix(filepath.Base(path), "~") {
			return nil
		}

		f, err := xlsx.OpenFile(path)
		if err != nil {
			return errors.Trace(err)
		}

		for _, sheet := range f.Sheets {
			s.Go(createSearchFunc(path, sheet))
		}
		return nil
	})
	panicIf(err)

	err = s.Wait()
	panicIf(err)

	close(resultCh)
}

func panicIf(err error) {
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

func selectResultChannel() {
	for {
		select {
		case r, ok := <-resultCh:
			if ok {
				fmt.Printf("%s:%s(%s) %s", r.XlsxPath, r.Sheet, r.Cell, r.Value)
				println("")
			} else {
				break
			}
		}
	}
}

func createSearchFunc(path string, sheet *xlsx.Sheet) func() error {
	s := semaphore.NewSemaphore(10)
	return func() error {
		for rowIndex, row := range sheet.Rows {
			s.Go(createSearchRowFunc(path, sheet, rowIndex, row))
		}
		return s.Wait()
	}
}

func createSearchRowFunc(path string, sheet *xlsx.Sheet, rowIndex int, row *xlsx.Row) func() error {
	return func() error {
		for cellIndex, cell := range row.Cells {
			if strings.Contains(cell.Value, searchStr) {
				cellIndexStr := cellIndexToAlphabet(cellIndex)
				resultCh <- Result{
					XlsxPath: path,
					Sheet:    sheet.Name,
					Cell:     fmt.Sprintf("%s%d", cellIndexStr, rowIndex+1),
					Value:    cell.String(),
				}
			}
		}
		return nil
	}
}

func cellIndexToAlphabet(index int) string {
	str := strconv.FormatInt(int64(index+1), 26)
	var a []string
	for _, code := range []byte(str) {
		if code <= 57 {
			code += 16
		} else {
			code -= 23
		}
		a = append(a, string(code))
	}
	return strings.Join(a, "")
}
