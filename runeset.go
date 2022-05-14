package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type Runeset [256][8]byte

var (
	FileReadError    = errors.New("unable to read file")
	BytesLengthError = errors.New("the byte slice is the wrong length (should be 2048)")
	IndexError       = errors.New("runeset index out of range")
)

func (r Runeset) toBytes() []byte {
	bytes := make([]byte, 2048)

	for i := 0; i < 256; i++ {
		for j := 0; j < 8; j++ {
			bytes[i*8+j] = r[i][j]
		}
	}

	return bytes
}

func (r Runeset) ToImg(index int) ([8][8]bool, error) {
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

func (r Runeset) Preview(index int) (string, error) {
	img, err := r.ToImg(index)

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
