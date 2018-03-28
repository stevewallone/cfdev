package launchd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type DaemonSpec struct {
	Label            string
	Program          string
	ProgramArguments []string
	RunAtLoad        bool
}

type Launchd struct {
	PListDir string
}

func (l *Launchd) AddDaemon(spec DaemonSpec, executable string) error {
	if err := l.copyExecutable(executable, spec.Program); err != nil {
		return err
	}
	plistPath := filepath.Join(l.PListDir, spec.Label+".plist")
	if err := l.writePlist(spec, plistPath); err != nil {
		return err
	}
	return l.load(plistPath)
}

func (l *Launchd) RemoveDaemon(spec DaemonSpec) error {
	plistPath := filepath.Join(l.PListDir, spec.Label+".plist")
	if err := l.unload(plistPath); err != nil {
		return err
	}
	if err := os.Remove(plistPath); err != nil {
		return err
	}
	return os.Remove(spec.Program)
}

func (l *Launchd) load(plistPath string) error {
	cmd := exec.Command("launchctl", "load", plistPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) unload(plistPath string) error {
	cmd := exec.Command("launchctl", "unload", plistPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) copyExecutable(src string, dest string) error {
	target, err := os.Create(dest)
	if err != nil {
		return err
	}

	if err = os.Chmod(dest, 0744); err != nil {
		return err
	}

	binData, err := os.Open(src)
	if err != nil {
		return err
	}

	_, err = io.Copy(target, binData)
	return err
}

func (l *Launchd) writePlist(spec DaemonSpec, dest string) error {
	tmplt := template.Must(template.New("plist").Parse(plistTemplate))
	plist, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer plist.Close()
	return tmplt.Execute(plist, spec)
}

var plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.Label}}</string>
  <key>Program</key>
  <string>{{.Program}}</string>
  <key>ProgramArguments</key>
  <array>
{{range .ProgramArguments}}    <string>{{.}}</string>{{"\n"}}{{end}}  </array>
  <key>RunAtLoad</key>
  <{{.RunAtLoad}}/>
</dict>
</plist>
`
