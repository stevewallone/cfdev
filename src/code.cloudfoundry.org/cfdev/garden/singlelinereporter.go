package garden

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

func singleLineReporter(out io.Writer) io.WriteCloser {
	pr, pw := io.Pipe()

	go func() {
		start := time.Now()
		lastLineLength := 0
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			txt := strings.TrimSpace(scanner.Text())
			if txt != "" {
				txt = fmt.Sprintf("  %s: %s", time.Now().Sub(start).Round(time.Second), txt)
				fmt.Fprintf(out, "\033[%dD%s\033[%dD%s", lastLineLength, strings.Repeat(" ", lastLineLength), lastLineLength, txt)
				lastLineLength = len(txt)
			}
		}
		fmt.Fprintf(out, "\033[%dD%s\033[%dD", lastLineLength, strings.Repeat(" ", lastLineLength), lastLineLength)
	}()

	return pw
}
