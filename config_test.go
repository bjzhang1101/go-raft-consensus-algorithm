package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func createTempFile(content string) (*os.File, error) {
	file, err := ioutil.TempFile("", "ut")
	if err != nil {
		return nil, err
	}

	if _, err = file.WriteString(content); err != nil {
		return nil, err
	}

	return file, nil
}

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		in  string
		out config
	}{
		{
			in: `
id: raft-01.raft.svc.cluster.local
quorum:
- raft-01.raft.svc.cluster.local
- raft-02.raft.svc.cluster.local
- raft-03.raft.svc.cluster.local
`,
			out: config{
				ID:     "raft-01.raft.svc.cluster.local",
				Quorum: []string{"raft-01.raft.svc.cluster.local", "raft-02.raft.svc.cluster.local", "raft-03.raft.svc.cluster.local"},
			},
		},
	}

	for i, c := range cases {
		file, err := createTempFile(c.in)
		if err != nil {
			t.Fatalf("failed to create temporary file: %v", err)
		}

		out, err := loadConfig(file.Name())
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if diff := cmp.Diff(c.out, out); len(diff) > 0 {
			t.Errorf("[%d] result mismatch (-want +got)\n%s", i, diff)
		}

		os.Remove(file.Name())
	}
}
