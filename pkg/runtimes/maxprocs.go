package runtimes

import (
	"os"
	"runtime"
	"strconv"

	"github.com/foredata/nova/pkg/runtimes/internal/maxprocs"
)

// GOMAXPROCS 通过cgroups自动设置GOMAXPROCS
func GOMAXPROCS() int {
	if maxstr, exists := os.LookupEnv("GOMAXPROCS"); exists {
		if max, err := strconv.Atoi(maxstr); err == nil {
			return runtime.GOMAXPROCS(max)
		}
	}

	procs, err := maxprocs.CPUQuotaToGOMAXPROCS()

	if err != nil {
		return runtime.GOMAXPROCS(0)
	}

	return runtime.GOMAXPROCS(procs)
}
