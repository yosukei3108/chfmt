package cmd

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type CLI struct {
	OutStream, ErrStream io.Writer
}

const (
	Version string = "v0.0.1"

	ExitCodeOK             = 0
	ExitCodeParseFlagError = 10 + iota
	ExitCodeTooManyArgs
	ExitCodeInvalidExtensionError
	ExitCodeFailedToGetCd
	ExitCodeFailedToExec
)

type Format string

func (c *CLI) Run(args []string) int {

	var flagVersion bool
	var flagSrcFormat string
	var flagDstFormat string

	flags := flag.NewFlagSet("chfmt", flag.ContinueOnError)
	flags.SetOutput(c.ErrStream)
	flags.BoolVar(&flagVersion, "version", false, "Print version information.")
	flags.StringVar(&flagSrcFormat, "src", "jpeg", "Specify the format of source image file(s). (\"jpg\" is not included in \"jpeg\" yet here :P)")
	flags.StringVar(&flagDstFormat, "dst", "png", "Specify the format of destination image file(s).")

	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeParseFlagError
	}

	if len(args) > 3 {
		fmt.Fprint(c.ErrStream, "Error: Too many arguments\n")
		return ExitCodeTooManyArgs
	}

	if flagVersion {
		fmt.Fprintf(c.OutStream, "chfmt version %s\n", Version)
		return ExitCodeOK
	}

	fmt.Fprint(c.OutStream, "Change formats...\n")

	currentDir, err := os.Getwd()
	if err != nil {
		return ExitCodeFailedToGetCd
	}

	if err := ChangeFormat(currentDir, Format(flagSrcFormat), Format(flagDstFormat)); err != nil {
		fmt.Fprintf(c.ErrStream, "Error: %v\n", err)
		return ExitCodeFailedToExec
	}

	return ExitCodeOK
}

func GetFormatFromExtention(path string) Format {
	ext := filepath.Ext(path)
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	}

	return "unknown"
}

// Decode decodes an image that has been encoded in the given Format.
// The string returned is the format name given in Reader r.
func Decode(r io.Reader) (image.Image, Format, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, "unknown", err
	}

	switch format {
	case "png":
		return img, "png", nil
	case "jpeg":
		return img, "jpeg", nil
	case "gif":
		return img, "gif", nil
	}

	return nil, "unknown", image.ErrFormat
}

// Encode writes the Image img to w in given Format f.
func Encode(w io.Writer, img image.Image, f Format) error {
	switch f {
	case "png":
		return png.Encode(w, img)
	case "jpeg":
		return jpeg.Encode(w, img, nil)
	case "gif":
		return gif.Encode(w, img, nil)
	}
	return image.ErrFormat
}

// ChangeExt changes the file extention to match as a given Format f
func ChangeExt(path string, f Format) string {
	if f == "unknown" {
		return path
	}

	ext := filepath.Ext(path)
	i := len(path) - len(ext)
	return path[:i] + "." + string(f)
}

// ChangeFormat makes images in Format dst from existing files in Format src.
// It searches files and dirs under current directory recursively.
func ChangeFormat(root string, src, dst Format) error {
	walkdirfunc := func(path string, info fs.DirEntry, err error) (rerr error) {

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := GetFormatFromExtention(path)
		if ext != src {
			return nil
		}

		sf, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sf.Close()

		img, format, err := Decode(sf)
		if err != nil {
			return err
		}

		if format != src {
			return nil
			// Run() がこの終了コードを返す、的なことがしたいのだが、、、
			// return ExitCodeInvalidExtensionError
		}

		dstPath := ChangeExt(path, dst)
		df, err := os.Create(dstPath)
		if err != nil {
			return err
		}

		defer func() {
			if err := df.Close(); rerr == nil && err != nil {
				rerr = err
			}
		}()

		if err := Encode(df, img, dst); err != nil {
			return err
		}

		return nil
	}

	return filepath.WalkDir(root, walkdirfunc)
}
