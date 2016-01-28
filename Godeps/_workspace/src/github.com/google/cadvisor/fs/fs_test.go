// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDiskStatsMap(t *testing.T) {
	diskStatsMap, err := getDiskStatsMap("test_resources/diskstats")
	if err != nil {
		t.Errorf("Error calling getDiskStatMap %s", err)
	}
	if len(diskStatsMap) != 30 {
		t.Errorf("diskStatsMap %+v not valid", diskStatsMap)
	}
	keySet := map[string]string{
		"/dev/sda":  "/dev/sda",
		"/dev/sdb":  "/dev/sdb",
		"/dev/sdc":  "/dev/sdc",
		"/dev/sdd":  "/dev/sdd",
		"/dev/sde":  "/dev/sde",
		"/dev/sdf":  "/dev/sdf",
		"/dev/sdg":  "/dev/sdg",
		"/dev/sdh":  "/dev/sdh",
		"/dev/sdb1": "/dev/sdb1",
		"/dev/sdb2": "/dev/sdb2",
		"/dev/sda1": "/dev/sda1",
		"/dev/sda2": "/dev/sda2",
		"/dev/sdc1": "/dev/sdc1",
		"/dev/sdc2": "/dev/sdc2",
		"/dev/sdc3": "/dev/sdc3",
		"/dev/sdc4": "/dev/sdc4",
		"/dev/sdd1": "/dev/sdd1",
		"/dev/sdd2": "/dev/sdd2",
		"/dev/sdd3": "/dev/sdd3",
		"/dev/sdd4": "/dev/sdd4",
		"/dev/sde1": "/dev/sde1",
		"/dev/sde2": "/dev/sde2",
		"/dev/sdf1": "/dev/sdf1",
		"/dev/sdf2": "/dev/sdf2",
		"/dev/sdg1": "/dev/sdg1",
		"/dev/sdg2": "/dev/sdg2",
		"/dev/sdh1": "/dev/sdh1",
		"/dev/sdh2": "/dev/sdh2",
		"/dev/dm-0": "/dev/dm-0",
		"/dev/dm-1": "/dev/dm-1",
	}

	for device := range diskStatsMap {
		if _, ok := keySet[device]; !ok {
			t.Errorf("Cannot find device %s", device)
		}
		delete(keySet, device)
	}
	if len(keySet) != 0 {
		t.Errorf("diskStatsMap %+v contains illegal keys %+v", diskStatsMap, keySet)
	}
}

func TestFileNotExist(t *testing.T) {
	_, err := getDiskStatsMap("/file_does_not_exist")
	if err != nil {
		t.Fatalf("getDiskStatsMap must not error for absent file: %s", err)
	}
}

func TestDirUsage(t *testing.T) {
	as := assert.New(t)
	fsInfo, err := NewFsInfo(Context{})
	as.NoError(err)
	dir, err := ioutil.TempDir(os.TempDir(), "")
	as.NoError(err)
	defer os.RemoveAll(dir)
	dataSize := 1024 * 100 //100 KB
	b := make([]byte, dataSize)
	f, err := ioutil.TempFile(dir, "")
	as.NoError(err)
	as.NoError(ioutil.WriteFile(f.Name(), b, 0700))
	fi, err := f.Stat()
	as.NoError(err)
	expectedSize := uint64(fi.Size())
	size, err := fsInfo.GetDirUsage(dir)
	as.NoError(err)
	as.True(expectedSize <= size, "expected dir size to be at-least %d; got size: %d", expectedSize, size)
}

var dmStatusTests = []struct {
	dmStatus    string
	used        uint64
	total       uint64
	errExpected bool
}{
	{`0 409534464 thin-pool 64085 3705/4161600 88106/3199488 - rw no_discard_passdown queue_if_no_space -`, 88106, 3199488, false},
	{`0 209715200 thin-pool 707 1215/524288 30282/1638400 - rw discard_passdown`, 30282, 1638400, false},
	{`Invalid status line`, 0, 0, false},
}

func TestParseDMStatus(t *testing.T) {
	for _, tt := range dmStatusTests {
		used, total, err := parseDMStatus(tt.dmStatus)
		if tt.errExpected && err != nil {
			t.Errorf("parseDMStatus(%q) expected error", tt.dmStatus)
		}
		if used != tt.used {
			t.Errorf("parseDMStatus(%q) wrong used value => %q, want %q", tt.dmStatus, used, tt.used)
		}
		if total != tt.total {
			t.Errorf("parseDMStatus(%q) wrong total value => %q, want %q", tt.dmStatus, total, tt.total)
		}
	}
}

var dmTableTests = []struct {
	dmTable     string
	major       uint
	minor       uint
	dataBlkSize uint
	errExpected bool
}{
	{`0 409534464 thin-pool 253:6 253:7 128 32768 1 skip_block_zeroing`, 253, 7, 128, false},
	{`0 409534464 thin-pool 253:6 258:9 512 32768 1 skip_block_zeroing otherstuff`, 258, 9, 512, false},
	{`Invalid status line`, 0, 0, 0, false},
}

func TestParseDMTable(t *testing.T) {
	for _, tt := range dmTableTests {
		major, minor, dataBlkSize, err := parseDMTable(tt.dmTable)
		if tt.errExpected && err != nil {
			t.Errorf("parseDMTable(%q) expected error", tt.dmTable)
		}
		if major != tt.major {
			t.Errorf("parseDMTable(%q) wrong major value => %q, want %q", tt.dmTable, major, tt.major)
		}
		if minor != tt.minor {
			t.Errorf("parseDMTable(%q) wrong minor value => %q, want %q", tt.dmTable, minor, tt.minor)
		}
		if dataBlkSize != tt.dataBlkSize {
			t.Errorf("parseDMTable(%q) wrong dataBlkSize value => %q, want %q", tt.dmTable, dataBlkSize, tt.dataBlkSize)
		}
	}
}