package scrape

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

const fetchScript = `
var system = require('system'),
    page = require("webpage").create();

// Workaround for https://github.com/ariya/phantomjs/issues/12697 since
// it doesn't seem like there will be another 1.9.x release fixing this
var phantomExit = function(exitCode) {
    page.close();
    setTimeout(function() { phantom.exit(exitCode); }, 0);
};

if( system.args.length !== 2 ) {
    system.stderr.writeLine("Usage: fetch.js URL");
    phantomExit(1);
}

var resourceWait  = 300,
    maxRenderWait = 10000,
    url           = system.args[1],
    count         = 0,
    forcedRenderTimeout,
    renderTimeout;

var doRender = function() {
    var c = page.evaluate(function() {
        return document.documentElement.outerHTML;
    });

    system.stdout.write(JSON.stringify({contents: c}));
    phantomExit();
}

page.onResourceRequested = function (req) {
    count += 1;
    system.stderr.writeLine('> ' + req.id + ' - ' + req.url);
    clearTimeout(renderTimeout);
};

page.onResourceReceived = function (res) {
    if (!res.stage || res.stage === 'end') {
        count -= 1;
        system.stderr.writeLine(res.id + ' ' + res.status + ' - ' + res.url);
        if (count === 0) {
            renderTimeout = setTimeout(doRender, resourceWait);
        }
    }
};

page.open(url, function (status) {
    if (status !== "success") {
        system.stderr.writeLine('Unable to load url');
        phantomExit(1);
    } else {
        forcedRenderTimeout = setTimeout(function () {
            console.log(count);
            doRender();
        }, maxRenderWait);
    }
});
`

var (
	// PhantomJS was not found on the system.  You should consider passing an
	// explicit path to NewPhantomJSFetcher().
	ErrNoPhantomJS = errors.New("PhantomJS was not found")

	// This error is returned when we try to use PhantomJS to perform a non-GET
	// request.
	ErrInvalidMethod = errors.New("invalid method")
)

func findPhantomJS() string {
	var path string
	var err error

	for _, nm := range []string{"phantomjs", "phantom"} {
		path, err = exec.LookPath(nm)
		if err == nil {
			return path
		}
	}

	return ""
}

// HasPhantomJS returns whether we can find a PhantomJS installation on this system.
// If this returns "false", creating a PhantomJSFetcher will fail.
func HasPhantomJS() bool {
	return findPhantomJS() != ""
}

// PhantomJSFetcher is a Fetcher that calls out to PhantomJS
// (http://phantomjs.org/) in order to fetch a page's content.  Since PhantomJS
// will evaluate Javascript in a page, this is the recommended Fetcher to use
// for Javascript-heavy pages.
type PhantomJSFetcher struct {
	binaryPath string
	tempDir    string
	scriptPath string

	// Arguments to pass to PhantomJS
	args []string
}

// NewPhantomJSFetcher will create a new instance of PhantomJSFetcher,
// searching the system's PATH for the appropriate binary.  If PhantomJS is not
// in the PATH, or you would like to use an alternate binary, then you can give
// an overridden path.
func NewPhantomJSFetcher(binary ...string) (*PhantomJSFetcher, error) {
	var path string

	// Find the PhantomJS binary
	if len(binary) == 0 || len(binary) == 1 && binary[0] == "" {
		path = findPhantomJS()
	} else if len(binary) == 1 {
		path = binary[0]
	} else {
		return nil, errors.New("invalid number of arguments")
	}

	if path == "" {
		return nil, ErrNoPhantomJS
	}

	// Create a temporary directory
	tdir, err := ioutil.TempDir("", "goscrape-phantom-")
	if err != nil {
		return nil, err
	}

	// Write our fetching script there (so it can be called)
	spath := filepath.Join(tdir, "fetch.js")
	err = ioutil.WriteFile(spath, []byte(fetchScript), 0600)
	if err != nil {
		return nil, err
	}

	ret := &PhantomJSFetcher{
		binaryPath: path,
		tempDir:    tdir,
		scriptPath: spath,
	}
	return ret, nil
}

func (pf *PhantomJSFetcher) Prepare() error {
	// TODO: configure ssl errors / web security
	// TODO: cookies file path might break if spaces
	pf.args = []string{
		"--ignore-ssl-errors=true",
		"--web-security=false",
		"--cookies-file=" + filepath.Join(pf.tempDir, "cookies.dat"),
		pf.scriptPath,
	}
	return nil
}

func (pf *PhantomJSFetcher) Fetch(method, url string) (io.ReadCloser, error) {
	if method != "GET" {
		return nil, ErrInvalidMethod
	}

	// Call the fetch script with these parameters.
	cmd := exec.Command(pf.binaryPath, append(pf.args, url)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// Load the resulting JSON.
	results := map[string]interface{}{}
	err = json.NewDecoder(&stdout).Decode(&results)
	if err != nil {
		return nil, err
	}

	// Return the contents
	contents, ok := results["contents"].(string)
	if !ok {
		return nil, fmt.Errorf("unknown type for 'contents': %T", results["contents"])
	}

	return newStringReadCloser(contents), nil
}

func (pf *PhantomJSFetcher) Close() {
	return
}

// Static type assertion
var _ Fetcher = &PhantomJSFetcher{}
