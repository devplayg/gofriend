package dff

import (
	"io"
	"os"
)

func isReadableDirs(dirs []string) error {
	for _, dir := range dirs {
		err := isValidDir(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func isValidDir(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		return err
	}
	return nil
}

func generateFileKey(path string) ([32]byte, error) {
	hash, err := getHighwayHash(path)
	//logrus.Debugf("%s - %s", filepath.Base(path), hex.EncodeToString(hash))
	if err != nil {
		return [32]byte{}, err
	}

	var key [32]byte
	copy(key[:], hash)

	return key, nil
}

func getHighwayHash(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	highwayHash.Reset()
	if _, err = io.Copy(highwayHash, file); err != nil {
		return nil, err
	}

	checksum := highwayHash.Sum(nil)
	return checksum, nil
}
