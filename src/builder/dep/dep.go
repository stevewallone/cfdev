package dep

import (
	"builder/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type FS interface {
	AddReader(name string, r io.Reader) error
	AddBytes(name string, data []byte) error
	AddURL(name, url string) error
	Exists(name string) bool
}

func DownloadReleases(stemcellVersion string, configs map[string]utils.Yaml, tf FS) error {
	compile := make(map[string]map[interface{}]interface{})
	for _, manifest := range configs {
		if releases, ok := manifest["releases"].([]interface{}); ok {
			for _, release := range releases {
				if r, ok := release.(utils.Yaml); ok {
					if url, ok := r["url"].(string); ok {
						filename := fmt.Sprintf("releases/%v-%v-%s.tgz", r["name"], r["version"], stemcellVersion)
						if tf.Exists(filename) {
							fmt.Println("Skip Download:", filename)
						} else {
							if r["stemcell"] != nil || strings.Contains(url, "-compiled-") {
								fmt.Println("Download:", filename)
								if err := tf.AddURL(filename, url); err != nil {
									return fmt.Errorf("download %s: %s", url, err)
								}
							} else {
								fmt.Println("Compile:", filename)
								compile[filename] = clone(r)
							}
						}
						r["url"] = fmt.Sprintf("file:///var/vcap/cache/%s", filename)
					}
				}
			}
		}
	}
	if len(compile) > 0 {
		if err := uploadReleases(stemcellVersion, compile); err != nil {
			return fmt.Errorf("upload releases: %s", err)
		}
		for filename, release := range compile {
			_, r, err := extractRelease(stemcellVersion, filename, release)
			if err != nil {
				return fmt.Errorf("extract release: %s", err)
			}
			if err := tf.AddReader(filename, r); err != nil {
				return fmt.Errorf("add reader: %s", err)
			}
		}
	}
	return nil
}

func uploadReleases(stemcellVersion string, releases map[string]map[interface{}]interface{}) error {
	manifest, err := yaml.Marshal(map[string]interface{}{
		"instance_groups": []interface{}{},
		"name":            "cf",
		"releases":        mapValues(releases),
		"stemcells": []interface{}{
			map[string]string{
				"alias":   "default",
				"os":      "ubuntu-trusty",
				"version": stemcellVersion,
			},
		},
		"update": map[string]interface{}{
			"canaries":          1,
			"canary_watch_time": "30000-1200000",
			"max_in_flight":     1,
			"update_watch_time": "5000-1200000",
		},
	})
	if err != nil {
		return fmt.Errorf("marhal yaml: %s", err)
	}
	cmd := exec.Command("bosh", "-n", "deploy", "-d", "cf", "-")
	cmd.Stdin = bytes.NewReader(manifest)
	txt, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(txt))
	}
	return err
}

func extractRelease(stemcellVersion, filename string, release map[interface{}]interface{}) (int64, io.ReadCloser, error) {
	fmt.Println("Export Release:", filename)
	tmpDir, err := ioutil.TempDir("", "extract-releases-")
	if err != nil {
		return 0, nil, err
	}
	if txt, err := exec.Command(
		"bosh", "-d", "cf", "export-release",
		fmt.Sprintf("%v/%v", release["name"], release["version"]),
		fmt.Sprintf("ubuntu-trusty/%s", stemcellVersion),
		"--dir", tmpDir,
	).CombinedOutput(); err != nil {
		fmt.Println(string(txt))
		return 0, nil, err
	}
	m, err := filepath.Glob(filepath.Join(tmpDir, "*"))
	if err != nil {
		return 0, nil, err
	}
	if len(m) != 1 {
		return 0, nil, fmt.Errorf("Could not find single file: %v", m)
	}
	fh, err := os.Open(m[0])
	if err != nil {
		return 0, nil, err
	}
	if err := os.RemoveAll(tmpDir); err != nil {
		return 0, nil, err
	}
	fi, err := fh.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), fh, nil
}

func WriteManifests(configs map[string]utils.Yaml, tf FS) error {
	for name, obj := range configs {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}
		if err := tf.AddBytes(name+".yml", data); err != nil {
			return err
		}
	}
	return nil
}

func mapValues(hash map[string]map[interface{}]interface{}) []interface{} {
	out := make([]interface{}, 0, len(hash))
	for _, v := range hash {
		out = append(out, v)
	}
	return out
}

func clone(hash map[interface{}]interface{}) map[interface{}]interface{} {
	hash2 := make(map[interface{}]interface{})
	for k, v := range hash {
		hash2[k] = v
	}
	return hash2
}

func StemcellFilename(stemcellVersion string) string {
	return fmt.Sprintf("bosh-stemcell-%s-warden-boshlite-ubuntu-trusty-go_agent.tgz", stemcellVersion)
}

func DownloadStemcell(stemcellVersion string, cftgz FS) error {
	path := StemcellFilename(stemcellVersion)
	if cftgz.Exists(path) {
		fmt.Println("Skip Download:", path)
		return nil
	} else {
		return cftgz.AddURL(path, "https://s3.amazonaws.com/bosh-core-stemcells/warden/"+path)
	}
}

func StemcellVersion(manifest string) (string, error) {
	txt, err := ioutil.ReadFile(manifest)
	if err != nil {
		return "", fmt.Errorf("read file: %s", err)
	}
	data := struct {
		Stemcells []struct {
			Version string `yaml:"version"`
		} `yaml:"stemcells"`
	}{}
	if err := yaml.Unmarshal(txt, &data); err != nil {
		return "", fmt.Errorf("parse file: %s", err)
	}
	if len(data.Stemcells) != 1 {
		return "", fmt.Errorf("manifest (%s) must contain 1 stemcell (not %d)", manifest, len(data.Stemcells))
	}

	stemcellVersion := data.Stemcells[0].Version

	if isUploaded, err := isStemcellUploaded(stemcellVersion); err != nil {
		return "", fmt.Errorf("is stemcell uploaded: %s", err)
	} else if isUploaded {
		fmt.Println("Stemcell:", stemcellVersion)
	} else {
		fmt.Println("Upload Stemcell:", stemcellVersion)
		if err := utils.UploadStemcell(stemcellVersion); err != nil {
			return "", fmt.Errorf("upload stemcell: %s: %s", stemcellVersion, err)
		}
	}

	return stemcellVersion, nil
}

func isStemcellUploaded(stemcellVersion string) (bool, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bosh", "stemcells", "--json")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", stderr.String())
		return false, err
	}
	data := struct {
		Tables []struct {
			Rows []struct {
				Version string `json:"version"`
			}
		}
	}{}
	if err := json.Unmarshal(stdout.Bytes(), &data); err != nil {
		return false, err
	}
	for _, table := range data.Tables {
		for _, row := range table.Rows {
			if row.Version == stemcellVersion || row.Version == fmt.Sprintf("%s*", stemcellVersion) {
				return true, nil
			}
		}
	}
	return false, nil
}
