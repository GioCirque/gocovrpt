package lib

import (
	"bytes"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

func MakeFileDir(filePath string) error {
	dir := path.Dir(filePath)
	return os.MkdirAll(dir, os.ModePerm)
}

func MakeFile(filePath string) (*os.File, error) {
	err := MakeFileDir(filePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func GetRelPathParts(commonRoot, filePath string) []PathTuple {
	if commonRoot == filePath {
		return []PathTuple{}
	}
	diffPath := filePath[len(commonRoot)+1:]
	pathParts := strings.Split(path.Dir(diffPath), string(os.PathSeparator))
	pathMap := make([]PathTuple, len(pathParts)+1)
	pathMap[0] = PathTuple{Name: path.Base(commonRoot), Path: strings.Repeat("../", len(pathParts))}
	pathMap[0].Path = pathMap[0].Path[0 : len(pathMap[0].Path)-1]
	if pathParts[len(pathParts)-1] == "." {
		pathMap = pathMap[0 : len(pathMap)-1]
		pathParts = pathParts[0 : len(pathParts)-1]
	}

	for i, part := range ReverseArray(pathParts) {
		pathMap[len(pathParts)-i] = PathTuple{part, path.Join(strings.Repeat("../", i+1), part)}
	}
	return pathMap
}

func GetRelRootPath(outFilePath, commonRoot string) string {
	absRoot, _ := filepath.Abs(commonRoot)
	absOutPath, _ := filepath.Abs(outFilePath)
	crLen := len(strings.Split(absRoot, string(os.PathSeparator)))
	ofLen := len(strings.Split(path.Dir(absOutPath), string(os.PathSeparator)))
	backLevels := ofLen - crLen
	relRootPath := strings.Repeat("../", backLevels)
	return relRootPath
}

func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetSourceFilePath(sourceDir string, sourceFile string) (string, error) {
	srcParts := strings.Split(sourceFile, string(os.PathSeparator))
	for i := 0; i < len(srcParts); i++ {
		possiblePath := path.Join(append([]string{sourceDir}, srcParts[i:]...)...)

		if FileExists(possiblePath) {
			return possiblePath, nil
		}
	}

	return "", UnresolvablePathError(sourceFile)
}

func GetOutPathInfo(outPath string, fileName string, ext string, root string) (dispPath, rprtPath string) {
	var newPath = fileName
	if root != "" {
		newPath = newPath[len(root):]
	}
	dispPath = newPath
	rprtPath = SwapFileExt(path.Join(outPath, newPath), ext)
	return
}

func SwapFileExt(filePath string, ext string) string {
	var extension = path.Ext(filePath)
	var name = filePath[0 : len(filePath)-len(extension)]
	return name + ext
}

func ReverseArray[T any](array []T) []T {
	var result = make([]T, len(array))
	for i, v := range array {
		result[len(array)-i-1] = v
	}
	return result
}

func InsertStringAt(str string, insert string, index int) string {
	safeDex := int(math.Min(float64(index), float64(len(str))))
	return str[:safeDex] + insert + str[safeDex:]
}

func GetCommonRoot(rfs []ReportedFile) string {
	var commonPath = ""
	var folders = make([][]string, len(rfs))
	for i, f := range rfs {
		folders[i] = strings.Split(f.FileName, string(os.PathSeparator))
	}

	for i := 0; i < len(folders); i++ {
		var thisFolder = folders[0][i]
		var allMatched = true
		for j := 1; j < len(folders) && allMatched; j++ {
			if len(folders[j]) < i {
				allMatched = false
				break
			}
			allMatched = allMatched && folders[j][i] == thisFolder
		}

		if allMatched {
			commonPath += thisFolder + string(os.PathSeparator)
		} else {
			break
		}

	}
	return commonPath
}

func Offset(reader *bytes.Reader) int {
	return int(reader.Size()) - reader.Len()
}

// Calculates the covered percentage for a profile, optionally multiplied by 100.
func GetCoveredPct(blocks []cover.ProfileBlock, multiplied bool) (result float64) {
	max := 0.0
	covCount := 0
	stmtCount := 0
	for _, b := range blocks {
		max = math.Max(max, float64(b.Count))
		covCount += b.Count
		stmtCount += b.NumStmt
	}
	if max > 1 {
		result = float64(covCount) / float64(stmtCount)
	} else {
		result = float64(covCount) / float64(len(blocks))
	}
	if multiplied {
		result *= 100
	}
	if math.IsNaN(result) {
		result = 0
	}
	return
}

// GetSourceFilePath returns the contents of a file.
func GetSourceCode(sourceFile string) (result string, err error) {
	file, err := os.Open(sourceFile)
	if err != nil {
		return "", err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	buf := make([]byte, fileInfo.Size())
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// GetLines returns the reported and covered lines for this file.
func GetProfiledLines(cp *cover.Profile) (reportedLines, coveredLines []ReportedBlock) {
	for _, b := range cp.Blocks {
		covered := b.Count > 0
		newBlock := newReportedBlock(&b, covered)
		if covered {
			coveredLines = append(coveredLines, newBlock)
		}
		reportedLines = append(reportedLines, newBlock)
	}
	return
}

func newReportedBlock(b *cover.ProfileBlock, covered bool) ReportedBlock {
	return ReportedBlock{
		StartLine: b.StartLine,
		StartCol:  b.StartCol,
		StopLine:  b.EndLine,
		StopCol:   b.EndCol,
		Covered:   covered,
	}
}
