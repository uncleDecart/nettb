package nettb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Logger interface {
	Functionf(format string, args ...interface{})
}

func NetDev(log Logger) ([]string, error) {
	content, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	procnetdev := strings.Split(string(content), "\n")
	if len(procnetdev) <= 2 {
		return nil, fmt.Errorf("/proc/net/dev does not contain any interfaces")
	}

	return processNetDev(procnetdev, log), nil
}

func processNetDev(lines []string, log Logger) []string {
	var ans []string
	for _, iface := range lines[2:] { // first two lines are headerlines
		idx := strings.Index(iface, ":")
		if idx > -1 {
			ans = append(ans, iface[:idx])
		} else {
			log.Functionf("Somethings wrong with %s", iface)
		}
	}
	return ans
}

func PciToIfNameMap(log Logger) (map[string]string, error) {
	netDevIfaces, err := NetDev(log)
	if err != nil {
		return nil, err
	}

	sysClassNetPath := "/sys/class/net"
	sysClassNetDevices, err := getDirNames(sysClassNetPath, log)
	if err != nil {
		return nil, err
	}

	ifaces := intersection(netDevIfaces, sysClassNetDevices)

	return getPciAddrsForDevices(sysClassNetPath, ifaces, log), nil
}

func getDirNames(path string, log Logger) ([]string, error) {
	var dirs []string
	getDirs := func(fs *[]string) filepath.WalkFunc {
		return func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Functionf("Error happend during walk %v", err)
				return err
			}
			if info.IsDir() {
				*fs = append(*fs, info.Name())
			}
			return nil
		}
	}
	err := filepath.Walk(path, getDirs(&dirs))
	if err != nil {
		return nil, err
	}
	return dirs[1:], nil // Since we don't want to inlclude root directory
}

func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		if hash[e] {
			inter = append(inter, e)
		}
	}
	return inter
}

func getPciAddrsForDevices(root string, devices []string, log Logger) map[string]string {
	pciBdfRe := regexp.MustCompile("[0-9a-f]{4}:[0-9a-f]{2,4}:[0-9a-f]{2}\\.[0-9a-f]")
	res := make(map[string]string)
	for _, d := range devices {
		path, err := filepath.EvalSymlinks(filepath.Join(root, d))
		if err != nil {
			log.Functionf("Cannot evaluate symlink for %s device. Error %s", d, err)
			continue
		}
		pci_addr := pciBdfRe.FindString(path)
		if pci_addr == "" {
			log.Functionf("PCI address is not in BFD notation for %s device", d)
			continue
		}
		res[pci_addr] = d
	}
	return res
}
