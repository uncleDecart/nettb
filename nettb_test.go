package nettb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestNetDev(t *testing.T) {
	// Not calling NetDev directly, to skip reading file
	input := []string{"HEADER LINE 1",
		"HEADER LINE 2",
		"lo: 1421 123 123 123",
		"enp0s5: 123 123 123 9 123 912"}
	want := []string{"lo", "enp0s5"}
	got := processNetDev(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Failed got %v, want %v", got, want)
	}
}

func TestIntersection(t *testing.T) {
	s1 := []string{"dummy0", "eth0", "eth1", "flow-mon-dummy", "ip6tnl0", "keth0", "keth1", "lo", "tunl0"}
	s2 := []string{"lo", "tunl0", "ip6tnl0", "keth0", "keth1", "eth0", "eth1", "dummy0", "flow-mon-dummy"}
	want := []string{"dummy0", "ip6tnl0", "eth0", "eth1", "keth0", "keth1", "lo", "tunl0", "flow-mon-dummy"}
	got := intersection(s1, s2)
	sort.Strings(want)
	sort.Strings(got)
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

	got := getPciAddrsForDevices(root, devs)

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Failed got %v, want %v", got, want)
	}

}
