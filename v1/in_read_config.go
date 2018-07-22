package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const DEFAULT_CONFIG_FILE = `screenshot-compare.json`

// NewConfig creates a new configuration struct with default values
func NewConfig() *Config {
	c := new(Config)
	c.ColorSpace = `RGB`
	c.Timeout = 0 * time.Second
	c.PreWait = 0 * time.Second
	c.AdmissibleDiffPixel = 0
	c.NoDimensionError = true
	return c
}

// FromEnv parses environment variables and stores its data in its Config struct
// If mode=1, all values must be set or an error is returned. Depending on the type, it might not be possible to distinguish between 'not set' and 'zero value'.
// If mode=2, values will be stored iff all required values are set. If mode=3, any non-zero value will be stored.
// The return values are warnings (value not set) and errors (value cannot be used/parsed).
// If the second return value is non-nil, Config will not be modified.
func (c *Config) FromEnv(mode int) (error, error) {
	s := os.Getenv(`SCMP_COLORS`)
	t := os.Getenv(`SCMP_TIMEOUT`)
	w := os.Getenv(`SCMP_WAIT`)
	d := os.Getenv(`SCMP_DIFFPIXEL`)
	n := os.Getenv(`SCMP_NODIMERROR`)
	b := os.Getenv(`SCMP_BASEIMG`)
	r := os.Getenv(`SCMP_REFIMG`)

	if s != "" && s != "Y'UV" && s != "RGB" {
		return nil, fmt.Errorf("unknown color space '%s'", s)
	}

	var err error
	var to, wa time.Duration
	if t != "" {
		to, err = parseDurationSpecifier(t)
		if err != nil {
			return nil, err
		}
	}
	if w != "" {
		wa, err = parseDurationSpecifier(w)
		if err != nil {
			return nil, err
		}
	}
	var diffpixel uint
	if d != "" {
		diffu64, err := strconv.ParseUint(d, 10, 32)
		if err != nil {
			return nil, err
		}
		diffpixel = uint(diffu64)
	}
	var nodimerr bool
	if strings.ToLower(n) == `true` || strings.ToLower(n) == `yes` {
		nodimerr = true
	} else if strings.ToLower(n) == `false` || strings.ToLower(n) == `no` || n == `` {
		nodimerr = false
	} else {
		return nil, fmt.Errorf(`invalid value for env variable SCMP_NODIMERROR, expected 'true' or 'false', got '%s'`, n)
	}

	switch mode {
	case 1:
		envs := []string{`SCMP_COLORS`, `SCMP_TIMEOUT`, `SCMP_WAIT`, `SCMP_DIFFPIXEL`, `SCMP_NODIMERROR`, `SCMP_BASEIMG`, `SCMP_REFIMG`}
		for _, env := range envs {
			if os.Getenv(env) == "" {
				return fmt.Errorf(`environment variable %s not set`, env), nil
			}
		}
		c.ColorSpace = s
		c.Timeout = to
		c.PreWait = wa
		c.AdmissibleDiffPixel = diffpixel
		c.NoDimensionError = nodimerr
		if err := c.BaseImg.FromFilepath(b); err != nil {
			return nil, err
		}
		if err := c.RefImg.FromFilepath(r); err != nil {
			return nil, err
		}

	case 2:
		if b == "" {
			return fmt.Errorf(`environment variable SCMP_BASEIMG not set`), nil
		} else if r == "" {
			return fmt.Errorf(`environment variable SCMP_REFIMG not set`), nil
		}
		c.ColorSpace = s
		c.Timeout = to
		c.PreWait = wa
		c.AdmissibleDiffPixel = diffpixel
		c.NoDimensionError = nodimerr
		if err := c.BaseImg.FromFilepath(b); err != nil {
			return nil, err
		}
		if err := c.RefImg.FromFilepath(r); err != nil {
			return nil, err
		}

	case 3:
		if s != "" {
			c.ColorSpace = s
		}
		if t != "" {
			c.Timeout = to
		}
		if w != "" {
			c.PreWait = wa
		}
		if d != "" {
			c.AdmissibleDiffPixel = diffpixel
		}
		if n != "" {
			c.NoDimensionError = nodimerr
		}
		if b != "" {
			if err := c.BaseImg.FromFilepath(b); err != nil {
				return nil, err
			}
		}
		if r != "" {
			if err := c.RefImg.FromFilepath(r); err != nil {
				return nil, err
			}
		}

	default:
		return nil, fmt.Errorf(`mode must be one of 1, 2, and 3; got '%d'`, mode)
	}

	return nil, nil
}

// FromArgs parses the given arguments and stores its data in its Config struct.
// If mode=1, all values must be set or an error is returned. Depending on the type, it might not be possible to distinguish between 'not set' and 'zero value'.
// If mode=2, values will be stored iff all required values are set. If mode=3, any non-zero value will be stored.
// The return values are warnings (value not set) and errors (value cannot be used/parsed).
// If the second return value is non-nil, Config will not be modified.
func (c *Config) FromArgs(args []string, usage string, mode int) (error, error) {
	var err error
	terminate := func(int) {
		err = fmt.Errorf(`invalid CLI call`)
	}

	// kingpin calls
	cli := kingpin.New(filepath.Base(args[0]), usage)
	colorSpace := cli.Flag("colors", `color space, one of "Y'UV" and "RGB"`).Default("RGB").Short('c').String()
	timeout := cli.Flag("timeout", `maximum time comparison is allowed to take, 0s is infinite, e.g. '1s'`).Default("0s").Short('t').Duration()
	preWait := cli.Flag("wait", `duration to wait before comparison starts, e.g. '200ms'`).Default("0s").Short('w').Duration()
	admissibleDiffPixel := cli.Flag("diffpixel", `fixed number of pixels with difference to ignore`).Short('d').Uint()
	nodimerror := cli.Flag("nodimerror", `if true, max diff will be returned if dimensions don't match instead of error`).Short('n').Bool()
	baseImg := cli.Arg("baseimg", `filepath to image to compare`).Required().String()
	refImg := cli.Arg("refimg", `filepath to image to compare with`).Required().String()

	cli.Version("1.2.0")
	cli.Terminate(terminate)
	_, err2 := cli.Parse(args[1:])
	if err2 != nil {
		return nil, err2
	}
	if err != nil {
		return nil, err
	}

	// no errors returned by kingpin, use the values
	if *colorSpace != "" && *colorSpace != "Y'UV" && *colorSpace != "RGB" {
		return nil, fmt.Errorf("unknown color space '%s'", *colorSpace)
	}

	switch mode {
	case 1:
		if *colorSpace == "" {
			return fmt.Errorf(`missing CLI argument --colors`), nil
		}
		if *baseImg == "" {
			return fmt.Errorf(`missing CLI argument --baseimg`), nil
		}
		if *refImg == "" {
			return fmt.Errorf(`missing CLI argument --refimg`), nil
		}
		c.ColorSpace = *colorSpace
		c.Timeout = *timeout
		c.PreWait = *preWait
		c.AdmissibleDiffPixel = *admissibleDiffPixel
		c.NoDimensionError = *nodimerror
		if err := c.BaseImg.FromFilepath(*baseImg); err != nil {
			return nil, err
		}
		if err := c.RefImg.FromFilepath(*refImg); err != nil {
			return nil, err
		}

	case 2:
		if *baseImg == "" {
			return fmt.Errorf(`CLI argument --baseimg not set`), nil
		} else if *refImg == "" {
			return fmt.Errorf(`CLI argument --refimg not set`), nil
		}
		c.ColorSpace = *colorSpace
		c.Timeout = *timeout
		c.PreWait = *preWait
		c.AdmissibleDiffPixel = *admissibleDiffPixel
		c.NoDimensionError = *nodimerror
		if err := c.BaseImg.FromFilepath(*baseImg); err != nil {
			return nil, err
		}
		if err := c.RefImg.FromFilepath(*refImg); err != nil {
			return nil, err
		}

	case 3:
		if *colorSpace != "" {
			c.ColorSpace = *colorSpace
		}
		if *timeout != 0 {
			c.Timeout = *timeout
		}
		if *preWait != 0 {
			c.PreWait = *preWait
		}
		if *admissibleDiffPixel != 0 {
			c.AdmissibleDiffPixel = *admissibleDiffPixel
		}
		if *nodimerror != false {
			c.NoDimensionError = *nodimerror
		}
		if *baseImg != "" {
			if err := c.BaseImg.FromFilepath(*baseImg); err != nil {
				return nil, err
			}
		}
		if *refImg != "" {
			if err := c.RefImg.FromFilepath(*refImg); err != nil {
				return nil, err
			}
		}

	default:
		return nil, fmt.Errorf(`mode must be one of 1, 2, and 3; got '%d'`, mode)
	}

	return nil, nil
}

// FromJSON retrieves the configuration parameters from a JSON file and stores its data in its Config struct.
// If `filepath` is empty, the default filepath will be used.
// If `silentMissingError` is true, FromJSON does not modify Config and returns nil if the JSON file does not exist.
// If mode=1, all values must be set or an error is returned. Depending on the type, it might not be possible to distinguish between 'not set' and 'zero value'.
// If mode=2, values will be stored iff all required values are set. If mode=3, any non-zero value will be stored.
// The return values are warnings (value not set) and errors (value cannot be used/parsed).
// If the second return value is non-nil, Config will not be modified.
func (c *Config) FromJSON(filepath string, silentMissingError bool, mode int) (error, error) {
	if filepath == "" {
		filepath = DEFAULT_CONFIG_FILE
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if silentMissingError {
			return nil, nil
		} else {
			return fmt.Errorf("configuration file '%s' does not exist", filepath), nil
		}
	}

	// json struct
	type jsonConfig struct {
		Colors     string `json:"colors,omitempty"`
		Timeout    string `json:"timeout,omitempty"`
		PreWait    string `json:"wait,omitempty"`
		DiffPixel  uint   `json:"diffpixel,omitempty"`
		NoDimError bool   `json:"nodimerror,omitempty"`
		BaseImg    string `json:"baseimg,omitempty"`
		RefImg     string `json:"refimg,omitempty"`
	}
	var jsonConf jsonConfig
	jBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jBytes, &jsonConf)
	if err != nil {
		return nil, err
	}

	var to, wa time.Duration
	if jsonConf.Timeout != "" {
		to, err = parseDurationSpecifier(jsonConf.Timeout)
		if err != nil {
			return nil, err
		}
	}
	if jsonConf.PreWait != "" {
		wa, err = parseDurationSpecifier(jsonConf.PreWait)
		if err != nil {
			return nil, err
		}
	}
	if jsonConf.Colors != "" && jsonConf.Colors != "Y'UV" && jsonConf.Colors != "RGB" {
		return nil, fmt.Errorf("unknown color space '%s'", jsonConf.Colors)
	}

	switch mode {
	case 1:
		if jsonConf.Colors == "" {
			return fmt.Errorf(`missing JSON parameter colors`), nil
		}
		if jsonConf.BaseImg == "" || jsonConf.RefImg == "" {
			return fmt.Errorf(`missing JSON parameter baseimg or refimg`), nil
		}
		c.ColorSpace = jsonConf.Colors
		c.Timeout = to
		c.PreWait = wa
		c.AdmissibleDiffPixel = jsonConf.DiffPixel
		c.NoDimensionError = jsonConf.NoDimError
		if err := c.BaseImg.FromFilepath(jsonConf.BaseImg); err != nil {
			return nil, err
		}
		if err := c.RefImg.FromFilepath(jsonConf.RefImg); err != nil {
			return nil, err
		}

	case 2:
		if jsonConf.BaseImg == "" {
			return fmt.Errorf(`missing JSON parameter baseimg`), nil
		} else if jsonConf.RefImg == "" {
			return fmt.Errorf(`missing JSON parameter refimg`), nil
		}
		c.ColorSpace = jsonConf.Colors
		c.Timeout = to
		c.PreWait = wa
		c.AdmissibleDiffPixel = jsonConf.DiffPixel
		c.NoDimensionError = jsonConf.NoDimError
		if err := c.BaseImg.FromFilepath(jsonConf.BaseImg); err != nil {
			return nil, err
		}
		if err := c.RefImg.FromFilepath(jsonConf.RefImg); err != nil {
			return nil, err
		}

	case 3:
		if jsonConf.Colors != "" {
			c.ColorSpace = jsonConf.Colors
		}
		if jsonConf.Timeout != "" {
			c.Timeout = to
		}
		if jsonConf.PreWait != "" {
			c.PreWait = wa
		}
		if jsonConf.DiffPixel != 0 {
			c.AdmissibleDiffPixel = jsonConf.DiffPixel
		}
		if jsonConf.NoDimError {
			c.NoDimensionError = jsonConf.NoDimError
		}
		if jsonConf.BaseImg != "" {
			if err := c.BaseImg.FromFilepath(jsonConf.BaseImg); err != nil {
				return nil, err
			}
		}
		if jsonConf.RefImg != "" {
			if err := c.RefImg.FromFilepath(jsonConf.RefImg); err != nil {
				return nil, err
			}
		}

	default:
		return nil, fmt.Errorf(`mode must be one of 1, 2, and 3; got '%d'`, mode)
	}

	return nil, nil
}

// parseDurationSpecifier takes a human-readable duration specifier
// like '12s' and returns `time.Second * 12`
func parseDurationSpecifier(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	errmsg := "invalid duration specifier; expected integer and one of 'ismh'; got '%s'"

	if s == "" {
		return time.Duration(0), fmt.Errorf(errmsg, s)
	}

	end := len(s) - 1
	if '0' <= s[end] && s[end] <= '9' {
		end++
	}

	val, err := strconv.Atoi(s[:end])
	if err != nil {
		return time.Duration(0), fmt.Errorf(errmsg, s)
	}

	if '0' <= s[len(s)-1] && s[len(s)-1] <= '9' {
		return time.Duration(val) * time.Second, nil
	}

	switch s[len(s)-1] {
	case 'i':
		return time.Duration(val) * time.Millisecond, nil
	case 's':
		return time.Duration(val) * time.Second, nil
	case 'm':
		return time.Duration(val) * time.Minute, nil
	case 'h':
		return time.Duration(val) * time.Hour, nil
	}

	return time.Duration(0), fmt.Errorf(errmsg, s)
}
