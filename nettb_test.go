package nettb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type StubLogger struct {
	log []string
}

func (l StubLogger) Functionf(format string, args ...interface{}) {
	l.log = append(l.log, fmt.Sprintf(format, args...))
}

func TestNetDev(t *testing.T) {
	// Not calling NetDev directly, to skip reading file
	input := []string{"HEADER LINE 1",
		"HEADER LINE 2",
		"lo: 1421 123 123 123",
		"enp0s5: 123 123 123 9 123 912"}
	var log StubLogger
	want := []string{"lo", "enp0s5"}
	got := processNetDev(input, log)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Failed got %v, want %v", got, want)
	}
}

func TestIntersection(t *testing.T) {
	s1 := []string{"a", "b", "c", "d"}
	s2 := []string{"d", "e", "f", "h"}
	want := []string{"d"}
	got := intersection(s1, s2)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Failed got %v, want %v", got, want)
	}
}

func setupTestGetDirNames(root string, folders []string) func() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for _, folder := range folders {
		os.MkdirAll(filepath.Join(wd, root, folder), 0755)
	}
	return func() {
		os.RemoveAll(filepath.Join(wd, root))
	}
}

func TestGetDirNames(t *testing.T) {
	root := "test/"
	folders := []string{"a", "b", "c", "d"}
	defer setupTestGetDirNames(root, folders)()

	var log StubLogger
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Got error during getting current directory %v", err)
	}
	got, err := getDirNames(filepath.Join(wd, root), log)
	if err != nil {
		t.Fatalf("Got error during function call %v", err)
	}

	if !reflect.DeepEqual(folders, got) {
		t.Fatalf("Failed got %v, want %v", got, folders)
	}

}

func setupTestPciToIfNameMap(root string, folders map[string]string) func() {
	err := os.Mkdir(root, 0755)
	if err != nil {
		panic(err)
	}
	for pci_addr, d := range folders {
		symlink := filepath.Join(root, pci_addr)
		device_path := filepath.Join(root, d)
		err = os.Mkdir(device_path, 0755)
		if err != nil {
			panic(err)
		}
		ioutil.WriteFile(symlink, []byte("Test\n"), 0644)
		os.Symlink(symlink, filepath.Join(device_path, "/device"))
	}
	return func() {
		os.RemoveAll(root)
	}
}

func TestPciToIfNameMap(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Got error during getting current directory %v", err)
	}
	root := filepath.Join(wd, "test/")
	want := map[string]string{"0000:00:00.0": "eno0", "0000:00:00.1": "eno1", "0000:00:01.0": "eno2", "0000:00:01.1": "eno3"}
	defer setupTestPciToIfNameMap(root, want)()

	var devs []string
	for _, value := range want {
		devs = append(devs, value)
	}

	var log StubLogger
	got := getPciAddrsForDevices(root, devs, log)

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Failed got %v, want %v. Log: %v", got, want, log)
	}

}
