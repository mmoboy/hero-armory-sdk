// Binary pack creates and validates deterministic Hero Armory .gtpack
// client plugin bundles from manifest.json, compiled client.wasm, and assets.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sdk "github.com/mmoboy/hero-armory-sdk"
)

func main() {
	srcDir := flag.String("dir", ".", "Source directory containing manifest.json, client.wasm, and optional assets/")
	outPath := flag.String("out", "client_plugin.gtpack", "Output .gtpack zip path")
	flag.Parse()

	if err := packDirectory(*srcDir, *outPath); err != nil {
		fmt.Fprintf(os.Stderr, "pack failed: %v\n", err)
		os.Exit(1)
	}
}

func packDirectory(srcDir, outPath string) error {
	rawMan, err := os.ReadFile(filepath.Join(srcDir, "manifest.json"))
	if err != nil {
		return fmt.Errorf("read manifest.json: %w", err)
	}
	var man sdk.PluginManifest
	if err := json.Unmarshal(rawMan, &man); err != nil {
		return fmt.Errorf("parse manifest.json: %w", err)
	}
	if err := sdk.ValidateManifest(&man); err != nil {
		return fmt.Errorf("validate manifest.json: %w", err)
	}

	type fileEntry struct {
		rel  string
		data []byte
	}
	var files []fileEntry
	seen := make(map[string]struct{})

	normMan, _ := json.MarshalIndent(man, "", "  ")
	normMan = append(normMan, '\n')
	files = append(files, fileEntry{rel: "manifest.json", data: normMan})
	seen["manifest.json"] = struct{}{}

	totalRaw := int64(len(normMan))
	err = filepath.WalkDir(srcDir, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		relFS, err := filepath.Rel(srcDir, p)
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(relFS)
		if rel == "manifest.json" || strings.HasSuffix(rel, "server.wasm") || strings.HasSuffix(rel, "server_plugin.wasm") {
			return nil
		}
		// Only include entry wasm + allowed asset files (skip Go source files cleanly).
		ext := strings.ToLower(filepath.Ext(rel))
		if ext != ".wasm" && ext != ".png" && ext != ".wav" && ext != ".ogg" && ext != ".json" && ext != ".glb" {
			return nil
		}
		cleanRel, cErr := sdk.CleanBundlePath(rel)
		if cErr != nil {
			return cErr
		}
		if _, dup := seen[cleanRel]; dup {
			return nil
		}
		data, rErr := os.ReadFile(p)
		if rErr != nil {
			return rErr
		}
		totalRaw += int64(len(data))
		if totalRaw > sdk.MaxExtractedBytes {
			return fmt.Errorf("extracted bundle exceeds %d bytes", sdk.MaxExtractedBytes)
		}
		seen[cleanRel] = struct{}{}
		files = append(files, fileEntry{rel: cleanRel, data: data})
		return nil
	})
	if err != nil {
		return err
	}
	if _, hasWasm := seen[man.EntryWasm]; !hasWasm {
		return fmt.Errorf("missing entry wasm %q in %s", man.EntryWasm, srcDir)
	}
	if len(files) > sdk.MaxEntries {
		return fmt.Errorf("too many bundle files (%d > %d)", len(files), sdk.MaxEntries)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].rel < files[j].rel })

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range files {
		hdr := &zip.FileHeader{
			Name:   f.rel,
			Method: zip.Deflate,
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if _, err := w.Write(f.data); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	outBytes := buf.Bytes()
	if len(outBytes) > sdk.MaxArchiveBytes {
		return fmt.Errorf("compressed .gtpack exceeds %d bytes (%d)", sdk.MaxArchiveBytes, len(outBytes))
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, outBytes, 0o644); err != nil {
		return err
	}
	sum := sha256.Sum256(outBytes)
	fmt.Printf("Packed %s (%d files, %d bytes, sha256=%s)\n",
		outPath, len(files), len(outBytes), hex.EncodeToString(sum[:]))
	return nil
}
