package main

import (
	"errors"
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
)

type Runeset [256][8]byte
type RuneBytes [8]byte

var (
	FileReadError    = errors.New("unable to read file")
	BytesLengthError = errors.New("the byte slice is the wrong length (should be 2048)")
	IndexError       = errors.New("runeset index out of range")
	NotFoundError    = errors.New("file not found")
)

func (r RuneBytes) Invert() RuneBytes {
	for i := 0; i < 8; i++ {
		r[i] = ^r[i]
	}

	return r
}

func (r RuneBytes) Reverse() RuneBytes {

	for i := 0; i < 8; i++ {
		r[i] = ((r[i] & 0x01) << 7) |
			((r[i] & 0x02) << 5) |
			((r[i] & 0x04) << 3) |
			((r[i] & 0x08) << 1) |
			((r[i] & 0x10) >> 1) |
			((r[i] & 0x20) >> 3) |
			((r[i] & 0x40) >> 5) |
			((r[i] & 0x80) >> 7)
	}

	return r
}

func (r *Runeset) ReadAt(index int) (RuneBytes, error) {
	if index < 0 || index >= 256 {
		return RuneBytes{}, IndexError
	}

	return r[index], nil
}

func (r *Runeset) SetAt(rune RuneBytes, index int) error {
	if index < 0 || index >= 256 {
		return IndexError
	}

	r[index] = rune

	return nil
}

func (r *Runeset) toBytes() []byte {
	bytes := make([]byte, 2048)

	for i := 0; i < 256; i++ {
		for j := 0; j < 8; j++ {
			bytes[i*8+j] = r[i][j]
		}
	}

	return bytes
}

func (r *Runeset) ToBitArray(index int) ([8][8]bool, error) {
	if index < 0 || index >= 256 {
		return [8][8]bool{}, IndexError
	}

	img := [8][8]bool{}

	c := r[index]

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			img[i][j] = (c[i]&(1<<j))>>j == 1
		}
	}

	return img, nil
}

func (r *Runeset) Preview(index int) (string, error) {
	img, err := r.ToBitArray(index)

	if err != nil {
		return "", err
	}

	lookup := []rune{' ', '▘', '▝', '▀', '▖', '▌', '▞', '▛', '▗', '▚', '▐', '▜', '▄', '▙', '▟', '█'}

	s := ""

	for i := 0; i < 4; i++ { // :(
		for j := 0; j < 4; j++ {
			n := 0
			for ii := 0; ii < 2; ii++ {
				for jj := 0; jj < 2; jj++ {
					if img[i*2+ii][j*2+jj] {
						n += 1 << (ii*2 + jj)
					}
				}
			}

			s += fmt.Sprintf("%c", lookup[n])
		}

		if i != 3 {
			s += "\n"
		}
	}

	return s, nil
}

func ReadRunesetFile(path string) (Runeset, error, bool) {
	if !checkFileExists(path) {
		return Runeset{}, nil, false
	}

	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		return Runeset{}, err, false
	}

	if len(bytes) != 2048 {
		return Runeset{}, FileReadError, true
	}

	c, err := bytesToCharset(bytes)

	return c, err, true
}

func WriteRunesetFile(r Runeset, path string) error {
	err := ioutil.WriteFile(path, r.toBytes(), 0755)

	if err != nil {
		return err
	}

	return nil
}

func readImage(path string) (Runeset, error) {
	if !checkFileExists(path) {
		return Runeset{}, NotFoundError
	}

	file, err := os.Open(path)

	if err != nil {
		return Runeset{}, err
	}

	img, err := png.Decode(file)

	if err != nil {
		return Runeset{}, err
	}

	r := Runeset{}

	for i := 0; i < 8; i++ {
		for j := 0; j < 32; j++ {
			var rb RuneBytes

			for ii := 0; ii < 8; ii++ {
				for jj := 0; jj < 8; jj++ {
					x, y := j*8+jj, i*8+ii
					rc, gc, bc, _ := img.At(x, y).RGBA()
					if rc == 0 && gc == 0 && bc == 0 {
						rb[ii] |= 1 << jj
					}
				}
			}

			r[i*32+j] = rb
		}
	}

	return r, nil
}

func bytesToCharset(bytes []byte) (Runeset, error) {
	if len(bytes) != 2048 {
		return Runeset{}, BytesLengthError
	}

	c := Runeset{}

	for i := 0; i < 256; i++ {
		for j := 0; j < 8; j++ {
			c[i][j] = bytes[i*8+j]
		}
	}

	return c, nil
}

func checkFileExists(path string) bool {
	_, err := os.Stat(path) // Does something (not sure what? should probably find out) but returns an error if file not found
	return !os.IsNotExist(err)
}
