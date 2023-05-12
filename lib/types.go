package lib

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/cover"
)

// AppConfig is the application configuration
type AppConfig struct {
	// The output format to use
	Format string `json:"format" yaml:"format" xml:"format"`
	// Whether to build a full report or summary
	Level string `json:"level" yaml:"level" xml:"level"`
	// If the output is a badge, this is the color to use
	Color string `json:"color" yaml:"color" xml:"color"`
	// The file or directory path to write the output to
	Output string `json:"output" yaml:"output" xml:"output"`
	// The input coverage file from `go test -coverprofile`
	Input []string `json:"input" yaml:"input" xml:"input"`
	// The source code folder location on disk
	SourceDir string `json:"source" yaml:"source" xml:"source"`
	// The display name of the package
	PackageName string `json:"packageName" yaml:"packageName" xml:"packageName"`
}

// The basic meta data for the report
type ReportMeta struct {
	// The display name of the package
	PackageName string `json:"packageName" yaml:"packageName" xml:"packageName"`
	// The common root directory path for all reported files
	CommonRoot string `json:"commonRoot" yaml:"commonRoot" xml:"commonRoot"`
	// The parent of the CommonRoot directory path
	ParentRoot string `json:"parentRoot" yaml:"parentRoot" xml:"parentRoot"`
}

type ReportContainer interface {
	// AddFolder adds a ReportedFolder to the container
	AddFolder(rf *ReportedFolder)
	// ContainsFolder returns true, and a pointer to ReportedFolder if the container contains a ReportedFolder with the given path
	ContainsFolder(folderPath string) (*ReportedFolder, bool)
	// AddFile adds a ReportedFile to the container
	AddFile(rf *ReportedFile)
	// ContainsFile returns true, and a pointer to ReportedFile if the container contains a ReportedFile with the given path
	ContainsFile(filePath string) (*ReportedFile, bool)
	// Updates the roll-up coverage value for the container and all children
	UpdateCoverage()
}

// The top-level context for coverage reporting
type ReportContext struct {
	// The application configuration
	Config AppConfig `json:"config" yaml:"config" xml:"config"`
	// The associated metadata for the report
	Meta ReportMeta `json:"meta" yaml:"meta" xml:"meta"`
	// All reported files from the coverage report
	ReportedFiles []*ReportedFile `json:"reportedFiles" yaml:"reportedFiles" xml:"reportedFiles"`
	// ReportedFolders is an array of ReportedFolder entries, each containing at least one reported file
	ReportedFolders []*ReportedFolder `json:"reportedFolders" yaml:"reportedFolders" xml:"reportedFolders"`
	// Whether this is a full report or summary
	IsFullReport bool `json:"isFullReport" yaml:"isFullReport" xml:"isFullReport"`
	// The file or directory path where the output will be written
	Output string `json:"output" yaml:"output" xml:"output"`
	// The coverage percentage for the entire report
	CoveredPct float64 `json:"coveredPct" yaml:"coveredPct" xml:"coveredPct"`
}

// Creates a new ReportContext
func NewReportContext(config AppConfig, meta ReportMeta, isFullRpt bool) ReportContext {
	absOutPath, err := filepath.Abs(config.Output)
	if err != nil {
		HandleStopError(UnresolvablePathError(config.Output))
	}

	return ReportContext{
		Config:          config,
		Meta:            meta,
		ReportedFiles:   make([]*ReportedFile, 0),
		ReportedFolders: make([]*ReportedFolder, 0),
		IsFullReport:    isFullRpt,
		Output:          absOutPath,
	}
}

// GetPseudoFolder returns a ReportedFolder that represents the root folder of the source code
func (rc *ReportContext) GetPseudoFolder() *ReportedFolder {
	noFiles := make([]*ReportedFile, 0)
	pseudoFolder := NewReportedFolder(rc, rc.Config.SourceDir, noFiles...)
	pseudoFolder.ReportedFolders = rc.ReportedFolders
	pseudoFolder.CoveredPct = rc.CoveredPct

	return &pseudoFolder
}

// AddProfile add a cover.Profile to the context.ReportedFiles as a ReportedFile
func (rc *ReportContext) AddProfile(profile *cover.Profile) {
	reportedFile := NewReportedFile(rc, profile)
	folderPath := path.Dir(reportedFile.SourceFile)
	rc.AddFolderFile(folderPath, &reportedFile)
}

// ContainsFolder returns true if the context.ReportedFolders contains a folder with the given path
func (rc *ReportContext) ContainsFolder(folderPath string) (*ReportedFolder, bool) {
	for i, folder := range rc.ReportedFolders {
		if strings.EqualFold(folder.FolderPath, folderPath) {
			return rc.ReportedFolders[i], true
		}
	}
	return nil, false
}

// ContainsFile returns true if the context.ReportedFiles contains a file with the given path
func (rc *ReportContext) ContainsFile(filePath string) (*ReportedFile, bool) {
	for i, file := range rc.ReportedFiles {
		if strings.EqualFold(file.SourceFile, filePath) {
			return rc.ReportedFiles[i], true
		}
	}
	return nil, false
}

// AddFolder adds a ReportedFolder to the context.ReportedFolders, creating the folder if it doesn't already exist.
func (rc *ReportContext) AddFolder(folder *ReportedFolder) {
	if _, exists := rc.ContainsFolder(folder.FolderPath); !exists {
		rc.ReportedFolders = append(rc.ReportedFolders, folder)
		sort.Slice(rc.ReportedFolders, func(i, j int) bool {
			return rc.ReportedFolders[i].FolderPath < rc.ReportedFolders[j].FolderPath
		})
	}
}

// AddFile adds a ReportedFile to the context.ReportedFiles, creating the folder if it doesn't already exist.
func (rc *ReportContext) AddFile(file *ReportedFile) {
	if _, exists := rc.ContainsFile(file.SourceFile); !exists {
		rc.ReportedFiles = append(rc.ReportedFiles, file)
	}
}

// AddFolderFile adds a ReportedFile to the context.ReportedFolders, creating the folder if it doesn't already exist.
func (rc *ReportContext) AddFolderFile(folderPath string, file *ReportedFile) {
	var node ReportContainer = rc
	relDirs := strings.Split(path.Dir(file.SourceFile)[len(rc.Config.SourceDir):], string(os.PathSeparator))[1:]
	for i := range relDirs {
		folderPath := path.Join(rc.Config.SourceDir, strings.Join(relDirs[:i+1], string(os.PathSeparator)))
		existing, exists := node.ContainsFolder(folderPath)
		if !exists {
			files := make([]*ReportedFile, 0)
			newFolder := NewReportedFolder(rc, folderPath, files...)
			node.AddFolder(&newFolder)
			node = &newFolder
		} else {
			node = existing
		}
		if i == len(relDirs)-1 {
			node.AddFile(file)
		}
	}
	rc.AddFile(file)
}

func (rc *ReportContext) GetAllFolders() []*ReportedFolder {
	folders := make([]*ReportedFolder, 0)
	for _, folder := range rc.ReportedFolders {
		folders = append(folders, folder)
		folders = append(folders, folder.GetAllFolders()...)
	}
	return folders
}

// UpdateCoverage updates the coverage percentage for each folder in the context.ReportedFolders based on the covered files.
func (rc *ReportContext) UpdateCoverage() {
	blocks := make([]cover.ProfileBlock, 0)
	for _, folder := range rc.ReportedFolders {
		folder.UpdateCoverage()
		blocks = append(blocks, folder.GetProfileBlocks()...)
	}
	rc.CoveredPct = GetCoveredPct(blocks, true)
}

// A ReportedFolder is a meta-level representation of a folder of ReportedFile entries
type ReportedFolder struct {
	// The associated metadata for the report
	Meta ReportMeta `json:"meta" yaml:"meta" xml:"meta"`
	// The reported files for this folder
	ReportedFiles []*ReportedFile `json:"sourceFiles" yaml:"sourceFiles" xml:"sourceFiles"`
	// The reported subfolder for this folder
	ReportedFolders []*ReportedFolder `json:"reportedFolders" yaml:"reportedFolders" xml:"reportedFolders"`
	// The resolved folder path for this folder
	FolderPath string `json:"folderPath" yaml:"folderPath" xml:"folderPath"`
	// The name for this folder
	FolderName string `json:"folderName" yaml:"folderName" xml:"folderName"`
	// The path parts for this folder
	PathParts []PathTuple `json:"pathParts" yaml:"pathParts" xml:"pathParts"`
	// The display path for this folder
	DisplayPath string `json:"displayPath" yaml:"displayPath" xml:"displayPath"`
	// The output file path for this folder
	OutFilePath string `json:"outFilePath" yaml:"outFilePath" xml:"outFilePath"`
	// The relative path to the assets folder from this folder
	AssetsPath string `json:"assetsPath" yaml:"assetsPath" xml:"assetsPath"`
	// The roll-up percentage of coverage for the files in this folder
	CoveredPct float64 `json:"coveredPct" yaml:"coveredPct" xml:"coveredPct"`
}

func NewReportedFolder(context *ReportContext, folderPath string, files ...*ReportedFile) ReportedFolder {
	folderFile := path.Join(folderPath, "index.html")
	absFolderPath, _ := filepath.Abs(folderFile)
	pathParts := GetRelPathParts(context.Meta.CommonRoot, absFolderPath)
	dispPath, outFilePath := GetOutPathInfo(context.Config.Output, folderFile, ".temp", context.Meta.CommonRoot)
	return ReportedFolder{
		Meta:          context.Meta,
		ReportedFiles: files,
		FolderPath:    folderPath,
		FolderName:    path.Base(folderPath),
		PathParts:     pathParts[:len(pathParts)-1],
		DisplayPath:   path.Dir(dispPath)[1:],
		OutFilePath:   outFilePath,
		AssetsPath:    GetRelRootPath(outFilePath, context.Config.Output),
		CoveredPct:    0,
	}
}

// ContainsFile returns true if the folder contains a file with the given path
func (rf *ReportedFolder) ContainsFile(filePath string) (*ReportedFile, bool) {
	for i, file := range rf.ReportedFiles {
		if file == nil {
			fmt.Printf("Found a nil file at index %d in %s while checking for %s\n", i, rf.FolderPath, filePath)
			continue
		}
		if strings.EqualFold(file.SourceFile, filePath) {
			return rf.ReportedFiles[i], true
		}
	}
	return nil, false
}

// AddFile adds a ReportedFile to the folder, if it doesn't already exist.
func (rf *ReportedFolder) AddFile(file *ReportedFile) {
	if _, exists := rf.ContainsFile(file.SourceFile); !exists {
		rf.ReportedFiles = append(rf.ReportedFiles, file)
	}
}

// ContainsFolder returns true, and a pointer to the ReportedFolder if the folder contains a folder with the given path
func (rf *ReportedFolder) ContainsFolder(folderPath string) (*ReportedFolder, bool) {
	for i, folder := range rf.ReportedFolders {
		if strings.EqualFold(folder.FolderPath, folderPath) {
			return rf.ReportedFolders[i], true
		}
	}
	return nil, false
}

// AddFolder adds a ReportedFolder to the folder, if it doesn't already exist.
func (rf *ReportedFolder) AddFolder(folder *ReportedFolder) {
	if _, exists := rf.ContainsFolder(folder.FolderPath); !exists {
		rf.ReportedFolders = append(rf.ReportedFolders, folder)
	}
}

func (rc *ReportedFolder) GetAllFolders() []*ReportedFolder {
	folders := make([]*ReportedFolder, 0)
	for _, folder := range rc.ReportedFolders {
		folders = append(folders, folder)
		folders = append(folders, folder.GetAllFolders()...)
	}
	return folders
}

// WithExtension gets the output file path for the folder with the specified extension.
func (rf *ReportedFolder) WithExtension(ext string) string {
	return SwapFileExt(rf.OutFilePath, ext)
}

// UpdateCoverage updates the coverage percentage for the folder based on the covered files.
func (rf *ReportedFolder) UpdateCoverage() {
	for _, folder := range rf.ReportedFolders {
		folder.UpdateCoverage()
	}
	rf.CoveredPct = GetCoveredPct(rf.GetProfileBlocks(), true)
}

// GetProfileBlocks gets the covered blocks for the folder, and all sub-folders.
func (rf *ReportedFolder) GetProfileBlocks() []cover.ProfileBlock {
	blocks := make([]cover.ProfileBlock, 0)
	for _, file := range rf.ReportedFiles {
		blocks = append(blocks, file.Profile.Blocks...)
	}
	for _, folder := range rf.ReportedFolders {
		blocks = append(blocks, folder.GetProfileBlocks()...)
	}
	return blocks
}

type ReportedBlock struct {
	// The start line number for this block
	StartLine int `json:"start"`
	// The start column number for this block
	StartCol int `json:"startCol"`
	// The stop line number for this block
	StopLine int `json:"stop"`
	// The stop column number for this block
	StopCol int `json:"stopCol"`
	// Whether or not this block is covered
	Covered bool `json:"covered"`
}

// PathTuple is a tuple of a displayable name and a navigable path
type PathTuple struct {
	// The displayable name for this path
	Name string `json:"name"`
	// The navigable path for this path
	Path string `json:"path"`
}

// ReportedFile is a wrapper around cover.Profile that includes additional metadata
type ReportedFile struct {
	// The associated metadata for the report
	Meta ReportMeta `json:"meta" yaml:"meta" xml:"meta"`
	// The resolved source file for this coverage profile
	SourceFile string `json:"sourceFile" yaml:"sourceFile" xml:"sourceFile"`
	// The reported lines, covered an uncovered for this file
	ReportedLines []ReportedBlock `json:"reportedLines" yaml:"reportedLines" xml:"reportedLines"`
	// The reported lines that are covered for this file
	CoveredLines []ReportedBlock `json:"coveredLines" yaml:"coveredLines" xml:"coveredLines"`
	// A map of the path parts for this file, the key is displayable and the value is navigable
	PathParts []PathTuple `json:"pathParts" yaml:"pathParts" xml:"pathParts"`
	// The display path for this file
	DisplayPath string `json:"displayPath" yaml:"displayPath" xml:"displayPath"`
	// The output file path for this file with a '.out' extension
	OutFilePath string `json:"outFilePath" yaml:"outFilePath" xml:"outFilePath"`
	// The file name for this file
	FileName string `json:"fileName" yaml:"fileName" xml:"fileName"`
	// The source code for this file
	SourceCode string `json:"sourceCode" yaml:"sourceCode" xml:"sourceCode"`
	// Whether or not the source has been read from disk
	isSourceRead bool `json:"-" yaml:"-" xml:"-"`
	// The relative path to the assets folder from this file
	AssetsPath string `json:"assetsPath" yaml:"assetsPath" xml:"assetsPath"`
	// The percentage of coverage for this file
	CoveredPct float64 `json:"coveredPct" yaml:"coveredPct" xml:"coveredPct"`
	// The coverage profile for this file
	Profile *cover.Profile `json:"-" yaml:"-" xml:"-"`
}

func NewReportedFile(context *ReportContext, profile *cover.Profile) ReportedFile {
	meta := context.Meta
	config := context.Config
	sourcePath, err := GetSourceFilePath(config.SourceDir, profile.FileName)
	if err != nil {
		HandleStopError(err)
	}
	dispPath, outFilePath := GetOutPathInfo(config.Output, sourcePath, ".temp", meta.CommonRoot)
	reportedLines, coveredLines := GetProfiledLines(profile)
	return ReportedFile{
		AssetsPath:    GetRelRootPath(outFilePath, config.Output),
		CoveredLines:  coveredLines,
		DisplayPath:   dispPath,
		FileName:      path.Base(profile.FileName),
		Meta:          meta,
		OutFilePath:   outFilePath,
		PathParts:     GetRelPathParts(meta.CommonRoot, sourcePath),
		Profile:       profile,
		ReportedLines: reportedLines,
		SourceFile:    sourcePath,
		CoveredPct:    GetCoveredPct(profile.Blocks, true),
	}
}

// WithExtension gets the output file path for the file with the specified extension.
func (rf *ReportedFile) WithExtension(ext string) string {
	return SwapFileExt(rf.OutFilePath, ext)
}

// GetCoveredPct returns the percentage of statements covered for all blocks in this file.
func (rf *ReportedFile) GetCoveredPct(multiplied bool) (result float64) {
	return GetCoveredPct(rf.Profile.Blocks, multiplied)
}

// GetSourceCode returns the source code for this file.
func (rf *ReportedFile) GetSourceCode() (result string, err error) {
	if !rf.isSourceRead {
		rf.SourceCode, err = GetSourceCode(rf.SourceFile)
		if err != nil {
			return "", err
		}
	}
	return rf.SourceCode, nil
}
