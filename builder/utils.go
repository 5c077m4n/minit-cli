package builder

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type fetchTarConfig struct {
	url      string
	checksum string
}

var (
	ErrNilConfig         = errors.New("the input is not an object")
	ErrNilConfigURL      = errors.New("the `url` input field was not supplied")
	ErrNilConfigChecksum = errors.New("the `checksum` input field was not supplied")
)

func newFetchTarConfig(tbl *lua.LTable) (*fetchTarConfig, error) {
	if tbl == nil {
		return nil, ErrNilConfig
	}

	url, ok := tbl.RawGetString("url").(lua.LString)
	if !ok {
		return nil, ErrNilConfigURL
	}
	checksum, ok := tbl.RawGetString("checksum").(lua.LString)
	if !ok {
		return nil, ErrNilConfigChecksum
	}

	return &fetchTarConfig{
		url:      url.String(),
		checksum: checksum.String(),
	}, nil
}

func fetchTar(l *lua.LState) int {
	configTbl := l.ToTable(1)
	tarConfig, err := newFetchTarConfig(configTbl)
	if err != nil {
		slog.Error(
			"Fetch tar config error",
			"error", err.Error(),
		)
		return 0
	}

	client := http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Get(tarConfig.url)
	if err != nil {
		slog.Error(
			"Get error",
			"Could not get the requested file", err,
		)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error(
			"Download error",
			"Status code", resp.StatusCode,
		)
		return 0
	}

	downloadFilename := filepath.Join(
		"/tmp",
		"minit",
		base64.URLEncoding.EncodeToString([]byte(tarConfig.url)),
	)

	out, err := os.Create(downloadFilename)
	if err != nil {
		slog.Error(
			"File creation error",
			"Could not create download file on OS", err,
		)
		return 0
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		slog.Error(
			"File creation error",
			"Could not copy content into OS file", err,
		)
		return 0
	}
	if _, err := out.Seek(0, 0); err != nil {
		slog.Error(
			"File rewind error",
			"Could not rewind file", err,
		)
		return 0
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, out); err != nil {
		slog.Error(
			"File to hash copy error",
			"Could not copy file content to hasher", err,
		)
		return 0
	}
	sum := string(hasher.Sum(nil))

	if tarConfig.checksum != sum {
		slog.Error(
			"Bad checksum",
			"got", sum,
			"expected", tarConfig.checksum,
		)
		return 0
	}

	// TODO: unpack tar + return unpacked final path

	return 0
}
