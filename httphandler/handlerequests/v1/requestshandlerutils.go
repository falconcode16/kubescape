package v1

import (
	"fmt"
	"os"
	"path/filepath"

	pkgcautils "github.com/armosec/utils-go/utils"

	"github.com/armosec/kubescape/core/cautils"
	"github.com/armosec/kubescape/core/cautils/getter"
	"github.com/armosec/kubescape/core/core"
)

func scan(scanRequest *PostScanRequest, scanID string) ([]byte, error) {
	scanInfo := getScanCommand(scanRequest, scanID)

	ks := core.NewKubescape()
	result, err := ks.Scan(scanInfo)
	if err != nil {
		f, e := os.Open(filepath.Join(FailedOutputDir, scanID))
		if e != nil {
			return []byte{}, fmt.Errorf("failed to scan. reason: '%s'. failed to save error in file. reason: %s", err.Error(), e.Error())
		}
		defer f.Close()
		f.Write([]byte(e.Error()))

	}
	result.HandleResults()
	b, err := result.ToJson()
	if err != nil {
		err = fmt.Errorf("failed to parse results to json, reason: %s", err.Error())
	}
	return b, err
}

func readResultsFile(fileID string) ([]byte, error) {
	if fileName := searchFile(fileID); fileName != "" {
		return os.ReadFile(fileName)
	}
	return nil, fmt.Errorf("file %s not found", fileID)
}

func removeResultDirs() {
	os.ReadDir(OutputDir)
	os.ReadDir(FailedOutputDir)
}
func removeResultsFile(fileID string) error {
	if fileName := searchFile(fileID); fileName != "" {
		return os.Remove(fileName)
	}
	return nil // no files found to delete
}
func searchFile(fileID string) string {
	if fileName, _ := findFile(OutputDir, fileID); fileName != "" {
		return fileName
	}
	if fileName, _ := findFile(FailedOutputDir, fileID); fileName != "" {
		return fileName
	}
	return ""
}

func findFile(targetDir string, fileName string) (string, error) {

	matches, err := filepath.Glob(filepath.Join(targetDir, fileName))
	if err != nil {
		return "", err
	}

	if len(matches) != 0 {
		return matches[0], nil
	}
	return "", nil
}

func getScanCommand(scanRequest *PostScanRequest, scanID string) *cautils.ScanInfo {

	scanInfo := scanRequest.ToScanInfo()
	scanInfo.ScanID = scanID

	// *** start ***
	// Set default format
	if scanInfo.Format == "" {
		scanInfo.Format = "json"
	}
	scanInfo.FormatVersion = "v2" // latest version
	// *** end ***

	// *** start ***
	// DO NOT CHANGE
	scanInfo.Output = filepath.Join(OutputDir, scanID)
	// *** end ***

	return scanInfo
}

func defaultScanInfo() *cautils.ScanInfo {
	scanInfo := &cautils.ScanInfo{}
	scanInfo.FailThreshold = 100
	scanInfo.Account = envToString("KS_ACCOUNT", "")                              // publish results to Kubescape SaaS
	scanInfo.ExcludedNamespaces = envToString("KS_EXCLUDE_NAMESPACES", "")        // namespace to exclude
	scanInfo.IncludeNamespaces = envToString("KS_INCLUDE_NAMESPACES", "")         // namespace to include
	scanInfo.FormatVersion = envToString("KS_FORMAT_VERSION", "v2")               // output format version
	scanInfo.Format = envToString("KS_FORMAT", "json")                            // default output should be json
	scanInfo.Submit = envToBool("KS_SUBMIT", false)                               // publish results to Kubescape SaaS
	scanInfo.HostSensorEnabled.SetBool(envToBool("KS_ENABLE_HOST_SCANNER", true)) // enable host scanner
	scanInfo.Local = envToBool("KS_KEEP_LOCAL", false)                            // do not publish results to Kubescape SaaS
	if !envToBool("KS_DOWNLOAD_ARTIFACTS", false) {
		scanInfo.UseArtifactsFrom = getter.DefaultLocalStore // Load files from cache (this will prevent kubescape fom downloading the artifacts every time)
	}
	return scanInfo
}

func envToBool(env string, defaultValue bool) bool {
	if d, ok := os.LookupEnv(env); ok {
		return pkgcautils.StringToBool(d)
	}
	return defaultValue
}

func envToString(env string, defaultValue string) string {
	if d, ok := os.LookupEnv(env); ok {
		return d
	}
	return defaultValue
}
