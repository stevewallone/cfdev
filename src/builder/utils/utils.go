package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func StemcellVersion(manifest string) (string, error) {
	txt, err := ioutil.ReadFile(manifest)
	if err != nil {
		return "", err
	}
	data := struct {
		Stemcells []struct {
			Version string `yaml:"version"`
		} `yaml:"stemcells"`
	}{}
	if err := yaml.Unmarshal(txt, &data); err != nil {
		return "", err
	}
	if len(data.Stemcells) != 1 {
		panic(fmt.Errorf("manifest (%s) must contain 1 stemcell (not %d)", manifest, len(data.Stemcells)))
		return "", fmt.Errorf("manifest (%s) must contain 1 stemcell (not %d)", manifest, len(data.Stemcells))
	}

	return data.Stemcells[0].Version, nil
}

type Yaml map[interface{}]interface{}

func BoshInt(manifest string, opsfiles []string, values map[string]string) (Yaml, error) {
	args := []string{"int", manifest}
	for _, opsfile := range opsfiles {
		args = append(args, "-o", opsfile)
	}
	for k, v := range values {
		args = append(args, "-v", fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("bosh", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", stderr.String())
		return nil, fmt.Errorf("Run bosh %s: %s", strings.Join(args, " "), err)
	}
	data := Yaml{}
	if err := yaml.Unmarshal(stdout.Bytes(), &data); err != nil {
		return nil, fmt.Errorf("parsing bosh int file: %s", err)
	}
	return data, nil
}

func UploadStemcell(stemcellVersion string) error {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(
		"bosh", "upload-stemcell",
		fmt.Sprintf("https://s3.amazonaws.com/bosh-gce-light-stemcells/light-bosh-stemcell-%s-google-kvm-ubuntu-trusty-go_agent.tgz", stemcellVersion),
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", stderr.String())
		return err
	}
	return nil
}
